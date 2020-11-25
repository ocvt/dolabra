package handler

import (
	"container/list"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/ocvt/dolabra/utils"
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/checkout/session"
	"github.com/stripe/stripe-go/v71/webhook"
)

/* Stripe payment flow:
 * 1. Member navigates to {FRONTEND}/dues
 * 2. Member selects item to buy
 * 3. When member clicks 'Submit', javascript does GET {API}/payment/{paymentOption}
 * 4. API returns Stripe Checkout session id
 * 5. javascript parses response and redirects to Stripe checkout
 * 6  Stripe sends a webhook to {API}/payment/success on the 'checkout.session.completed' event
 * 7. API scknowledges event
 * 8. Stripe redirects to {FRONTEND}/dues/success or {FRONTEND}/dues/cancel
 */

/* Manual payment flow:
 * 1. Officer manually creates an order with webtools
 * 2. Member is given secret code
 * 3. Member redeems code
 */

/* Only for redeeming codes */
type simpleStoreCodeStruct struct {
	Code string `json:"code"`
}

// Start new Stripe payment
func GetPayment(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId, paymentOption
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}
	paymentOption := chi.URLParam(r, "paymentOption")

	// Permissions
	if paymentOption != "dues" && paymentOption != "duesShirt" &&
		paymentOption != "freshmanSpecial" {
		respondError(w, http.StatusNotFound, "Payment option does not exist.")
		return
	}

	// Used for us
	membershipYears := 0
	shirt := true
	// Used for stripe
	var amount int
	description := ""
	if paymentOption == "dues" {
		membershipYears = 1
		shirt = false
		amount = 20
		description = "Dues for 1 year"
	} else if paymentOption == "duesShirt" {
		membershipYears = 1
		amount = 30
		description = "Dues for 1 year + shirt"
	} else {
		membershipYears = 4
		amount = 65
		description = "Dues for 4 years + shirt"
	}
	amount *= 100

	// Create checkout session
	// https://stripe.com/docs/payments/checkout/accept-a-payment#create-checkout-session
	stripe.Key = utils.GetConfig().StripeSecretKey
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		Mode: stripe.String("payment"),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("usd"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(description),
					},
					UnitAmount: stripe.Int64(int64(amount)),
				},
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(utils.GetConfig().FrontendUrl + "/dues/success"),
		CancelURL:  stripe.String(utils.GetConfig().FrontendUrl + "/dues/cancel"),
	}
	session, err := session.New(params)
	if !checkError(w, err) {
		return
	}

	// Insert payment
	if membershipYears > 0 && !dbInsertPayment(
		w, 8000000, "", memberId, "MEMBERSHIP", membershipYears, amount, "STRIPE",
		session.ID, false) {
		return
	}
	if shirt && !dbInsertPayment(
		w, 8000000, "", memberId, "SHIRT", 1, amount, "STRIPE", session.ID, false) {
		return
	}

	// Pass checkout session id
	// https://stripe.com/docs/payments/checkout/accept-a-payment#pass-the-session-id
	respondJSON(w, http.StatusOK, map[string]string{"sessionId": session.ID})
}

// Redeem manually generated store codes
func PostPaymentRedeem(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId, paymentOption
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}

	// Get request body
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var simpleStoreCode simpleStoreCodeStruct
	err := decoder.Decode(&simpleStoreCode)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	stmt := `
		SELECT *
		FROM store_code
		WHERE redeemed_datetime is NULL AND code = ?`
	rows, err := db.Query(stmt, simpleStoreCode.Code)
	if !checkError(w, err) {
		return
	}
	defer rows.Close()

	codes := list.New()
	for rows.Next() {
		var storeCode storeCodeStruct
		err = rows.Scan(
			&storeCode.Id,
			&storeCode.CreateDatetime,
			&storeCode.GeneratedById,
			&storeCode.Note,
			&storeCode.StoreItemId,
			&storeCode.StoreItemCount,
			&storeCode.Amount,
			&storeCode.Code,
			&storeCode.Completed,
			&storeCode.RedeemedDatetime)
		if !checkError(w, err) {
			return
		}
		codes.PushBack(storeCode)
	}
	err = rows.Err()
	if !checkError(w, err) {
		return
	}

	if codes.Len() == 0 {
		respondError(w, http.StatusForbidden, "Code does not exist or has already been redeemed.")
		return
	}

	for c := codes.Front(); c != nil; c = c.Next() {
		code := c.Value.(storeCodeStruct)

		// Impossible to be MEMBERSHIP AND not redeemed AND not completed
		completed := code.Completed
		if code.StoreItemId == "MEMBERSHIP" {
			if !dbExtendMembership(w, memberId, code.StoreItemCount) {
				return
			}
			completed = true
		}

		ctx := context.Background()
		tx, err := db.BeginTx(ctx, nil)
		if !checkError(w, err) {
			return
		}

		// Transfer to be in proper payment table associated with member
		stmt = `
			INSERT INTO payment (
				create_datetime,
				entered_by_id,
				note,
				member_id,
				store_item_id,
				store_item_count,
				amount,
				payment_method,
				payment_id,
				completed)
			VALUES (?, ?, ?, ?, ?, ?, ?, 'MANUAL', ?, ?)`
		_, err = tx.ExecContext(
			ctx,
			stmt,
			code.CreateDatetime,
			code.GeneratedById,
			code.Note,
			memberId,
			code.StoreItemId,
			code.StoreItemCount,
			code.Amount,
			code.Code,
			completed)
		if !checkError(w, err) {
			tx.Rollback()
			return
		}

		// Prevent from redeeming item again
		stmt = `
			UPDATE store_code
			SET redeemed_datetime = datetime('now')
			WHERE id = ?`
		_, err = tx.ExecContext(ctx, stmt, code.Id)
		if !checkError(w, err) {
			tx.Rollback()
			return
		}

		err = tx.Commit()
		if !checkError(w, err) {
			return
		}
	}

	respondJSON(w, http.StatusNoContent, nil)
}

// Complete Stripe payment
// https://stripe.com/docs/payments/checkout/accept-a-payment#payment-success
func PostPaymentSuccess(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	body, err := ioutil.ReadAll(r.Body)
	if !checkError(w, err) {
		return
	}

	// Verify event
	event, err := webhook.ConstructEvent(body,
		r.Header.Get("Stripe-Signature"), utils.GetConfig().StripeWebhookSecret)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// We only accept checkout.session.completed event types so no need to check
	var session stripe.CheckoutSession
	err = json.Unmarshal(event.Data.Raw, &session)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Add years and complete payment
	memberId, membershipYears, ok := dbGetItemCount(w, "MEMBERSHIP", "STRIPE", session.ID)
	if !ok {
		return
	}

	if !dbExtendMembership(w, memberId, membershipYears) {
		return
	}

	stmt := `
		UPDATE payment
		SET completed = true
		WHERE
			store_item_id = 'MEMBERSHIP'
			AND payment_method = 'STRIPE'
			AND payment_id = ?`
	_, err = db.Exec(stmt, session.ID)
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}
