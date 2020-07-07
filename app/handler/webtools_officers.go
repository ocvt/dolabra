package handler

import (
	"encoding/json"
	"net/http"
)

type officerStruct struct {
	/* Managed server side */
	// from member table
	CellNumber string `json:"cellNumber,omitempty"`
	Email      string `json:"email,omitempty"`
	FirstName  string `json:"firstName,omitempty"`
	LastName   string `json:"lastName,omitempty"`
	/* Required fields for creating a trip */
	MemberId       int    `json:"memberId"`
	ExpireDatetime string `json:"expireDatetime"`
	Position       string `json:"position"`
	Security       int    `json:"security"`
}

func DeleteWebtoolsOfficers(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId, officerId
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}
	officerId, ok := getURLIntParam(w, r, "memberId")
	if !ok {
		return
	}

	// Permissions
	// TODO Don't allow officers with less privileges to modify officers with more privileges
	if memberId == officerId {
		respondError(w, http.StatusForbidden, "Cannot remove yourself from officers.")
		return
	}

	stmt := `
	 DELETE FROM officer
	 WHERE member_id = ?`
	_, err := db.Exec(stmt, officerId)
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func GetWebtoolsOfficers(w http.ResponseWriter, r *http.Request) {
	// Permissions
	// TODO Don't allow officers with less privileges to modify officers with more privileges

	stmt := `
		SELECT
			member.id,
			member.email,
			member.first_name,
			member.last_name,
			member.cell_number,
			officer.expire_datetime,
			officer.position,
			officer.security
		FROM member
		INNER JOIN officer ON officer.member_id = member.id`
	rows, err := db.Query(stmt)
	if !checkError(w, err) {
		return
	}
	defer rows.Close()

	var officers = []*officerStruct{}
	i := 0
	for rows.Next() {
		officers = append(officers, &officerStruct{})
		err = rows.Scan(
			&officers[i].MemberId,
			&officers[i].Email,
			&officers[i].FirstName,
			&officers[i].LastName,
			&officers[i].CellNumber,
			&officers[i].ExpireDatetime,
			&officers[i].Position,
			&officers[i].Security)
		if !checkError(w, err) {
			return
		}
		i++
	}

	err = rows.Err()
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, map[string][]*officerStruct{"officers": officers})
}

func PostWebtoolsOfficers(w http.ResponseWriter, r *http.Request) {
	// Get request body
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var newOfficer officerStruct
	err := decoder.Decode(&newOfficer)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Permissions
	// TODO Don't allow officers with less privileges to modify officers with more privileges
	if !dbEnsureMemberIdExists(w, newOfficer.MemberId) ||
		!dbEnsureNotOfficer(w, newOfficer.MemberId) {
		return
	}

	stmt := `
		INSERT INTO officer (
			member_id,
			create_datetime,
			expire_datetime,
			position,
			security)
		VALUES (?, datetime('now'), datetime(?), ?, ?)`
	_, err = db.Exec(
		stmt,
		newOfficer.MemberId,
		newOfficer.ExpireDatetime,
		newOfficer.Position,
		newOfficer.Security)
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}
