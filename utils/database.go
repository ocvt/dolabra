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
      notification_preference INTEGER NOT NULL
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
      email TEXT NOT NULL COLLATE NOCASE
    );
  `)

  /* Trip related tables */
  execHelper(db, `
    CREATE TABLE IF NOT EXISTS trip_type (
      id TEXT PRIMARY KEY NOT NULL UNIQUE,
      name TEXT UNIQUE NOT NULL COLLATE NOCASE,
      description TEXT NOT NULL COLLATE NOCASE
    );
  `)

  execHelper(db, `
    CREATE TABLE IF NOT EXISTS trip_attending_code (
      id TEXT NOT NULL COLLATE NOCASE UNIQUE PRIMARY KEY,
      description TEXT NOT NULL COLLATE NOCASE UNIQUE
    );
  `)

  //TODO don't append trip_ to every field??
  execHelper(db, `
    CREATE TABLE IF NOT EXISTS trip (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      create_datetime DATETIME NOT NULL,
      cancel BOOLEAN NOT NULL,
      publish BOOLEAN NOT NULL,
      member_id INTEGER REFERENCES member (id) NOT NULL,
      members_only BOOLEAN NOT NULL,
      allow_late_signups BOOLEAN NOT NULL,
      driving_required BOOLEAN NOT NULL,
      has_cost BOOLEAN NOT NULL,
      cost_description TEXT NOT NULL COLLATE NOCASE,
      max_people INTEGER NOT NULL,
      name TEXT NOT NULL COLLATE NOCASE,
      trip_type_id TEXT NOT NULL REFERENCES trip_type (id),
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
      car_capacity_total INTEGER,
      notes TEXT NOT NULL COLLATE NOCASE,
      pet BOOLEAN NOT NULL,
      attended BOOLEAN
    );
  `)

  /* Notification related tables */
  // clarification about type & subtype TODO
  execHelper(db, `
    CREATE TABLE IF NOT EXISTS email (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      type TEXT NOT NULL COLLATE NOCASE,
      subtype TEXT NOT NULL COLLATE NOCASE,
      subject TEXT NOT NULL COLLATE NOCASE,
      content TEXT NOT NULL COLLATE NOCASE,
      member_id INTEGER REFERENCES member (id) NOT NULL,
      reply_to TEXT NOT NULL COLLATE NOCASE,
      return_path TEXT NOT NULL COLLATE NOCASE,
      sent BOOLEAN NOT NULL,
      create_datetime DATETIME NOT NULL
    );
  `)

  execHelper(db, `
    CREATE TABLE IF NOT EXISTS news (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      member_id INTEGER REFERENCES member (id) NOT NULL,
      create_datetime DATETIME NOT NULL,
      title TEXT NOT NULL COLLATE NOCASE,
      content TEXT NOT NULL COLLATE NOCASE
    );
  `)

  /* Inventory related tables */
  //TODO
  // Contains items like shirts
//  execHelper(db, `
//    CREATE TABLE IF NOT EXISTS store_item (TODO
//    );
//  `)

  //TODO
  //  Equipment inventory
//  execHelper(db, `
//    CREATE TABLE IF NOT EXISTS equipment (
//      inventory_id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
//      TODO
//    );
//  `)
}

func insertData(db *sql.DB) {
  // Populate trip types
  execHelper(db, `
    INSERT OR IGNORE INTO trip_type (id, name, description)
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
    INSERT OR IGNORE INTO trip_attending_code (id, description)
    VALUES
      ('ATTEND', 'User is attending'),
      ('BOOT', 'User has been manually booted'),
      ('WAIT', 'User is on waiting list'),
      ('CANCEL', 'User has chosen to cancel '),
      ('FORCE', 'User is force added')
  `)
}

func DBMigrate(db *sql.DB) {
  createTables(db);
  insertData(db);
}
