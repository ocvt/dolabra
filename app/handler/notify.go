package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

func PostTripsNotifySignup(w http.ResponseWriter, r *http.Request) {
	_, subject, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId, tripId, signupId
	memberId, ok := dbGetActiveMemberId(w, subject)
	if !ok {
		return
	}
	tripId, ok := checkURLParam(w, r, "tripId")
	if !ok {
		return
	}
	signupId, ok := checkURLParam(w, r, "signupId")
	if !ok {
		return
	}

	// Get email fields
	decoder := json.NewDecoder(r.Body)
	var jsonBody map[string]string
	err := decoder.Decode(&jsonBody)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	emailBody := jsonBody["emailBody"]
	emailSubject := jsonBody["emailSubject"]

	// Permissions
	if !dbEnsurePublishedTrip(w, tripId) ||
		!dbEnsureOfficerOrTripLeader(w, tripId, memberId) {
		return
	}

	// Notify signup
	signupIdStr := strconv.Itoa(signupId)
	if !stageEmail(w, signupIdStr, tripId, memberId, emailSubject, emailBody) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func PostTripsNotifyGroup(w http.ResponseWriter, r *http.Request) {
	_, subject, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId, tripId, groupId
	memberId, ok := dbGetActiveMemberId(w, subject)
	if !ok {
		return
	}
	tripId, ok := checkURLParam(w, r, "tripId")
	if !ok {
		return
	}
	groupId := chi.URLParam(r, "groupId")

	// Get email fields
	decoder := json.NewDecoder(r.Body)
	var jsonBody map[string]string
	err := decoder.Decode(&jsonBody)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	emailSubject := jsonBody["emailSubject"]
	emailBody := jsonBody["emailBody"]

	// Permissions
	if !dbEnsurePublishedTrip(w, tripId) ||
		!dbEnsureOfficerOrTripLeader(w, tripId, memberId) {
		return
	}

	if groupId != "all" && groupId != "attending" && groupId != "waitlist" {
		respondError(w, http.StatusBadRequest, "Invalid group id")
		return
	}

	// Notify signups
	if groupId == "all" {
		if !stageEmail(w, "TRIP_ALERT_ALL", tripId, memberId, emailSubject,
			emailBody) {
			return
		}
	} else if groupId == "attending" {
		if !stageEmail(w, "TRIP_ALERT_ATTEND", tripId, memberId, emailSubject,
			emailBody) {
			return
		}
	} else {
		if !stageEmail(w, "TRIP_ALERT_WAIT", tripId, memberId, emailSubject,
			emailBody) {
			return
		}
	}

	memberIdStr := strconv.Itoa(memberId)
	emailSubject = "Notification of OCVT notification"
	emailBody = "You are receiving this because you sent a notification for a" +
		" trip you are the trip leader of."
	if !stageEmail(w, memberIdStr, tripId, memberId, emailSubject, emailBody) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}
