package handler

import (
  "fmt"
  "net/http"
//  "net/smtp"
  "strconv"
  "strings"
)

var SMTP_PASSWORD string
var SMTP_USERNAME string
var SMTP_HOSTNAME string
var SMTP_PORT string
var SMTP_FROM_FIRST_NAME_DEFAULT string
var SMTP_FROM_LAST_NAME_DEFAULT string
var SMTP_FROM_EMAIL_DEFAULT string

/*
 * Actually send an email
 */
func sendEmail(w http.ResponseWriter, replyName string, replyEmail string,
    toName string, toEmail string, subject string, body string) bool {
  fullFromName := SMTP_FROM_FIRST_NAME_DEFAULT + " " + SMTP_FROM_LAST_NAME_DEFAULT
  message := fmt.Sprintf("From: %s <%s>\n" +
                         "Reply-To: %s <%s>\n" +
                         "To: %s <%s>\n" +
                         "Subject: %s\n\n" +
                         "%s",
                         fullFromName,
                         SMTP_FROM_EMAIL_DEFAULT,
                         replyName,
                         replyEmail,
                         toName,
                         toEmail,
                         subject,
                         body)

//  auth := smtp.PlainAuth("", SMTP_USERNAME, SMTP_PASSWORD, SMTP_HOSTNAME)
  fmt.Printf("MESSAGE: %s\n",  message)
//  err := smtp.SendMail(fmt.Sprintf("%s:%s", SMTP_HOSTNAME, SMTP_PORT), auth,
//      SMTP_FROM_EMAIL_DEFAULT, []string{toEmail}, []byte(message))
//  if err != nil {
//    respondError(w, http.StatusInternalServerError, err.Error())
//    return false
//  }

  return true
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
 *   otherwise it is purely for logging the relevant trip
 */
func stageEmail(w http.ResponseWriter, notificationType string, tripId int,
    replyToId int, subject string, body string) bool {

  // fromId should always member id of default Websystem account
  fromId := 0
  subject = "[OCVT] " + subject

  isTripAlert := strings.HasPrefix(notificationType, "TRIP_ALERT_")
  _, err := strconv.Atoi(notificationType)
  if !isTripAlert && err != nil {
    body = body +
           "=========================================\n" +
           "You received this message because you\n" +
           "are on the OCVT email list. <Unsubscribe>"
  }

  stmt := `
    INSERT INTO email (
      create_datetime,
      sent,
      notification_type_id,
      trip_id,
      from_id,
      reply_to_id,
      subject,
      body)
    VALUES (datetime('now'), false, ?, ?, ?, ?, ?, ?)`
  _, err = db.Exec(
    stmt,
    notificationType,
    tripId,
    fromId,
    replyToId,
    subject,
    body)
  if !checkError(w, err) {
    return false
  }

  return true
}

/* HELPERS */
func stageEmailNewTrip(w http.ResponseWriter, tripId int) bool {
  return true
}
