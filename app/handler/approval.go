package handler

import (
	"fmt"
	"net/http"
)

// guid table -> id, guid, memberId, tripId, status -> {default to 'NONE', change to 'APPROVE', 'DECLINE'
// for i in approved_people -> create guid, send emails (TODO create PATCH /tripApproval/{approvalGuid}/{status}, status must be APPROVE|DECLINE
// in PatchWebtoolsApproval:
// - if !dbEnsureTripNoApproval(w, tripId) -> ensure not exists with
//	 tripId AND status != NONE, otherwise return bad request
//		- DECLINE -> return no conent
//		- APPROVE -> stageEmailNewTrip(w, tripId)
const GUID_LENGTH = 64

/* HELPERS */
func approveNewTrip(w http.ResponseWriter, tripId int) bool {
	email := emailStruct{
		NotificationTypeId: "TRIP_APPROVAL",
		ReplyToId:          0,
		ToId:               0,
		TripId:             tripId,
	}

	guidCode := generateCode(GUID_LENGTH)
	title, date, description, ok := dbGetTripApprovalSummary(w, tripId)
	if !ok {
		return false
	}
	email.Subject = fmt.Sprintf(
		"[Trip Approval] ID: %d, Title: %s", tripId, title)
	email.Body = fmt.Sprintf(
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

	if !stageEmail(w, email) {
		return false
	}

	return true
}

/* MAIN FUNCTIONS */
