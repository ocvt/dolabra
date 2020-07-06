package handler

import (
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
	/* Required fields for creating an account */
	Email           string `json:"email"`
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	CellNumber      string `json:"cellNumber"`
	Gender          string `json:"gender"`
	Birthyear       int    `json:"birthyear"`
	MedicalCond     bool   `json:"medicalCond"`
	MedicalCondDesc string `json:"medicalCondDesc"`
	/* Allow independent updates of member & emergency info */
	EmergencyContactName         string `json:"emergencyContactName,omitempty"`
	EmergencyContactNumber       string `json:"emergencyContactNumber,omitempty"`
	EmergencyContactRelationship string `json:"emergencyContactRelationship,omitempty"`
}

// Separate struct for updating emergency info independently
type emergencyStruct struct {
	EmergencyContactName         string `json:"emergencyContactName"`
	EmergencyContactNumber       string `json:"emergencyContactNumber"`
	EmergencyContactRelationship string `json:"emergencyContactRelationship"`
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

	stmt := `
		UPDATE member
		SET
			email = '',
			first_name = '',
			last_name = '',
			create_datetime = 0,
			cell_number = '',
			gender = '',
			birth_year = 0,
			active = false,
			medical_cond = false,
			medical_cond_desc = '',
			paid_expire_datetime = 0,
			notification_preference = ?
		WHERE id = ?`
	_, err = db.Exec(stmt, notificationsStr, memberId)
	if !checkError(w, err) {
		return
	}

	stmt = `
		UPDATE emergency_contact
		SET
			name = '',
			number = '',
			relationship = ''
		WHERE member_id = ?`
	_, err = db.Exec(stmt, memberId)
	if !checkError(w, err) {
		return
	}

	stmt = `
		UPDATE auth
		SET
			sub = '',
			idp = '',
			idp_sub = ''
		WHERE member_id = ?`
	_, err = db.Exec(stmt, memberId)
	if !checkError(w, err) {
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
	_, err = db.Exec(stmt, memberId)
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
			member.id,
			member.email,
			member.first_name,
			member.last_name,
			member.create_datetime,
			member.cell_number,
			member.gender,
			member.birth_year,
			member.active,
			member.medical_cond,
			member.medical_cond_desc,
			member.paid_expire_datetime,
			emergency_contact.name,
			emergency_contact.number,
			emergency_contact.relationship
		FROM member
		INNER JOIN emergency_contact ON emergency_contact.member_id = member.id
		WHERE member.id = ?`
	var member memberStruct
	err := db.QueryRow(stmt, memberId).Scan(
		&member.Id,
		&member.Email,
		&member.FirstName,
		&member.LastName,
		&member.CreateDatetime,
		&member.CellNumber,
		&member.Gender,
		&member.Birthyear,
		&member.Active,
		&member.MedicalCond,
		&member.MedicalCondDesc,
		&member.PaidExpireDatetime,
		&member.EmergencyContactName,
		&member.EmergencyContactNumber,
		&member.EmergencyContactRelationship)
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

	isOfficer, err := dbIsOfficer(w, memberId)
	if err != nil {
		return
	}

	stmt := `
		SELECT first_name
		FROM member
		WHERE id = ?`
	var firstName string
	err = db.QueryRow(stmt, memberId).Scan(&firstName)
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"firstName": firstName, "officer": isOfficer})
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
			first_name = ?,
			last_name = ?,
			cell_number = ?,
			gender = ?,
			birth_year = ?,
			medical_cond = ?,
			medical_cond_desc = ?
		WHERE id = ?`
	_, err = db.Exec(
		stmt,
		member.Email,
		member.FirstName,
		member.LastName,
		member.CellNumber,
		member.Gender,
		member.Birthyear,
		member.MedicalCond,
		member.MedicalCondDesc,
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

func PatchMyAccountEmergency(w http.ResponseWriter, r *http.Request) {
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
	var emergency emergencyStruct
	err := decoder.Decode(&emergency)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	stmt := `
		UPDATE emergency_contact
		SET
			name = ?,
			number = ?,
			relationship = ?
		WHERE member_id = ?`
	_, err = db.Exec(stmt,
		emergency.EmergencyContactName,
		emergency.EmergencyContactNumber,
		emergency.EmergencyContactRelationship,
		memberId)
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

	stmt := `
		INSERT INTO member (
			email,
			first_name,
			last_name,
			create_datetime,
			cell_number,
			gender,
			birth_year,
			active,
			medical_cond,
			medical_cond_desc,
			paid_expire_datetime,
			notification_preference)
		VALUES (?, ?, ?, datetime('now'), ?, ?, ?, 1, ?, ?, datetime('now'), ?)`
	result, err := db.Exec(
		stmt,
		member.Email,
		member.FirstName,
		member.LastName,
		member.CellNumber,
		member.Gender,
		member.Birthyear,
		member.MedicalCond,
		member.MedicalCondDesc,
		notificationsStr)
	if !checkError(w, err) {
		return
	}

	// Get new member id
	memberId, err := result.LastInsertId()
	if !checkError(w, err) {
		return
	}

	// Insert placeholder values for emergency contact
	stmt = `
		INSERT INTO emergency_contact (
			member_id,
			name,
			number,
			relationship)
		VALUES (?, '', '', '')`
	_, err = db.Exec(
		stmt,
		memberId)
	if !checkError(w, err) {
		return
	}

	// Insert new auth
	stmt = `
		UPDATE auth
			SET member_id = ?
		WHERE sub = ?`
	_, err = db.Exec(stmt, memberId, sub)
	if !checkError(w, err) {
		return
	}

	// Remove from quicksignup table
	stmt = `
		DELETE FROM quick_signup
		WHERE email = ?`
	_, err = db.Exec(stmt, member.Email)
	if !checkError(w, err) {
		return
	}

	// Mark as officer if first person
	initOfficerId := int64(2)
	if len(os.Getenv("DEV")) > 0 {
		initOfficerId = 4
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
		_, err = db.Exec(stmt, memberId)
		if !checkError(w, err) {
			return
		}
	}

	respondJSON(w, http.StatusCreated, member)
}
