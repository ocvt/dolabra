package handler

import (
	"net/http"
)

type tripName struct {
	Id   int `json:"id"`
	Name int `json:"name"`
}

func GetWebtoolsMembers(w http.ResponseWriter, r *http.Request) {
	stmt := `
    SELECT
      member.id,
      member.email,
      member.first_name,
      member.last_name,
      member.create_datetime,
      member.cell_number,
      member.gender,
      member.birth_year,
      member.active,
      member.medical_cond,
      member.medical_cond_desc,
      member.paid_expire_datetime,
      emergency_contact.name,
      emergency_contact.number,
      emergency_contact.relationship
    FROM member
    INNER JOIN emergency_contact ON emergency_contact.member_id = member.id`
	rows, err := db.Query(stmt)
	if !checkError(w, err) {
		return
	}
	defer rows.Close()

	var members = []*memberStruct{}
	i := 0
	for rows.Next() {
		members = append(members, &memberStruct{})
		err = rows.Scan(
			&members[i].Id,
			&members[i].Email,
			&members[i].FirstName,
			&members[i].LastName,
			&members[i].CreateDatetime,
			&members[i].CellNumber,
			&members[i].Gender,
			&members[i].Birthyear,
			&members[i].Active,
			&members[i].MedicalCond,
			&members[i].MedicalCondDesc,
			&members[i].PaidExpireDatetime,
			&members[i].EmergencyContactName,
			&members[i].EmergencyContactNumber,
			&members[i].EmergencyContactRelationship)
		if !checkError(w, err) {
			return
		}
		i++
	}

	err = rows.Err()
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, map[string][]*memberStruct{"members": members})
}

func GetWebtoolsMembersTrips(w http.ResponseWriter, r *http.Request) {
	memberId, ok := checkURLParam(w, r, "memberId")
	if !ok {
		return
	}

	stmt := `
    SELECT
      id,
      name
    FROM trip
    WHERE trip.member_id = ?`
	rows, err := db.Query(stmt, memberId)
	if !checkError(w, err) {
		return
	}
	defer rows.Close()

	var trips = []*tripName{}
	i := 0
	for rows.Next() {
		trips = append(trips, &tripName{})
		err = rows.Scan(
			&trips[i].Id,
			&trips[i].Name)
		if !checkError(w, err) {
			return
		}
		i++
	}

	err = rows.Err()
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, map[string][]*tripName{"trips": trips})
}

func GetWebtoolsMembersAttendance(w http.ResponseWriter, r *http.Request) {
	memberId, ok := checkURLParam(w, r, "memberId")
	if !ok {
		return
	}

	stmt := `
    SELECT
      trip.id,
      trip.name
    FROM trip_signup
    INNER JOIN trip ON trip.id = trip_signup.trip_id
    WHERE trip_signup.id = ? AND trip_signup.attended = true`
	rows, err := db.Query(stmt, memberId)
	if !checkError(w, err) {
		return
	}
	defer rows.Close()

	var trips = []*tripName{}
	i := 0
	for rows.Next() {
		trips = append(trips, &tripName{})
		err = rows.Scan(
			&trips[i].Id,
			&trips[i].Name)
		if !checkError(w, err) {
			return
		}
		i++
	}

	err = rows.Err()
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, map[string][]*tripName{"trips": trips})
}

func PatchWebtoolsDuesGrant(w http.ResponseWriter, r *http.Request) {
	memberId, ok := checkURLParam(w, r, "memberId")
	if !ok {
		return
	}

	stmt := `
    UPDATE member
    SET paid_expire_datetime = datetime(paid_expire_datetime, '+1 year')
    WHERE id = ?`
	_, err := db.Exec(stmt, memberId)
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func PatchWebtoolsDuesRevoke(w http.ResponseWriter, r *http.Request) {
	memberId, ok := checkURLParam(w, r, "memberId")
	if !ok {
		return
	}

	stmt := `
    UPDATE member
    SET paid_expire_datetime = datetime('now')
    WHERE id = ?`
	_, err := db.Exec(stmt, memberId)
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}
