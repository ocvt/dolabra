package handler

import (
  "encoding/json"
  "fmt"
  "net/http"

  "github.com/go-chi/chi"
)

func PostTripsNotifySignup(w http.ResponseWriter, r * http.Request) {
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

  // Send email
  if !sendEmailToMember(w, memberId, signupId, emailSubject, emailBody) {
    return
  }

  respondJSON(w, http.StatusNoContent, nil)
}

func PostTripsNotifyGroup(w http.ResponseWriter, r * http.Request) {
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
  emailBody := jsonBody["emailBody"]
  emailSubject := jsonBody["emailSubject"]

  // Permissions
  if !dbEnsurePublishedTrip(w, tripId) ||
     !dbEnsureOfficerOrTripLeader(w, tripId, memberId) {
     return
  }

  if groupId != "all" && groupId != "attending" && groupId != "waitlist" {
    respondError(w, http.StatusBadRequest, "Invalid group id")
    return
  }

  // Get emails
  var signups = []int{}
   if groupId == "attending" || groupId == "all" {
    if !dbGetTripSignupGroup(w, tripId, "ATTEND", &signups) {
      return
    }
    if !dbGetTripSignupGroup(w, tripId, "FORCE", &signups) {
      return
    }
  } else if groupId == "waitlist" || groupId == "all" {
    if !dbGetTripSignupGroup(w, tripId, "WAIT", &signups) {
      return
    }
  }

  for i := 0; i < len(signups); i++ {
    if signups[i] == memberId {
      continue
    }

    if !sendEmailToMember(w, memberId, signups[i], emailSubject, emailBody) {
      return
    }
  }

  emailBody = fmt.Sprintf("You are receiving this message because you sent it\n\n%s", emailBody)
  if !sendEmailToMember(w, 0, memberId, emailSubject, emailBody) {
    return
  }

  respondJSON(w, http.StatusNoContent, nil)
}
