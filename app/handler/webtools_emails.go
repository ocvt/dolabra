package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ocvt/dolabra/utils"
)

func GetWebtoolsEmails(w http.ResponseWriter, r *http.Request) {
	stmt := `
		SELECT *
		FROM email
		ORDER BY datetime(create_datetime) DESC`
	rows, err := db.Query(stmt)
	if !checkError(w, err) {
		return
	}
	defer rows.Close()

	var emails = []*emailStruct{}
	i := 0
	for rows.Next() {
		emails = append(emails, &emailStruct{})
		err = rows.Scan(
			&emails[i].Id,
			&emails[i].CreateDatetime,
			&emails[i].SentDatetime,
			&emails[i].NotificationTypeId,
			&emails[i].TripId,
			&emails[i].ToId,
			&emails[i].ReplyToId,
			&emails[i].Subject,
			&emails[i].Body)
		if !checkError(w, err) {
			return
		}
		i++
	}

	err = rows.Err()
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, map[string][]*emailStruct{"emails": emails})
}

func PostWebtoolsEmails(w http.ResponseWriter, r *http.Request) {
	// Get request body
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var email emailStruct
	err := decoder.Decode(&email)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	email.NotificationTypeId = "GENERAL_ANNOUNCEMENTS"
	email.TripId = 3000
	email.ReplyToId = 8000000
	email.ToId = 8000000

	label := utils.GetConfig().EmailLabel
	url := utils.GetConfig().FrontendUrl
	email.Body += fmt.Sprintf(
		"<br>"+
			"<br>"+
			"<br>"+
			"<hr>"+
			"This message has been sent via the %s Websystem.<br>"+
			"You can modify your notification and account settings "+
			"<a href=\"%s/myocvt\">here</a>.<br> You can also click "+
			"<a href=\"%s/unsubscribe?email=EMAIL_HERE&sig=SIG_HERE\">here</a> to unsubscribe.<br>"+
			"<hr>", label, url, url)

	if !stageEmail(w, email) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

type directEmailStruct struct {
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	MemberIds []int  `json:"memberIds"`
}

/* Send an email directly to specific members, e.g. for testing */
func PostWebtoolsEmailsDirect(w http.ResponseWriter, r *http.Request) {
	// Get request body
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var direct directEmailStruct
	err := decoder.Decode(&direct)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if len(direct.MemberIds) == 0 || len(direct.MemberIds) > 100 {
		respondError(w, http.StatusBadRequest, "Select between 1 and 100 members")
		return
	}

	label := utils.GetConfig().EmailLabel
	url := utils.GetConfig().FrontendUrl
	body := direct.Body + fmt.Sprintf(
		"<br>"+
			"<br>"+
			"<br>"+
			"<hr>"+
			"This message has been sent via the %s Websystem.<br>"+
			"You can modify your notification and account settings "+
			"<a href=\"%s/myocvt\">here</a>.<br> You can also click "+
			"<a href=\"%s/unsubscribe?email=EMAIL_HERE&sig=SIG_HERE\">here</a> to unsubscribe.<br>"+
			"<hr>", label, url, url)

	for _, memberId := range direct.MemberIds {
		email := emailStruct{
			NotificationTypeId: "DIRECT",
			TripId:             3000,
			ReplyToId:          8000000,
			ToId:               memberId,
			Subject:            direct.Subject,
			Body:               body,
		}
		if !stageEmail(w, email) {
			return
		}
	}

	respondJSON(w, http.StatusNoContent, nil)
}
