package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
)

// Similar to memberStruct but only used to migrate users from old site
type oldMemberStruct struct {
	Id int `json:"id,omitempty"`
	/* Converted values from old site */
	Name     string `json:"name,omitempty"`
	Pronouns string `json:"pronouns,omitempty"`
	/* Required fields to lookup oldsite data */
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Birthyear int    `json:"birthyear"`
	Gender    string `json:"gender"`
}

func PostMyAccountMigrate(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get request body
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var oldMember oldMemberStruct
	err := decoder.Decode(&oldMember)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Convert oldsite M/F and name to new site format
	oldMember.Name = oldMember.FirstName + " " + oldMember.LastName
	oldMember.Pronouns = "he/him"
	if oldMember.Gender == "F" {
		oldMember.Pronouns = "she/her"
	}

	// Ensure user doesn't already exist
	if !dbEnsureMemberDoesNotExist(w, sub) {
		return
	}

	// Lookup all old member data
	member, notificationsStr, payments, ok := lookupOldMember(w, oldMember)
	if !ok {
		return
	}

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if !checkError(w, err) {
		return
	}

	// Copy member data
	stmt := `
		INSERT INTO member (
			email,
			name,
			create_datetime,
			cell_number,
			pronouns,
			birth_year,
			active,
			medical_cond,
			medical_cond_desc,
			paid_expire_datetime,
			ec_name,
			ec_number,
			ec_relationship,
			notification_preference)
		VALUES (?, ?, ?, ?, ?, ?, true, ?, ?, ?, ?, ?, ?, ?)`
	result, err := tx.ExecContext(
		ctx,
		stmt,
		member.Email,
		member.Name,
		member.CreateDatetime,
		member.CellNumber,
		member.Pronouns,
		member.Birthyear,
		member.MedicalCond,
		member.MedicalCondDesc,
		member.PaidExpireDatetime,
		member.ECName,
		member.ECNumber,
		member.ECRelationship,
		notificationsStr)
	if !checkError(w, err) {
		tx.Rollback()
		return
	}

	// Get new member id
	memberId, err := result.LastInsertId()
	if !checkError(w, err) {
		tx.Rollback()
		return
	}

	// Insert new auth
	stmt = `
		UPDATE auth
			SET member_id = ?
		WHERE sub = ?`
	_, err = tx.ExecContext(ctx, stmt, memberId, sub)
	if !checkError(w, err) {
		tx.Rollback()
		return
	}

	// Remove from quicksignup table
	stmt = `
		DELETE FROM quick_signup
		WHERE email = ?`
	_, err = tx.ExecContext(ctx, stmt, member.Email)
	if !checkError(w, err) {
		tx.Rollback()
		return
	}

	// Mark as officer if first person
	initOfficerId := int64(8000001)
	if os.Getenv("DEV") == "1" {
		initOfficerId = 8000004
	}
	if memberId == initOfficerId {
		stmt = `
			INSERT INTO officer (
				member_id,
				create_datetime,
				expire_datetime,
				position,
				security)
			VALUES (?, datetime('now'), datetime('now', '+1000 years'), 'Super Admin', 100)`
		_, err = tx.ExecContext(ctx, stmt, memberId)
		if !checkError(w, err) {
			tx.Rollback()
			return
		}
	}

	// Copy all payments (membership is already applied in migrate.py script)
	// All fields are copied directly except for:
	//  - entered_by_id: always system user (8000000) since the actual user who entered may not exist
	//  - member id: different upon migration
	for p := payments.Front(); p != nil; p = p.Next() {
		payment := p.Value.(paymentStruct)
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
			VALUES (?, 8000000, ?, ?, ?, ?, ?, ?, ?, ?)`
		_, err = tx.ExecContext(ctx, stmt,
			payment.CreateDatetime,
			payment.Note,
			memberId,
			payment.StoreItemId,
			payment.StoreItemCount,
			payment.Amount,
			payment.PaymentMethod,
			payment.PaymentId,
			payment.Completed)
		if !checkError(w, err) {
			tx.Rollback()
			return
		}
	}

	// Delete old member data to prevent the same person being migrated twice
	// member.Id is old member id (new id is memberId)
	stmt = `
		DELETE FROM oldsite_payment
		WHERE member_id = ?`
	_, err = tx.ExecContext(ctx, stmt, member.Id)
	if !checkError(w, err) {
		tx.Rollback()
		return
	}

	stmt = `
		DELETE FROM oldsite_member
		WHERE id = ?`
	_, err = tx.ExecContext(ctx, stmt, member.Id)
	if !checkError(w, err) {
		tx.Rollback()
		return
	}

	err = tx.Commit()
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusCreated, member)
}
