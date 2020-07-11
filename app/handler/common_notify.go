package handler

import (
	"fmt"
	"net/http"
	//	"net/smtp"
)

var SMTP_PASSWORD string
var SMTP_USERNAME string
var SMTP_HOSTNAME string
var SMTP_PORT string
var SMTP_FROM_FIRST_NAME_DEFAULT string
var SMTP_FROM_LAST_NAME_DEFAULT string
var SMTP_FROM_EMAIL_DEFAULT string

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

/*
 * Actually send an email
 */
func sendEmail(fromName string, fromEmail string, replyName string,
	replyEmail string, toName string, toEmail string, subject string,
	body string) {
	message := fmt.Sprintf(
		"From: %s <%s>\n"+
			"Reply-To: %s <%s>\n"+
			"To: %s <%s>\n"+
			"Subject: %s\n\n"+
			"%s",
		fromName,
		fromEmail,
		replyName,
		replyEmail,
		toName,
		toEmail,
		subject,
		body)

	//	auth := smtp.PlainAuth("", SMTP_USERNAME, SMTP_PASSWORD, SMTP_HOSTNAME)
	fmt.Printf("MESSAGE: %s\n", message)
	//	err := smtp.SendMail(fmt.Sprintf("%s:%s", SMTP_HOSTNAME, SMTP_PORT), auth,
	//			SMTP_FROM_EMAIL_DEFAULT, []string{toEmail}, []byte(message))
	//	if err != nil {
	//		log.Fatal(err)
	//	}
}

/*
 * Process db fields and send email(s)
 * - Ratelimits according SES TODO
 * - MAY take a long time
 * - Should be called from separate thread checking for emails every 5 minutes
 */
func processAndSendEmail(w http.ResponseWriter, emailId int) bool {
	// Lookup + process all email FIELDS TODO
	// Send emails
	// If type is starts with TRIP_ALERT, send separate email to fromId TODO
	//emailBody = fmt.Sprintf("You are receiving this message because you sent it:\n\n%s", emailBody)
	return true
}

/*
 * Insert entry into email table to eventually send
 * - TRIP_ALERT_* are special types to indicate direct trip alerts
 * - tripId field is used ONLY with TRIP_ALERT_* types
 *	 otherwise it is purely for logging the relevant trip
 *   or not all for a non trip related email
 */
func stageEmail(w http.ResponseWriter, email emailStruct) bool {

	// fromId should always member id of default Websystem account
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
