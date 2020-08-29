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
			&emails[i].Sent,
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
	email.TripId = 0
	email.ReplyToId = 0

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
			"<a href=\"%s/unsubscribe\">here</a> to unsubscribe.<br>"+
			"<hr>", label, url, url)

	if !stageEmail(w, email) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}
