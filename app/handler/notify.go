package handler

import (
	"encoding/json"
	"net/http"
)

func PostTripsNotifyGroup(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId, tripId, groupId
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}
	tripId, ok := getURLIntParam(w, r, "tripId")
	if !ok {
		return
	}

	// Get email fields
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var email emailStruct
	err := decoder.Decode(&email)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	email.ToId = 0
	email.TripId = tripId
	email.ReplyToId = memberId

	// Permissions
	if !dbEnsurePublishedTrip(w, tripId) ||
		!dbEnsureOfficerOrTripLeader(w, tripId, memberId) {
		return
	}

	if email.NotificationTypeId != "TRIP_MESSAGE_NOTIFY" &&
		email.NotificationTypeId != "TRIP_MESSAGE_ATTEND" &&
		email.NotificationTypeId != "TRIP_MESSAGE_WAIT" {
		respondError(w, http.StatusBadRequest, "Invalid notification type id")
		return
	}

	if !stageEmail(w, email) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

// Send message to specific signup
func PostTripsNotifySignup(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId, tripId, signupId
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
	decoder.DisallowUnknownFields()
	var email emailStruct
	err := decoder.Decode(&email)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	email.NotificationTypeId = "TRIP_ALERT"
	email.ReplyToId = memberId
	email.TripId = tripId

	// Permissions
	if !dbEnsurePublishedTrip(w, tripId) ||
		!dbEnsureOfficerOrTripLeader(w, tripId, memberId) {
		return
	}

	if !stageEmail(w, email) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func PostTripsReminder(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get member id, trip id
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

	if !dbEnsureTripLeader(w, tripId, memberId) {
		return
	}

	err := stageEmailTripReminder(tripId)
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}
