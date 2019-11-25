package handler

import (
  "encoding/json"
  "net/http"
  "strconv"
  "strings"

  "github.com/go-chi/chi"
)

type tripStruct struct {
  Id int `json:"id,omitempty"`
  CreateDatetime string `json:"createDatetime,omitempty"`
  Cancel *bool `json:"cancel,omitempty"`
  Publish *bool `json:"publish,omitempty"`
  MemberName string `json:"memberName,omitempty"` // Used client side
  MemberId int `json:"memberId,omitempty"` // Used server side
  /* Required fields for creating a trip */
  MembersOnly bool `json:"membersOnly"`
  AllowLateSignups bool `json:"allowLateSignups"`
  DrivingRequired bool `json:"drivingRequired"`
  HasCost bool `json:"hasCost"`
  CostDescription string `json:"costDescription"`
  MaxPeople int `json:"maxPeople"`
  Name string `json:"name"`
  TripTypeId string `json:"tripTypeId"`
  StartDatetime string `json:"startDatetime"`
  EndDatetime string `json:"endDatetime"`
  Summary string `json:"summary"`
  Description string `json:"description"`
  Location string `json:"location"`
  LocationDirections string `json:"locationDirections"`
  MeetupLocation string `json:"MeetupLocation"`
  Distance string `json:"distance"`
  Difficulty int `json:"difficulty"`
  DifficultyDescription string `json:"difficultyDescription"`
  Instructions string `json:"instructions"`
  PetsAllowed bool `json:"petsAllowed"`
  PetsDescription string `json:"petsDescription"`
}

type tripSignupStruct struct {
  SignupDatetime string `json:"signupDatetime,omitempty"`
  /* Required fields for signing up for a trip */
  ShortNotice bool `json:"shortNotice"`
  Driver bool `json:"driver"`
  Carpool bool `json:"carpool"`
  CarCapacityTotal int `json:carCapacityTotal"`
  Notes string `json:"notes"`
  Pet bool `json:"pet"`
}

func GetTrips(w http.ResponseWriter, r *http.Request) {
  stmt := `
    SELECT *
    FROM trip
    WHERE cancel = false AND datetime(start_datetime) >= datetime('now')
    ORDER BY datetime(start_datetime) DESC`
  rows, err := db.Query(stmt)
  if !checkError(w, err) {
    return
  }
  defer rows.Close() // TODO needed?

  var trips = []*tripStruct{}
  tripIndex := 0
  for rows.Next() {
    trips = append(trips, &tripStruct{})
    err = rows.Scan(
      &trips[tripIndex].Id,
      &trips[tripIndex].CreateDatetime,
      &trips[tripIndex].Cancel,
      &trips[tripIndex].Publish,
      &trips[tripIndex].MemberId,
      &trips[tripIndex].MembersOnly,
      &trips[tripIndex].AllowLateSignups,
      &trips[tripIndex].DrivingRequired,
      &trips[tripIndex].HasCost,
      &trips[tripIndex].CostDescription,
      &trips[tripIndex].MaxPeople,
      &trips[tripIndex].Name,
      &trips[tripIndex].TripTypeId,
      &trips[tripIndex].StartDatetime,
      &trips[tripIndex].EndDatetime,
      &trips[tripIndex].Summary,
      &trips[tripIndex].Description,
      &trips[tripIndex].Location,
      &trips[tripIndex].LocationDirections,
      &trips[tripIndex].MeetupLocation,
      &trips[tripIndex].Distance,
      &trips[tripIndex].Difficulty,
      &trips[tripIndex].DifficultyDescription,
      &trips[tripIndex].Instructions,
      &trips[tripIndex].PetsAllowed,
      &trips[tripIndex].PetsDescription)
    if !checkError(w, err) {
      return
    }
    tripIndex++
  }

  err = rows.Err()
  if !checkError(w, err) {
    return
  }

  respondJSON(w, http.StatusOK, map[string][]*tripStruct{"trips": trips})
}

func GetTripsArchive(w http.ResponseWriter, r *http.Request) {
  path := chi.URLParam(r, "*")
  pathVars := strings.Split(path, "/")

  tripStartId := MAX_INT
  tripsPerPage := 20
  if len(pathVars) > 0 || len(pathVars) < 3 {
    i, _ := strconv.Atoi(pathVars[0])
    if i != 0 {
      tripStartId = i
    }
  }
  if len(pathVars) == 2 {
    i, _ := strconv.Atoi(pathVars[1])
    if i != 0 {
      tripsPerPage = i
    }
  }

  stmt := `
    SELECT *
    FROM trip
    WHERE id <= ?
    ORDER BY datetime(start_datetime) DESC
    LIMIT ?`
  rows, err := db.Query(stmt, tripStartId, tripsPerPage)
  if !checkError(w, err) {
    return
  }

  defer rows.Close() // TODO needed?

  var trips = []*tripStruct{}
  tripIndex := 0
  for rows.Next() {
    trips = append(trips, &tripStruct{})
    err = rows.Scan(
      &trips[tripIndex].Id,
      &trips[tripIndex].CreateDatetime,
      &trips[tripIndex].Cancel,
      &trips[tripIndex].Publish,
      &trips[tripIndex].MemberId,
      &trips[tripIndex].MembersOnly,
      &trips[tripIndex].AllowLateSignups,
      &trips[tripIndex].DrivingRequired,
      &trips[tripIndex].HasCost,
      &trips[tripIndex].CostDescription,
      &trips[tripIndex].MaxPeople,
      &trips[tripIndex].Name,
      &trips[tripIndex].TripTypeId,
      &trips[tripIndex].StartDatetime,
      &trips[tripIndex].EndDatetime,
      &trips[tripIndex].Summary,
      &trips[tripIndex].Description,
      &trips[tripIndex].Location,
      &trips[tripIndex].LocationDirections,
      &trips[tripIndex].MeetupLocation,
      &trips[tripIndex].Distance,
      &trips[tripIndex].Difficulty,
      &trips[tripIndex].DifficultyDescription,
      &trips[tripIndex].Instructions,
      &trips[tripIndex].PetsAllowed,
      &trips[tripIndex].PetsDescription)
    if !checkError(w, err) {
      return
    }
    tripIndex++
  }

  err = rows.Err()
  if !checkError(w, err) {
    return
  }

  respondJSON(w, http.StatusOK, map[string][]*tripStruct{"trips": trips})
}

func GetTripsTypes(w http.ResponseWriter, r *http.Request) {
  stmt := `
    SELECT *
    FROM trip_type`
  rows, err := db.Query(stmt)
  if !checkError(w, err) {
    return
  }
  defer rows.Close() // TODO needed

  var tripTypes = map[string]map[string]string{}
  for rows.Next() {
    var id, name, description string
    err = rows.Scan(&id, &name, &description)
    if !checkError(w, err) {
      return
    }

    tripTypes[id] = map[string]string{
      "typeName": name,
      "typeDescription": description,
    }
  }

  err = rows.Err()
  if !checkError(w, err) {
    return
  }

  respondJSON(w, http.StatusOK, tripTypes)
}

func PatchTripsCancel(w http.ResponseWriter, r *http.Request) {
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

  // Check if admin
  isCreator, err := dbIsTripCreator(w, tripId, memberId)
  if err != nil {
    return
  }
  isLeader, err := dbIsTripLeader(w, tripId, memberId)
  if err != nil {
    return
  }
  isOfficer, err := dbIsOfficer(w, memberId)
  if err != nil {
    return
  }

  if !isCreator && !isLeader && !isOfficer {
    respondError(w, http.StatusUnauthorized, "Not authorized to cancel trip.")
    return
  }

  stmt := `
    UPDATE trip
    SET cancel = true
    WHERE id = ?`
  _, err = db.Exec(stmt, tripId)
  if !checkError(w, err) {
    return
  }

  respondJSON(w, http.StatusNoContent, nil)
}

func PostTrips(w http.ResponseWriter, r *http.Request) {
  _, subject, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Get request body
  decoder := json.NewDecoder(r.Body)
  var trip tripStruct
  err := decoder.Decode(&trip)
  if err != nil {
    respondError(w, http.StatusBadRequest, err.Error())
    return
  }

  // Get member id
  memberId, ok := dbGetActiveMemberId(w, subject)
  if !ok {
    return
  }

  // Insert new trip
  stmt := `
    INSERT INTO trip (
      create_datetime,
      cancel,
      publish,
      member_id,
      members_only,
      allow_late_signups,
      driving_required,
      has_cost,
      cost_description,
      max_people,
      name,
      trip_type_id,
      start_datetime,
      end_datetime,
      summary,
      description,
      location,
      location_directions,
      meetup_location,
      distance,
      difficulty,
      difficulty_description,
      instructions,
      pets_allowed,
      pets_description)
    VALUES (datetime('now'), false, false, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
  result, err := db.Exec(
    stmt,
    memberId,
    trip.MembersOnly,
    trip.AllowLateSignups,
    trip.DrivingRequired,
    trip.HasCost,
    trip.CostDescription,
    trip.MaxPeople,
    trip.Name,
    trip.TripTypeId,
    trip.StartDatetime,
    trip.EndDatetime,
    trip.Summary,
    trip.Description,
    trip.Location,
    trip.LocationDirections,
    trip.MeetupLocation,
    trip.Distance,
    trip.Difficulty,
    trip.DifficultyDescription,
    trip.Instructions,
    trip.PetsAllowed,
    trip.PetsDescription)
  if !checkError(w, err) {
    return
  }

  // Get new trip id
  tripId, err := result.LastInsertId()
  if !checkError(w, err) {
    return
  }

  respondJSON(w, http.StatusCreated, map[string]int64{"tripId": tripId})
}
