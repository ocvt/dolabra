package handler

import (
  "encoding/json"
  "net/http"
  "strconv"

  "github.com/go-chi/chi"
)

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
  tripId, err := strconv.Atoi(chi.URLParam(r, "tripId"))
  if err != nil {
    respondError(w, http.StatusBadRequest, err.Error())
    return
  }

  // Get request body
  decoder := json.NewDecoder(r.Body)
  var tripSignup tripSignupStruct
  err = decoder.Decode(&tripSignup)
  if err != nil {
    respondError(w, http.StatusBadRequest, err.Error())
    return
  }

  // Get trip data
  var trip tripStruct
  stmt := `
    SELECT *
    FROM trip
    WHERE id = ?`
  err = db.QueryRow(stmt, tripId).Scan(
    &trip.Id,
    &trip.CreateDatetime,
    &trip.Cancel,
    &trip.Publish,
    &trip.MemberId,
    &trip.MembersOnly,
    &trip.AllowLateSignups,
    &trip.DrivingRequired,
    &trip.HasCost,
    &trip.CostDescription,
    &trip.MaxPeople,
    &trip.Name,
    &trip.TripTypeId,
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
    return
  }

  /* Mandatory checks for any trip signup or modification */
  exists, err := dbTripExists(w, tripId)
  if err != nil {
    return
  }
  if !exists {
    respondError(w, http.StatusBadRequest, "Trip does not exist.")
    return
  }

  isCanceled, err := dbIsTripCanceled(w, tripId)
  if err != nil {
    return
  }
  if isCanceled {
    respondError(w, http.StatusBadRequest, "Trip is canceled.")
    return
  }

  inPast, err := dbIsTripInPast(w, tripId)
  if err != nil {
    return
  }
  if inPast {
    respondError(w, http.StatusBadRequest, "Trip is in past.")
    return
  }

  /* Checks only applicable for new signups */
  onTrip, err := dbIsMemberOnTrip(w, tripId, memberId)
  if err != nil {
    return
  }
  if onTrip {
    respondError(w, http.StatusBadRequest, "Member is already on trip.")
    return
  }

  lateSignup, err := dbIsLateSignup(w, tripId)
  if err != nil {
    return
  }
  if lateSignup && !trip.AllowLateSignups {
    respondError(w, http.StatusBadRequest, "Past trip signup deadline.")
    return
  }

  if tripSignup.Carpool && !tripSignup.Driver {
    respondError(w, http.StatusBadRequest, "Cannot carpool without being a driver.")
    return
  }
  if tripSignup.CarCapacityTotal < 0 {
    respondError(w, http.StatusBadRequest, "Cannot have negative car capacity.")
    return
  }

  if tripSignup.Pet && !trip.PetsAllowed {
    respondError(w, http.StatusBadRequest, "Cannot bring pet on trip.")
    return
  }

  /* Booleans for insertion */
  isCreator := memberId == trip.MemberId

  isPaid, err := dbIsPaidMember(w, memberId)
  if err != nil {
    return
  }

  // Insert new trip person
  stmt = `
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
    VALUES (?, ?, datetime('now'), ?, ?, ?, "", ?, ?, ?, ?, ?, ?, false)`
  _, err = db.Exec(
    stmt,
    tripId,
    memberId,
    isCreator,
    isPaid,
    "ATTEN",
    tripSignup.ShortNotice,
    tripSignup.Driver,
    tripSignup.Carpool,
    tripSignup.CarCapacityTotal,
    tripSignup.Notes,
    tripSignup.Pet)
  if !checkError(w, err) {
    return
  }

  respondJSON(w, http.StatusCreated, map[string]int{"tripId": tripId})
}
