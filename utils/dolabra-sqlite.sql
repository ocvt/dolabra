--
-- File generated with SQLiteStudio v3.2.1 on Sat Jan 11 10:33:58 2020
--
-- Text encoding used: UTF-8
--

-- Table: auth
CREATE TABLE IF NOT EXISTS auth (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      member_id INTEGER REFERENCES member (id) UNIQUE NOT NULL,
      type TEXT NOT NULL COLLATE NOCASE,
      subject TEXT NOT NULL COLLATE NOCASE
    );

-- Table: email
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

-- Table: emergency_contact
CREATE TABLE IF NOT EXISTS emergency_contact (
      id INTEGER UNIQUE NOT NULL PRIMARY KEY AUTOINCREMENT,
      member_id INTEGER REFERENCES member (id) NOT NULL UNIQUE,
      name TEXT NOT NULL COLLATE NOCASE,
      number TEXT NOT NULL COLLATE NOCASE,
      relationship TEXT NOT NULL COLLATE NOCASE
    );

-- Table: equipment
CREATE TABLE IF NOT EXISTS equipment (
      id TEXT PRIMARY KEY NOT NULL UNIQUE,
      create_datetime DATETIME NOT NULL,
      description TEXT UNIQUE NOT NULL COLLATE NOCASE,
      count INTEGER NOT NULL
    );

-- Table: guid
CREATE TABLE IF NOT EXISTS guid (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      code TEXT COLLATE NOCASE NOT NULL,
      member_id INTEGER REFERENCES member (id) NOT NULL,
      trip_id INTEGER REFERENCES trip (id) NOT NULL,
      status TEXT COLLATE NOCASE NOT NULL
    );

-- Table: member
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

-- Table: member_log
CREATE TABLE IF NOT EXISTS member_log (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      member_id INTEGER REFERENCES member (id) UNIQUE NOT NULL,
      log TEXT NOT NULL COLLATE NOCASE
    );

-- Table: news
CREATE TABLE IF NOT EXISTS news (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      member_id INTEGER REFERENCES member (id) NOT NULL,
      create_datetime DATETIME NOT NULL,
      publish BOOLEAN NOT NULL,
      title TEXT NOT NULL COLLATE NOCASE,
      summary TEXT NOT NULL COLLATE NOCASE,
      content TEXT NOT NULL COLLATE NOCASE
    );

-- Table: notification_type
CREATE TABLE IF NOT EXISTS notification_type (
      id TEXT PRIMARY KEY NOT NULL UNIQUE,
      name TEXT UNIQUE NOT NULL COLLATE NOCASE,
      description TEXT NOT NULL COLLATE NOCASE
    );

-- Table: officer
CREATE TABLE IF NOT EXISTS officer (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      member_id INTEGER REFERENCES member (id) UNIQUE NOT NULL,
      create_datetime DATETIME NOT NULL,
      expire_datetime DATETIME NOT NULL,
      position TEXT NOT NULL UNIQUE COLLATE NOCASE,
      security INTEGER NOT NULL
    );

-- Table: payment
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

-- Table: quick_signup
CREATE TABLE IF NOT EXISTS quick_signup (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      create_datetime DATETIME NOT NULL,
      expire_datetime DATETIME NOT NULL,
      email TEXT NOT NULL COLLATE NOCASE UNIQUE
    );

-- Table: store_code
CREATE TABLE IF NOT EXISTS store_code (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      create_datetime DATETIME NOT NULL,
      generated_by_id INTEGER REFERENCES member (id) NOT NULL,
      note TEXT NOT NULL COLLATE NOCASE,
      store_item_id TEXT REFERENCES store_item (id) NOT NULL,
      store_item_count INTEGER NOT NULL,
      amount INTEGER NOT NULL,
      code TEXT NOT NULL,
      completed BOOLEAN NOT NULL,
      redeemed BOOLEAN NOT NULL
    );

-- Table: store_item
CREATE TABLE IF NOT EXISTS store_item (
      id TEXT PRIMARY KEY NOT NULL UNIQUE,
      name TEXT UNIQUE NOT NULL COLLATE NOCASE,
      description TEXT NOT NULL COLLATE NOCASE
    );

-- Table: trip
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

-- Table: trip_approver
CREATE TABLE IF NOT EXISTS trip_approver (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      member_id INTEGER REFERENCES member (id) UNIQUE NOT NULL,
      create_datetime DATETIME NOT NULL,
      expire_datetime DATETIME NOT NULL
    );

-- Table: trip_attending_code
CREATE TABLE IF NOT EXISTS trip_attending_code (
      id TEXT NOT NULL COLLATE NOCASE UNIQUE PRIMARY KEY,
      description TEXT NOT NULL COLLATE NOCASE UNIQUE
    );

-- Table: trip_signup
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
