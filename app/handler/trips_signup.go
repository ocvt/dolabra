package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

func GetTripMyStatus(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId, tripId
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}
	tripId, ok := getURLIntParam(w, r, "tripId")
	if !ok {
		return
	}

	if !dbEnsureTripExists(w, tripId) {
		return
	}

	// Statuses
	signedUp, err := dbIsMemberOnTrip(w, tripId, memberId)
	if err != nil {
		return
	}
	tripCreator, err := dbIsTripCreator(w, tripId, memberId)
	if err != nil {
		return
	}
	tripLeader, err := dbIsTripLeader(w, tripId, memberId)
	if err != nil {
		return
	}
	attendingCode := ""
	if signedUp {
		attendingCode, err = dbGetTripSignupStatus(w, tripId, memberId)
		if err != nil {
			return
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"attendingCode": attendingCode, "signedUp": signedUp, "tripCreator": tripCreator, "tripLeader": tripLeader})
}

func GetTripsSignup(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId, tripId
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}
	tripId, ok := getURLIntParam(w, r, "tripId")
	if !ok {
		return
	}

	// Permissions
	if !dbEnsureMemberIsOnTrip(w, tripId, memberId) {
		return
	}

	stmt := `
		SELECT *
		FROM trip_signup
		WHERE trip_id = ? AND member_id = ?`
	var tripSignup tripSignupStruct
	err := db.QueryRow(stmt, tripId, memberId).Scan(
		&tripSignup.Id,
		&tripSignup.TripId,
		&tripSignup.MemberId,
		&tripSignup.Leader,
		&tripSignup.SignupDatetime,
		&tripSignup.PaidMember,
		&tripSignup.AttendingCode,
		&tripSignup.BootReason,
		&tripSignup.ShortNotice,
		&tripSignup.Driver,
		&tripSignup.Carpool,
		&tripSignup.CarCapacity,
		&tripSignup.Notes,
		&tripSignup.Pet,
		&tripSignup.Attended)
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, tripSignup)
}

func PatchTripsSignupAbsent(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId, tripId, signupMemberId
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}
	tripId, ok := getURLIntParam(w, r, "tripId")
	if !ok {
		return
	}
	signupMemberId, ok := getURLIntParam(w, r, "memberId")
	if !ok {
		return
	}

	// Permissions
	if !dbEnsureIsTrip(w, tripId) ||
		!dbEnsureMemberIsOnTrip(w, tripId, signupMemberId) ||
		!dbEnsureOfficerOrTripLeader(w, tripId, memberId) {
		return
	}

	stmt := `
		UPDATE trip_signup
		SET attended = false
		WHERE trip_id = ? and member_id = ?`
	_, err := db.Exec(stmt, tripId, signupMemberId)
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func PatchTripsSignupBoot(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId, tripId, signupMemberId
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}
	tripId, ok := getURLIntParam(w, r, "tripId")
	if !ok {
		return
	}
	signupMemberId, ok := getURLIntParam(w, r, "memberId")
	if !ok {
		return
	}

	// Get request body
	decoder := json.NewDecoder(r.Body)
	var tripSignupBoot tripSignupBootStruct
	err := decoder.Decode(&tripSignupBoot)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if tripSignupBoot.BootReason == "" {
		respondError(w, http.StatusForbidden, "BOOT action must have reason.")
		return
	}

	// Permissions
	if !dbEnsurePublishedTrip(w, tripId) ||
		!dbEnsureMemberIsOnTrip(w, tripId, signupMemberId) ||
		!dbEnsureNotTripCreator(w, tripId, signupMemberId) ||
		!dbEnsureOfficerOrTripLeader(w, tripId, memberId) ||
		!dbEnsureNotSignupCode(w, tripId, signupMemberId, "CANCEL") {
		return
	}

	stmt := `
		UPDATE trip_signup
		SET
			leader = false,
			attending_code = 'BOOT',
			boot_reason = ?,
			attended = false
		WHERE trip_id = ? AND member_id = ?`
	_, err = db.Exec(stmt, tripSignupBoot.BootReason, tripId, signupMemberId)
	if !checkError(w, err) {
		return
	}

	tripName, ok := dbGetTripName(w, tripId)
	if !ok {
		return
	}

	email := emailStruct{
		NotificationTypeId: "TRIP_ALERT_BOOT",
		ReplyToId:          memberId,
		ToId:               signupMemberId,
		TripId:             tripId,
		Subject:            "You have been Booted from the trip " + tripName,
	}
	email.Body =
		"This email is a notification that you have been Booted from the trip " +
			tripName + " with the message " + tripSignupBoot.BootReason
	if !stageEmail(w, email) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func PatchTripsSignupCancel(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId, tripId
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}
	tripId, ok := getURLIntParam(w, r, "tripId")
	if !ok {
		return
	}

	// Permissions
	if !dbEnsurePublishedTrip(w, tripId) ||
		!dbEnsureMemberIsOnTrip(w, tripId, memberId) ||
		!dbEnsureNotTripCreator(w, tripId, memberId) ||
		!dbEnsureNotSignupCode(w, tripId, memberId, "CANCEL") ||
		!dbEnsureNotSignupCode(w, tripId, memberId, "BOOT") {
		return
	}

	stmt := `
		UPDATE trip_signup
		SET
			leader = false,
			attending_code = 'CANCEL',
			attended = false
		WHERE trip_id = ? AND member_id = ?`
	_, err := db.Exec(stmt, tripId, memberId)
	if !checkError(w, err) {
		return
	}

	tripName, ok := dbGetTripName(w, tripId)
	if !ok {
		return
	}

	email := emailStruct{
		NotificationTypeId: "TRIP_ALERT_CANCEL",
		ReplyToId:          0,
		ToId:               memberId,
		TripId:             tripId,
		Subject:            "You have canceled your signup for trip " + tripName + ".",
	}
	email.Body =
		"This email is a notification that you have canceled your signup on " +
			"trip " + tripName + ". Note, you cannot signup again after you have " +
			"canceled."

	if !stageEmail(w, email) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func PatchTripsSignupForceadd(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId, tripid, signupMemberId
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}
	tripId, ok := getURLIntParam(w, r, "tripId")
	if !ok {
		return
	}
	signupMemberId, ok := getURLIntParam(w, r, "memberId")
	if !ok {
		return
	}

	// Permissions
	if !dbEnsurePublishedTrip(w, tripId) ||
		!dbEnsureMemberIsOnTrip(w, tripId, signupMemberId) ||
		!dbEnsureOfficerOrTripLeader(w, tripId, memberId) {
		return
	}

	// Change to FORCE code
	stmt := `
		UPDATE trip_signup
		SET attending_code = 'FORCE'
		WHERE trip_id = ? AND member_id = ?`
	_, err := db.Exec(stmt, tripId, signupMemberId)
	if !checkError(w, err) {
		return
	}

	tripName, ok := dbGetTripName(w, tripId)
	if !ok {
		return
	}

	email := emailStruct{
		NotificationTypeId: "TRIP_ALERT_FORCE",
		ReplyToId:          memberId,
		ToId:               signupMemberId,
		TripId:             tripId,
		Subject:            "You have been Force Added to the trip " + tripName,
	}
	email.Body =
		"This email is a notification that you have been Force Added to the " +
			"trip " + tripName + "."
	if !stageEmail(w, email) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func PatchTripsSignupTripLeaderPromote(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId, tripId, signupMemberId, promote
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}
	tripId, ok := getURLIntParam(w, r, "tripId")
	if !ok {
		return
	}
	signupMemberId, ok := getURLIntParam(w, r, "memberId")
	if !ok {
		return
	}
	promote, err := strconv.ParseBool(chi.URLParam(r, "promote"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
	}

	// Permissions
	if !dbEnsurePublishedTrip(w, tripId) ||
		!dbEnsureOfficerOrTripLeader(w, tripId, memberId) ||
		!dbEnsureMemberIsOnTrip(w, tripId, signupMemberId) ||
		!dbEnsureNotTripCreator(w, tripId, signupMemberId) ||
		!dbEnsureNotSignupCode(w, tripId, signupMemberId, "CANCEL") ||
		!dbEnsureNotSignupCode(w, tripId, signupMemberId, "BOOT") {
		return
	}

	stmt := `
		UPDATE trip_signup
		SET
			leader = ?,
			attending_code = 'FORCE'
		WHERE trip_id = ? AND member_id = ?`
	_, err = db.Exec(stmt, promote, tripId, signupMemberId)
	if !checkError(w, err) {
		return
	}

	// Notify signup
	tripName, ok := dbGetTripName(w, tripId)
	if !ok {
		return
	}
	email := emailStruct{
		NotificationTypeId: "TRIP_ALERT_LEADER",
		ReplyToId:          memberId,
		ToId:               signupMemberId,
		TripId:             tripId,
	}
	email.Subject = "You have been promoted to Trip Leader for the trip " + tripName
	email.Body =
		"This email is a notification that you have been promoted to Trip " +
			"Leader for the trip " + tripName + "."
	if !stageEmail(w, email) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func PostTripsSignup(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId, tripId
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}
	tripId, ok := getURLIntParam(w, r, "tripId")
	if !ok {
		return
	}

	// Get request body
	decoder := json.NewDecoder(r.Body)
	var tripSignup tripSignupStruct
	err := decoder.Decode(&tripSignup)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Booleans for insertion
	isCreator, err := dbIsTripCreator(w, tripId, memberId)
	if err != nil {
		return
	}
	isPaid, err := dbIsPaidMember(w, memberId)
	if err != nil {
		return
	}

	// Permissions
	if !dbEnsureMemberIsNotOnTrip(w, tripId, memberId) {
		return
	}

	// Permissions if not creator
	attendingCode := "FORCE"
	attended := true
	if !isCreator {
		attendingCode = "ATTEND"

		if !dbEnsureActiveTrip(w, tripId) {
			return
		}

		if !dbEnsureValidSignup(w, tripId, memberId, tripSignup.Carpool,
			tripSignup.Driver, tripSignup.CarCapacity, tripSignup.Pet) {
			return
		}

		// TODO check for waitlist, member_only, max people
	}

	stmt := `
		INSERT INTO trip_signup (
			trip_id,
			member_id,
			leader,
			signup_datetime,
			paid_member,
			attending_code,
			boot_reason,
			short_notice,
			driver,
			carpool,
			car_capacity_total,
			notes,
			pet,
			attended)
		VALUES (?, ?, ?, datetime('now'), ?, ?, '', ?, ?, ?, ?, ?, ?, ?)`
	_, err = db.Exec(
		stmt,
		tripId,
		memberId,
		isCreator,
		isPaid,
		attendingCode,
		tripSignup.ShortNotice,
		tripSignup.Driver,
		tripSignup.Carpool,
		tripSignup.CarCapacity,
		tripSignup.Notes,
		tripSignup.Pet,
		attended)
	if !checkError(w, err) {
		return
	}

	tripName, ok := dbGetTripName(w, tripId)
	if !ok {
		return
	}

	email := emailStruct{
		NotificationTypeId: "TRIP_ALERT_ATTEND",
		ReplyToId:          0,
		ToId:               memberId,
		TripId:             tripId,
		Subject:            "You have been added to the trip " + tripName,
	}
	email.Body =
		"This email is a notification that you have been added to the roster " +
			"for the trip " + tripName + "."
	if !stageEmail(w, email) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}
