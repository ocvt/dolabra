package handler

import (
  "net/http"
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

/* Error Checking & Handling */
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

/* Getters */
func dbGetMemberId(w http.ResponseWriter, subject string) (int, error) {
  stmt := `
    SELECT member_id
    FROM auth
    WHERE subject = ?`
  var memberId int
  err := db.QueryRow(stmt, subject).Scan(&memberId)
  checkError(w, err)
  return memberId, err
}

/* Misc checkers */
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

/* EXISTS checkers */
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

func dbIsMemberWithMemberId(w http.ResponseWriter, memberId int) (bool, error) {
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

func dbTripExists(w http.ResponseWriter, tripId int) (bool, error) {
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

func dbIsTripLeader(w http.ResponseWriter, tripId int, memberId int) (bool, error) {
  stmt := `
    SELECT EXISTS (
      SELECT 1
      FROM trip_signup
      WHERE (trip_id = ? AND member_id = ? AND leader = true))`
  var isLeader bool
  err := db.QueryRow(stmt, tripId, memberId).Scan(&isLeader)
  checkError(w, err)
  return isLeader, err
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

