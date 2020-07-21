package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gitlab.com/ocvt/dolabra/utils"
)

// guid table -> id, guid, memberId, tripId, status -> {default to 'NONE', change to 'APPROVE', 'DECLINE'
// for i in approved_people -> create guid, send emails (TODO create PATCH /tripApproval/{approvalGuid}/{status}, status must be APPROVE|DECLINE
// in PatchWebtoolsApproval:
// - if !dbEnsureTripNoApproval(w, tripId) -> ensure not exists with
//	 tripId AND status != NONE, otherwise return bad request
//		- DECLINE -> return no conent
//		- APPROVE -> stageEmailNewTrip(w, tripId)
const GUID_LENGTH = 64

type approverStruct struct {
	/* Managed server side */
	// from member table
	CellNumber string `json:"cellNumber,omitempty"`
	Email      string `json:"email,omitempty"`
	FirstName  string `json:"firstName,omitempty"`
	LastName   string `json:"lastName,omitempty"`
	/* Required fields for creating a trip */
	MemberId       int    `json:"memberId"`
	ExpireDatetime string `json:"expireDatetime"`
}

/* HELPERS */
func approveNewTrip(w http.ResponseWriter, tripId int) bool {
	email := emailStruct{
		NotificationTypeId: "TRIP_APPROVAL",
		ReplyToId:          0,
		ToId:               0,
		TripId:             tripId,
	}

	url := utils.GetConfig().FrontendUrl
	guidCode := generateCode(GUID_LENGTH)
	trip, ok := dbGetTrip(w, tripId)
	if !ok {
		return false
	}
	email.Subject = fmt.Sprintf(
		"[Trip Approval] - ID: %d, Title: %s", tripId, trip.Name)
	email.Body = fmt.Sprintf(
		"The following trip needs approval:<br>"+
			"<br>"+
			"Title: %s<br>"+
			"<br>"+
			"Scheduled for: %s<br>"+
			"<br>"+
			"Summary: %s<br>"+
			"<br>"+
			"Description: %s<br>"+
			"<br>"+
			"<br>"+
			"To View this trip go <a href=\"%s/trips/%d\">here</a><br>"+
			"To Administer or cancel this trip go <a href=\"%s/trips/%d/admin\">here</a><br>"+
			"<br>"+
			"<a href=\"%s/tripapproval/%s/approve\">Approve Trip</a><br>"+
			"<br>"+
			"<a href=\"%s/tripapproval/%s/deny\">Deny Trip</a><br>",
		trip.Name, trip.CreateDatetime, trip.Summary, trip.Description, url, tripId, url, tripId, url, guidCode, url, guidCode)

	if !stageEmail(w, email) {
		return false
	}

	return true
}

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
			member.first_name,
			member.last_name,
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
			&approvers[i].FirstName,
			&approvers[i].LastName,
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
