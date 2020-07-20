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
	trip, ok := dbGetTrip(w, tripId)
	if !ok {
		return false
	}
	email.Subject = fmt.Sprintf(
		"[OCVT Trip Approval] - ID: %d, Title: %s", tripId, trip.Name)
	email.Body = fmt.Sprintf(
		"The following trip needs approval:\n"+
			"\n"+
			"Title: %s\n"+
			"\n"+
			"Scheduled for: %s\n"+
			"\n"+
			"Summary: %s\n"+
			"\n"+
			"Description: %s\n"+
			"\n"+
			"\n"+
			"To View this trip go <a href=\"???/trips/%d\">here</a>\n"+
			"To Administer or cancel this trip go <a href=\"???/trips/%d/admin\">here</a>\n"+
			"\n"+
			"<a href=\"???/tripapproval/%d/APPROVE\">Approve Trip</a>\n"+
			"\n"+
			"<a href=\"???/tripapproval/%d/DENY\">Deny Trip</a>\n",
		tripId, trip.Name, trip.CreateDatetime, trip.Summary, trip.Description, tripId, tripId, guidCode, guidCode)

	if !stageEmail(w, email) {
		return false
	}

	return true
}

/* MAIN FUNCTIONS */
