package handler

import (
  "database/sql"
  "encoding/json"
  "net/http"
)

type accountDataStruct struct {
  MemberId int `json:"memberId,omitempty"`
  Email string `json:"email"`
  FirstName string `json:"firstName"`
  LastName string `json:"lastName"`
  DatetimeCreated string `json:"datetimeCreated"`
  CellNumber string `json:"cellNumber"`
  Gender string `json:"gender"`
  Birthyear int `json:"birthyear"`
  Active bool `json:"active"`
  MedicalCond bool `json:"medicalCond"`
  MedicalCondDesc string `json:"medicalCondDesc"`
  PaidExpireDatetime string `json:"paidExpirationDatetime"` // int?? TODO
  NotificationPreference int `json:"notificationPreference"`
  EmergencyContactName string `json:"emergencyContactName"`
  EmergencyContactNumber string `json:"emergencyContactNumber"`
  EmergencyContactRelationship string `json:"emergencyContactRelationship"`
}

func checkLogin(w http.ResponseWriter, r *http.Request) (interface{}, interface{}, bool) {
  idp := r.Context().Value("idp")
  userSub := r.Context().Value("userSub")
  if idp == nil || userSub == nil {
    respondError(w, http.StatusUnauthorized, "User is not authenticated")
    return nil, nil, false
  }
  return idp, userSub, true
}

func dbCheckError(w http.ResponseWriter, err error) bool {
  if err != nil && err != sql.ErrNoRows {
    respondError(w, http.StatusInternalServerError, err.Error())
    return false
  }
  return true
}

func DeleteMyAccountDelete(w http.ResponseWriter, r *http.Request) {

  // TODO
  respondError(w, http.StatusNotImplemented, "not implemented")
}

func GetMyAccount(w http.ResponseWriter, r *http.Request) {
  _, userSub, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Get account data
  stmt := `
  SELECT
    ocvt_users.member_id,
    ocvt_users.email,
    ocvt_users.name_first,
    ocvt_users.name_last,
    ocvt_users.datetime_created,
    ocvt_users.cell_number,
    ocvt_users.gender,
    ocvt_users.birth_year,
    ocvt_users.active,
    ocvt_users.medical_cond,
    ocvt_users.medical_cond_desc,
    ocvt_users.paid_expire_datetime,
    ocvt_users.notification_preference,
    emergency_contacts.contact_name,
    emergency_contacts.contact_number,
    emergency_contacts.contact_relationship
  FROM ocvt_auth
  INNER JOIN
    ocvt_users ON ocvt_users.member_id = ocvt_auth.member_id,
    emergency_contacts ON emergency_contacts.member_id = ocvt_auth.member_id
  WHERE ocvt_auth.auth_sub = ?`

  var accountData accountDataStruct
  err := db.QueryRow(stmt, userSub).Scan(
    &accountData.MemberId,
    &accountData.Email,
    &accountData.FirstName,
    &accountData.LastName,
    &accountData.DatetimeCreated,
    &accountData.CellNumber,
    &accountData.Gender,
    &accountData.Birthyear,
    &accountData.Active,
    &accountData.MedicalCond,
    &accountData.MedicalCondDesc,
    &accountData.PaidExpireDatetime,
    &accountData.NotificationPreference,
    &accountData.EmergencyContactName,
    &accountData.EmergencyContactNumber,
    &accountData.EmergencyContactRelationship)
  if err != nil && err == sql.ErrNoRows {
    respondError(w, http.StatusNotFound, "User is not registered")
    return
  } else if ok := dbCheckError(w, err); !ok {
    return
  }

  respondJSON(w, http.StatusOK, accountData)
}

func GetMyAccountName(w http.ResponseWriter, r *http.Request) {
  _, userSub, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Only get users first name
  stmt := `
    SELECT ocvt_users.name_first
    FROM ocvt_auth
    INNER JOIN ocvt_users ON ocvt_auth.member_id = ocvt_users.member_id
    WHERE ocvt_auth.auth_sub = ?`

  var firstname string
  err := db.QueryRow(stmt, userSub).Scan(&firstname)
  if err != nil && err == sql.ErrNoRows {
    respondError(w, http.StatusNotFound, "User is not registered")
    return
  } else if ok := dbCheckError(w, err); !ok {
    return
  }

  respondJSON(w, http.StatusOK, map[string]string{"firstname": firstname})
}

func PatchMyAccountDeactivate(w http.ResponseWriter, r *http.Request) {
  _, userSub, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Set active to 0
  stmt := `
    UPDATE ocvt_users
    SET active = 0
    WHERE EXISTS (
      SELECT member_id
      FROM ocvt_auth
      WHERE ocvt_auth.auth_sub = ?)`

  _, err := db.Exec(stmt, userSub)
  if err != nil && err == sql.ErrNoRows {
    respondError(w, http.StatusNotFound, "User is not registered")
    return
  } else if ok := dbCheckError(w, err); !ok {
    return
  }

  respondJSON(w, http.StatusOK, map[string]string{})
}

func PatchMyAccountReactivate(w http.ResponseWriter, r *http.Request) {
  _, userSub, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Set active to 1
  stmt := `
    UPDATE ocvt_users
    SET active = 1
    WHERE EXISTS (
      SELECT member_id
      FROM ocvt_auth
      WHERE ocvt_auth.auth_sub = ?)`

  _, err := db.Exec(stmt, userSub)
  if err != nil && err == sql.ErrNoRows {
    respondError(w, http.StatusNotFound, "User is not registered")
    return
  } else if ok := dbCheckError(w, err); !ok {
    return
  }

  respondJSON(w, http.StatusOK, map[string]string{})
}

func PostMyAccountRegister(w http.ResponseWriter, r *http.Request) {
  idp, userSub, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Get request body
  decoder := json.NewDecoder(r.Body)
  decoder.DisallowUnknownFields()
  var accountData accountDataStruct
  err := decoder.Decode(&accountData)
  if err != nil {
    respondError(w, http.StatusBadRequest, err.Error())
    return
  }

  // Ensure user doesn't already exist
  stmt := `
    SELECT member_id
    FROM ocvt_auth
    WHERE auth_sub = ?`
  var member_id_tmp int
  err = db.QueryRow(stmt, userSub).Scan(&member_id_tmp)
  if ok := dbCheckError(w, err); !ok {
    return
  } else if err == nil {
    respondError(w, http.StatusConflict, "User is already registered")
    return
  }

  // Insert user into ocvt_users
  stmt = `
    INSERT INTO ocvt_users (
      email,
      name_first,
      name_last,
      datetime_created,
      cell_number,
      gender,
      birth_year,
      active,
      medical_cond,
      medical_cond_desc,
      paid_expire_datetime,
      notification_preference)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
  result, err := db.Exec(
    stmt,
    accountData.Email,
    accountData.FirstName,
    accountData.LastName,
    accountData.DatetimeCreated,
    accountData.CellNumber,
    accountData.Gender,
    accountData.Birthyear,
    accountData.Active,
    accountData.MedicalCond,
    accountData.MedicalCondDesc,
    accountData.PaidExpireDatetime,
    accountData.NotificationPreference)
  if ok := dbCheckError(w, err); !ok {
    return
  }

  // Get new member id
  member_id, err := result.LastInsertId()
  if ok := dbCheckError(w, err); !ok {
      return
  }

  // Insert user into emergency_contacts
  stmt = `
    INSERT INTO emergency_contacts (
      member_id,
      contact_name,
      contact_number,
      contact_relationship)
    VALUES (?, ?, ?, ?)`
  _, err = db.Exec(
    stmt,
    member_id,
    accountData.EmergencyContactName,
    accountData.EmergencyContactNumber,
    accountData.EmergencyContactRelationship)
  if ok := dbCheckError(w, err); !ok {
    return
  }

  // Insert user into ocvt_auth
  stmt = `
    INSERT INTO ocvt_auth (
      member_id,
      auth_type,
      auth_sub)
    VALUES (?, ?, ?)`
  _, err = db.Exec(stmt, member_id, idp, userSub)
  if ok := dbCheckError(w, err); !ok {
    return
  }

  respondJSON(w, http.StatusCreated, accountData)
}
