package utils

import (
	"database/sql"
	"io/ioutil"
	"log"
	"strings"
)

func execHelper(db *sql.DB, sql string) {
	_, err := db.Exec(sql)
	if err != nil {
		log.Fatal(err)
	}
}

/* Create table based on sql file */
// TODO member_log table
func createTables(db *sql.DB) {
	file, err := ioutil.ReadFile("utils/dolabra-sqlite.sql")
	if err != nil {
		log.Fatal(err)
	}

	requests := strings.Split(string(file), ";")

	for _, request := range requests {
		_, err = db.Exec(request)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func insertData(db *sql.DB) {
	// Populate notification types
	execHelper(db, `
		INSERT OR REPLACE INTO notification_type (id, name, description)
		VALUES
			('DIRECT', 'Direct Message', 'Email sent directly to specific members'),
			('GENERAL_ANNOUNCEMENTS', 'Club Meetings / News / Events', 'Important Club Announcements'),
			('TRIP_APPROVAL', 'Trip Approval Alert', 'Alert asking member (probably officer) to approve a trip'),
			('TRIP_ALERT_ATTEND', 'Signup status changed to attending', 'Signup status changed to attending'),
			('TRIP_ALERT_BOOT', 'Signup status changed to booted', 'Signup status changed to booted'),
			('TRIP_ALERT_CANCEL', 'Signup status changed to cancel', 'Signup status changed to cancel'),
			('TRIP_ALERT_FORCE', 'Signup status changed to force-added', 'Signup status changed to force-added'),
			('TRIP_ALERT_LEADER', 'Signup status changed to trip leader', 'Signup status changed to trip leader'),
			('TRIP_ALERT_WAIT', 'Signup status changed to waitlisted', 'Signup status changed to waitlisted'),
			('TRIP_MESSAGE_DIRECT', 'Direct message related to trip', 'Direct message related to trip'),
			('TRIP_MESSAGE_NOTIFY', 'Trip message for force-added, attendees, and waitlist', 'Trip message for force-added, attendees, and waitlist'),
			('TRIP_MESSAGE_ATTEND', 'Trip message to force-added and attendees', 'Trip message to force-added and attendees'),
			('TRIP_MESSAGE_WAIT', 'Trip message to waitlist', 'Trip message to waitlist'),
			('TRIP_BACKPACKING', 'Backpacking', 'Multi day hikes.'),
			('TRIP_BIKING', 'Biking', 'Road or mountain biking.'),
			('TRIP_CAMPING', 'Camping', 'Single overnight trips.'),
			('TRIP_CLIMBING', 'Climbing', 'Rock climbing or bouldering.'),
			('TRIP_DAYHIKE', 'Dayhike', 'In and out on the same day.'),
			('TRIP_LASER_TAG', 'Laser Tag', 'Laser Tag with LCAT'),
			('TRIP_OFFICIAL_MEETING', 'Official Meeting', 'An official meeting'),
			('TRIP_OTHER', 'Other', 'Anything else not covered. '),
			('TRIP_RAFTING_CANOEING_KAYAKING', 'Rafting / Canoeing / Kayaking', 'Rafting / Canoeing / Kayaking'),
			('TRIP_ROAD_TRIP', 'Road Trip', 'Just getting out and about, Ex a trip to Busch Gardens or DC etc'),
			('TRIP_SKIING_SNOWBOARDING', 'Skiing / Snowboarding', 'Skiing / Snowboarding'),
			('TRIP_SNOW_OTHER', 'Snow / Other', 'Sledding snowshoeing etc'),
			('TRIP_SOCIAL', 'Social', 'Strictly social, potluck, movie nights, games or other casual gatherings'),
			('TRIP_SPECIAL_EVENT', 'Special Event', 'A special event.'),
			('TRIP_TEAM_SPORTS_MISC', 'Team Sports / Misc.', 'Football, basketball ultimate Frisbee etc.'),
			('TRIP_WATER_OTHER', 'Water / Other', 'Swimming, tubing anything else in the water.'),
			('TRIP_WORK_TRIP', 'Worktrip', 'Trail work or other maintenance.')
	`)

	// Populate store items
	execHelper(db, `
		INSERT OR REPLACE INTO store_item (id, name, description)
		VALUES
			('CUSTOM', 'Custom Amount', ''),
			('MEMBERSHIP', '1 Year of membership', ''),
			('SHIRT', '1 Shirt', 'Size determined at pickup')
	`)

	// Populate attending codes
	execHelper(db, `
		INSERT OR REPLACE INTO trip_attending_code (id, description)
		VALUES
			('ATTEND', 'User is attending'),
			('BOOT', 'User has been manually booted'),
			('CANCEL', 'User has chosen to cancel '),
			('FORCE', 'User is force added'),
			('WAIT', 'User is on waiting list')
	`)

	// Populate allowed pronouns. (We don't care about the actual pronouns, this is only for simplicity in programming)
	execHelper(db, `
		INSERT OR REPLACE INTO pronouns (id)
		VALUES
			('he/him'),
			('she/her'),
			('they/them'),
			('prefer not to say')
	`)

}

/*
 * Strip surrounding whitespace from stored addresses. A trailing space still
 * matches the LIKE filter in cleanInvalidEmails but is rejected at send time,
 * so the member stays active while silently receiving nothing.
 */
func trimEmails(db *sql.DB) {
	rows, err := db.Query(`
		SELECT id, name, email
		FROM member
		WHERE email != trim(email)`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, email string
		err = rows.Scan(&id, &name, &email)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Trimming whitespace from member email [id: %d] [name: %s] [email: %q]", id, name, email)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	execHelper(db, `
		UPDATE member
		SET email = trim(email)
		WHERE email != trim(email)`)

	// quick_signup.email is UNIQUE, so trimming can collide with an existing
	// row; OR REPLACE drops the duplicate instead of aborting startup
	execHelper(db, `
		UPDATE OR REPLACE quick_signup
		SET email = trim(email)
		WHERE email != trim(email)`)
}

/* Clean up rows with undeliverable email addresses */
// Members are deactivated (not deleted) so they can log in, fix their
// email, and reactivate; quick signups are just an email list so bad
// rows are deleted
func cleanInvalidEmails(db *sql.DB) {
	rows, err := db.Query(`
		SELECT id, name, email
		FROM member
		WHERE active = true AND email NOT LIKE '%_@_%'`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, email string
		err = rows.Scan(&id, &name, &email)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Deactivating member with invalid email [id: %d] [name: %s] [email: %s]", id, name, email)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	execHelper(db, `
		UPDATE member
		SET active = false
		WHERE active = true AND email NOT LIKE '%_@_%'`)

	execHelper(db, `
		DELETE FROM quick_signup
		WHERE email NOT LIKE '%_@_%'`)
}

/*
 * A completed MEMBERSHIP row should mean the member got the years. Webtools has
 * a "completed" tick that only flips the flag, and during the Aug-Sep 2026
 * webhook outage every stranded row was ticked by hand, which would hide them
 * from the reconcile sweep. Re-open rows from that window whose member's
 * expiry shows the payment never took effect; the sweep settles them against
 * Stripe. A granted payment always leaves the expiry at or past
 * create_datetime + years, so those are left alone.
 */
func resetUnappliedStripePayments(db *sql.DB) {
	rows, err := db.Query(`
		SELECT payment.id, payment.member_id, member.name, payment.create_datetime
		FROM payment
		INNER JOIN member ON member.id = payment.member_id
		WHERE payment.payment_method = 'STRIPE'
			AND payment.store_item_id = 'MEMBERSHIP'
			AND payment.completed = true
			AND datetime(payment.create_datetime) > datetime('2026-08-06')
			AND datetime(member.paid_expire_datetime) <
				datetime(payment.create_datetime, '+' || payment.store_item_count || ' years')`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, memberId int
		var name, created string
		err = rows.Scan(&id, &memberId, &name, &created)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Re-opening Stripe payment never applied to membership [payment: %d] [member: %d %s] [created: %s]", id, memberId, name, created)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	execHelper(db, `
		UPDATE payment
		SET completed = false
		WHERE payment_method = 'STRIPE'
			AND store_item_id = 'MEMBERSHIP'
			AND completed = true
			AND datetime(create_datetime) > datetime('2026-08-06')
			AND datetime((SELECT paid_expire_datetime FROM member WHERE member.id = payment.member_id)) <
				datetime(create_datetime, '+' || store_item_count || ' years')`)
}

func DBMigrate(db *sql.DB) {
	createTables(db)
	insertData(db)
	trimEmails(db)
	cleanInvalidEmails(db)
	resetUnappliedStripePayments(db)
}
