package handler

import (
  "fmt"
  "net/http"
//  "net/smtp"
)

var SMTP_PASSWORD string
var SMTP_USERNAME string
var SMTP_HOSTNAME string
var SMTP_PORT string
var SMTP_FROM_NAME_DEFAULT string
var SMTP_FROM_EMAIL_DEFAULT string

func sendEmail(w http.ResponseWriter, replyName string, replyEmail string,
    toName string, toEmail string, subject string, body string) bool {
  if replyEmail == "" {
    replyEmail = SMTP_FROM_EMAIL_DEFAULT
  }

  subject = fmt.Sprintf("[OCVT] %s", subject)

  // TODO add unsubscribe body footer

  message := fmt.Sprintf("From: %s <%s>\n" +
                         "Reply-To: %s <%s>\n" +
                         "To: %s <%s>\n" +
                         "Subject: %s\n\n" +
                         "%s",
                         SMTP_FROM_NAME_DEFAULT,
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

/* HELPERS */
func sendEmailToMember(w http.ResponseWriter, fromId int, toId int,
    subject string, body string) bool {

  fromName := ""
  fromEmail := ""
  if fromId != 0 {
    var ok bool
    fromName, ok = dbGetMemberName(w, fromId)
    if !ok {
      return false
    }
    fromEmail, ok = dbGetMemberEmail(w, fromId)
    if !ok {
      return false
    }
  }

  toName, ok := dbGetMemberName(w, toId)
  if !ok {
    return false
  }
  toEmail, ok := dbGetMemberEmail(w, toId)
  if !ok {
    return false
  }

  return sendEmail(w, fromName, fromEmail, toName, toEmail, subject, body)
}

func sendEmailToTripSignup(w http.ResponseWriter, fromId int, toId int,
    tripId int, subject string, body string) bool {

  tripName, ok := dbGetTripName(w, tripId)
  if !ok {
    return false
  }

  subject = fmt.Sprintf(subject, tripName)
  body = fmt.Sprintf(subject, tripName)

  return sendEmailToMember(w, fromId, toId, subject, body)
}
