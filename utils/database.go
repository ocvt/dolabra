package utils

import (
  "database/sql"
  "log"
)

func execHelper(db *sql.DB, sql string) {
  _, err := db.Exec(sql)
  if err != nil {
    log.Fatal("Failed to execute sql: %s", err.Error())
  }
}

func createTables(db *sql.DB) {
  /* Account related tables */
  execHelper(db, `
    CREATE TABLE IF NOT EXISTS member (
      id INTEGER PRIMARY KEY UNIQUE NOT NULL,
      email TEXT NOT NULL COLLATE NOCASE,
      first_name TEXT NOT NULL COLLATE NOCASE,
      last_name TEXT NOT NULL COLLATE NOCASE,
      create_datetime DATETIME NOT NULL,
      cell_number TEXT NOT NULL COLLATE NOCASE,
      gender TEXT NOT NULL COLLATE NOCASE,
      birth_year INTEGER NOT NULL,
      active BOOLEAN NOT NULL,
      medical_cond BOOLEAN NOT NULL,
      medical_cond_desc TEXT NOT NULL COLLATE NOCASE,
      paid_expire_datetime DATETIME NOT NULL,
      notification_preference TEXT NOT NULL
    );
  `)

  execHelper(db, `
    CREATE TABLE IF NOT EXISTS emergency_contact (
      id INTEGER UNIQUE NOT NULL PRIMARY KEY AUTOINCREMENT,
      member_id INTEGER REFERENCES member (id) NOT NULL UNIQUE,
      name TEXT NOT NULL COLLATE NOCASE,
      number TEXT NOT NULL COLLATE NOCASE,
      relationship TEXT NOT NULL COLLATE NOCASE
    );
  `)

  execHelper(db, `
    CREATE TABLE IF NOT EXISTS member_log (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      member_id INTEGER REFERENCES member (id) UNIQUE NOT NULL,
      log TEXT NOT NULL COLLATE NOCASE
    );
  `)

  execHelper(db, `
    CREATE TABLE IF NOT EXISTS officer (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      member_id INTEGER REFERENCES member (id) UNIQUE NOT NULL,
      create_datetime DATETIME NOT NULL,
      expire_datetime DATETIME NOT NULL,
      position TEXT NOT NULL UNIQUE COLLATE NOCASE,
      security INTEGER NOT NULL
    );
  `)

  // Keep track of users who can approve trips
  execHelper(db, `
    CREATE TABLE IF NOT EXISTS trip_approver (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      member_id INTEGER REFERENCES member (id) UNIQUE NOT NULL,
      create_datetime DATETIME NOT NULL,
      expire_datetime DATETIME NOT NULL
    );
  `)

  // Keep track of approved trips
  //  execHelper(db, `
//    CREATE TABLE IF NOT EXISTS trip_approval (
//      approval_seqno INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
//      TODO
//    );
//  `)

  execHelper(db, `
    CREATE TABLE IF NOT EXISTS auth (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      member_id INTEGER REFERENCES member (id) UNIQUE NOT NULL,
      type TEXT NOT NULL COLLATE NOCASE,
      subject TEXT NOT NULL COLLATE NOCASE
    );
  `)

  execHelper(db, `
    CREATE TABLE IF NOT EXISTS quick_signup (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      create_datetime DATETIME NOT NULL,
      expire_datetime DATETIME NOT NULL,
      email TEXT NOT NULL COLLATE NOCASE
    );
  `)

  /* Trip related tables */
  execHelper(db, `
    CREATE TABLE IF NOT EXISTS trip_attending_code (
      id TEXT NOT NULL COLLATE NOCASE UNIQUE PRIMARY KEY,
      description TEXT NOT NULL COLLATE NOCASE UNIQUE
    );
  `)

  execHelper(db, `
    CREATE TABLE IF NOT EXISTS trip (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      create_datetime DATETIME NOT NULL,
      cancel BOOLEAN NOT NULL,
      publish BOOLEAN NOT NULL,
      reminder_sent BOOLEAN NOT NULL,
      member_id INTEGER REFERENCES member (id) NOT NULL,
      members_only BOOLEAN NOT NULL,
      allow_late_signups BOOLEAN NOT NULL,
      driving_required BOOLEAN NOT NULL,
      has_cost BOOLEAN NOT NULL,
      cost_description TEXT NOT NULL COLLATE NOCASE,
      max_people INTEGER NOT NULL,
      name TEXT NOT NULL COLLATE NOCASE,
      notification_type_id TEXT NOT NULL REFERENCES notification_type (id),
      start_datetime DATETIME NOT NULL,
      end_datetime DATETIME NOT NULL,
      summary TEXT NOT NULL COLLATE NOCASE,
      description TEXT NOT NULL COLLATE NOCASE,
      location TEXT NOT NULL COLLATE NOCASE,
      location_directions TEXT NOT NULL COLLATE NOCASE,
      meetup_location TEXT NOT NULL COLLATE NOCASE,
      distance DECIMAL (5, 2) NOT NULL,
      difficulty INTEGER NOT NULL,
      difficulty_description TEXT NOT NULL COLLATE NOCASE,
      instructions TEXT NOT NULL COLLATE NOCASE,
      pets_allowed BOOLEAN NOT NULL,
      pets_description TEXT NOT NULL COLLATE NOCASE
    );
  `)

  //TODO
  // Purpose: include UUID email for anonymously linking to a member & trip id approval pair
//  execHelper(db, `
//    CREATE TABLE IF NOT EXISTS guid (TODO
//    );
//  `)

  execHelper(db, `
    CREATE TABLE IF NOT EXISTS trip_signup (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      trip_id INTEGER REFERENCES trip (id) NOT NULL,
      member_id INTEGER REFERENCES member (id) NOT NULL,
      leader BOOLEAN NOT NULL,
      signup_datetime DATETIME NOT NULL,
      paid_member BOOLEAN NOT NULL,
      attending_code TEXT NOT NULL REFERENCES trip_attending_code (id) COLLATE NOCASE,
      boot_reason TEXT COLLATE NOCASE,
      short_notice BOOLEAN NOT NULL,
      driver BOOLEAN NOT NULL,
      carpool BOOLEAN NOT NULL,
      car_capacity_total INTEGER NOT NULL,
      notes TEXT NOT NULL COLLATE NOCASE,
      pet BOOLEAN NOT NULL,
      attended BOOLEAN NOT NULL
    );
  `)

  /* Notification & Announcement related tables */
  // Note: notification_type_id is a member id for direct alerts
  execHelper(db, `
    CREATE TABLE IF NOT EXISTS email (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      create_datetime DATETIME NOT NULL,
      sent_datetime DATETIME,
      sent BOOLEAN NOT NULL,
      notification_type_id TEXT NOT NULL,
      trip_id INTEGER REFERENCES trip (id) NOT NULL,
      from_id INTEGER REFERENCES member (id) NOT NULL,
      reply_to_id INTEGER REFERENCES member (id) NOT NULL,
      subject TEXT NOT NULL COLLATE NOCASE,
      body TEXT NOT NULL COLLATE NOCASE
    );
  `)

  execHelper(db, `
    CREATE TABLE IF NOT EXISTS news (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      member_id INTEGER REFERENCES member (id) NOT NULL,
      create_datetime DATETIME NOT NULL,
      publish BOOLEAN NOT NULL,
      title TEXT NOT NULL COLLATE NOCASE,
      summary TEXT NOT NULL COLLATE NOCASE,
      content TEXT NOT NULL COLLATE NOCASE
    );
  `)

  // TODO ensure announcement & trips use correct notification type
  execHelper(db, `
    CREATE TABLE IF NOT EXISTS notification_type (
      id TEXT PRIMARY KEY NOT NULL UNIQUE,
      name TEXT UNIQUE NOT NULL COLLATE NOCASE,
      description TEXT NOT NULL COLLATE NOCASE
    );
  `)

  /* Inventory & payment related tables */
  // TODO Note: a payment method+id pair can be valid for multiple items
  execHelper(db, `
    CREATE TABLE IF NOT EXISTS payment (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      create_datetime DATETIME NOT NULL,
      entered_by_id INTEGER REFERENCES member (id) NOT NULL,
      note TEXT NOT NULL COLLATE NOCASE,
      member_id INTEGER REFERENCES member (id) NOT NULL,
      store_item_id TEXT REFERENCES store_item (id) NOT NULL,
      store_item_count INTEGER NOT NULL,
      amount INTEGER NOT NULL,
      payment_method TEXT NOT NULL COLLATE NOCASE,
      payment_id TEXT NOT NULL,
      completed BOOLEAN NOT NULL
    );
  `)

  // TODO Note: a single code can be valid for multiple items
  execHelper(db, `
    CREATE TABLE IF NOT EXISTS store_code (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      create_datetime DATETIME NOT NULL,
      generated_by_id INTEGER REFERENCES member (id) NOT NULL,
      note TEXT NOT NULL COLLATE NOCASE,
      store_item_id TEXT REFERENCES store_item (id) NOT NULL,
      store_item_count INTEGER NOT NULL,
      amount INTEGER NOT NULL,
      code TEXT NOT NULL
      completed BOOLEAN NOT NULL,
    );
  `)

  execHelper(db, `
    CREATE TABLE IF NOT EXISTS store_item (
      id TEXT PRIMARY KEY NOT NULL UNIQUE,
      name TEXT UNIQUE NOT NULL COLLATE NOCASE,
      description TEXT NOT NULL COLLATE NOCASE
    );
  `)

  execHelper(db, `
    CREATE TABLE IF NOT EXISTS equipment (
      id TEXT PRIMARY KEY NOT NULL UNIQUE,
      description TEXT UNIQUE NOT NULL COLLATE NOCASE
    );
  `)
}

func insertData(db *sql.DB) {
  // Populate notification types
  execHelper(db, `
    INSERT OR REPLACE INTO notification_type (id, name, description)
    VALUES
      ('GENERAL_EVENTS', 'General Events', 'Gobblerfest, parade, etc'),
      ('GENERAL_IMPORTANT', 'General Important Items', 'Important Club Announcements'),
      ('GENERAL_ITEMS_FOR_SALE', 'General Items for Sale', 'Items for sale through the club'),
      ('GENERAL_MEETINGS', 'General Meeting', 'Announcements about Club Meetings'),
      ('GENERAL_NEWS', 'General News', 'News from the Club'),
      ('GENERAL_OTHER', 'General Other', 'Miscellaneous Club Announcements'),
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
  createTables(db);
  insertData(db);
}
