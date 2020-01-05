package handler

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net/http"

  "github.com/go-chi/chi"
  "github.com/stripe/stripe-go"
  "github.com/stripe/stripe-go/paymentintent" 
  "github.com/stripe/stripe-go/webhook"
)

var STRIPE_SECRET_KEY string
var STRIPE_WEBOOK_SECRET string

func GetPayment(w http.ResponseWriter, r *http.Request) {
  _, subject, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Get memberId, paymentOption
  memberId, ok := dbGetActiveMemberId(w, subject)
  if !ok {
    return
  }
  paymentOption := chi.URLParam(r, "paymentOption")

  // Permissions
  if paymentOption != "dues" && paymentOption != "duesShirt" &&
     paymentOption != "freshmanSpecial" {
    respondError(w, http.StatusBadRequest, "Invalid payment option.")
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

  // Create paymentIntent and send to client
  stripe.LogLevel = 1
  stripe.Key = STRIPE_SECRET_KEY
  params := &stripe.PaymentIntentParams{
    Amount: stripe.Int64(int64(amount)),
    Currency: stripe.String(string(stripe.CurrencyUSD)),
    Description: &description,
  }
  myPI, err := paymentintent.New(params)
  if !checkError(w, err) {
    return
  }

  // Insert payment
  if membershipYears > 0 && !dbInsertPayment(
      w, 0, "", memberId, "MEMBERSHIP", membershipYears, amount, myPI.ID) {
    return
  }
  if shirt && !dbInsertPayment(
      w, 0, "", memberId, "SHIRT", 1, amount, myPI.ID) {
    return
  }

  response := map[string]string{
    "stripeClientSecret": myPI.ClientSecret,
  }
  respondJSON(w, http.StatusOK, response)
}

// Mostly copied from
// https://stripe.com/docs/payments/payment-intents/verifying-status#webhooks
func PostPaymentSucceeded(w http.ResponseWriter, r *http.Request) {
  // Get POST body
  const MaxBodyBytes = int64(65536)
  r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
  body, err := ioutil.ReadAll(r.Body)
  if !checkError(w, err) {
    return
  }

  // Verify event
  event, err := webhook.ConstructEvent(body,
      r.Header.Get("Stripe-Signature"), STRIPE_WEBOOK_SECRET)
  if err != nil {
    respondError(w, http.StatusBadRequest, err.Error())
    return
  }

  // We only accept payment_intent.succeeded event types so no need to check
  var myPI stripe.PaymentIntent
  err = json.Unmarshal(event.Data.Raw, &myPI)
  if err != nil {
    respondError(w, http.StatusBadRequest, err.Error())
    return
  }
  fmt.Printf("myPI ID: %s", myPI.ID)

  // TODO Detect payment failure?

  // Add years and complete payment
  memberId, membershipYears, ok := dbGetItemCount(w, "MEMBERSHIP", myPI.ID)
  if !ok {
    return
  }

  stmt := `
    UPDATE member
    SET paid_expire_datetime = datetime(paid_expire_datetime, '+? years')
    WHERE member.id = ?`
  _, err = db.Exec(stmt, membershipYears, memberId)
  if !checkError(w, err) {
    return
  }

  stmt = `
    UPDATE payment
    SET completed = true
    WHERE store_item_id = 'MEMBERSHIP' AND stripe_payment_id = ?`
  _, err = db.Exec(stmt, myPI.ID)
  if !checkError(w, err) {
    return
  }

  respondJSON(w, http.StatusNoContent, nil)
}
