package handler

import (
  "net/http"
  "strconv"

  "github.com/go-chi/chi"
)

var MAX_INT = 9223372036854775807

/* Not DB but sort of related */
func checkLogin(w http.ResponseWriter, r *http.Request) (string, string, bool) {
  idp := r.Context().Value("idp")
  subject := r.Context().Value("subject")
  if idp == nil || subject == nil {
    respondError(w, http.StatusUnauthorized, "Member is not authenticated")
    return "", "", false
  }
  return idp.(string), subject.(string), true
}

func checkError(w http.ResponseWriter, err error) bool {
  if err != nil {
    respondError(w, http.StatusInternalServerError, err.Error())
    return false
  }
  return true
}

func checkURLParam(w http.ResponseWriter, r *http.Request, param string) (int, bool) {
  paramInt, err := strconv.Atoi(chi.URLParam(r, param))
  if err != nil {
    respondError(w, http.StatusBadRequest, err.Error())
    return 0, false
  }
  return paramInt, true
}

func redactDataIfOldTrip(w http.ResponseWriter, tripId int, tripSignup *tripSignupStruct) bool {
  stmt := `
    SELECT EXISTS (
      SELECT 1
      FROM trip
      WHERE id = ? AND datetime(start_datetime) < datetime('now', '-1 month'))`
  var exists bool
  err := db.QueryRow(stmt, tripId).Scan(&exists)
  if !checkError(w, err) {
    return false
  }

  if exists {
    tripSignup.Email = "redacted@redacted.com"
    tripSignup.FirstName = "Red"
    tripSignup.LastName = "Acted"
    tripSignup.CellNumber = "(RED) ACT-EDDD"
    tripSignup.Gender = "redacted"
    tripSignup.BirthYear = 1990
    tripSignup.MedicalCond = false
    tripSignup.MedicalCondDesc = "redacted"
  }
  return true
}

/* Error Checking & Handling */
func dbGetMemberId(w http.ResponseWriter, subject string) (int, bool) {
  if !dbEnsureMemberExists(w, subject) {
    return 0, false
  }

  stmt := `
    SELECT member_id
    FROM auth
    WHERE subject = ?`
  var memberId int
  err := db.QueryRow(stmt, subject).Scan(&memberId)
  if !checkError(w, err) {
    return 0, false
  }
  return memberId, true
}

func dbGetActiveMemberId(w http.ResponseWriter, subject string) (int, bool) {
  memberId, ok := dbGetMemberId(w, subject)
  if !ok {
    return 0, false
  }

  isActive, err := dbIsActiveMember(w, memberId)
  if err != nil {
    return 0, false
  }
  if !isActive {
    respondError(w, http.StatusBadRequest, "Member is not active.")
    return 0, false
  }
  return memberId, true
}

func dbEnsureMemberExists(w http.ResponseWriter, subject string) bool {
  exists, err := dbIsMemberWithSubject(w, subject)
  if err == nil && !exists {
      respondError(w, http.StatusNotFound, "Member is not registered.")
  }

  if err == nil {
    return exists
  }
  return false
}

func dbEnsureMemberDoesNotExist(w http.ResponseWriter, subject string) bool {
  exists, err := dbIsMemberWithSubject(w, subject)
  if err == nil && exists {
    respondError(w, http.StatusBadRequest, "Member is already registered.")
  }

  if err == nil {
    return !exists
  }
  return false
}

func dbEnsureMemberCanModifySignup(w http.ResponseWriter, tripId int, memberId int, signupId int) bool {
  if !dbEnsureActiveTrip(w, tripId) {
    return false
  }

  signupExists, err := dbIsMemberOnTrip(w, tripId, signupId)
  if err != nil {
    return false
  }
  if !signupExists {
    respondError(w, http.StatusBadRequest, "Member is not on trip.")
    return false
  }

  isCreator, err := dbIsTripCreator(w, tripId, signupId)
  if err != nil {
    return false
  }
  if isCreator {
    respondError(w, http.StatusBadRequest, "Cannot modify trip creator status.")
    return false
  }

  return true
}

// Mandatory checks for any trip signup or modification
func dbEnsureActiveTrip(w http.ResponseWriter, tripId int) bool {
  exists, err := dbIsTrip(w, tripId)
  if err != nil {
    return false
  }
  if !exists {
    respondError(w, http.StatusBadRequest, "Trip does not exist.")
    return false
  }

  isCanceled, err := dbIsTripCanceled(w, tripId)
  if err != nil {
    return false
  }
  if isCanceled {
    respondError(w, http.StatusBadRequest, "Trip is canceled.")
    return false
  }

  inPast, err := dbIsTripInPast(w, tripId)
  if err != nil {
    return false
  }
  if inPast {
    respondError(w, http.StatusBadRequest, "Trip has already occured.")
    return false
  }

  isPublished, err := dbIsTripPublished(w, tripId)
  if err != nil {
    return false
  }
  if !isPublished {
    respondError(w, http.StatusBadRequest, "Trip is not published.")
    return false
  }

  return true
}

func dbEnsureOfficerOrTripLeader(w http.ResponseWriter, tripId int, memberId int) bool {
  isOfficer, err := dbIsOfficer(w, memberId)
  if err != nil {
    return false
  }
  isTripLeader, err := dbIsTripLeader(w, tripId, memberId)
  if err != nil {
    return false
  }
  if !isOfficer && !isTripLeader {
    respondError(w, http.StatusBadRequest, "Must be officer or leader.")
    return false
  }
  return true
}

// Checks only applicable for new signups
func dbEnsureValidSignup(w http.ResponseWriter, tripId int, memberId int,
    carpool bool, driver bool, carCapacityTotal int, pet bool) bool {
  onTrip, err := dbIsMemberOnTrip(w, tripId, memberId)
  if err != nil {
    return false
  }
  if onTrip {
    respondError(w, http.StatusBadRequest, "Member is already on trip.")
    return false
  }

  lateSignup, err := dbIsLateSignup(w, tripId)
  if err != nil {
    return false
  }
  lateSignupsAllowed, err := dbIsLateSignupAllowed(w, tripId)
  if err != nil {
    return false
  }
  if lateSignup && !lateSignupsAllowed {
    respondError(w, http.StatusBadRequest, "Past trip signup deadline.")
    return false
  }

  if carpool && !driver {
    respondError(w, http.StatusBadRequest, "Cannot carpool without being a driver.")
    return false
  }
  if carCapacityTotal < 0 {
    respondError(w, http.StatusBadRequest, "Cannot have negative car capacity.")
    return false
  }

  petAllowed, err := dbIsPetAllowed(w, tripId)
  if err != nil {
    return false
  }
  if pet && !petAllowed {
    respondError(w, http.StatusBadRequest, "Cannot bring pet on trip.")
    return false
  }

  return true
}

func dbEnsureTripSignupNotCanceled(w http.ResponseWriter, tripId int, signupId int) bool {
  cancel, err := dbCheckTripSignupCode(w, tripId, signupId, "CANCEL")
  if err != nil {
    return false
  }
  if cancel {
    respondError(w, http.StatusBadRequest, "Member status is canceled.")
    return false
  }
  return true
}

/* EXISTS checkers */
// Account related
func dbIsActiveMember(w http.ResponseWriter, memberId int) (bool, error) {
  stmt := `
    SELECT EXISTS (
      SELECT 1
      FROM member
      WHERE id = ? AND active = true)`
  var exists bool
  err := db.QueryRow(stmt, memberId).Scan(&exists)
  checkError(w, err)
  return exists, err
}

func dbIsPaidMember(w http.ResponseWriter, memberId int) (bool, error) {
  stmt := `
    SELECT EXISTS (
      SELECT 1
      FROM member
      WHERE id = ? AND date(paid_expire_datetime) > datetime('now'))`
  var exists bool
  err := db.QueryRow(stmt, memberId).Scan(&exists)
  checkError(w, err)
  return exists, err
}

func dbIsMemberWithSubject(w http.ResponseWriter, subject string) (bool, error) {
  stmt := `
    SELECT EXISTS (
      SELECT 1
      FROM auth
      WHERE subject = ?)`
  var exists bool
  err := db.QueryRow(stmt, subject).Scan(&exists)
  checkError(w, err)
  return exists, err
}

func dbIsActiveMemberWithMemberId(w http.ResponseWriter, memberId int) (bool, error) {
  stmt := `
    SELECT EXISTS (
      SELECT 1
      FROM auth
      WHERE member_id = ?)`
  var exists bool
  err := db.QueryRow(stmt, memberId).Scan(&exists)
  checkError(w, err)
  return exists, err
}

func dbIsMemberOnTrip(w http.ResponseWriter, tripId int, memberId int) (bool, error) {
  stmt := `
    SELECT EXISTS (
      SELECT 1
      FROM trip_signup
      WHERE (trip_id = ? AND member_id = ?))`
  var exists bool
  err := db.QueryRow(stmt, tripId, memberId).Scan(&exists)
  checkError(w, err)
  return exists, err
}

func dbIsOfficer(w http.ResponseWriter, memberId int) (bool, error) {
  stmt := `
    SELECT EXISTS (
      SELECT 1
      FROM officer
      WHERE member_id = ?)`
  var exists bool
  err := db.QueryRow(stmt, memberId).Scan(&exists)
  checkError(w, err)
  return exists, err
}

// Trips related
func dbCheckTripSignupCode(w http.ResponseWriter, tripId int, memberId int, code string) (bool, error) {
  stmt := `
    SELECT EXISTS (
      SELECT 1
      FROM trip_signup
      WHERE trip_id = ? AND member_id = ? AND attending_code = ?)`
  var exists bool
  err := db.QueryRow(stmt, tripId, memberId, code).Scan(&exists)
  checkError(w, err)
  return exists, err
}

func dbIsLateSignupAllowed(w http.ResponseWriter, tripId int) (bool, error) {
  stmt := `
    SELECT EXISTS (
      SELECT 1
      FROM trip
      WHERE id = ? AND allow_late_signups = true)`
  var exists bool
  err := db.QueryRow(stmt, tripId).Scan(&exists)
  checkError(w, err)
  return exists, err
}

func dbIsPetAllowed(w http.ResponseWriter, tripId int) (bool, error) {
  stmt := `
    SELECT EXISTS (
      SELECT 1
      FROM trip
      WHERE id = ? AND pets_allowed = true)`
  var exists bool
  err := db.QueryRow(stmt, tripId).Scan(&exists)
  checkError(w, err)
  return exists, err
}

func dbIsTrip(w http.ResponseWriter, tripId int) (bool, error) {
  stmt := `
    SELECT EXISTS (
      SELECT 1
      FROM trip
      WHERE id = ?)`
  var exists bool
  err := db.QueryRow(stmt, tripId).Scan(&exists)
  checkError(w, err)
  return exists, err
}

func dbIsTripCreator(w http.ResponseWriter, tripId int, memberId int) (bool, error) {
  stmt := `
    SELECT EXISTS (
      SELECT 1
      FROM trip
      WHERE id = ? AND member_id = ?)`
  var exists bool
  err := db.QueryRow(stmt, tripId, memberId).Scan(&exists)
  checkError(w, err)
  return exists, err
}

func dbIsTripCanceled(w http.ResponseWriter, tripId int) (bool, error) {
  stmt := `
    SELECT EXISTS (
      SELECT 1
      FROM trip
      WHERE id = ? AND cancel = true)`
  var exists bool
  err := db.QueryRow(stmt, tripId).Scan(&exists)
  checkError(w, err)
  return exists, err
}

func dbIsTripLeader(w http.ResponseWriter, tripId int, memberId int) (bool, error) {
  stmt := `
    SELECT EXISTS (
      SELECT 1
      FROM trip_signup
      WHERE (trip_id = ? AND member_id = ? AND leader = true))`
  var isTripLeader bool
  err := db.QueryRow(stmt, tripId, memberId).Scan(&isTripLeader)
  checkError(w, err)
  return isTripLeader, err
}

func dbIsTripInFuture(w http.ResponseWriter, tripId int) (bool, error) {
  stmt := `
    SELECT EXISTS (
      SELECT 1
      FROM trip
      WHERE id = ? AND datetime('now') < datetime(start_datetime))`
  var exists bool
  err := db.QueryRow(stmt, tripId).Scan(&exists)
  checkError(w, err)
  return exists, err
}

func dbIsTripInPast(w http.ResponseWriter, tripId int) (bool, error) {
  stmt := `
    SELECT EXISTS (
      SELECT 1
      FROM trip
      WHERE id = ? AND datetime(start_datetime) < datetime('now'))`
  var exists bool
  err := db.QueryRow(stmt, tripId).Scan(&exists)
  checkError(w, err)
  return exists, err
}

func dbIsTripPublished(w http.ResponseWriter, tripId int) (bool, error) {
  stmt := `
    SELECT EXISTS (
      SELECT 1
      FROM trip
      WHERE id = ? AND publish = true)`
  var exists bool
  err := db.QueryRow(stmt, tripId).Scan(&exists)
  checkError(w, err)
  return exists, err
}

func dbIsLateSignup(w http.ResponseWriter, tripId int) (bool, error) {
  stmt := `
    SELECT EXISTS (
      SELECT 1
      FROm trip
      WHERE id = ? AND datetime(start_datetime) < datetime('now', '+12 hours'))`
  var exists bool
  err := db.QueryRow(stmt, tripId).Scan(&exists)
  checkError(w, err)
  return exists, err
}

