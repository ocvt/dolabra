package handler

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

/* Contains:
- General helpers
- "Ensure" helpers to guarantee a specific result
- EXISTS helpers
*/

const MAX_INT = 4294967295
const LETTER_BYTES = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

/* General non-db helpers */
func checkError(w http.ResponseWriter, err error) bool {
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return false
	}
	return true
}

func checkLogin(w http.ResponseWriter, r *http.Request) (string, bool) {
	sub := r.Context().Value("sub")
	if sub == nil {
		respondError(w, http.StatusInternalServerError, "Error getting sub from context")
		return "", false
	}
	return sub.(string), true
}

func getURLIntParam(w http.ResponseWriter, r *http.Request, param string) (int, bool) {
	paramInt, err := strconv.Atoi(chi.URLParam(r, param))
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return 0, false
	}
	return paramInt, true
}

// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
func generateCode(n int) string {
	byteArr := make([]byte, n)

	for i := range byteArr {
		byteArr[i] = LETTER_BYTES[rand.Intn(len(LETTER_BYTES))]
	}
	return string(byteArr)
}

/* General db helpers */
func dbCreateNullTrip() error {
	stmt := `
		INSERT OR REPLACE INTO trip
		VALUES (0, '1990-01-01 00:00:00', true, false, 0, false, false, false,
						false, false, '', 0, 'Null Trip for Announcement logs',
						'TRIP_OTHER', '1990-01-02 00:00:00', '1990-01-03 00:00:00', '', '',
						'', '', '', 0, 0, '', '', false, '')`
	_, err := db.Exec(stmt) // sqlvet: ignore
	return err
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
		notificationsStr) // sqlvet: ignore

	return err
}

/* "Ensure" helpers */
func dbCheckMemberWantsNotification(memberId int, notificationType string) bool {
	stmt := `
		SELECT notification_preference
		FROM member
		WHERE id = ?`
	var notificationsStr string
	err := db.QueryRow(stmt, memberId).Scan(&notificationsStr)
	if err != nil {
		log.Fatal(err)
	}

	notifications := map[string]bool{}
	err = json.Unmarshal([]byte(notificationsStr), &notifications)
	if err != nil {
		log.Fatal(err)
	}
	return notifications[notificationType]
}

func dbEnsureMemberDoesNotExist(w http.ResponseWriter, sub string) bool {
	exists, err := dbIsMemberWithSub(w, sub)
	if err != nil {
		return false
	}

	if exists {
		respondError(w, http.StatusBadRequest, "Member is already registered.")
		return false
	}

	return true
}

func dbEnsureMemberExists(w http.ResponseWriter, sub string) bool {
	exists, err := dbIsMemberWithSub(w, sub)
	if err != nil {
		return false
	}

	if !exists {
		respondError(w, http.StatusNotFound, "Member is not registered.")
		return false
	}

	return true
}

func dbEnsureMemberIdExists(w http.ResponseWriter, memberId int) bool {
	exists, err := dbIsMemberWithMemberId(w, memberId)
	if err != nil {
		return false
	}

	if !exists {
		respondError(w, http.StatusNotFound, "Member is not registered.")
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

func dbEnsureTripExists(w http.ResponseWriter, tripId int) bool {
	exists, err := dbIsTrip(w, tripId)
	if err != nil {
		return false
	}

	if !exists {
		respondError(w, http.StatusNotFound, "Trip does not exist.")
		return false
	}

	return true
}

func dbExtendMembership(w http.ResponseWriter, memberId int, years int) bool {
	for i := 0; i < years; i++ {
		stmt := `
			UPDATE member
			SET paid_expire_datetime = datetime(paid_expire_datetime, '+1 year')
			WHERE id = ?`
		_, err := db.Exec(stmt, memberId)
		if !checkError(w, err) {
			return false
		}
	}
	return true
}

func dbGetActiveMemberId(w http.ResponseWriter, sub string) (int, bool) {
	memberId, ok := dbGetMemberId(w, sub)
	if !ok {
		return 0, false
	}

	isActive, err := dbIsActiveMember(w, memberId)
	if err != nil {
		return 0, false
	}
	if !isActive {
		respondError(w, http.StatusForbidden, "Member is not active.")
		return 0, false
	}
	return memberId, true
}

func dbGetItemCount(w http.ResponseWriter, storeItemId string,
	paymentMethod string, paymentId string) (int, int, bool) {
	stmt := `
		SELECT
			member_id,
			store_item_count
		FROM payment
		WHERE payment_method = ? AND payment_id = ?`
	var memberId int
	var storeItemCount int
	err := db.QueryRow(stmt, paymentMethod, paymentId).Scan(&memberId, &storeItemCount)
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

func dbGetMemberId(w http.ResponseWriter, sub string) (int, bool) {
	if !dbEnsureMemberExists(w, sub) {
		return 0, false
	}

	stmt := `
		SELECT member_id
		FROM auth
		WHERE sub = ?`
	var memberId int
	err := db.QueryRow(stmt, sub).Scan(&memberId)
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

func dbGetMemberNameEmail(memberId int) (string, string) {
	stmt := `
		SELECT
			email,
			first_name || ' ' || last_name AS full_name
		FROM member
		WHERE id = ?`
	var name, email string
	err := db.QueryRow(stmt, memberId).Scan(&email, &name)
	if err != nil {
		log.Fatal(err)
	}
	return name, email
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

func dbGetMemberSubWithIdp(w http.ResponseWriter, idp string, idpSub string) (string, bool) {
	stmt := `
		SELECT sub
		FROM auth
		WHERE idp = ? AND idp_sub = ?`
	var sub string
	err := db.QueryRow(stmt, idp, idpSub).Scan(&sub)
	if !checkError(w, err) {
		return "", false
	}
	return sub, true
}

func dbInsertPayment(w http.ResponseWriter, enteredById int, note string,
	memberId int, storeItemId string, storeItemCount int, amount int,
	paymentMethod string, paymentId string, completed bool) bool {
	stmt := `
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
		VALUES (datetime('now'), ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(stmt,
		enteredById,
		note,
		memberId,
		storeItemId,
		storeItemCount,
		amount,
		paymentMethod,
		paymentId,
		completed)
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

func dbIsMemberWithIdp(w http.ResponseWriter, idp string, idpSub string) (bool, error) {
	stmt := `
		SELECT EXISTS (
			SELECT 1
			FROM auth
			WHERE idp = ? AND idp_sub = ?)`
	var exists bool
	err := db.QueryRow(stmt, idp, idpSub).Scan(&exists)
	checkError(w, err)
	return exists, err
}

func dbIsMemberWithMemberId(w http.ResponseWriter, memberId int) (bool, error) {
	stmt := `
		SELECT EXISTS (
			SELECT 1
			FROM auth
			WHERE member_id > 0 AND member_id = ?)`
	var exists bool
	err := db.QueryRow(stmt, memberId).Scan(&exists)
	checkError(w, err)
	return exists, err
}

func dbIsMemberWithSub(w http.ResponseWriter, sub string) (bool, error) {
	stmt := `
		SELECT EXISTS (
			SELECT 1
			FROM auth
			WHERE member_id > 0 AND sub = ?)`
	var exists bool
	err := db.QueryRow(stmt, sub).Scan(&exists)
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
