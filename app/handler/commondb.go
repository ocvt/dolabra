package handler

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/ocvt/dolabra/utils"
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

func dbCreateNullTrip() error {
	stmt := `
		INSERT OR REPLACE INTO trip
		VALUES (3000, '1990-01-01 00:00:00', true, false, false, 8000000, false, false,
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
		VALUES (8000000, ?, ?, ?, datetime('now'), '555-555-5555', 'Prefer not to say',
						1990, false, true, 'Require frequent oiling',
						datetime('now', '+100 years'), '', '', '', ?)`
	_, err = db.Exec(
		stmt,
		utils.GetConfig().SmtpFromEmailDefault,
		utils.GetConfig().SmtpFromFirstNameDefault,
		utils.GetConfig().SmtpFromLastNameDefault,
		notificationsStr) // sqlvet: ignore

	return err
}

func dbExtendMembership(w http.ResponseWriter, memberId int, years int) bool {
	isPaidMember, ok := dbIsPaidMember(w, memberId)
	if !ok {
		return false
	}

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

	if !isPaidMember && years > 0 {
		// Mark as paid on any upcoming trips
		stmt := `
			UPDATE trip_signup
			SET paid_member = true
			WHERE member_id = ? AND EXISTS(
				SELECT id
				FROM trip
				WHERE id = trip_signup.trip_id AND datetime('now') < datetime(start_datetime)
			)`
		_, err := db.Exec(stmt, memberId)
		if !checkError(w, err) {
			return false
		}

		// Bump from waitlist if applicable
		if !dbBumpMemberFromWaitlists(w, memberId) {
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

	isActive, ok := dbIsActiveMember(w, memberId)
	if !ok {
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

func dbGetMemberSubWithIdp(w http.ResponseWriter, idp string, idpHash string) (string, bool) {
	stmt := `
		SELECT sub
		FROM auth
		WHERE idp = ? AND idp_hash = ?`
	var sub string
	err := db.QueryRow(stmt, idp, idpHash).Scan(&sub)
	if !checkError(w, err) {
		return "", false
	}
	return sub, true
}

func dbGetSecurity(w http.ResponseWriter, memberId int) (int, bool) {
	stmt := `
		SELECT security
		FROM officer
		WHERE member_id = ?`
	var security int
	err := db.QueryRow(stmt, memberId).Scan(&security)
	if !checkError(w, err) {
		return 0, false
	}
	return security, true
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

/* "Ensure" helpers */
func dbEnsureMemberDoesNotExist(w http.ResponseWriter, sub string) bool {
	exists, ok := dbIsMemberWithSub(w, sub)
	if !ok {
		return false
	}

	if exists {
		respondError(w, http.StatusBadRequest, "Member is already registered.")
		return false
	}

	return true
}

func dbEnsureMemberExists(w http.ResponseWriter, sub string) bool {
	exists, ok := dbIsMemberWithSub(w, sub)
	if !ok {
		return false
	}

	if !exists {
		respondError(w, http.StatusNotFound, "Member is not registered.")
		return false
	}

	return true
}

func dbEnsureMemberIdExists(w http.ResponseWriter, memberId int) bool {
	exists, ok := dbIsMemberWithMemberId(w, memberId)
	if !ok {
		return false
	}

	if !exists {
		respondError(w, http.StatusNotFound, "Member is not registered.")
		return false
	}

	return true
}

func dbEnsureNotApprover(w http.ResponseWriter, memberId int) bool {
	isApprover, ok := dbIsApprover(w, memberId)
	if !ok {
		return false
	}
	if isApprover {
		respondError(w, http.StatusBadRequest, "Must not be approver.")
		return false
	}
	return true
}

func dbEnsureNotOfficer(w http.ResponseWriter, memberId int) bool {
	isOfficer, ok := dbIsOfficer(w, memberId)
	if !ok {
		return false
	}
	if isOfficer {
		respondError(w, http.StatusBadRequest, "Must not be officer.")
		return false
	}
	return true
}

func dbEnsureOfficer(w http.ResponseWriter, memberId int) bool {
	isOfficer, ok := dbIsOfficer(w, memberId)
	if !ok {
		return false
	}
	if !isOfficer {
		respondError(w, http.StatusBadRequest, "Must be officer.")
		return false
	}
	return true
}

func dbEnsureTripExists(w http.ResponseWriter, tripId int) bool {
	exists, ok := dbIsTrip(w, tripId)
	if !ok {
		return false
	}

	if !exists {
		respondError(w, http.StatusNotFound, "Trip does not exist.")
		return false
	}

	return true
}

/* EXISTS helpers */
func dbIsActiveMember(w http.ResponseWriter, memberId int) (bool, bool) {
	stmt := `
		SELECT EXISTS (
			SELECT 1
			FROM member
			WHERE id = ? AND active = true)`
	var exists bool
	err := db.QueryRow(stmt, memberId).Scan(&exists)
	return exists, checkError(w, err)
}

func dbIsApprover(w http.ResponseWriter, memberId int) (bool, bool) {
	stmt := `
		SELECT EXISTS (
			SELECT 1
			FROM trip_approver
			WHERE member_id = ?)`
	var exists bool
	err := db.QueryRow(stmt, memberId).Scan(&exists)
	return exists, checkError(w, err)
}

func dbIsMemberWithIdp(w http.ResponseWriter, idp string, idpHash string) (bool, bool) {
	stmt := `
		SELECT EXISTS (
			SELECT 1
			FROM auth
			WHERE idp = ? AND idp_hash = ?)`
	var exists bool
	err := db.QueryRow(stmt, idp, idpHash).Scan(&exists)
	return exists, checkError(w, err)
}

func dbIsMemberWithMemberId(w http.ResponseWriter, memberId int) (bool, bool) {
	stmt := `
		SELECT EXISTS (
			SELECT 1
			FROM auth
			WHERE member_id > 8000000 AND member_id = ?)`
	var exists bool
	err := db.QueryRow(stmt, memberId).Scan(&exists)
	return exists, checkError(w, err)
}

func dbIsMemberWithSub(w http.ResponseWriter, sub string) (bool, bool) {
	stmt := `
		SELECT EXISTS (
			SELECT 1
			FROM auth
			WHERE member_id > 8000000 AND sub = ?)`
	var exists bool
	err := db.QueryRow(stmt, sub).Scan(&exists)
	return exists, checkError(w, err)
}

func dbIsOfficer(w http.ResponseWriter, memberId int) (bool, bool) {
	stmt := `
		SELECT EXISTS (
			SELECT 1
			FROM officer
			WHERE member_id = ?)`
	var exists bool
	err := db.QueryRow(stmt, memberId).Scan(&exists)
	return exists, checkError(w, err)
}

func dbIsPaidMember(w http.ResponseWriter, memberId int) (bool, bool) {
	stmt := `
		SELECT EXISTS (
			SELECT 1
			FROM member
			WHERE id = ? AND datetime(paid_expire_datetime) > datetime('now'))`
	var exists bool
	err := db.QueryRow(stmt, memberId).Scan(&exists)
	return exists, checkError(w, err)
}
