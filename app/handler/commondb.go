package handler

import (
  "net/http"
  "strconv"

  "github.com/go-chi/chi"
)

/* Contains:
   - General helpers
   - "Ensure" helpers to guarantee a specific result
   - EXISTS helpers
*/

var MAX_INT = 9223372036854775807

/* General helpers */
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

/* "Ensure" helpers */
func dbEnsureOfficer(w http.ResponseWriter, memberId int) bool {
  isOfficer, err := dbIsOfficer(w, memberId)
  if err != nil {
    return false
  }
  if !isOfficer {
    respondError(w, http.StatusBadRequest, "Must be officer.")
    return false
  }
  return true
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

func dbGetMemberEmail(w http.ResponseWriter, memberId int) (string, bool) {
  stmt := `
    SELECT email
    FROM member
    WHERE id = ?`
  var email string
  err := db.QueryRow(stmt, memberId).Scan(&email)
  if !checkError(w, err) {
    return "", false
  }
  return email, true
}

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

func dbGetMemberName(w http.ResponseWriter, memberId int) (string, bool) {
  stmt := `
    SELECT first_name || ' ' || last_name AS full_name
    FROM member
    WHERE id = ?`
  var fullName string
  err := db.QueryRow(stmt, memberId).Scan(&fullName)
  if !checkError(w, err) {
    return "", false
  }
  return fullName, true
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

/* EXISTS helpers */
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

func dbIsOfficer(w http.ResponseWriter, memberId int) (bool, error) {
  stmt := `
    SELECT EXISTS (
      SELECT 1
      FROM officer
      WHERE member_id = ? AND date(expire_datetime) > datetime('now'))`
  var exists bool
  err := db.QueryRow(stmt, memberId).Scan(&exists)
  checkError(w, err)
  return exists, err
}
