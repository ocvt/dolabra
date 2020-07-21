package handler

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/service/ses"
	"gopkg.in/mail.v2"
)

type emailStruct struct {
	/* Only used to GET already sent emails */
	SentDatetime string `json:"sentDatetime,omitempty"`
	/* Managed server side */
	Id                 int    `json:"id,omitempty"`
	CreateDatetime     string `json:"createDatetime,omitempty"`
	Sent               bool   `json:"sent,omitempty"`
	NotificationTypeId string `json:"notificationTypeId,omitempty"`
	TripId             int    `json:"tripId,omitempty"`
	ReplyToId          int    `json:"replyTo,omitempty"`
	ToId               int    `json:"toId,omitempty"` // 0 if not direct message
	/* Required fields for creating announcements */
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type rawEmailStruct struct {
	FromName     string
	FromEmail    string
	ReplyToEmail string
	ReplyToName  string
	ToName       string
	ToEmail      string
	Subject      string
	Body         string
}

/*
 * Actually send an email
 */
func sendEmail(sesService *ses.SES, email rawEmailStruct) (*ses.SendRawEmailOutput, error) {
	msg := mail.NewMessage()
	msg.SetHeader("From", fmt.Sprintf("%s <%s>", email.FromName, email.FromEmail))
	msg.SetHeader("To", fmt.Sprintf("%s <%s>", email.ToName, email.ToEmail))
	msg.SetHeader("Subject", email.Subject)
	msg.SetBody("text/html", email.Body)

	var rawMsg bytes.Buffer
	msg.WriteTo(&rawMsg)

	input := &ses.SendRawEmailInput{
		RawMessage: &ses.RawMessage{
			Data: rawMsg.Bytes(),
		},
	}

	output, err := sesService.SendRawEmail(input)
	return output, err
}

/*
 * Insert entry into email table to eventually send
 * - TRIP_ALERT_* are special types to indicate direct trip alerts
 * - tripId field is used ONLY with TRIP_ALERT_* types
 *	 otherwise it is purely for logging the relevant trip
 *   or not all for a non trip related email
 */
func stageEmail(w http.ResponseWriter, email emailStruct) bool {
	email.Subject = "[OCVT] " + email.Subject

	stmt := `
		INSERT INTO email (
			create_datetime,
			sent_datetime,
			sent,
			notification_type_id,
			trip_id,
			to_id,
			reply_to_id,
			subject,
			body)
		VALUES (datetime('now'), datetime(0, 'unixepoch'), false, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(
		stmt,
		email.NotificationTypeId,
		email.TripId,
		email.ToId,
		email.ReplyToId,
		email.Subject,
		email.Body)
	if !checkError(w, err) {
		return false
	}

	return true
}

/* HELPERS */
func stageEmailNewTrip(w http.ResponseWriter, tripId int) bool {
	return true
}

func stageEmailTripReminder(tripId int) error {
	return nil
}
