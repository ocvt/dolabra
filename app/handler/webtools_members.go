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
			id,
			email,
			name,
			create_datetime,
			cell_number,
			pronouns,
			birth_year,
			active,
			medical_cond,
			medical_cond_desc,
			paid_expire_datetime,
			ec_name,
			ec_number,
			ec_relationship
		FROM member`
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
			&members[i].Name,
			&members[i].CreateDatetime,
			&members[i].CellNumber,
			&members[i].Pronouns,
			&members[i].Birthyear,
			&members[i].Active,
			&members[i].MedicalCond,
			&members[i].MedicalCondDesc,
			&members[i].PaidExpireDatetime,
			&members[i].ECName,
			&members[i].ECNumber,
			&members[i].ECRelationship)
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

func GetWebtoolsMembersAttendance(w http.ResponseWriter, r *http.Request) {
	memberId, ok := getURLIntParam(w, r, "memberId")
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

func GetWebtoolsMembersTrips(w http.ResponseWriter, r *http.Request) {
	memberId, ok := getURLIntParam(w, r, "memberId")
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

func PostWebtoolsDuesGrant(w http.ResponseWriter, r *http.Request) {
	memberId, ok := getURLIntParam(w, r, "memberId")
	if !ok {
		return
	}

	if !dbExtendMembership(w, memberId, 1) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func PostWebtoolsDuesRevoke(w http.ResponseWriter, r *http.Request) {
	memberId, ok := getURLIntParam(w, r, "memberId")
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
