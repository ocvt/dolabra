package handler

import (
  "encoding/json"
  "net/http"
  "strconv"
)

func GetTripsSignup(w http.ResponseWriter, r *http.Request) {
  _, subject, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Get memberId, tripId
  memberId, ok := dbGetActiveMemberId(w, subject)
  if !ok {
    return
  }
  tripId, ok := checkURLParam(w, r, "tripId")
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
    &tripSignup.CarCapacityTotal,
    &tripSignup.Notes,
    &tripSignup.Pet,
    &tripSignup.Attended)
  if !checkError(w, err) {
    return
  }

  respondJSON(w, http.StatusOK, tripSignup)
}

func PatchTripsSignupAbsent(w http.ResponseWriter, r *http.Request) {
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

  // Permissions
  if !dbEnsureIsTrip(w, tripId) ||
     !dbEnsureMemberIsOnTrip(w, tripId, signupId) ||
     !dbEnsureNotTripCreator(w, tripId, signupId) ||
     !dbEnsureOfficerOrTripLeader(w, tripId, memberId) {
     return
   }

  stmt := `
    UPDATE trip_signup
    SET attended = false
    WHERE trip_id = ? and member_id = ?`
  _, err := db.Exec(stmt, tripId, signupId)
  if !checkError(w, err) {
    return
  }

  respondJSON(w, http.StatusNoContent, nil)
}

func PatchTripsSignupBoot(w http.ResponseWriter, r *http.Request) {
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

  // Get request body
  decoder := json.NewDecoder(r.Body)
  var tripSignupBoot tripSignupBootStruct
  err := decoder.Decode(&tripSignupBoot)
  if err != nil {
    respondError(w, http.StatusBadRequest, err.Error())
    return
  }
  if tripSignupBoot.BootReason == "" {
    respondError(w, http.StatusBadRequest, "BOOT action must have reason.")
    return
  }

  // Permissions
  if !dbEnsurePublishedTrip(w, tripId) ||
     !dbEnsureMemberIsOnTrip(w, tripId, signupId) ||
     !dbEnsureNotTripCreator(w, tripId, signupId) ||
     !dbEnsureOfficerOrTripLeader(w, tripId, memberId) ||
     !dbEnsureNotSignupCode(w, tripId, signupId, "CANCEL") ||
     !dbEnsureNotSignupCode(w, tripId, signupId, "BOOT") {
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
  _, err = db.Exec(stmt, tripSignupBoot.BootReason, tripId, signupId)
  if !checkError(w, err) {
    return
  }

  // Notify signup
  signupIdStr := strconv.Itoa(signupId)
  emailSubject :=
      "You have been Booted from the trip \"%s\""
  emailBody :=
      "This email is a notification that you have been Booted from the trip " +
      "\"%s\" with the message " + tripSignupBoot.BootReason
  if !stageEmail(w, signupIdStr, tripId, memberId, emailSubject, emailBody) {
    return
  }

  respondJSON(w, http.StatusNoContent, nil)
}

func PatchTripsSignupCancel(w http.ResponseWriter, r *http.Request) {
  _, subject, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Get memberId, tripId
  memberId, ok := dbGetActiveMemberId(w, subject)
  if !ok {
    return
  }
  tripId, ok := checkURLParam(w, r, "tripId")
  if !ok {
    return
  }

  // Permissions
  if !dbEnsurePublishedTrip(w, tripId) ||
     !dbEnsureOfficerOrTripLeader(w, tripId, memberId) ||
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

  // Notify member
  memberIdStr := strconv.Itoa(memberId)
  emailSubject :=
      "You have canceled your signup for trip \"%s\""
  emailBody :=
      "This email is a notification that you have canceled your signup on " +
      "trip \"%s\". Note, you cannot signup again after you have canceled."
  if !stageEmail(w, memberIdStr, tripId, 0, emailSubject, emailBody) {
    return
  }

  respondJSON(w, http.StatusNoContent, nil)
}

func PatchTripsSignupForceadd(w http.ResponseWriter, r *http.Request) {
  _, subject, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Get memberId, tripid, signupId
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

  // Permissions
  if !dbEnsurePublishedTrip(w, tripId) ||
     !dbEnsureMemberIsOnTrip(w, tripId, signupId) ||
     !dbEnsureOfficerOrTripLeader(w, tripId, memberId) ||
     !dbEnsureNotTripCreator(w, tripId, signupId) ||
     !dbEnsureNotSignupCode(w, tripId, memberId, "CANCEL") ||
     !dbEnsureNotSignupCode(w, tripId, memberId, "FORCE") {
     return
  }

  // Change to FORCE code
  stmt := `
    UPDATE trip_signup
    SET attending_code = 'FORCE'
    WHERE trip_id = ? AND member_id = ?`
  _, err := db.Exec(stmt, tripId, signupId)
  if !checkError(w, err) {
    return
  }

  // Notify signup
  signupIdStr := strconv.Itoa(signupId)
  emailSubject :=
      "You have been Force Added to the trip \"%s\""
  emailBody :=
      "This email is a notification that you have been Force Added to the " +
      "trip \"%s\"."
  if !stageEmail(w, signupIdStr, tripId, memberId, emailSubject, emailBody) {
    return
  }

  respondJSON(w, http.StatusNoContent, nil)
}

func PatchTripsSignupTripLeaderPromote(w http.ResponseWriter, r *http.Request) {
  _, subject, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Get memberId, tripId, signupId, promote
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
  promote, ok := checkURLParam(w, r, "promote")
  if !ok {
    return
  }

  // Permissions
  if !dbEnsurePublishedTrip(w, tripId) ||
     !dbEnsureOfficerOrTripLeader(w, tripId, memberId) ||
     !dbEnsureMemberIsOnTrip(w, tripId, signupId) ||
     !dbEnsureNotTripCreator(w, tripId, signupId) ||
     !dbEnsureNotSignupCode(w, tripId, signupId, "CANCEL") ||
     !dbEnsureNotSignupCode(w, tripId, signupId, "BOOT") {
     return
  }

  stmt := `
    UPDATE trip_signup
    SET
      leader = ?,
      attending_code = 'FORCE'
    WHERE trip_id = ? AND member_id = ?`
  _, err := db.Exec(stmt, promote, tripId, signupId)
  if !checkError(w, err) {
    return
  }

  // Notify signup
  signupIdStr := strconv.Itoa(signupId)
  emailSubject :=
      "You have been promoted to Trip Leader for the trip \"%s\""
  emailBody :=
      "This email is a notification that you have been promoted to Trip " +
      "Leader for the trip \"%s\""
  if !stageEmail(w, signupIdStr, tripId, memberId, emailSubject, emailBody) {
    return
  }

  respondJSON(w, http.StatusNoContent, nil)
}

func PostTripsSignup(w http.ResponseWriter, r *http.Request) {
  _, subject, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Get memberId, tripId
  memberId, ok := dbGetActiveMemberId(w, subject)
  if !ok {
    return
  }
  tripId, ok := checkURLParam(w, r, "tripId")
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

  // Permissions if not creator
  attendingCode := "FORCE"
  attended := true
  if !isCreator {
    attendingCode = "ATTEND"

    if !dbEnsureActiveTrip(w, tripId) {
      return
    }

    if !dbEnsureValidSignup(w, tripId, memberId, tripSignup.Carpool,
        tripSignup.Driver, tripSignup.CarCapacityTotal, tripSignup.Pet) {
      return
    }

    // TODO check for waitlist
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
    tripSignup.CarCapacityTotal,
    tripSignup.Notes,
    tripSignup.Pet,
    attended)
  if !checkError(w, err) {
    return
  }

  respondJSON(w, http.StatusCreated, map[string]int{"tripId": tripId})
}
