package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
)

// All fields are returned to client
type memberStruct struct {
	Id                 int    `json:"id,omitempty"`
	CreateDatetime     string `json:"createDatetime,omitempty"`
	Active             bool   `json:"active,omitempty"`
	PaidExpireDatetime string `json:"paidExpireDatetime,omitempty"`
	ECName             string `json:"ECName,omitempty"`
	ECNumber           string `json:"ECNumber,omitempty"`
	ECRelationship     string `json:"ECRelationship,omitempty"`
	/* Required fields for creating an account */
	Email           string `json:"email"`
	Name            string `json:"name"`
	CellNumber      string `json:"cellNumber"`
	Pronouns        string `json:"pronouns"`
	Birthyear       int    `json:"birthyear"`
	MedicalCond     bool   `json:"medicalCond"`
	MedicalCondDesc string `json:"medicalCondDesc"`
}

func DeleteMyAccountDelete(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}

	// Get new notifications
	notificationsArr, err := json.Marshal(notificationsStruct{})
	if !checkError(w, err) {
		return
	}
	notificationsStr := string(notificationsArr)

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if !checkError(w, err) {
		return
	}

	stmt := `
		UPDATE member
		SET
			email = '',
			name = '',
			create_datetime = 0,
			cell_number = '',
			pronouns = '',
			birth_year = 0,
			active = false,
			medical_cond = false,
			medical_cond_desc = '',
			paid_expire_datetime = 0,
			ec_name = '',
			ec_number = '',
			ec_relationship = '',
			notification_preference = ?
		WHERE id = ?`
	_, err = tx.ExecContext(ctx, stmt, notificationsStr, memberId)
	if !checkError(w, err) {
		tx.Rollback()
		return
	}

	stmt = `
		UPDATE auth
		SET
			sub = '',
			idp = '',
			idp_hash = ''
		WHERE member_id = ?`
	_, err = tx.ExecContext(ctx, stmt, memberId)
	if !checkError(w, err) {
		tx.Rollback()
		return
	}

	stmt = `
		UPDATE trip_signup
		SET
			trip_id = 0,
			leader = false,
			signup_datetime = 0,
			paid_member = false,
			attending_code = 'ATTEN',
			boot_reason = '',
			short_notice = false,
			driver = false,
			carpool = false,
			car_capacity_total = 0,
			notes = '',
			pet = false,
			attended = false
		WHERE member_id = ?`
	_, err = tx.ExecContext(ctx, stmt, memberId)
	if !checkError(w, err) {
		tx.Rollback()
		return
	}

	err = tx.Commit()
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func GetLogout(w http.ResponseWriter, r *http.Request) {
	deleteAuthCookies(w)

	http.Redirect(w, r, r.URL.Query().Get("state"), http.StatusTemporaryRedirect)
}

func GetMyAccount(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}

	stmt := `
		SELECT
			id,
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
			ec_relationship
		FROM member
		WHERE id = ?`
	var member memberStruct
	err := db.QueryRow(stmt, memberId).Scan(
		&member.Id,
		&member.Email,
		&member.Name,
		&member.CreateDatetime,
		&member.CellNumber,
		&member.Pronouns,
		&member.Birthyear,
		&member.Active,
		&member.MedicalCond,
		&member.MedicalCondDesc,
		&member.PaidExpireDatetime,
		&member.ECName,
		&member.ECNumber,
		&member.ECRelationship)
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, member)
}

func GetMyAccountName(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}

	isOfficer, ok := dbIsOfficer(w, memberId)
	if !ok {
		return
	}

	stmt := `
		SELECT name
		FROM member
		WHERE id = ?`
	var name string
	err := db.QueryRow(stmt, memberId).Scan(&name)
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"name": name, "officer": isOfficer})
}

func PatchMyAccount(w http.ResponseWriter, r *http.Request) {
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
	var member memberStruct
	err := decoder.Decode(&member)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	stmt := `
		UPDATE member
		SET
			email = ?,
			name = ?,
			cell_number = ?,
			pronouns = ?,
			birth_year = ?,
			medical_cond = ?,
			medical_cond_desc = ?,
			ec_name = ?,
			ec_number = ?,
			ec_relationship = ?
		WHERE id = ?`
	_, err = db.Exec(
		stmt,
		member.Email,
		member.Name,
		member.CellNumber,
		member.Pronouns,
		member.Birthyear,
		member.MedicalCond,
		member.MedicalCondDesc,
		member.ECName,
		member.ECNumber,
		member.ECRelationship,
		memberId)
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func PatchMyAccountDeactivate(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}

	stmt := `
		UPDATE member
		SET active = false
		WHERE id = ?`
	_, err := db.Exec(stmt, memberId)
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func PatchMyAccountReactivate(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId
	memberId, ok := dbGetMemberId(w, sub)
	if !ok {
		return
	}

	stmt := `
		UPDATE member
		SET active = 1
		WHERE id = ?`
	_, err := db.Exec(stmt, memberId)
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func PostMyAccount(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get request body
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var member memberStruct
	err := decoder.Decode(&member)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Ensure user doesn't already exist
	if !dbEnsureMemberDoesNotExist(w, sub) {
		return
	}

	// Default to prefer all notifications
	notifications := setAllPreferences()
	notificationsArr, err := json.Marshal(notifications)
	if !checkError(w, err) {
		return
	}
	notificationsStr := string(notificationsArr)

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if !checkError(w, err) {
		return
	}

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
		VALUES (?, ?, datetime('now'), ?, ?, ?, true, ?, ?, datetime('now'), '', '', '', ?)`
	result, err := tx.ExecContext(
		ctx,
		stmt,
		member.Email,
		member.Name,
		member.CellNumber,
		member.Pronouns,
		member.Birthyear,
		member.MedicalCond,
		member.MedicalCondDesc,
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

	err = tx.Commit()
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusCreated, member)
}
