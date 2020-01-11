package handler

import (
	"fmt"
	"net/http"
	"strconv"
)

// guid table -> id, guid, memberId, tripId, status -> {default to 'NONE', change to 'APPROVE', 'DECLINE'
// for i in approved_people -> create guid, send emails (TODO create PATCH /tripApproval/{approvalGuid}/{status}, status must be APPROVE|DECLINE
// in PatchWebtoolsApproval:
// - if !dbEnsureTripNoApproval(w, tripId) -> ensure not exists with
//   tripId AND status != NONE, otherwise return bad request
//    - DECLINE -> return no conent
//    - APPROVE -> stageEmailNewTrip(w, tripId)
const GUID_LENGTH = 64

/* HELPERS */
func approveNewTrip(w http.ResponseWriter, tripId int) bool {
	stmt := `
    SELECT member_id
    FROM trip_approver
    WHERE datetime('now') < datetime(expireDatetime)`
	rows, err := db.Query(stmt)
	if !checkError(w, err) {
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var memberId int
		err = rows.Scan(&memberId)
		if !checkError(w, err) {
			return false
		}

		guidCode := generateCode(GUID_LENGTH)
		stmt = `
      INSERT INTO guid (
        code,
        member_id,
        trip_id,
        status)
      VALUES (?, ?, ?, 'NONE')`
		_, err = db.Exec(stmt, guidCode, memberId, tripId)
		if !checkError(w, err) {
			return false
		}

		// Notify approver
		memberIdStr := strconv.Itoa(memberId)
		title, date, description, ok := dbGetTripApprovalSummary(w, tripId)
		if !ok {
			return false
		}

		// TODO email links
		emailSubject := fmt.Sprintf(
			"[Trip Approval] ID: %d, Title: %s", tripId, title)
		emailBody := fmt.Sprintf(
			"The following trip needs approval:\n"+
				"\n"+
				"%s\n"+
				"\n"+
				"Scheduled for: %s\n"+
				"Description: %s\n"+
				"\n"+
				"To View this trip go <here>\n"+
				"To Administer or cancel this trip go <here>\n"+
				"\n"+
				"<Approve Trip%s>\n"+
				"\n"+
				"<Deny Trip%s>\n", tripId, title, date, description, guidCode, guidCode)
		if !stageEmail(w, memberIdStr, tripId, 0, emailSubject, emailBody) {
			return false
		}
	}

	err = rows.Err()
	if !checkError(w, err) {
		return false
	}

	return true
}

/* MAIN FUNCTIONS */
