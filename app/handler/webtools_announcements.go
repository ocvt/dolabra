package handler

import (
	"encoding/json"
	"net/http"
)

type emailStruct struct {
	/* Managed server side */
	Id             int    `json:"id,omitempty"`
	CreateDatetime string `json:"createDatetime,omitempty"`
	SentDatetime   string `json:"sentDatetime,omitempty"`
	Sent           bool   `json:"sent,omitempty"`
	TripId         int    `json:"tripId,omitempty"`
	FromId         string `json:"fromId,omitempty"`
	ReplyToId      string `json:"replyTo"`
	/* Required fields for creating announcements */
	NotificationTypeId string `json:"notificationTypeId"`
	Subject            string `json:"subject"`
	Body               string `json:"body"`
}

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
			&emails[i].FromId,
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

func PostWebtoolsAnnouncements(w http.ResponseWriter, r *http.Request) {
	// Get request body
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var announcement emailStruct
	err := decoder.Decode(&announcement)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Notify members
	if !stageEmail(w,
		announcement.NotificationTypeId, 0, 0,
		announcement.Subject,
		announcement.Body) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}
