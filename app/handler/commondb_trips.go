package handler

import (
	"net/http"
)

/* Contains Trip (including signups) specific functions:
- General helpers
- "Ensure" helpers to guarantee a specific result
- EXISTS helpers
*/

/* General helpers */
func dbEnsureActiveTrip(w http.ResponseWriter, tripId int) bool {
	if !dbEnsureIsTrip(w, tripId) {
		return false
	}

	if !dbEnsureTripNotCanceled(w, tripId) {
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

	if !dbEnsurePublishedTrip(w, tripId) {
		return false
	}

	return true
}

func dbEnsureIsTrip(w http.ResponseWriter, tripId int) bool {
	exists, err := dbIsTrip(w, tripId)
	if err != nil {
		return false
	}
	if !exists {
		respondError(w, http.StatusBadRequest, "Trip does not exists.")
		return false
	}
	return true
}

func dbEnsureMemberIsOnTrip(w http.ResponseWriter, tripId int, memberId int) bool {
	exists, err := dbIsMemberOnTrip(w, tripId, memberId)
	if err != nil {
		return false
	}
	if !exists {
		respondError(w, http.StatusBadRequest, "Member is not on trip.")
		return false
	}
	return true
}

func dbEnsureMemberIsNotOnTrip(w http.ResponseWriter, tripId int, memberId int) bool {
	exists, err := dbIsMemberOnTrip(w, tripId, memberId)
	if err != nil {
		return false
	}
	if exists {
		respondError(w, http.StatusBadRequest, "Member is on trip (or has canceled or been booted).")
		return false
	}
	return true
}

func dbEnsureNotSignupCode(w http.ResponseWriter, tripId int, memberId int, code string) bool {
	exists, err := dbCheckSignupCode(w, tripId, memberId, code)
	if err != nil {
		return false
	}
	if exists {
		respondError(w, http.StatusBadRequest, "Signup status is "+code+".")
		return false
	}
	return true
}

func dbEnsureNotTripCreator(w http.ResponseWriter, tripId int, memberId int) bool {
	exists, err := dbIsTripCreator(w, tripId, memberId)
	if err != nil {
		return false
	}
	if exists {
		respondError(w, http.StatusForbidden, "Cannot modify trip creator status.")
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
		respondError(w, http.StatusForbidden, "Must be officer or trip leader.")
		return false
	}
	return true
}

func dbEnsurePublishedTrip(w http.ResponseWriter, tripId int) bool {
	exists, err := dbIsTripPublished(w, tripId)
	if err != nil {
		return false
	}
	if !exists {
		respondError(w, http.StatusBadRequest, "Trip is not published.")
		return false
	}
	return true
}

func dbEnsureTripLeader(w http.ResponseWriter, tripId int, memberId int) bool {
	exists, err := dbIsTripLeader(w, tripId, memberId)
	if err != nil {
		return false
	}
	if !exists {
		respondError(w, http.StatusForbidden, "Not a trip leader.")
		return false
	}
	return true
}

func dbEnsureTripNotCanceled(w http.ResponseWriter, tripId int) bool {
	isCanceled, err := dbIsTripCanceled(w, tripId)
	if err != nil {
		return false
	}
	if isCanceled {
		respondError(w, http.StatusBadRequest, "Trip is canceled.")
		return false
	}
	return true
}

func dbEnsureValidSignup(w http.ResponseWriter, tripId int, memberId int,
	carpool bool, driver bool, carCapacity int, pet bool) bool {
	tooLateSignup, err := dbIsTooLateSignup(w, tripId)
	if err != nil {
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

	petAllowed, err := dbIsPetAllowed(w, tripId)
	if err != nil {
		return false
	}
	if pet && !petAllowed {
		respondError(w, http.StatusForbidden, "Cannot bring pet on trip.")
		return false
	}

	return true
}

func dbGetTrip(w http.ResponseWriter, tripId int) (*tripStruct, bool) {
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
	if !checkError(w, err) {
		return nil, false
	}

	return &trip, true
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
	if !checkError(w, err) {
		return false
	}

	return true
}

func dbGetTripSignupStatus(w http.ResponseWriter, tripId int, memberId int) (string, error) {
	stmt := `
		SELECT attending_code
		FROM trip_signup
		WHERE trip_id = ? AND member_id = ?`
	var code string
	err := db.QueryRow(stmt, tripId, memberId).Scan(&code)
	if !checkError(w, err) {
		return "", err
	}
	return code, nil
}

/* EXISTS helpers */
func dbCheckSignupCode(w http.ResponseWriter, tripId int, memberId int, code string) (bool, error) {
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

func dbIsTooLateSignup(w http.ResponseWriter, tripId int) (bool, error) {
	stmt := `
		SELECT EXISTS (
			SELECT 1
			FROM trip
			WHERE id = ? AND datetime(start_datetime) < datetime('now', '+12 hours') AND allow_late_signups = true)`
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
			WHERE id = ? AND datetime(end_datetime) < datetime('now'))`
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
