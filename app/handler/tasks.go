package handler

import (
  "log"
  "strconv"
)

func DoTasks() {
  /* Remove expired quick signups */
  stmt := `
    DELETE FROM quick_signup
    WHERE datetime(expire_datetime) < datetime('now')`
  _, err := db.Exec(stmt)
  if err != nil {
    log.Fatal(err)
  }

  /* Stage trip reminder email */
  stmt = `
    SELECT id
    FROM trip
    WHERE datetime(create_datetime) < datetime(start_datetime, '+3 days') AND
      cancel = false AND
      publish = true AND
      reminder_sent = false`
  rows, err := db.Query(stmt)
  if err != nil {
    log.Fatal(err)
  }
  defer rows.Close()

  for rows.Next() {
    var tripId int
    err = rows.Scan(&tripId)
    if err != nil {
      log.Fatal(err)
    }

    // Stage email
    err = stageEmailTripReminder(tripId)
    if err != nil {
      log.Fatal(err)
    }

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

  /* Sent any un-sent emails  */
  // TODO SES rate limiting
  // Get emails
  stmt = `
    SELECT
      id,
      notification_type_id,
      trip_id,
      from_id,
      reply_to_id,
      subject,
      body
    FROM email
    WHERE sent = false`
  rows, err = db.Query(stmt)
  if err != nil {
    log.Fatal(err)
  }
  defer rows.Close()

  for rows.Next() {
    var id, tripId, fromId, replyToId int
    var notificationTypeId, subject, body string
    err = rows.Scan(
      &id,
      &notificationTypeId,
      &tripId,
      &fromId,
      &replyToId,
      &subject,
      &body)
    if err != nil {
      log.Fatal(err)
    }

    fromName, fromEmail := dbGetMemberNameEmail(fromId)
    replyToName, replyToEmail := dbGetMemberNameEmail(replyToId)

    stmt := ""
    // Direct Email
    if toId, err := strconv.Atoi(notificationTypeId); err == nil {
      toName, toEmail := dbGetMemberNameEmail(toId)
      sendEmail(fromName, fromEmail, replyToName, replyToEmail, toName, toEmail, subject, body)
    // Trip Alerts
    } else if notificationTypeId == "TRIP_ALERT_ALL" {
      stmt = `
        SELECT id
        FROM trip_signup
        WHERE trip_id = ? AND
          (attending_code = 'ATTEND' OR
           attending_code = 'FORCE' OR
           attending_code = 'WAIT')`
    } else if notificationTypeId == "TRIP_ALERT_ATTEND" {
      stmt = `
        SELECT id
        FROM trip_signup
        WHERE trip_id = ? AND
          (attending_code = 'ATTEND' OR attending_code = 'FORCE')`
    } else if notificationTypeId == "TRIP_ALERT_WAIT" {
      stmt = `
        SELECT id
        FROM trip_signup
        WHERE trip_id = ? AND attending_code = 'WAIT'`
    // All other email types
    } else {
      // Send to all ACTIVE members with notification preference set
      stmt := `
        SELECT id
        FROM member
        WHERE active = true`
      rows, err := db.Query(stmt)
      if err != nil {
        log.Fatal(err)
      }

      for rows.Next() {
        var toId int
        err = rows.Scan(&toId)
        if err != nil {
          log.Fatal(err)
        }

        if !dbCheckMemberWantsNotification(toId, notificationTypeId) {
          continue
        }
        toName, toEmail := dbGetMemberNameEmail(toId)
        sendEmail(fromName, fromEmail, replyToName, replyToEmail, toName, toEmail, subject, body)
      }
    }

    // Send TRIP_ALERT_* emails
    if stmt != "" {
      rows, err = db.Query(stmt, tripId)
      if err != nil {
        log.Fatal(err)
      }

      for rows.Next() {
        var toId int
        err = rows.Scan(&toId)
        if err != nil {
          log.Fatal(err)
        }

        toName, toEmail := dbGetMemberNameEmail(toId)
        sendEmail(fromName, fromEmail, replyToName, replyToEmail, toName, toEmail, subject, body)
      }
    }

    // Mark email as sent
    stmt = `
      UPDATE email
      SET sent = true
      WHERE id = ?`
    _, err = db.Exec(stmt, id)
    if err != nil {
      log.Fatal(err)
    }
  }
}
