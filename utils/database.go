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
      ('GENERAL_ANNOUNCEMENTS', 'Club Updates / News / Events', 'Important Club Announcements'),
      ('GENERAL_MEETINGS', 'Official Meetings', 'Announcements about Club Meetings'),
      ('TRIP_ALERTS', 'Trip Status Alerts', 'Important alerts about your status on trips'),
      ('TRIP_BACKPACKING', 'Backpacking', 'Multi day hikes.'),
      ('TRIP_BIKING', 'Biking', 'Road or mountain biking.'),
      ('TRIP_CAMPING', 'Camping', 'Single overnight trips.'),
      ('TRIP_CLIMBING', 'Climbing', 'Rock climbing or bouldering.'),
      ('TRIP_DAYHIKE', 'Dayhike', 'In and out on the same day.'),
      ('TRIP_LASER_TAG', 'Laser Tag', 'Laser Tag with LCAT'),
      ('TRIP_OFFICIAL_MEETING', 'Official Meeting', 'An official OCVT meeting'),
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
}

func DBMigrate(db *sql.DB) {
	createTables(db)
	insertData(db)
}
