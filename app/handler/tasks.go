package handler

import (
	"bytes"
	"container/list"
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/ocvt/dolabra/utils"
	"github.com/stripe/stripe-go/v71"
	checkoutsession "github.com/stripe/stripe-go/v71/checkout/session"
)

/* Periodic housekeeping: expired rows + trip reminders */
func DoTasks() {
	/* Remove expired trip approvers */
	stmt := `
		DELETE FROM trip_approver
		WHERE datetime(expire_datetime) < datetime('now')`
	_, err := db.Exec(stmt)
	if err != nil {
		log.Fatal(err)
	}
	/*********************************/

	/* Remove expired officers */
	stmt = `
		DELETE FROM officer
		WHERE datetime(expire_datetime) < datetime('now')`
	_, err = db.Exec(stmt)
	if err != nil {
		log.Fatal(err)
	}
	/***************************/

	/* Remove expired quick signups */
	stmt = `
		DELETE FROM quick_signup
		WHERE datetime(expire_datetime) < datetime('now')`
	_, err = db.Exec(stmt)
	if err != nil {
		log.Fatal(err)
	}
	/********************************/

	/* Stage trip reminder email */
	// Sends 1 day before trip as long as trip was created >= 3 days before start
	stmt = `
		SELECT id
		FROM trip
		WHERE
			datetime(create_datetime) < datetime(start_datetime, '-3 days') AND
			datetime(start_datetime, '-1 day') < datetime('now') AND
			datetime('now') < datetime(start_datetime) AND
			cancel = false AND
			publish = true AND
			reminder_sent = false`
	rows, err := db.Query(stmt)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	tripIds := list.New()
	for rows.Next() {
		var tripId int
		err = rows.Scan(&tripId)
		if err != nil {
			log.Fatal(err)
		}
		tripIds.PushBack(tripId)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	for t := tripIds.Front(); t != nil; t = t.Next() {
		tripId := t.Value.(int)
		// Stage email
		stageEmailTripReminder(tripId)

		// Mark as sent
		stmt = `
			UPDATE trip
			SET reminder_sent = true
			WHERE id = ?`
		_, err = db.Exec(stmt, tripId)
		if err != nil {
			log.Fatal(err)
		}
	}
	/*****************************/

	/* Complete card payments the Stripe webhook never confirmed */
	reconcileStripePayments()
	/*************************************************************/

	/* Process emails: expand staged emails to recipients, then send */
	expandStagedEmails()
	sendPendingEmails()
	/*****************************************************************/
}

/*
 * The webhook is the only thing that completes a card payment, and after the
 * Aug 2026 proxy migration it silently delivered nothing for a month. Poll
 * Stripe for sessions the webhook has not confirmed and complete the ones that
 * actually paid. Runs every 5 minutes; skips sessions under 10 minutes old so
 * the webhook gets first crack, and stops looking after 45 days (Checkout
 * sessions expire after 24h, so anything older is an abandoned checkout).
 */
var lastStripeReconcile time.Time

func reconcileStripePayments() {
	stripeKey := utils.GetConfig().StripeSecretKey
	if stripeKey == "" || time.Since(lastStripeReconcile) < 5*time.Minute {
		return
	}
	lastStripeReconcile = time.Now()

	stmt := `
		SELECT DISTINCT payment_id
		FROM payment
		WHERE
			payment_method = 'STRIPE'
			AND (store_item_id = 'MEMBERSHIP' OR store_item_id = 'CUSTOM')
			AND completed = false
			AND datetime(create_datetime) > datetime('now', '-45 days')
			AND datetime(create_datetime) < datetime('now', '-10 minutes')`
	rows, err := db.Query(stmt)
	if err != nil {
		log.Print("ERROR loading unconfirmed Stripe sessions: " + err.Error())
		return
	}
	defer rows.Close()

	sessionIds := []string{}
	for rows.Next() {
		var sessionId string
		err = rows.Scan(&sessionId)
		if err != nil {
			log.Print("ERROR loading unconfirmed Stripe sessions: " + err.Error())
			return
		}
		sessionIds = append(sessionIds, sessionId)
	}
	err = rows.Err()
	if err != nil {
		log.Print("ERROR loading unconfirmed Stripe sessions: " + err.Error())
		return
	}

	stripe.Key = stripeKey
	for _, sessionId := range sessionIds {
		params := &stripe.CheckoutSessionParams{}
		params.AddExpand("payment_intent")
		checkoutSession, err := checkoutsession.Get(sessionId, params)
		if err != nil {
			log.Printf("ERROR fetching Stripe session %s: %s", sessionId, err.Error())
			continue
		}
		if checkoutSession.PaymentIntent == nil ||
			checkoutSession.PaymentIntent.Status != stripe.PaymentIntentStatusSucceeded ||
			stripeSessionRefunded(checkoutSession) {
			continue
		}

		rw := newDiscardResponseWriter()
		applied, ok := completeStripeSession(rw, sessionId)
		if !ok {
			log.Printf("ERROR completing Stripe session %s: %s", sessionId, rw.body.String())
			continue
		}
		if applied {
			log.Printf("Reconciled Stripe session %s (webhook never confirmed it)", sessionId)
		}
	}
}

/*
 * A refund leaves the payment intent reading "succeeded", so check the charge.
 * Refunding a duplicate checkout in Stripe is then enough on its own: the
 * sweep will not turn the refunded session into membership years.
 */
func stripeSessionRefunded(s *stripe.CheckoutSession) bool {
	if s.PaymentIntent == nil || s.PaymentIntent.Charges == nil {
		return false
	}
	for _, charge := range s.PaymentIntent.Charges.Data {
		if charge.Refunded {
			return true
		}
	}
	return false
}

/*
 * Lets the request-scoped db helpers run from the task ticker, where there is
 * no client to answer. The error body they would have sent is kept for the log.
 */
type discardResponseWriter struct {
	header http.Header
	body   bytes.Buffer
}

func newDiscardResponseWriter() *discardResponseWriter {
	return &discardResponseWriter{header: http.Header{}}
}

func (d *discardResponseWriter) Header() http.Header         { return d.header }
func (d *discardResponseWriter) Write(b []byte) (int, error) { return d.body.Write(b) }
func (d *discardResponseWriter) WriteHeader(int)             {}

type recipientStruct struct {
	name  string
	email string
}

/*
 * Bulk mail is anything a member opted into via notification preferences, as
 * opposed to transactional mail tied to a specific trip or signup. Only bulk
 * mail carries unsubscribe headers and an unsubscribe footer.
 */
func isBulkNotification(notificationTypeId string) bool {
	return notificationTypeId != "DIRECT" &&
		notificationTypeId != "TRIP_APPROVAL" &&
		!strings.HasPrefix(notificationTypeId, "TRIP_ALERT") &&
		!strings.HasPrefix(notificationTypeId, "TRIP_MESSAGE")
}

/*
 * Resolve a staged email to its recipient list based on notification type
 */
func resolveRecipients(email emailStruct) ([]recipientStruct, bool) {
	doQuickSignup := false
	var rows *sql.Rows
	var err error
	if email.NotificationTypeId == "DIRECT" {
		stmt := `
			SELECT id
			FROM member
			WHERE id = ? AND active = true`
		rows, err = db.Query(stmt, email.ToId)
	} else if email.NotificationTypeId == "TRIP_APPROVAL" {
		stmt := `
			SELECT member_id
			FROM trip_approver
			WHERE member_id = ?`
		rows, err = db.Query(stmt, email.ToId)
	} else if strings.HasPrefix(email.NotificationTypeId, "TRIP_ALERT") ||
		email.NotificationTypeId == "TRIP_MESSAGE_DIRECT" {
		stmt := `
			SELECT member_id
			FROM trip_signup
			WHERE member_id = ? AND trip_id = ?`
		rows, err = db.Query(stmt, email.ToId, email.TripId)
	} else if email.NotificationTypeId == "TRIP_MESSAGE_NOTIFY" {
		stmt := `
			SELECT member_id
			FROM trip_signup
			WHERE trip_id = ? AND
				(attending_code = 'ATTEND' OR
				 attending_code = 'FORCE' OR
				 attending_code = 'WAIT')`
		rows, err = db.Query(stmt, email.TripId)
	} else if email.NotificationTypeId == "TRIP_MESSAGE_ATTEND" {
		stmt := `
			SELECT member_id
			FROM trip_signup
			WHERE trip_id = ? AND
				(attending_code = 'ATTEND' OR attending_code = 'FORCE')`
		rows, err = db.Query(stmt, email.TripId)
	} else if email.NotificationTypeId == "TRIP_MESSAGE_WAIT" {
		stmt := `
			SELECT member_id
			FROM trip_signup
			WHERE trip_id = ? AND attending_code = 'WAIT'`
		rows, err = db.Query(stmt, email.TripId)
	} else {
		doQuickSignup = true
		// Send to all ACTIVE members with notification preference set and quicksignups
		stmt := `
			SELECT id
			FROM member
			WHERE active = true`
		rows, err = db.Query(stmt)
	}

	if err != nil {
		log.Print("ERROR resolving recipients: " + err.Error())
		return nil, false
	}
	defer rows.Close()

	memberIds := []int{}
	for rows.Next() {
		var memberId int
		err = rows.Scan(&memberId)
		if err != nil {
			log.Print("ERROR resolving recipients: " + err.Error())
			return nil, false
		}
		memberIds = append(memberIds, memberId)
	}
	err = rows.Err()
	if err != nil {
		log.Print("ERROR resolving recipients: " + err.Error())
		return nil, false
	}

	recipients := []recipientStruct{}
	for _, memberId := range memberIds {
		if isBulkNotification(email.NotificationTypeId) &&
			!dbCheckMemberWantsNotification(memberId, email.NotificationTypeId) {
			continue
		}
		toName, toEmail := dbGetMemberNameEmail(memberId)
		recipients = append(recipients, recipientStruct{name: toName, email: toEmail})
	}

	if doQuickSignup {
		stmt := `
			SELECT DISTINCT email
			FROM quick_signup`
		qsRows, err := db.Query(stmt)
		if err != nil {
			log.Print("ERROR resolving quicksignups: " + err.Error())
			return nil, false
		}
		defer qsRows.Close()

		for qsRows.Next() {
			var emailAddress string
			err = qsRows.Scan(&emailAddress)
			if err != nil {
				log.Print("ERROR resolving quicksignups: " + err.Error())
				return nil, false
			}
			recipients = append(recipients, recipientStruct{name: "", email: emailAddress})
		}
		err = qsRows.Err()
		if err != nil {
			log.Print("ERROR resolving quicksignups: " + err.Error())
			return nil, false
		}
	}

	return recipients, true
}

/*
 * Expand unexpanded staged emails into per-recipient rows. Expansion and
 * marking the email row are one transaction; the UNIQUE(email_id, to_email)
 * constraint dedups members who are also quick signups.
 */
func expandStagedEmails() {
	stmt := `
		SELECT
			id,
			notification_type_id,
			trip_id,
			to_id,
			reply_to_id,
			subject,
			body
		FROM email
		WHERE sent_datetime is NULL`
	rows, err := db.Query(stmt)
	if err != nil {
		log.Print("ERROR loading staged emails: " + err.Error())
		return
	}
	defer rows.Close()

	emails := []emailStruct{}
	for rows.Next() {
		email := emailStruct{}
		err = rows.Scan(
			&email.Id,
			&email.NotificationTypeId,
			&email.TripId,
			&email.ToId,
			&email.ReplyToId,
			&email.Subject,
			&email.Body)
		if err != nil {
			log.Print("ERROR loading staged emails: " + err.Error())
			return
		}
		emails = append(emails, email)
	}
	err = rows.Err()
	if err != nil {
		log.Print("ERROR loading staged emails: " + err.Error())
		return
	}

	for _, email := range emails {
		recipients, ok := resolveRecipients(email)
		if !ok {
			continue
		}

		tx, err := db.Begin()
		if err != nil {
			log.Print("ERROR expanding email: " + err.Error())
			continue
		}
		ok = true
		for _, recipient := range recipients {
			if !validEmail(recipient.email) {
				log.Printf("Skipping invalid recipient address [name: %s] [email: %s]", recipient.name, recipient.email)
				continue
			}
			stmt := `
				INSERT OR IGNORE INTO email_recipient (
					email_id,
					to_name,
					to_email)
				VALUES (?, ?, ?)`
			_, err = tx.Exec(stmt, email.Id, recipient.name, recipient.email)
			if err != nil {
				log.Print("ERROR expanding email: " + err.Error())
				tx.Rollback()
				ok = false
				break
			}
		}
		if !ok {
			continue
		}

		stmt := `
			UPDATE email
			SET sent_datetime = datetime('now')
			WHERE id = ?`
		_, err = tx.Exec(stmt, email.Id)
		if err != nil {
			log.Print("ERROR expanding email: " + err.Error())
			tx.Rollback()
			continue
		}
		err = tx.Commit()
		if err != nil {
			log.Print("ERROR expanding email: " + err.Error())
		}
	}
}

/*
 * Send unsent recipient rows, marking each sent only after SES accepts it.
 * Rate limiting stops the pass; the next kick or tick resumes it.
 */
func sendPendingEmails() {
	stmt := `
		SELECT
			email_recipient.id,
			email_recipient.to_name,
			email_recipient.to_email,
			email.reply_to_id,
			email.notification_type_id,
			email.subject,
			email.body
		FROM email_recipient
		JOIN email ON email.id = email_recipient.email_id
		WHERE email_recipient.sent_datetime IS NULL AND email_recipient.failed = 0
		ORDER BY email_recipient.id`
	rows, err := db.Query(stmt)
	if err != nil {
		log.Print("ERROR loading pending emails: " + err.Error())
		return
	}
	defer rows.Close()

	type pendingStruct struct {
		id                 int
		toName             string
		toEmail            string
		replyToId          int
		notificationTypeId string
		subject            string
		body               string
	}
	pending := []pendingStruct{}
	for rows.Next() {
		p := pendingStruct{}
		err = rows.Scan(&p.id, &p.toName, &p.toEmail, &p.replyToId, &p.notificationTypeId, &p.subject, &p.body)
		if err != nil {
			log.Print("ERROR loading pending emails: " + err.Error())
			return
		}
		pending = append(pending, p)
	}
	err = rows.Err()
	if err != nil {
		log.Print("ERROR loading pending emails: " + err.Error())
		return
	}

	if len(pending) == 0 {
		return
	}

	// Always send from System Account
	fromName, fromEmail := dbGetMemberNameEmail(8000000)
	replyToCache := map[int][2]string{}

	sesSession, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		log.Print("ERROR: " + err.Error())
		return
	}
	sesService := ses.New(sesSession)

	for _, p := range pending {
		replyTo, ok := replyToCache[p.replyToId]
		if !ok {
			replyToName, replyToEmail := dbGetMemberNameEmail(p.replyToId)
			replyTo = [2]string{replyToName, replyToEmail}
			replyToCache[p.replyToId] = replyTo
		}

		rawEmail := rawEmailStruct{
			FromName:     fromName,
			FromEmail:    fromEmail,
			ReplyToName:  replyTo[0],
			ReplyToEmail: replyTo[1],
			ToName:       p.toName,
			ToEmail:      p.toEmail,
			Subject:      p.subject,
			Body:         p.body,
			Bulk:         isBulkNotification(p.notificationTypeId),
		}

		_, err = sendEmail(sesService, rawEmail)
		if err == nil {
			stmt := `
				UPDATE email_recipient
				SET sent_datetime = datetime('now')
				WHERE id = ?`
			_, err = db.Exec(stmt, p.id)
			if err != nil {
				log.Print("ERROR marking email sent: " + err.Error())
			}
		} else if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == ses.ErrCodeLimitExceededException {
			// Rate limited, resume on next kick or tick
			return
		} else {
			stmt := `
				UPDATE email_recipient
				SET failed = 1
				WHERE id = ?`
			_, execErr := db.Exec(stmt, p.id)
			if execErr != nil {
				log.Print("ERROR marking email failed: " + execErr.Error())
			}

			// Attempt to send error to system email, otherwise log error
			nameSystem := utils.GetConfig().SmtpFromNameDefault
			emailSystem := utils.GetConfig().SmtpFromEmailDefault
			rawEmail := rawEmailStruct{
				FromName:     nameSystem,
				FromEmail:    emailSystem,
				ReplyToName:  nameSystem,
				ReplyToEmail: emailSystem,
				ToName:       nameSystem,
				ToEmail:      emailSystem,
				Subject:      "Error sending email [name: " + p.toName + "] [email: " + p.toEmail + "]",
				Body:         "Error occured sending email: " + err.Error(),
			}
			_, err = sendEmail(sesService, rawEmail)
			if err != nil {
				log.Print("ERROR: " + err.Error())
			}
		}
	}
}
