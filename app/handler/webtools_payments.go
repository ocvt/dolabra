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
	Email         string `json:"email,omitempty"`
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
	Id              int    `json:"id,omitempty"`
	CreateDatetime  string `json:"createDatetime,omitempty"`
	GeneratedById   int    `json:"generatedById,omitempty"`
	GeneratedByName string `json:"generatedByName,omitempty"`
	Redeemed        bool   `json:"redeemed,omitempty"`
	/* Required fields */
	Note           string `json:"note"`
	StoreItemId    string `json:"storeItemId"`
	StoreItemCount int    `json:"storeItemCount"`
	Amount         int    `json:"amount"`
	Code           string `json:"code"`
	Completed      bool   `json:"completed"`
}

func GetWebtoolsCodes(w http.ResponseWriter, r *http.Request) {
	stmt := `
		SELECT *
		FROM store_code
		WHERE redeemed = false`
	rows, err := db.Query(stmt)
	if !checkError(w, err) {
		return
	}
	defer rows.Close()

	var codes = []*storeCodeStruct{}
	i := 0
	for rows.Next() {
		codes = append(codes, &storeCodeStruct{})
		err = rows.Scan(
			&codes[i].Id,
			&codes[i].CreateDatetime,
			&codes[i].GeneratedById,
			&codes[i].Note,
			&codes[i].StoreItemId,
			&codes[i].StoreItemCount,
			&codes[i].Amount,
			&codes[i].Code,
			&codes[i].Completed,
			&codes[i].Redeemed)
		if !checkError(w, err) {
			return
		}

		generatedByName, ok := dbGetMemberName(w, codes[i].GeneratedById)
		if !ok {
			return
		}
		codes[i].GeneratedByName = generatedByName

		i++
	}

	err = rows.Err()
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, map[string][]*storeCodeStruct{"codes": codes})
}

func GetWebtoolsPayments(w http.ResponseWriter, r *http.Request) {
	stmt := `
		SELECT
			member.first_name || ' ' || member.last_name AS full_name,
			member.email,
			payment.id,
			payment.create_datetime,
			payment.note,
			payment.member_id,
			payment.store_item_id,
			payment.store_item_count,
			payment.amount,
			payment.payment_method,
			payment.payment_id,
			payment.completed
		FROM payment
		INNER JOIN member ON member.id = payment.member_id
		ORDER BY datetime(payment.create_datetime) DESC`
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
			&payments[i].MemberName,
			&payments[i].Email,
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
		payments[i].EnteredByName = enteredByName

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

	// If multiple items associated with a single purchase, re-use same code (with same purchase amount)
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

	// Get memberId
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

	// If multiple items associated with a single purchase, re-use same id (with same purchase amount)
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
		VALUES (datetime('now'), ?, ?, ?, ?, ?, ?, 'MANUAL', ?, ?)`
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
