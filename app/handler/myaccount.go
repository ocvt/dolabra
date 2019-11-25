package handler

import (
  "encoding/json"
  "net/http"
)

// All fields are returned to client
type memberStruct struct {
  Id int `json:"id,omitempty"`
  CreateDatetime string `json:"createDatetime,omitempty"`
  /* Required fields for creating an account */
  Email string `json:"email"`
  FirstName string `json:"firstName"`
  LastName string `json:"lastName"`
  CellNumber string `json:"cellNumber"`
  Gender string `json:"gender"`
  Birthyear int `json:"birthyear"`
  Active bool `json:"active"`
  MedicalCond bool `json:"medicalCond"`
  MedicalCondDesc string `json:"medicalCondDesc"`
  PaidExpireDatetime string `json:"paidExpireDatetime"`
  NotificationPreference int `json:"notificationPreference"`
  EmergencyContactName string `json:"emergencyContactName"`
  EmergencyContactNumber string `json:"emergencyContactNumber"`
  EmergencyContactRelationship string `json:"emergencyContactRelationship"`
}

func DeleteMyAccountDelete(w http.ResponseWriter, r *http.Request) {
  _, subject, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Ensure user exists
  if !dbEnsureMemberExists(w, subject) {
    return
  }

  // Get memberId
  memberId, err := dbGetMemberId(w, subject)
  if err != nil {
    return
  }

  // Clear member fields
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
      active = 0,
      medical_cond = 0,
      medical_cond_desc = '',
      paid_expire_datetime = 0,
      notification_preference = 0
    WHERE id = ?`
  _, err = db.Exec(stmt, memberId)
  if !checkError(w, err) {
    return
  }

  // Clear emergency_contact entry
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

  // Clear auth data
  stmt = `
    UPDATE auth
    SET
      type = '',
      subject = ''
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

func GetMyAccount(w http.ResponseWriter, r *http.Request) {
  _, subject, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Ensure user exists
  if !dbEnsureMemberExists(w, subject) {
    return
  }

  // Get account data
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
    member.notification_preference,
    emergency_contact.name,
    emergency_contact.number,
    emergency_contact.relationship
  FROM auth
  INNER JOIN
    member ON member.id = auth.member_id,
    emergency_contact ON emergency_contact.member_id = auth.member_id
  WHERE auth.subject = ?`
  var member memberStruct
  err := db.QueryRow(stmt, subject).Scan(
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
    &member.NotificationPreference,
    &member.EmergencyContactName,
    &member.EmergencyContactNumber,
    &member.EmergencyContactRelationship)
  if !checkError(w, err) {
    return
  }

  respondJSON(w, http.StatusOK, member)
}

func GetMyAccountName(w http.ResponseWriter, r *http.Request) {
  _, subject, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Ensure user exists
  if !dbEnsureMemberExists(w, subject) {
    return
  }

  // Only get users first name
  stmt := `
    SELECT member.first_name
    FROM auth
    INNER JOIN member ON auth.member_id = member.id
    WHERE auth.subject = ?`
  var firstname string
  err := db.QueryRow(stmt, subject).Scan(&firstname)
  if !checkError(w, err) {
    return
  }

  respondJSON(w, http.StatusOK, map[string]string{"firstName": firstname})
}

func PatchMyAccountDeactivate(w http.ResponseWriter, r *http.Request) {
  _, subject, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Ensure user exists
  if !dbEnsureMemberExists(w, subject) {
    return
  }

  // Set active to 0
  stmt := `
    UPDATE member
    SET active = 0
    WHERE EXISTS (
      SELECT member_id
      FROM auth
      WHERE auth.subject = ?)`
  _, err := db.Exec(stmt, subject)
  if !checkError(w, err) {
    return
  }

  respondJSON(w, http.StatusNoContent, nil)
}

func PatchMyAccountReactivate(w http.ResponseWriter, r *http.Request) {
  _, subject, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Ensure user exists
  if !dbEnsureMemberExists(w, subject) {
    return
  }

  // Set active to 1
  stmt := `
    UPDATE member
    SET active = 1
    WHERE EXISTS (
      SELECT member_id
      FROM auth
      WHERE auth.subject = ?)`
  _, err := db.Exec(stmt, subject)
  if !checkError(w, err) {
    return
  }

  respondJSON(w, http.StatusNoContent, nil)
}

func PostMyAccount(w http.ResponseWriter, r *http.Request) {
  idp, subject, ok := checkLogin(w, r)
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
  if !dbEnsureMemberDoesNotExist(w, subject) {
    return
  }

  // Insert new member
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
    VALUES (?, ?, ?, datetime('now'), ?, ?, ?, ?, ?, ?, ?, ?)`
  result, err := db.Exec(
    stmt,
    member.Email,
    member.FirstName,
    member.LastName,
    member.CellNumber,
    member.Gender,
    member.Birthyear,
    member.Active,
    member.MedicalCond,
    member.MedicalCondDesc,
    member.PaidExpireDatetime,
    member.NotificationPreference)
  if !checkError(w, err) {
    return
  }

  // Get new member id
  memberId, err := result.LastInsertId()
  if !checkError(w, err) {
      return
  }

  // Insert new emergency contact
  stmt = `
    INSERT INTO emergency_contact (
      member_id,
      name,
      number,
      relationship)
    VALUES (?, ?, ?, ?)`
  _, err = db.Exec(
    stmt,
    memberId,
    member.EmergencyContactName,
    member.EmergencyContactNumber,
    member.EmergencyContactRelationship)
  if !checkError(w, err) {
    return
  }

  // Insert new auth
  stmt = `
    INSERT INTO auth (
      member_id,
      type,
      subject)
    VALUES (?, ?, ?)`
  _, err = db.Exec(stmt, memberId, idp, subject)
  if !checkError(w, err) {
    return
  }

  respondJSON(w, http.StatusCreated, member)
}
