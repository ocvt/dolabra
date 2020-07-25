package handler

import (
	"container/list"
	"database/sql"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

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

	/* Load staged emails into queue to send */
	stmt = `
		SELECT *
		FROM email
		WHERE sent = false`
	rows, err = db.Query(stmt)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	emails := list.New()
	for rows.Next() {
		email := emailStruct{}
		err = rows.Scan(
			&email.Id,
			&email.CreateDatetime,
			&email.SentDatetime,
			&email.Sent,
			&email.NotificationTypeId,
			&email.TripId,
			&email.ToId,
			&email.ReplyToId,
			&email.Subject,
			&email.Body)
		if err != nil {
			log.Fatal(err)
		}
		emails.PushBack(email)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	for e := emails.Front(); e != nil; e = e.Next() {
		email := e.Value.(emailStruct)
		// Always send from System Account
		fromName, fromEmail := dbGetMemberNameEmail(0)
		replyToName, replyToEmail := dbGetMemberNameEmail(email.ReplyToId)

		var rows *sql.Rows
		var err error
		if email.NotificationTypeId == "TRIP_APPROVAL" {
			stmt := `
				SELECT member_id
				FROM trip_approver
				WHERE member_id = ?`
			rows, err = db.Query(stmt, email.ToId)
		} else if email.NotificationTypeId == "TRIP_MESSAGE_DIRECT" {
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
			// Send to all ACTIVE members with notification preference set
			stmt := `
				SELECT id
				FROM member
				WHERE active = true`
			rows, err = db.Query(stmt)
		}

		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		memberIds := list.New()
		for rows.Next() {
			var memberId int
			err = rows.Scan(&memberId)
			if err != nil {
				log.Fatal(err)
			}
			memberIds.PushBack(memberId)
		}

		for m := memberIds.Front(); m != nil; m = m.Next() {
			memberId := m.Value.(int)
			if email.NotificationTypeId != "TRIP_APPROVAL" &&
				!strings.HasPrefix(email.NotificationTypeId, "TRIP_ALERT") &&
				!strings.HasPrefix(email.NotificationTypeId, "TRIP_MESSAGE") &&
				!dbCheckMemberWantsNotification(memberId, email.NotificationTypeId) {
				continue
			}
			toName, toEmail := dbGetMemberNameEmail(memberId)

			// Put into queue
			rawEmail := rawEmailStruct{
				FromName:     fromName,
				FromEmail:    fromEmail,
				ReplyToEmail: replyToEmail,
				ReplyToName:  replyToName,
				ToName:       toName,
				ToEmail:      toEmail,
				Subject:      email.Subject,
				Body:         email.Body,
			}

			emailQueue.PushBack(rawEmail)
		}
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}

		// Mark email as sent
		stmt = `
			UPDATE email
			SET
				sent_datetime = datetime('now'),
				sent = true
			WHERE id = ?`
		_, err = db.Exec(stmt, email.Id)
		if err != nil {
			log.Fatal(err)
		}
	}
	/***************************/

	/* Send emails from queue */
	sesSession, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	sesService := ses.New(sesSession)

	var next *list.Element
	for e := emailQueue.Front(); e != nil; e = next {
		next = e.Next()
		email := e.Value.(rawEmailStruct)
		_, err = sendEmail(sesService, email)
		if err == nil {
			emailQueue.Remove(e)
		} else if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == ses.ErrCodeMessageRejected {
			log.Print(ses.ErrCodeMessageRejected, awsErr.Error())
			break
		} else {
			log.Print(err.Error())
		}
	}
	/**************************/
}
