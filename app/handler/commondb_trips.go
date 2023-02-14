package handler

import (
	"container/list"
	"database/sql"
	"log"
	"net/http"
	"time"
)

/* Contains Trip (including signups) specific functions:
- General helpers
- "Ensure" helpers to guarantee a specific result
- EXISTS helpers
*/

/* Non-DB Helpers */
func prettyPrintDate(date string) string {
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		log.Fatal(err.Error())
	}
	t = t.Local()
	return t.Format("Monday, Jan 2, 2006")
}

/* General helpers */
func dbBumpMemberFromWaitlists(w http.ResponseWriter, memberId int) bool {
	// Get trips where member is on waitlist and eligible to be bumped up
	stmt := `
		SELECT trip_signup.trip_id
		FROM trip_signup
		INNER JOIN trip ON trip.id = trip_signup.trip_id
		WHERE trip_signup.member_id = ?
			AND trip_signup.attending_code = 'WAIT'
			AND datetime('now', '+ 1 day') < datetime(trip.start_datetime)`
	rows, err := db.Query(stmt, memberId)
	if err != nil && err == sql.ErrNoRows {
		return true
	}
	if !checkError(w, err) {
		return false
	}
	defer rows.Close()

	tripIds := list.New()
	for rows.Next() {
		var tripId int
		err = rows.Scan(&tripId)
		if !checkError(w, err) {
			return false
		}
		tripIds.PushBack(tripId)
	}
	err = rows.Err()
	if !checkError(w, err) {
		return false
	}

	// Check each upcoming trip to see if member can be bumped up
	for t := tripIds.Front(); t != nil; t = t.Next() {
		tripId := t.Value.(int)
		// Get most recent unpaid signup
		memberIdToChange, ok := dbGetRecentUnpaidSignup(w, tripId)
		if !ok {
			return false
		}

		// id 0 is internal systems account and guaranteed to not be signed up
		if memberIdToChange > 0 {
			if !dbSetSignupStatus(w, tripId, memberIdToChange, "WAIT") ||
				!dbSetSignupStatus(w, tripId, memberId, "ATTEND") {
				return false
			}
		}
	}

	return true
}

func dbGetNextWaitlist(w http.ResponseWriter, tripId int) (int, bool) {
	// id 0 is internal systems account and guaranteed to not be signed up
	memberId := 0

	// Get waitlisted signups then sort by paid -> not paid then take the first result
	stmt := `
		SELECT
			datetime('now') < datetime(member.paid_expire_datetime) AS paid,
			trip_signup.member_id
		FROM trip_signup
		INNER JOIN member ON member.id = trip_signup.member_id
		WHERE trip_signup.trip_id = ? AND trip_signup.attending_code = 'WAIT'
		ORDER BY paid DESC
		LIMIT 1`
	rows, err := db.Query(stmt, tripId)
	if err != nil && err == sql.ErrNoRows {
		return memberId, true
	}
	if !checkError(w, err) {
		return memberId, false
	}
	defer rows.Close()

	for rows.Next() {
		var isPaidOnlyUsedInSqlQuery bool
		err = rows.Scan(&isPaidOnlyUsedInSqlQuery, &memberId)
		if !checkError(w, err) {
			return memberId, false
		}
	}

	err = rows.Err()
	if !checkError(w, err) {
		return memberId, false
	}

	return memberId, true
}

func dbGetRecentUnpaidSignup(w http.ResponseWriter, tripId int) (int, bool) {
	// id 0 is internal systems account and guaranteed to not be signed up
	memberId := 0

	stmt := `
		SELECT trip_signup.member_id
		FROM trip_signup
		INNER JOIN member ON member.id = trip_signup.member_id
		WHERE trip_signup.trip_id = ?
			AND trip_signup.attending_code = 'ATTEND'
			AND datetime(member.paid_expire_datetime) < datetime('now')`
	rows, err := db.Query(stmt, tripId)
	if err != nil && err == sql.ErrNoRows {
		return memberId, true
	}
	if !checkError(w, err) {
		return memberId, false
	}
	defer rows.Close()

	// Last member id in rows is most recently signed up unpaid signup
	for rows.Next() {
		err = rows.Scan(&memberId)
		if !checkError(w, err) {
			return memberId, false
		}
	}

	err = rows.Err()
	if !checkError(w, err) {
		return memberId, false
	}

	return memberId, true
}

func dbGetSignupMemberId(w http.ResponseWriter, tripId int, signupId int) (int, bool) {
	stmt := `
		SELECT member_id
		FROM trip_signup
		WHERE
			id = ? AND trip_id = ?`
	var memberId int
	err := db.QueryRow(stmt, signupId, tripId).Scan(&memberId)
	if !checkError(w, err) {
		return 0, false
	}
	return memberId, true
}

func dbGetTrip(w http.ResponseWriter, tripId int) (*tripStruct, bool) {
	trip, err := dbGetTripPlain(tripId)
	if !checkError(w, err) {
		return nil, false
	}
	return trip, true
}

func dbGetTripPlain(tripId int) (*tripStruct, error) {
	stmt := `
		SELECT *
		FROM trip
		WHERE
			id = ?`
	var trip tripStruct
	err := db.QueryRow(stmt, tripId).Scan(
		&trip.Id,
		&trip.CreateDatetime,
		&trip.Cancel,
		&trip.Publish,
		&trip.ReminderSent,
		&trip.MemberId,
		&trip.MembersOnly,
		&trip.AllowLateSignups,
		&trip.DrivingRequired,
		&trip.HasCost,
		&trip.CostDescription,
		&trip.MaxPeople,
		&trip.Name,
		&trip.NotificationTypeId,
		&trip.StartDatetime,
		&trip.EndDatetime,
		&trip.Summary,
		&trip.Description,
		&trip.Location,
		&trip.LocationDirections,
		&trip.MeetupLocation,
		&trip.Distance,
		&trip.Difficulty,
		&trip.DifficultyDescription,
		&trip.Instructions,
		&trip.PetsAllowed,
		&trip.PetsDescription)

	return &trip, err
}

func dbGetTripName(w http.ResponseWriter, tripId int) (string, bool) {
	stmt := `
		SELECT name
		FROM trip
		WHERE trip.id = ?`
	var name string
	err := db.QueryRow(stmt, tripId).Scan(&name)
	if !checkError(w, err) {
		return "", false
	}

	return name, true
}

func dbGetTripSignupGroup(w http.ResponseWriter, tripId int, groupId string, signups *[]int) bool {
	stmt := `
		SELECT member_id
		FROM trip_signup
		WHERE trip_signup.trip_id = ? AND trip_signup.attending_code = ?`
	rows, err := db.Query(stmt, tripId, groupId)
	if !checkError(w, err) {
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var memberId int
		err = rows.Scan(&memberId)
		if !checkError(w, err) {
			return false
		}

		*signups = append(*signups, memberId)
	}

	err = rows.Err()
	return checkError(w, err)
}

func dbGetTripSignupStatus(w http.ResponseWriter, tripId int, memberId int) (string, bool) {
	stmt := `
		SELECT attending_code
		FROM trip_signup
		WHERE trip_id = ? AND member_id = ?`
	var code string
	err := db.QueryRow(stmt, tripId, memberId).Scan(&code)
	if !checkError(w, err) {
		return "", false
	}
	return code, true
}

func dbIsTripFull(w http.ResponseWriter, tripId int) (bool, bool) {
	stmt := `
		SELECT max_people
		FROM trip
		WHERE id = ?`
	var maxPeople int
	err := db.QueryRow(stmt, tripId).Scan(&maxPeople)
	if !checkError(w, err) {
		return false, false
	}

	stmt = `
		SELECT COUNT(*)
		FROM trip_signup
		WHERE trip_id = ? AND attending_code = 'ATTEND'`
	var count int
	err = db.QueryRow(stmt, tripId).Scan(&count)
	if !checkError(w, err) {
		return false, false
	}

	if count == maxPeople {
		return true, true
	}
	return false, true
}

func dbSetSignupStatus(w http.ResponseWriter, tripId int, memberId int, status string) bool {
	stmt := `
		UPDATE trip_signup
		SET attending_code = ?
		WHERE trip_id = ? and member_id = ?`
	_, err := db.Exec(stmt, status, tripId, memberId)
	if !checkError(w, err) {
		return false
	}

	tripName, ok := dbGetTripName(w, tripId)
	if !ok {
		return false
	}
	if status == "ATTEND" {
		return stageEmailSignupAttend(w, tripId, tripName, memberId)
	}
	if status == "BOOT" || status == "CANCEL" {
		respondError(w, http.StatusInternalServerError, "BOOT or CANCEL should not be able to be passed to this function.")
		return false
	}
	if status == "FORCE" {
		return stageEmailSignupForce(w, tripId, tripName, memberId)
	}
	// WAIT
	return stageEmailSignupWait(w, tripId, tripName, memberId)
}

/* "Ensure" helpers */
func dbEnsureActiveTrip(w http.ResponseWriter, tripId int) bool {
	if !dbEnsureIsTrip(w, tripId) {
		return false
	}

	if !dbEnsureTripNotCanceled(w, tripId) {
		return false
	}

	inPast, ok := dbIsTripInPast(w, tripId)
	if !ok {
		return false
	}
	if inPast {
		respondError(w, http.StatusBadRequest, "Trip has already occured.")
		return false
	}

	if !dbEnsurePublishedTrip(w, tripId) {
		return false
	}

	return true
}

func dbEnsureIsTrip(w http.ResponseWriter, tripId int) bool {
	exists, ok := dbIsTrip(w, tripId)
	if !ok {
		return false
	}
	if !exists {
		respondError(w, http.StatusBadRequest, "Trip does not exists.")
		return false
	}
	return true
}

func dbEnsureMemberIsOnTrip(w http.ResponseWriter, tripId int, memberId int) bool {
	exists, ok := dbIsMemberOnTrip(w, tripId, memberId)
	if !ok {
		return false
	}
	if !exists {
		respondError(w, http.StatusBadRequest, "Member is not on trip.")
		return false
	}
	return true
}

func dbEnsureMemberIsNotOnTrip(w http.ResponseWriter, tripId int, memberId int) bool {
	exists, ok := dbIsMemberOnTrip(w, tripId, memberId)
	if !ok {
		return false
	}
	if exists {
		respondError(w, http.StatusBadRequest, "Member is on trip (or has canceled or been booted).")
		return false
	}
	return true
}

func dbEnsureNotSignupCode(w http.ResponseWriter, tripId int, memberId int, code string) bool {
	exists, ok := dbCheckSignupCode(w, tripId, memberId, code)
	if !ok {
		return false
	}
	if exists {
		respondError(w, http.StatusBadRequest, "Signup status is "+code+".")
		return false
	}
	return true
}

func dbEnsureNotTripCreator(w http.ResponseWriter, tripId int, memberId int) bool {
	exists, ok := dbIsTripCreator(w, tripId, memberId)
	if !ok {
		return false
	}
	if exists {
		respondError(w, http.StatusForbidden, "Cannot modify trip creator status.")
		return false
	}
	return true
}

func dbEnsureOfficerOrTripLeader(w http.ResponseWriter, tripId int, memberId int) bool {
	isOfficer, ok := dbIsOfficer(w, memberId)
	if !ok {
		return false
	}
	isTripLeader, ok := dbIsTripLeader(w, tripId, memberId)
	if !ok {
		return false
	}
	if !isOfficer && !isTripLeader {
		respondError(w, http.StatusForbidden, "Must be officer or trip leader.")
		return false
	}
	return true
}

func dbEnsurePublishedTrip(w http.ResponseWriter, tripId int) bool {
	exists, ok := dbIsTripPublished(w, tripId)
	if !ok {
		return false
	}
	if !exists {
		respondError(w, http.StatusBadRequest, "Trip is not published.")
		return false
	}
	return true
}

func dbEnsureTripLeader(w http.ResponseWriter, tripId int, memberId int) bool {
	exists, ok := dbIsTripLeader(w, tripId, memberId)
	if !ok {
		return false
	}
	if !exists {
		respondError(w, http.StatusForbidden, "Not a trip leader.")
		return false
	}
	return true
}

func dbEnsureTripNotCanceled(w http.ResponseWriter, tripId int) bool {
	isCanceled, ok := dbIsTripCanceled(w, tripId)
	if !ok {
		return false
	}
	if isCanceled {
		respondError(w, http.StatusBadRequest, "Trip is canceled.")
		return false
	}
	return true
}

func dbEnsureUnpaidTrip(w http.ResponseWriter, tripId int) bool {
	isPaidOnly, ok := dbIsTripPaidOnly(w, tripId)
	if !ok {
		return false
	}
	if isPaidOnly {
		respondError(w, http.StatusForbidden, "Cannot join a paid only trip as an unpaid member.")
		return false
	}
	return true
}

func dbEnsureValidSignup(w http.ResponseWriter, tripId int, memberId int,
	carpool bool, driver bool, carCapacity int, pet bool) bool {
	tooLateSignup, ok := dbIsTooLateSignup(w, tripId)
	if !ok {
		return false
	}
	if tooLateSignup {
		respondError(w, http.StatusForbidden, "Cannot signup past trip deadline.")
		return false
	}

	if carpool && !driver {
		respondError(w, http.StatusForbidden, "Cannot carpool without being a driver.")
		return false
	}
	if carCapacity < 0 {
		respondError(w, http.StatusForbidden, "Cannot have negative car capacity.")
		return false
	}

	petAllowed, ok := dbIsPetAllowed(w, tripId)
	if !ok {
		return false
	}
	if pet && !petAllowed {
		respondError(w, http.StatusForbidden, "Cannot bring pet on trip.")
		return false
	}

	return true
}

/* EXISTS helpers */
func dbCheckSignupCode(w http.ResponseWriter, tripId int, memberId int, code string) (bool, bool) {
	stmt := `
		SELECT EXISTS (
			SELECT 1
			FROM trip_signup
			WHERE trip_id = ? AND member_id = ? AND attending_code = ?)`
	var exists bool
	err := db.QueryRow(stmt, tripId, memberId, code).Scan(&exists)
	return exists, checkError(w, err)
}

func dbIsMemberOnTrip(w http.ResponseWriter, tripId int, memberId int) (bool, bool) {
	stmt := `
		SELECT EXISTS (
			SELECT 1
			FROM trip_signup
			WHERE (trip_id = ? AND member_id = ?))`
	var exists bool
	err := db.QueryRow(stmt, tripId, memberId).Scan(&exists)
	return exists, checkError(w, err)
}

func dbIsPetAllowed(w http.ResponseWriter, tripId int) (bool, bool) {
	stmt := `
		SELECT EXISTS (
			SELECT 1
			FROM trip
			WHERE id = ? AND pets_allowed = true)`
	var exists bool
	err := db.QueryRow(stmt, tripId).Scan(&exists)
	return exists, checkError(w, err)
}

func dbIsTooLateSignup(w http.ResponseWriter, tripId int) (bool, bool) {
	stmt := `
		SELECT EXISTS (
			SELECT 1
			FROM trip
			WHERE id = ? AND datetime(start_datetime) < datetime('now', '+12 hours') AND allow_late_signups = false)`
	var exists bool
	err := db.QueryRow(stmt, tripId).Scan(&exists)
	return exists, checkError(w, err)
}

func dbIsTrip(w http.ResponseWriter, tripId int) (bool, bool) {
	stmt := `
		SELECT EXISTS (
			SELECT 1
			FROM trip
			WHERE id = ?)`
	var exists bool
	err := db.QueryRow(stmt, tripId).Scan(&exists)
	return exists, checkError(w, err)
}

func dbIsTripCanceled(w http.ResponseWriter, tripId int) (bool, bool) {
	stmt := `
		SELECT EXISTS (
			SELECT 1
			FROM trip
			WHERE id = ? AND cancel = true)`
	var exists bool
	err := db.QueryRow(stmt, tripId).Scan(&exists)
	return exists, checkError(w, err)
}

func dbIsTripCreator(w http.ResponseWriter, tripId int, memberId int) (bool, bool) {
	stmt := `
		SELECT EXISTS (
			SELECT 1
			FROM trip
			WHERE id = ? AND member_id = ?)`
	var exists bool
	err := db.QueryRow(stmt, tripId, memberId).Scan(&exists)
	return exists, checkError(w, err)
}

func dbIsTripInPast(w http.ResponseWriter, tripId int) (bool, bool) {
	stmt := `
		SELECT EXISTS (
			SELECT 1
			FROM trip
			WHERE id = ? AND datetime(end_datetime) < datetime('now'))`
	var exists bool
	err := db.QueryRow(stmt, tripId).Scan(&exists)
	return exists, checkError(w, err)
}

func dbIsTripLeader(w http.ResponseWriter, tripId int, memberId int) (bool, bool) {
	stmt := `
		SELECT EXISTS (
			SELECT 1
			FROM trip_signup
			WHERE (trip_id = ? AND member_id = ? AND leader = true))`
	var exists bool
	err := db.QueryRow(stmt, tripId, memberId).Scan(&exists)
	return exists, checkError(w, err)
}

func dbIsTripPaidOnly(w http.ResponseWriter, tripId int) (bool, bool) {
	stmt := `
		SELECT EXISTS (
			SELECT 1
			FROM trip
			WHERE id = ? AND members_only = true)`
	var exists bool
	err := db.QueryRow(stmt, tripId).Scan(&exists)
	return exists, checkError(w, err)
}

func dbIsTripPublished(w http.ResponseWriter, tripId int) (bool, bool) {
	stmt := `
		SELECT EXISTS (
			SELECT 1
			FROM trip
			WHERE id = ? AND publish = true)`
	var exists bool
	err := db.QueryRow(stmt, tripId).Scan(&exists)
	return exists, checkError(w, err)
}
