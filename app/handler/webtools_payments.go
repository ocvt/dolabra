package handler

import (
	"encoding/json"
	"net/http"
)

const MANUAL_PAYMENT_ID_LENGTH = 64

type paymentStruct struct {
	/* Managed Server side */
	Id             int    `json:"id,omitempty"`
	CreateDatetime string `json:"createDatetime,omitempty"`
	EnteredById    int    `json:"enteredById,omitempty"`
	MemberId       int    `json:"memberId,omitempty"`
	PaymentMethod  string `json:"paymentMethod,omitempty"`
	// Only for client view when returning data
	EnteredByName string `json:"enteredByName,omitempty"`
	MemberName    string `json:"memberName,omitempty"`
	/* Required fields */
	Note           string `json:"note"`
	StoreItemId    string `json:"storeItemId"`
	StoreItemCount int    `json:"storeItemCount"`
	Amount         int    `json:"amount"`
	// Only used if multiple items in 1 payment
	PaymentId string `json:"paymentId"`
	Completed bool   `json:"completed"`
}

// This is somewhat similar to paymentStruct but different enough to justify
// a separate struct
type storeCodeStruct struct {
	/* Managed Server side */
	Id             int    `json:"id,omitempty"`
	CreateDatetime string `json:"createDatetime,omitempty"`
	GeneratedById  int    `json:"generatedById,omitempty"`
	Redeemed       bool   `json:"redeemed,omitempty"`
	/* Required fields */
	Note           string `json:"note"`
	StoreItemId    string `json:"storeItemId"`
	StoreItemCount int    `json:"storeItemCount"`
	Amount         int    `json:"amount"`
	// Only used if multiple items in 1 payment
	Code      string `json:"code"`
	Completed bool   `json:"completed"`
}

func GetWebtoolsPayments(w http.ResponseWriter, r *http.Request) {
	stmt := `
		SELECT *
		FROM payment
		ORDER BY datetime(create_datetim) DESC`
	rows, err := db.Query(stmt)
	if !checkError(w, err) {
		return
	}
	defer rows.Close()

	var payments = []*paymentStruct{}
	i := 0
	for rows.Next() {
		payments = append(payments, &paymentStruct{})
		err = rows.Scan(
			&payments[i].Id,
			&payments[i].CreateDatetime,
			&payments[i].EnteredById,
			&payments[i].Note,
			&payments[i].MemberId,
			&payments[i].StoreItemId,
			&payments[i].StoreItemCount,
			&payments[i].Amount,
			&payments[i].PaymentMethod,
			&payments[i].PaymentId,
			&payments[i].Completed)
		if !checkError(w, err) {
			return
		}

		enteredByName, ok := dbGetMemberName(w, payments[i].EnteredById)
		if !ok {
			return
		}
		memberName, ok := dbGetMemberName(w, payments[i].MemberId)
		if !ok {
			return
		}

		payments[i].EnteredByName = enteredByName
		payments[i].MemberName = memberName

		i++
	}

	err = rows.Err()
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, map[string][]*paymentStruct{"payments": payments})
}

func PatchWebtoolsPaymentsCompleted(w http.ResponseWriter, r *http.Request) {
	// Get paymentRowId
	// NOTE: uses id (primary key) field, NOT payment_id field
	paymentRowId, ok := getURLIntParam(w, r, "paymentRowId")
	if !ok {
		return
	}

	stmt := `
		UPDATE payment
		SET completed = true
		WHERE id = ?`
	_, err := db.Exec(stmt, paymentRowId)
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func PostWebtoolsGenerateCode(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}

	// Get request body
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var storeCode storeCodeStruct
	err := decoder.Decode(&storeCode)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	code := storeCode.Code
	if code == "" {
		code = generateCode(MANUAL_PAYMENT_ID_LENGTH)
	}

	completed := storeCode.Completed
	if storeCode.StoreItemId == "MEMBERSHIP" {
		completed = true
	}

	stmt := `
		INSERT INTO store_code (
			create_datetime,
			generated_by_id,
			note,
			store_item_id,
			store_item_count,
			amount,
			code,
			completed,
			redeemed)
		VALUES (datetime('now'), ?, ?, ?, ?, ?, ?, ?, false)`
	_, err = db.Exec(stmt,
		memberId,
		storeCode.Note,
		storeCode.StoreItemId,
		storeCode.StoreItemCount,
		storeCode.Amount,
		code,
		completed)
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"code": code})
}

func PostWebtoolsPayments(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get officerId, memberId
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}
	paymentMemberId, ok := getURLIntParam(w, r, "paymentMemberId")
	if !ok {
		return
	}

	// Get request body
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var payment paymentStruct
	err := decoder.Decode(&payment)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	paymentId := payment.PaymentId
	if paymentId == "" {
		paymentId = generateCode(MANUAL_PAYMENT_ID_LENGTH)
	}

	// Complete order if possible
	completed := payment.Completed
	if payment.StoreItemId == "MEMBERSHIP" {
		if !dbExtendMembership(w, paymentMemberId, payment.StoreItemCount) {
			return
		}
		completed = true
	}

	stmt := `
		INSERT INTO payment (
			create_datetime,
			created_by_id,
			note,
			member_id,
			store_item_id,
			store_item_count,
			amount,
			payment_method,
			payment_id,
			completed)
		VALUES (datetime('now'), ?, ?, ?, ?, ?, ?, 'METHOD', ?, ?)`
	_, err = db.Exec(stmt,
		memberId,
		payment.Note,
		paymentMemberId,
		payment.StoreItemId,
		payment.StoreItemCount,
		payment.Amount,
		paymentId,
		completed)
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"paymentId": paymentId})
}
