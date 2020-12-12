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
	Name       string `json:"name,omitempty"`
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
	if memberId == officerId {
		respondError(w, http.StatusForbidden, "Cannot remove yourself from officers.")
		return
	}

	memberSecurity, ok := dbGetSecurity(w, memberId)
	if !ok {
		return
	}
	officerSecurity, ok := dbGetSecurity(w, officerId)
	if !ok {
		return
	}
	if memberSecurity <= officerSecurity {
		respondError(w, http.StatusForbidden, "Cannot modify officer with equal or higher security.")
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
	stmt := `
		SELECT
			member.id,
			member.email,
			member.name,
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
			&officers[i].Name,
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
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId
	memberId, ok := dbGetActiveMemberId(w, sub)
	if !ok {
		return
	}

	// Get request body
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var officer officerStruct
	err := decoder.Decode(&officer)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if !dbEnsureMemberIdExists(w, officer.MemberId) ||
		!dbEnsureNotOfficer(w, officer.MemberId) {
		return
	}

	memberSecurity, ok := dbGetSecurity(w, memberId)
	if !ok {
		return
	}
	if memberSecurity < officer.Security {
		respondError(w, http.StatusForbidden, "Cannot add officer with higher security.")
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
		officer.MemberId,
		officer.ExpireDatetime,
		officer.Position,
		officer.Security)
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}
