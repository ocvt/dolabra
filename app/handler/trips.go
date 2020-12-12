package handler

import (
	"encoding/json"
	"net/http"
)

type attendanceInfoStruct struct {
	Attend int `json:"attend"`
	Boot   int `json:"boot"`
	Cancel int `json:"cancel"`
	Force  int `json:"force"`
	Wait   int `json:"wait"`
}

/*
	Detailed trip info. Used for:
	- POST /trips
	- GET /trips/{tripId}
	Some information is redacted for:
	- GET /noauth/trips
	- GET /noauth/trips/{tripId}
*/
type tripStruct struct {
	/* Managed server side, used for GET /trips/{tripId} */
	Id             int    `json:"id,omitempty"`
	CreateDatetime string `json:"createDatetime,omitempty"`
	Cancel         bool   `json:"cancel,omitempty"`
	Publish        bool   `json:"publish,omitempty"`
	ReminderSent   bool   `json:"reminderSent,omitempty"`
	MemberId       int    `json:"memberId,omitempty"`
	MemberName     string `json:"memberName,omitempty"`
	/* Required fields for creating a trip, used for both methods */
	MembersOnly           bool    `json:"membersOnly"`
	AllowLateSignups      bool    `json:"allowLateSignups"`
	DrivingRequired       bool    `json:"drivingRequired"`
	HasCost               bool    `json:"hasCost"`
	CostDescription       string  `json:"costDescription"`
	MaxPeople             int     `json:"maxPeople"`
	Name                  string  `json:"name"`
	NotificationTypeId    string  `json:"notificationTypeId"`
	StartDatetime         string  `json:"startDatetime"`
	EndDatetime           string  `json:"endDatetime"`
	Summary               string  `json:"summary"`
	Description           string  `json:"description"`
	Location              string  `json:"location"`
	LocationDirections    string  `json:"locationDirections"`
	MeetupLocation        string  `json:"meetupLocation"`
	Distance              float32 `json:"distance"`
	Difficulty            int     `json:"difficulty"`
	DifficultyDescription string  `json:"difficultyDescription"`
	Instructions          string  `json:"instructions"`
	PetsAllowed           bool    `json:"petsAllowed"`
	PetsDescription       string  `json:"petsDescription"`
}

type tripSignupStruct struct {
	/* Managed server side */
	// from trip_signup table
	Id             int    `json:"id,omitempty"`
	TripId         int    `json:"tripId,omitempty"`
	MemberId       int    `json:"memberId,omitempty"`
	Leader         bool   `json:"leader,omitempty"`
	SignupDatetime string `json:"signupDatetime,omitempty"`
	PaidMember     bool   `json:"paidMember,omitempty"`
	AttendingCode  string `json:"attendingCode,omitempty"`
	BootReason     string `json:"bootReason,omitempty"`
	Attended       bool   `json:"attended,omitempty"`
	// from member table
	Email           string `json:"email,omitempty"`
	FirstName       string `json:"firstName,omitempty"`
	LastName        string `json:"lastName,omitempty"`
	CellNumber      string `json:"cellNumber,omitempty"`
	Pronouns        string `json:"pronouns,omitempty"`
	BirthYear       int    `json:"birthYear,omitempty"`
	MedicalCond     bool   `json:"medicalCond,omitempty"`
	MedicalCondDesc string `json:"medicalCondDesc,omitempty"`
	ECName          string `json:"ECName,omitempty"`
	ECNumber        string `json:"ECNumber,omitempty"`
	ECRelationship  string `json:"ECRelationship,omitempty"`
	/* Required fields for signing up for a trip */
	ShortNotice bool   `json:"shortNotice"`
	Driver      bool   `json:"driver"`
	Carpool     bool   `json:"carpool"`
	CarCapacity int    `json:"carCapacity"`
	Notes       string `json:"notes"`
	Pet         bool   `json:"pet"`
}

type tripSignupBootStruct struct {
	BootReason string `json:"bootReason"`
}

func GetMyAttendance(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get member id and trip id
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}

	stmt := `
		SELECT *
		FROM trip_signup
		INNER JOIN trip ON trip.id = trip_signup.trip_id
		WHERE trip_signup.member_id = ?`
	rows, err := db.Query(stmt, memberId)
	if !checkError(w, err) {
		return
	}
	defer rows.Close()

	var trips = []*tripStruct{}
	var tripSignups = []*tripSignupStruct{}
	i := 0
	for rows.Next() {
		trips = append(trips, &tripStruct{})
		tripSignups = append(tripSignups, &tripSignupStruct{})
		err = rows.Scan(
			&tripSignups[i].Id,
			&tripSignups[i].TripId,
			&tripSignups[i].MemberId,
			&tripSignups[i].Leader,
			&tripSignups[i].SignupDatetime,
			&tripSignups[i].PaidMember,
			&tripSignups[i].AttendingCode,
			&tripSignups[i].BootReason,
			&tripSignups[i].ShortNotice,
			&tripSignups[i].Driver,
			&tripSignups[i].Carpool,
			&tripSignups[i].CarCapacity,
			&tripSignups[i].Notes,
			&tripSignups[i].Pet,
			&tripSignups[i].Attended,
			&trips[i].Id,
			&trips[i].CreateDatetime,
			&trips[i].Cancel,
			&trips[i].Publish,
			&trips[i].ReminderSent,
			&trips[i].MemberId,
			&trips[i].MembersOnly,
			&trips[i].AllowLateSignups,
			&trips[i].DrivingRequired,
			&trips[i].HasCost,
			&trips[i].CostDescription,
			&trips[i].MaxPeople,
			&trips[i].Name,
			&trips[i].NotificationTypeId,
			&trips[i].StartDatetime,
			&trips[i].EndDatetime,
			&trips[i].Summary,
			&trips[i].Description,
			&trips[i].Location,
			&trips[i].LocationDirections,
			&trips[i].MeetupLocation,
			&trips[i].Distance,
			&trips[i].Difficulty,
			&trips[i].DifficultyDescription,
			&trips[i].Instructions,
			&trips[i].PetsAllowed,
			&trips[i].PetsDescription)
		if !checkError(w, err) {
			return
		}

		var ok bool
		trips[i].MemberName, ok = dbGetMemberName(w, trips[i].MemberId)
		if !ok {
			return
		}
		i++
	}

	err = rows.Err()
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"trips": trips, "tripSignups": tripSignups})
}

func GetTrip(w http.ResponseWriter, r *http.Request) {
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
	isOfficer, ok := dbIsOfficer(w, memberId)
	if !ok {
		return
	}
	onTrip, ok := dbIsMemberOnTrip(w, tripId, memberId)
	if !ok {
		return
	}

	if !isOfficer && !onTrip {
		respondError(w, http.StatusUnauthorized, "Must be officer or on trip.")
		return
	}

	trip, ok := dbGetTrip(w, tripId)
	if !ok {
		return
	}

	trip.MemberName, ok = dbGetMemberName(w, trip.MemberId)
	if !ok {
		return
	}

	respondJSON(w, http.StatusOK, trip)
}

func GetTripSummary(w http.ResponseWriter, r *http.Request) {
	tripId, ok := getURLIntParam(w, r, "tripId")
	if !ok {
		return
	}

	if !dbEnsureTripExists(w, tripId) || !dbEnsurePublishedTrip(w, tripId) {
		return
	}

	trip, ok := dbGetTrip(w, tripId)
	if !ok {
		return
	}

	// Hide location to public
	trip.MeetupLocation = ""

	trip.MemberName, ok = dbGetMemberName(w, trip.MemberId)
	if !ok {
		return
	}

	respondJSON(w, http.StatusOK, trip)
}

func GetTripsAdmin(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get member id and trip id
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}
	tripId, ok := getURLIntParam(w, r, "tripId")
	if !ok {
		return
	}

	// Permissions
	if !dbEnsureOfficerOrTripLeader(w, tripId, memberId) {
		return
	}

	stmt := `
		SELECT *
		FROM trip_signup
		WHERE trip_id = ?
		ORDER BY datetime(signup_datetime) ASC`
	rows, err := db.Query(stmt, tripId)
	if !checkError(w, err) {
		return
	}
	defer rows.Close()

	var tripSignups = []*tripSignupStruct{}
	i := 0
	for rows.Next() {
		tripSignups = append(tripSignups, &tripSignupStruct{})
		err = rows.Scan(
			&tripSignups[i].Id,
			&tripSignups[i].TripId,
			&tripSignups[i].MemberId,
			&tripSignups[i].Leader,
			&tripSignups[i].SignupDatetime,
			&tripSignups[i].PaidMember,
			&tripSignups[i].AttendingCode,
			&tripSignups[i].BootReason,
			&tripSignups[i].ShortNotice,
			&tripSignups[i].Driver,
			&tripSignups[i].Carpool,
			&tripSignups[i].CarCapacity,
			&tripSignups[i].Notes,
			&tripSignups[i].Pet,
			&tripSignups[i].Attended)
		if !checkError(w, err) {
			return
		}

		stmt = `
			SELECT
				email,
				first_name,
				last_name,
				cell_number,
				pronouns,
				birth_year,
				medical_cond,
				medical_cond_desc,
				ec_name,
				ec_number,
				ec_relationship
			FROM member
			WHERE id = ?`
		err := db.QueryRow(stmt, tripSignups[i].MemberId).Scan(
			&tripSignups[i].Email,
			&tripSignups[i].FirstName,
			&tripSignups[i].LastName,
			&tripSignups[i].CellNumber,
			&tripSignups[i].Pronouns,
			&tripSignups[i].BirthYear,
			&tripSignups[i].MedicalCond,
			&tripSignups[i].MedicalCondDesc,
			&tripSignups[i].ECName,
			&tripSignups[i].ECNumber,
			&tripSignups[i].ECRelationship)
		if !checkError(w, err) {
			return
		}

		i++
	}

	err = rows.Err()
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, map[string][]*tripSignupStruct{"tripSignups": tripSignups})
}

func GetTripsAdminAttendanceInfo(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get member id and trip id
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}
	tripId, ok := getURLIntParam(w, r, "tripId")
	if !ok {
		return
	}

	// Permissions
	if !dbEnsureOfficerOrTripLeader(w, tripId, memberId) {
		return
	}

	stmt := `
		SELECT attending_code
		FROM trip_signup
		WHERE trip_id = ?
		ORDER BY datetime(signup_datetime) ASC`
	rows, err := db.Query(stmt, tripId)
	if !checkError(w, err) {
		return
	}
	defer rows.Close()

	attendanceInfo := attendanceInfoStruct{}
	for rows.Next() {
		var attendingCode string
		err = rows.Scan(&attendingCode)
		if !checkError(w, err) {
			return
		}

		if attendingCode == "ATTEND" {
			attendanceInfo.Attend += 1
		} else if attendingCode == "BOOT" {
			attendanceInfo.Boot += 1
		} else if attendingCode == "CANCEL" {
			attendanceInfo.Cancel += 1
		} else if attendingCode == "FORCE" {
			attendanceInfo.Force += 1
		} else {
			// WAIT
			attendanceInfo.Wait += 1
		}
	}

	err = rows.Err()
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, map[string]attendanceInfoStruct{"attendanceInfo": attendanceInfo})
}

func GetTripsArchive(w http.ResponseWriter, r *http.Request) {
	tripStartId, ok := getURLIntParam(w, r, "startId")
	if !ok {
		return
	}
	tripsPerPage, ok := getURLIntParam(w, r, "perPage")
	if !ok {
		return
	}

	stmt := `
		SELECT id
		FROM trip
		WHERE id > 0 AND id <= ? AND publish = true
		ORDER BY datetime(end_datetime) DESC
		LIMIT ?`
	rows, err := db.Query(stmt, tripStartId, tripsPerPage)
	if !checkError(w, err) {
		return
	}
	defer rows.Close()

	var trips = []*tripStruct{}
	i := 0
	for rows.Next() {
		var tripId int
		trips = append(trips, &tripStruct{})
		err = rows.Scan(&tripId)
		if !checkError(w, err) {
			return
		}

		trip, ok := dbGetTrip(w, tripId)
		if !ok {
			return
		}
		trips[i] = trip

		trips[i].MemberName, ok = dbGetMemberName(w, trips[i].MemberId)
		if !ok {
			return
		}
		i++
	}

	err = rows.Err()
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, map[string][]*tripStruct{"trips": trips})
}

func GetTripsArchiveDefault(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, r.URL.RequestURI()+"/100000/20", http.StatusPermanentRedirect)
}

func GetTripsMyTrips(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get member id and trip id
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}

	stmt := `
		SELECT id
		FROM trip
		WHERE
			member_id = ?`
	rows, err := db.Query(stmt, memberId)
	if !checkError(w, err) {
		return
	}
	defer rows.Close()

	var trips = []*tripStruct{}
	i := 0
	for rows.Next() {
		var tripId int
		trips = append(trips, &tripStruct{})
		err = rows.Scan(&tripId)
		if !checkError(w, err) {
			return
		}

		trip, ok := dbGetTrip(w, tripId)
		if !ok {
			return
		}
		trips[i] = trip

		trips[i].MemberName, ok = dbGetMemberName(w, trips[i].MemberId)
		if !ok {
			return
		}
		i++
	}

	err = rows.Err()
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, map[string][]*tripStruct{"trips": trips})
}

func GetTripsSummary(w http.ResponseWriter, r *http.Request) {
	stmt := `
		SELECT id
		FROM trip
		WHERE
			cancel = false
			AND publish = true
			AND datetime(end_datetime) >= datetime('now')
		ORDER BY datetime(start_datetime) DESC`
	rows, err := db.Query(stmt)
	if !checkError(w, err) {
		return
	}
	defer rows.Close()

	var trips = []*tripStruct{}
	i := 0
	for rows.Next() {
		var tripId int
		trips = append(trips, &tripStruct{})
		err = rows.Scan(&tripId)
		if !checkError(w, err) {
			return
		}

		trip, ok := dbGetTrip(w, tripId)
		if !ok {
			return
		}
		trips[i] = trip

		// Hide location to public
		trips[i].MeetupLocation = ""

		trips[i].MemberName, ok = dbGetMemberName(w, trips[i].MemberId)
		if !ok {
			return
		}
		i++
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
		FROM notification_type`
	rows, err := db.Query(stmt)
	if !checkError(w, err) {
		return
	}
	defer rows.Close()

	var notificationTypes = map[string]map[string]string{}
	for rows.Next() {
		var id, name, description string
		err = rows.Scan(&id, &name, &description)
		if !checkError(w, err) {
			return
		}

		notificationTypes[id] = map[string]string{
			"id":          id,
			"name":        name,
			"description": description,
		}
	}

	err = rows.Err()
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, notificationTypes)
}

func PatchTripsCancel(w http.ResponseWriter, r *http.Request) {
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

	// Permissions
	if !dbEnsureOfficerOrTripLeader(w, tripId, memberId) ||
		!dbEnsureTripNotCanceled(w, tripId) {
		return
	}

	stmt := `
		UPDATE trip
		SET cancel = true
		WHERE id = ?`
	_, err := db.Exec(stmt, tripId)
	if !checkError(w, err) {
		return
	}

	// Notify signups
	var signups = []int{}
	if !dbGetTripSignupGroup(w, tripId, "ATTEND", &signups) {
		return
	}
	if !dbGetTripSignupGroup(w, tripId, "FORCE", &signups) {
		return
	}
	if !dbGetTripSignupGroup(w, tripId, "WAIT", &signups) {
		return
	}

	if !stageEmailTripCancel(w, tripId) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func PatchTripsPublish(w http.ResponseWriter, r *http.Request) {
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

	// Only creator can signup while not published
	if !dbEnsureTripLeader(w, tripId, memberId) {
		return
	}

	stmt := `
		UPDATE trip
		SET publish = true
		WHERE id = ?`
	_, err := db.Exec(stmt, tripId)
	if !checkError(w, err) {
		return
	}

	// Approve trip
	if !approveNewTrip(w, tripId) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func PostTrips(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get member id
	memberId, ok := dbGetActiveMemberId(w, sub)
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

	if trip.Difficulty < 0 || trip.Difficulty > 5 {
		respondError(w, http.StatusForbidden, "Trip difficulty must be between 0 and 5.")
		return
	}

	// Insert new trip
	stmt := `
		INSERT INTO trip (
			create_datetime,
			cancel,
			publish,
			reminder_sent,
			member_id,
			members_only,
			allow_late_signups,
			driving_required,
			has_cost,
			cost_description,
			max_people,
			name,
			notification_type_id,
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
		VALUES (datetime('now'), false, false, false, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
						?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
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
		trip.NotificationTypeId,
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
