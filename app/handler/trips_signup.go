package handler

import (
  "encoding/json"
  "net/http"
)

func GetTripsSignup(w http.ResponseWriter, r *http.Request) {
  _, subject, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Get member id and trip id
  memberId, ok := dbGetActiveMemberId(w, subject)
  if !ok {
    return
  }
  tripId, ok := checkURLParam(w, r, "tripId")
  if !ok {
    return
  }

  // Ensure user is on trip
  signupExists, err := dbIsMemberOnTrip(w, tripId, memberId)
  if err != nil {
    return
  }
  if !signupExists {
    respondError(w, http.StatusBadRequest, "Not on on trip.")
    return
  }

  // Get signup info
  stmt := `
    SELECT *
    FROM trip_signup
    WHERE trip_id = ? AND member_id = ?`
  var tripSignup tripSignupStruct
  err = db.QueryRow(stmt, tripId, memberId).Scan(
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

  // Get member id
  memberId, ok := dbGetActiveMemberId(w, subject)
  if !ok {
    return
  }

  // Get trip id and validate
  tripId, ok := checkURLParam(w, r, "tripId")
  if !ok {
    return
  }
  signupId, ok := checkURLParam(w, r, "signupId")
  if !ok {
    return
  }

  isTrip, err := dbIsTrip(w, tripId)
  if err != nil {
    return
  }
  if !isTrip {
    return
  }

  if !dbEnsureMemberCanModifySignup(w, tripId, memberId, signupId) {
    return
  }
  if !dbEnsureOfficerOrTripLeader(w, tripId, memberId) {
    return
  }

  // Mark as absent
  stmt := `
    UPDATE trip_signup
    SET attended = false
    WHERE trip_id = ? and member_id = ?`
  _, err = db.Exec(stmt, tripId, signupId)
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

  // Get member id
  memberId, ok := dbGetActiveMemberId(w, subject)
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

  // Get trip id, signup id
  tripId, ok := checkURLParam(w, r, "tripId")
  if !ok {
    return
  }
  signupId, ok := checkURLParam(w, r, "signupId")
  if !ok {
    return
  }

  if !dbEnsureMemberCanModifySignup(w, tripId, memberId, signupId) {
    return
  }
  if !dbEnsureOfficerOrTripLeader(w, tripId, memberId) {
    return
  }

  boot, err := dbCheckTripSignupCode(w, tripId, signupId, "BOOT")
  if err != nil {
    return
  }
  if boot {
    respondError(w, http.StatusBadRequest, "User is already booted.")
    return
  }
  if !dbEnsureTripSignupNotCanceled(w, tripId, signupId) {
    return
  }

  // Change to BOOT code
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

  respondJSON(w, http.StatusNoContent, nil)
}

func PatchTripsSignupCancel(w http.ResponseWriter, r *http.Request) {
  _, subject, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Get member id
  memberId, ok := dbGetActiveMemberId(w, subject)
  if !ok {
    return
  }

  // Get trip id and validate
  tripId, ok := checkURLParam(w, r, "tripId")
  if !ok {
    return
  }

  if !dbEnsureMemberCanModifySignup(w, tripId, memberId, memberId) {
    return
  }

  // Ensure not already canceled
  cancel, err := dbCheckTripSignupCode(w, tripId, memberId, "CANCEL")
  if err != nil {
    return
  }
  if cancel {
    respondError(w, http.StatusBadRequest, "User is already canceled.")
    return
  }

  // Change to CANCEL code
  stmt := `
    UPDATE trip_signup
    SET
      leader = false,
      attending_code = 'CANCEL',
      attended = false
    WHERE trip_id = ? AND member_id = ?`
  _, err = db.Exec(stmt, tripId, memberId)
  if !checkError(w, err) {
    return
  }

  respondJSON(w, http.StatusNoContent, nil)
}

func PatchTripsSignupForceadd(w http.ResponseWriter, r *http.Request) {
  _, subject, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Get member id
  memberId, ok := dbGetActiveMemberId(w, subject)
  if !ok {
    return
  }

  // Get trip id, signup id
  tripId, ok := checkURLParam(w, r, "tripId")
  if !ok {
    return
  }
  signupId, ok := checkURLParam(w, r, "signupId")
  if !ok {
    return
  }

  if !dbEnsureMemberCanModifySignup(w, tripId, memberId, signupId) {
    return
  }
  if !dbEnsureOfficerOrTripLeader(w, tripId, memberId) {
    return
  }

  // Validate not already forceadded or canceled
  force, err := dbCheckTripSignupCode(w, tripId, signupId, "FORCE")
  if err != nil {
    return
  }
  if force {
    respondError(w, http.StatusBadRequest, "User is already force-added or canceled.")
    return
  }
  if !dbEnsureTripSignupNotCanceled(w, tripId, signupId) {
    return
  }

  // Change to FORCE code
  stmt := `
    UPDATE trip_signup
    SET attending_code = 'FORCE'
    WHERE trip_id = ? AND member_id = ?`
  _, err = db.Exec(stmt, tripId, signupId)
  if !checkError(w, err) {
    return
  }

  respondJSON(w, http.StatusNoContent, nil)
}

func PatchTripsSignupTripLeaderPromote(w http.ResponseWriter, r *http.Request) {
  _, subject, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Get member id
  memberId, ok := dbGetActiveMemberId(w, subject)
  if !ok {
    return
  }

  // Get trip id, signup id
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

  if !dbEnsureMemberCanModifySignup(w, tripId, memberId, signupId) {
    return
  }
  if !dbEnsureOfficerOrTripLeader(w, tripId, memberId) {
    return
  }

  // Validate not already canceled
  if !dbEnsureTripSignupNotCanceled(w, tripId, signupId) {
    return
  }

  // Change to promote user
  stmt := `
    UPDATE trip_signup
    SET
      leader = ?,
      attending_code = 'FORCE'
    WHERE trip_id = ? AND member_id = ?`
  _, err = db.Exec(stmt, promote, tripId, signupId)
  if !checkError(w, err) {
    return
  }

  respondJSON(w, http.StatusNoContent, nil)
}

func PostTripsSignup(w http.ResponseWriter, r *http.Request) {
  _, subject, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Get member id, trip id
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

  // Validate signup
  isCreator, err := dbIsTripCreator(w, tripId, memberId)
  if err != nil {
    return
  }

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
  }

  isPaid, err := dbIsPaidMember(w, memberId)
  if err != nil {
    return
  }

  // Insert signup
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
