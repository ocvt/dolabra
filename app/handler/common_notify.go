package handler

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/ocvt/dolabra/utils"
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

func stageEmail(w http.ResponseWriter, email emailStruct) bool {
	err := stageEmailPlain(email)
	return checkError(w, err)
}

/*
 * Insert entry into email table to eventually send
 * - TRIP_ALERT_* are special types to indicate direct trip alerts
 * - tripId field is used ONLY with TRIP_ALERT_* types
 *	 otherwise it is purely for logging the relevant trip
 *   or not all for a non trip related email
 */
func stageEmailPlain(email emailStruct) error {
	label := utils.GetConfig().EmailLabel
	email.Subject = "[" + label + "] " + email.Subject

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

	return err
}

/* HELPERS */
func stageEmailNewTrip(w http.ResponseWriter, tripId int) bool {
	label := utils.GetConfig().EmailLabel
	url := utils.GetConfig().FrontendUrl
	trip, ok := dbGetTrip(w, tripId)
	if !ok {
		return false
	}

	date := prettyPrintDate(trip.StartDatetime)
	email := emailStruct{
		NotificationTypeId: trip.NotificationTypeId,
		ReplyToId:          0,
		ToId:               0,
		TripId:             tripId,
	}
	email.Subject = fmt.Sprintf("New Trip: %s", trip.Name)
	email.Body = fmt.Sprintf(
		"A new trip has been posted to %s scheduled for %s:<br>"+
			"<h3>%s</h3>"+
			"<br>"+
			"Trip Summary: %s<br>"+
			"<br>"+
			"Location Directions: %s<br>"+
			"<br>"+
			"<br>"+
			"Full details and the signup form can be found at "+
			"<a href=\"%s/trips/%d\">%s/trips/%d</a><br>"+
			"<br>"+
			"<br>"+
			"<br>"+
			"<hr>"+
			"This message has been sent via the %s Websystem.<br>"+
			"You can modify your notification and account settings "+
			"<a href=\"%s/myocvt\">here</a>.<br> You can also click "+
			"<a href=\"%s/unsubscribe\">here</a> to unsubscribe.<br>"+
			"<hr>",
		label, date, trip.Name, trip.Summary, trip.LocationDirections,
		url, tripId, url, tripId, label, url, url)

	return stageEmail(w, email)
}

func stageEmailTripApproval(w http.ResponseWriter, tripId int, memberId int, guidCode string) bool {
	url := utils.GetConfig().FrontendUrl
	trip, ok := dbGetTrip(w, tripId)
	if !ok {
		return false
	}

	email := emailStruct{
		NotificationTypeId: "TRIP_APPROVAL",
		ReplyToId:          0,
		ToId:               memberId,
		TripId:             tripId,
	}
	email.Subject = fmt.Sprintf(
		"Trip Approval - ID: %d, Title: %s", tripId, trip.Name)
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

	return stageEmail(w, email)
}

func stageEmailTripReminder(tripId int) {
	url := utils.GetConfig().FrontendUrl
	trip, err := dbGetTripPlain(tripId)
	if err != nil {
		log.Fatal(err)
	}

	email := emailStruct{
		NotificationTypeId: "TRIP_MESSAGE_NOTIFY",
		ReplyToId:          trip.MemberId,
		ToId:               0,
		TripId:             tripId,
	}
	email.Subject = "Trip Reminder: " + trip.Name
	email.Body = fmt.Sprintf(
		"This is a reminder for the trip scheduled tomorrow:<br"+
			"<h3>%s</h3>"+
			"<br>"+
			"Full trip details can be found at "+
			"<a href=\"%s/trips/%d\">%s/trips/%d</a><br>"+
			"<br>"+
			"If you're not planning on attending please log in and cancel your "+
			"attendance so someone else can take your spot.<br>"+
			"<br>",
		trip.Name, url, tripId, url, tripId)

	err = stageEmailPlain(email)
	if err != nil {
		log.Print(err)
	}
}

func stageEmailTripCancel(w http.ResponseWriter, tripId int) bool {
	trip, ok := dbGetTrip(w, tripId)
	if !ok {
		return false
	}

	email := emailStruct{
		NotificationTypeId: "TRIP_MESSAGE_NOTIFY",
		ReplyToId:          0,
		ToId:               0,
		TripId:             tripId,
	}
	email.Subject = "Trip CANCELED: " + trip.Name
	email.Body = fmt.Sprintf(
		"You are receiving this message becaused you are signed up for this trip<br>"+
			"<br>"+
			"This trip has been canceled:<br>"+
			"<h3>%s</h3>"+
			"<br>",
		trip.Name)

	return stageEmail(w, email)
}
