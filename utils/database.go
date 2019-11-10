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
    CREATE TABLE IF NOT EXISTS ocvt_users (
      member_id INTEGER PRIMARY KEY UNIQUE NOT NULL,
      email TEXT NOT NULL COLLATE NOCASE,
      name_first TEXT NOT NULL COLLATE NOCASE,
      name_last TEXT NOT NULL COLLATE NOCASE,
      datetime_created DATETIME NOT NULL,
      cell_number TEXT NOT NULL COLLATE NOCASE,
      gender TEXT NOT NULL COLLATE NOCASE,
      birth_year INTEGER NOT NULL,
      active BOOLEAN NOT NULL,
      medical_cond BOOLEAN NOT NULL,
      medical_cond_desc TEXT NOT NULL COLLATE NOCASE,
      paid_expire_datetime DATETIME NOT NULL,
      notification_preference INTEGER NOT NULL
    );
  `)

  execHelper(db, `
    CREATE TABLE IF NOT EXISTS emergency_contacts (
      contact_seqno INTEGER UNIQUE NOT NULL PRIMARY KEY AUTOINCREMENT,
      member_id INTEGER REFERENCES ocvt_users (member_id) NOT NULL UNIQUE,
      contact_name TEXT NOT NULL COLLATE NOCASE,
      contact_number TEXT NOT NULL COLLATE NOCASE,
      contact_relationship TEXT NOT NULL COLLATE NOCASE
    );
  `)

  //TODO
//  execHelper(db, `
//    CREATE TABLE IF NOT EXISTS ocvt_cars(
//      TODO
//    );
//  `)

  execHelper(db, `
    CREATE TABLE IF NOT EXISTS ocvt_officers (
      officer_seqno INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      member_id INTEGER REFERENCES ocvt_users (member_id) UNIQUE NOT NULL,
      create_datetime DATETIME NOT NULL,
      expire_datetime DATETIME NOT NULL,
      officer_position TEXT NOT NULL UNIQUE COLLATE NOCASE,
      officer_security INTEGER NOT NULL
    );
  `)

  execHelper(db, `
    CREATE TABLE IF NOT EXISTS ocvt_approvers (
      approver_seqno INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      member_id INTEGER REFERENCES ocvt_users (member_id) UNIQUE NOT NULL,
      create_date DATETIME NOT NULL,
      expire_date DATETIME NOT NULL
    );
  `)

  execHelper(db, `
    CREATE TABLE IF NOT EXISTS ocvt_auth (
      auth_seqno INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      member_id INTEGER REFERENCES ocvt_users (member_id) UNIQUE NOT NULL,
      auth_type TEXT NOT NULL COLLATE NOCASE,
      auth_sub TEXT NOT NULL COLLATE NOCASE
     );
  `)

  execHelper(db, `
    CREATE TABLE IF NOT EXISTS quick_signup (
      signup_seqno INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      create_date DATETIME NOT NULL,
      email TEXT NOT NULL COLLATE NOCASE
    );
  `)

  /* Trip related tables */
  execHelper(db, `
    CREATE TABLE IF NOT EXISTS trip_types (
      trip_type_id TEXT PRIMARY KEY NOT NULL UNIQUE,
      trip_type_name TEXT UNIQUE NOT NULL COLLATE NOCASE,
      trip_type_description TEXT NOT NULL COLLATE NOCASE UNIQUE
    );
  `)

  execHelper(db, `
    CREATE TABLE IF NOT EXISTS trip_attending_codes (
      code TEXT NOT NULL COLLATE NOCASE UNIQUE PRIMARY KEY,
      description TEXT NOT NULL COLLATE NOCASE UNIQUE
    );
  `)

  execHelper(db, `
    CREATE TABLE IF NOT EXISTS ocvt_trips (
      trip_id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      member_id INTEGER REFERENCES ocvt_users (member_id) NOT NULL,
      trip_publish BOOLEAN NOT NULL,
      trip_name TEXT NOT NULL COLLATE NOCASE,
      trip_type_id TEXT NOT NULL REFERENCES trip_types (trip_type_id),
      trip_create_datetime DATETIME NOT NULL,
      trip_members_only BOOLEAN NOT NULL,
      trip_driver_required BOOLEAN NOT NULL,
      trip_max_people INTEGER NOT NULL,
      trip_age_min INTEGER NOT NULL,
      trip_age_max INTEGER NOT NULL,
      trip_summary TEXT NOT NULL COLLATE NOCASE,
      trip_description TEXT NOT NULL COLLATE NOCASE,
      trip_location TEXT NOT NULL COLLATE NOCASE,
      trip_directions TEXT NOT NULL COLLATE NOCASE,
      trip_end_datetime DATETIME NOT NULL,
      trip_start_datetime DATETIME NOT NULL,
      trip_distance DECIMAL (5, 2) NOT NULL,
      trip_difficulty INTEGER NOT NULL,
      trip_difficulty_description TEXT NOT NULL COLLATE NOCASE,
      trip_has_cost BOOLEAN NOT NULL,
      trip_cost_description TEXT NOT NULL COLLATE NOCASE,
      trip_gear TEXT NOT NULL COLLATE NOCASE,
      trip_pets_allowed BOOLEAN NOT NULL,
      trip_pets_description TEXT NOT NULL COLLATE NOCASE
    );
  `)

  //TODO
  // Purpose: include UUID email for anonymously linking to a member & trip id approval pair
//  execHelper(db, `
//    CREATE TABLE IF NOT EXISTS ocvt_guids (TODO
//    );
//  `)

  execHelper(db, `
    CREATE TABLE IF NOT EXISTS ocvt_trip_people (
      trip_people_seqno INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      trip_id INTEGER REFERENCES ocvt_trips (trip_id) NOT NULL,
      member_id INTEGER REFERENCES ocvt_users (member_id) NOT NULL,
      admin BOOLEAN NOT NULL,
      signup_datetime DATETIME NOT NULL,
      paid_member BOOLEAN NOT NULL,
      attending_code TEXT NOT NULL REFERENCES trip_attending_codes (code) COLLATE NOCASE,
      trip_paid BOOLEAN NOT NULL,
      driver BOOLEAN NOT NULL,
      car_capacity_total INTEGER,
      notes TEXT NOT NULL COLLATE NOCASE,
      boot_reason TEXT COLLATE NOCASE,
      pet BOOLEAN NOT NULL,
      short_notice BOOLEAN NOT NULL,
      attended BOOLEAN
    );
  `)

  //TODO Finalize table name
  // Purpose: Keep track of trips that have been approved and who approved them
//  execHelper(db, `
//    CREATE TABLE IF NOT EXISTS ocvt_trip_approvals (
//      approval_seqno INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
//      TODO
//    );
//  `)

  /* Notification related tables */
  execHelper(db, `
    CREATE TABLE IF NOT EXISTS ocvt_emails (
      email_seqno INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      email_type TEXT NOT NULL COLLATE NOCASE,
      email_subtype TEXT NOT NULL COLLATE NOCASE,
      email_subject TEXT NOT NULL COLLATE NOCASE,
      email_content TEXT NOT NULL COLLATE NOCASE,
      member_id INTEGER REFERENCES ocvt_users (member_id) NOT NULL,
      email_replyto TEXT NOT NULL COLLATE NOCASE,
      email_returnpath TEXT NOT NULL COLLATE NOCASE,
      email_sent BOOLEAN NOT NULL,
      email_create_datetime DATETIME NOT NULL
    );
  `)

  execHelper(db, `
    CREATE TABLE IF NOT EXISTS ocvt_news (
      news_seqno INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      member_id INTEGER REFERENCES ocvt_users (member_id) NOT NULL,
      create_datetime DATETIME NOT NULL,
      news_title TEXT NOT NULL COLLATE NOCASE,
      news_content TEXT NOT NULL COLLATE NOCASE
    );
  `)

  /* Inventory related tables */
  //TODO
  // Contains items like shirts
//  execHelper(db, `
//    CREATE TABLE IF NOT EXISTS ocvt_member_items (TODO
//    );
//  `)

  //TODO
  //  Equipment inventory
//  execHelper(db, `
//    CREATE TABLE IF NOT EXISTS ocvt_inventory (
//      inventory_id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
//      TODO
//    );
//  `)
}

func insertData(db *sql.DB) {
  // Populate trip types
  execHelper(db, `
    INSERT OR IGNORE INTO trip_types (trip_type_id, trip_type_name, trip_type_description)
    VALUES
      ('TR01', 'Dayhike', 'In and out on the same day.'),
      ('TR02', 'Worktrip', 'Trail work or other maintenance.'),
      ('TR03', 'Backpacking', 'Multi day hikes.'),
      ('TR04', 'Camping', 'Single overnight trips. 	'),
      ('TR05', 'Official Meeting', 'An official OCVT meeting'),
      ('TR06', 'Social', 'Strictly social, potluck, movie nights, games or other casual gatherings'),
      ('TR07', 'Rafting / Canoeing / Kayaking', 'Rafting / Canoeing / Kayaking'),
      ('TR08', 'Water / Other', 'Swimming, tubing anything else in the water.'),
      ('TR09', 'Biking', 'Road or mountain biking.'),
      ('TR10', 'Team Sports / Misc.', 'Football, basketball ultimate Frisbee etc.'),
      ('TR11', 'Climbing', 'Rock climbing or bouldering.'),
      ('TR12', 'Skiing / Snowboarding', 'Skiing / Snowboarding'),
      ('TR13', 'Snow / Other', 'Sledding snowshoeing etc'),
      ('TR14', 'Road Trip', 'Just getting out and about, Ex a trip to Busch Gardens or DC etc'),
      ('TR15', 'Special Event', 'A special event.'),
      ('TR16', 'Other', 'Anything else not covered. '),
      ('TR17', 'Laser Tag', 'Laser Tag with LCAT');
  `)

  // Populate attending codes
  execHelper(db, `
    INSERT OR IGNORE INTO trip_attending_codes (code, description)
    VALUES
      ('ATTEN', 'User is attending'),
      ('BOOT', 'User has been manually booted'),
      ('WAIT', 'User is on waiting list'),
      ('CANCL', 'User has chosen to cancel '),
      ('FORCE', 'User is force added'),
      ('TBD', 'User status is to be determined');
  `)
}

func DBMigrate(db *sql.DB) {
  createTables(db);
  insertData(db);
}
