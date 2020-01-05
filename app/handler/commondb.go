package handler

import (
  "encoding/json"
  "net/http"
  "strconv"

  "github.com/go-chi/chi"
)

/* Contains:
   - General helpers
   - "Ensure" helpers to guarantee a specific result
   - EXISTS helpers
*/

var MAX_INT = 4294967295

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

func dbCreateSystemMember() error {
  notifications := setAllPreferences()
  notificationsArr, err := json.Marshal(notifications)
  if err != nil {
    return err
  }
  notificationsStr := string(notificationsArr)

  stmt := `
    INSERT OR REPLACE INTO member
    VALUES (0, ?, ?, ?, datetime('now'), '555-555-5555', 'Prefer not to say',
            1990, true, 'Robot', 'Require frequent oiling',
            datetime('now', '+100 years'), ?)`
  _, err = db.Exec(
    stmt,
    SMTP_FROM_EMAIL_DEFAULT,
    SMTP_FROM_FIRST_NAME_DEFAULT,
    SMTP_FROM_LAST_NAME_DEFAULT,
    notificationsStr)

  return err
}

func dbCreateNullTrip() error {
  stmt := `
    INSERT OR REPLACE INTO trip
    VALUES (0, datetime('now'), true, false, 0, false, false, false, false,
            false, "", 0, "Null Trip for Announcement logs", "TRIP_OTHER",
            datetime('now'), datetime('now'), "", "", "", "", "", 0, 0, "",
            "", false, "")`
  _, err := db.Exec(stmt)
  return err
}

/* "Ensure" helpers */
// TODO organize A-Z
func dbEnsureMemberWantsNotification(w http.ResponseWriter, memberId int, notificationType string) bool {
  notifications, ok := dbGetMemberNotifications(w, memberId)
  if !ok {
    return false
  }

  notificationsArr, err := json.Marshal(notifications)
  if !checkError(w, err) {
    return false
  }

  var notificationsStrMap = map[string]bool{}
  err = json.Unmarshal(notificationsArr, &notificationsStrMap)
  if !checkError(w, err) {
    return false
  }

  return notificationsStrMap[notificationType]
}

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

func dbEnsureNotOfficer(w http.ResponseWriter, memberId int) bool {
  isOfficer, err := dbIsOfficer(w, memberId)
  if err != nil {
    return false
  }
  if isOfficer {
    respondError(w, http.StatusBadRequest, "Must not be officer.")
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

func dbGetItemCount(w http.ResponseWriter, storeItemId string,
    stripePaymentId string) (int, int, bool) {
  stmt := `
    SELECT
      member_id,
      store_item_count
    FROM payment
    WHERE strip_payment_id = ?`
  var memberId int
  var storeItemCount int
  err := db.QueryRow(stmt, stripePaymentId).Scan(&memberId, &storeItemCount)
  if !checkError(w, err) {
    return 0, 0, false
  }

  return memberId, storeItemCount, true
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

func dbGetMemberNotifications(w http.ResponseWriter, memberId int) (notificationsStruct, bool) {
  stmt := `
    SELECT notification_preference
    FROM member
    WHERE id = ?`
  var notificationsStr string
  err := db.QueryRow(stmt, memberId).Scan(&notificationsStr)
  if !checkError(w, err) {
    return notificationsStruct{}, false
  }

  var notifications = notificationsStruct{}
  err = json.Unmarshal([]byte(notificationsStr), &notifications)
  if !checkError(w, err) {
    return notificationsStruct{}, false
  }

  return notifications, true
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

func dbEnsureMemberIdExists(w http.ResponseWriter, memberId int) bool {
  exists, err := dbIsMemberWithMemberId(w, memberId)
  if err == nil && !exists {
      respondError(w, http.StatusNotFound, "Member is not registered.")
  }

  if err == nil {
    return exists
  }
  return false
}

func dbInsertPayment(w http.ResponseWriter, enteredById int, note string,
    memberId int, storeItemId string, storeItemCount int, amount int,
    paymentMethod string, paymentId string, completed bool) bool {
  stmt := `
    INSERT INTO payment
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
    VALUES (datetime('now'), ?, ?, ?, ?, ?, ?, 'STRIPE', ?, ?)`
  _, err := db.Exec(stmt, enteredById, note, memberId, storeItemId,
      storeItemCount, amount, paymentId, completed)
  return checkError(w, err)
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
