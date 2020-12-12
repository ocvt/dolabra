package handler

import (
	"encoding/json"
	"net/http"
)

/* MAIN FUNCTIONS */
func DeleteWebtoolsApprovers(w http.ResponseWriter, r *http.Request) {
	approverId, ok := getURLIntParam(w, r, "memberId")
	if !ok {
		return
	}

	stmt := `
	 DELETE FROM trip_approver
	 WHERE member_id = ?`
	_, err := db.Exec(stmt, approverId)
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

func GetWebtoolsApprovers(w http.ResponseWriter, r *http.Request) {
	stmt := `
		SELECT
			member.id,
			member.email,
			member.name,
			member.cell_number,
			trip_approver.expire_datetime
		FROM member
		INNER JOIN trip_approver ON trip_approver.member_id = member.id`
	rows, err := db.Query(stmt)
	if !checkError(w, err) {
		return
	}
	defer rows.Close()

	var approvers = []*approverStruct{}
	i := 0
	for rows.Next() {
		approvers = append(approvers, &approverStruct{})
		err = rows.Scan(
			&approvers[i].MemberId,
			&approvers[i].Email,
			&approvers[i].Name,
			&approvers[i].CellNumber,
			&approvers[i].ExpireDatetime)
		if !checkError(w, err) {
			return
		}
		i++
	}

	err = rows.Err()
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, map[string][]*approverStruct{"approvers": approvers})
}

func PostWebtoolsApprovers(w http.ResponseWriter, r *http.Request) {
	// Get request body
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var approver approverStruct
	err := decoder.Decode(&approver)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Permissions
	if !dbEnsureMemberIdExists(w, approver.MemberId) ||
		!dbEnsureNotApprover(w, approver.MemberId) {
		return
	}

	stmt := `
		INSERT INTO trip_approver (
			member_id,
			create_datetime,
			expire_datetime)
		VALUES (?, datetime('now'), datetime(?))`
	_, err = db.Exec(
		stmt,
		approver.MemberId,
		approver.ExpireDatetime)
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}
