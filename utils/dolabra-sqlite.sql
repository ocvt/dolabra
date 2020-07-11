--
-- File generated with SQLiteStudio v3.2.1 on Sat Jan 11 10:33:58 2020
--
-- Text encoding used: UTF-8
--

-- Table: auth
CREATE TABLE IF NOT EXISTS auth (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      member_id INTEGER NOT NULL,
      sub TEXT NOT NULL,
      idp TEXT NOT NULL,
      idp_sub TEXT NOT NULL,
      FOREIGN KEY(member_id) REFERENCES member(id)
);

-- Table: email
CREATE TABLE IF NOT EXISTS email (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      create_datetime DATETIME NOT NULL,
      sent_datetime DATETIME NOT NULL,
      sent BOOLEAN NOT NULL,
      notification_type_id TEXT NOT NULL,
      trip_id INTEGER NOT NULL,
      to_id INTEGER  NOT NULL,
      reply_to_id INTEGER  NOT NULL,
      subject TEXT NOT NULL,
      body TEXT NOT NULL,
      FOREIGN KEY(notification_type_id) REFERENCES notification_type(id),
      FOREIGN KEY(trip_id) REFERENCES trip(id),
      FOREIGN KEY(to_id) REFERENCES member(id),
      FOREIGN KEY(reply_to_id) REFERENCES member(id)
);

-- Table: emergency_contact
CREATE TABLE IF NOT EXISTS emergency_contact (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      member_id INTEGER UNIQUE NOT NULL,
      name TEXT NOT NULL COLLATE NOCASE,
      number TEXT NOT NULL,
      relationship TEXT NOT NULL,
      FOREIGN KEY(member_id) REFERENCES member(id)
);

-- Table: equipment
CREATE TABLE IF NOT EXISTS equipment (
      id TEXT PRIMARY KEY UNIQUE NOT NULL,
      create_datetime DATETIME NOT NULL,
      description TEXT UNIQUE NOT NULL,
      count INTEGER NOT NULL
);

-- Table: guid
CREATE TABLE IF NOT EXISTS guid (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      code TEXT COLLATE NOCASE NOT NULL,
      member_id INTEGER NOT NULL,
      trip_id INTEGER NOT NULL,
      status TEXT NOT NULL,
      FOREIGN KEY(member_id) REFERENCES member(id),
      FOREIGN KEY(trip_id) REFERENCES trip(id)
);

-- Table: member
CREATE TABLE IF NOT EXISTS member (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      email TEXT NOT NULL,
      first_name TEXT NOT NULL COLLATE NOCASE,
      last_name TEXT NOT NULL COLLATE NOCASE,
      create_datetime DATETIME NOT NULL,
      cell_number TEXT NOT NULL,
      gender TEXT NOT NULL,
      birth_year INTEGER NOT NULL,
      active BOOLEAN NOT NULL,
      medical_cond BOOLEAN NOT NULL,
      medical_cond_desc TEXT NOT NULL,
      paid_expire_datetime DATETIME NOT NULL,
      notification_preference TEXT NOT NULL
);

-- Table: member_log
CREATE TABLE IF NOT EXISTS member_log (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      member_id INTEGER NOT NULL,
      log TEXT NOT NULL,
      FOREIGN KEY(member_id) REFERENCES member(id)
);

-- Table: news
CREATE TABLE IF NOT EXISTS news (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      member_id INTEGER NOT NULL,
      create_datetime DATETIME NOT NULL,
      publish BOOLEAN NOT NULL,
      title TEXT NOT NULL,
      summary TEXT NOT NULL,
      content TEXT NOT NULL,
      FOREIGN KEY(member_id) REFERENCES member(id)
);

-- Table: notification_type
CREATE TABLE IF NOT EXISTS notification_type (
      id TEXT PRIMARY KEY UNIQUE NOT NULL,
      name TEXT UNIQUE NOT NULL,
      description TEXT NOT NULL
);

-- Table: officer
CREATE TABLE IF NOT EXISTS officer (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      member_id INTEGER UNIQUE NOT NULL,
      create_datetime DATETIME NOT NULL,
      expire_datetime DATETIME NOT NULL,
      position TEXT UNIQUE NOT NULL,
      security INTEGER NOT NULL,
      FOREIGN KEY(member_id) REFERENCES member(id)
);

-- Table: payment
CREATE TABLE IF NOT EXISTS payment (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      create_datetime DATETIME NOT NULL,
      entered_by_id INTEGER NOT NULL,
      note TEXT NOT NULL,
      member_id INTEGER NOT NULL,
      store_item_id TEXT NOT NULL,
      store_item_count INTEGER NOT NULL,
      amount INTEGER NOT NULL,
      payment_method TEXT NOT NULL,
      payment_id TEXT NOT NULL,
      completed BOOLEAN NOT NULL,
      FOREIGN KEY(member_id) REFERENCES member(id),
      FOREIGN KEY(store_item_id) REFERENCES store_item(id)
);

-- Table: quick_signup
CREATE TABLE IF NOT EXISTS quick_signup (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      create_datetime DATETIME NOT NULL,
      expire_datetime DATETIME NOT NULL,
      email TEXT UNIQUE NOT NULL
);

-- Table: store_code
CREATE TABLE IF NOT EXISTS store_code (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      create_datetime DATETIME NOT NULL,
      generated_by_id INTEGER NOT NULL,
      note TEXT NOT NULL,
      store_item_id TEXT NOT NULL,
      store_item_count INTEGER NOT NULL,
      amount INTEGER NOT NULL,
      code TEXT NOT NULL,
      completed BOOLEAN NOT NULL,
      redeemed BOOLEAN NOT NULL,
      FOREIGN KEY(generated_by_id) REFERENCES member(id),
      FOREIGN KEY(store_item_id) REFERENCES store_item(id)
);

-- Table: store_item
CREATE TABLE IF NOT EXISTS store_item (
      id TEXT PRIMARY KEY UNIQUE NOT NULL,
      name TEXT UNIQUE NOT NULL,
      description TEXT NOT NULL
    );

-- Table: trip
CREATE TABLE IF NOT EXISTS trip (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      create_datetime DATETIME NOT NULL,
      cancel BOOLEAN NOT NULL,
      publish BOOLEAN NOT NULL,
      reminder_sent BOOLEAN NOT NULL,
      member_id INTEGER NOT NULL,
      members_only BOOLEAN NOT NULL,
      allow_late_signups BOOLEAN NOT NULL,
      driving_required BOOLEAN NOT NULL,
      has_cost BOOLEAN NOT NULL,
      cost_description TEXT NOT NULL,
      max_people INTEGER NOT NULL,
      name TEXT NOT NULL,
      notification_type_id TEXT NOT NULL,
      start_datetime DATETIME NOT NULL,
      end_datetime DATETIME NOT NULL,
      summary TEXT NOT NULL,
      description TEXT NOT NULL,
      location TEXT NOT NULL,
      location_directions TEXT NOT NULL,
      meetup_location TEXT NOT NULL,
      distance DECIMAL (5, 2) NOT NULL,
      difficulty INTEGER NOT NULL,
      difficulty_description TEXT NOT NULL,
      instructions TEXT NOT NULL,
      pets_allowed BOOLEAN NOT NULL,
      pets_description TEXT NOT NULL,
      FOREIGN KEY(member_id) REFERENCES member(id),
      FOREIGN KEY(notification_type_id) REFERENCES notification_type(id)
);

-- Table: trip_approver
CREATE TABLE IF NOT EXISTS trip_approver (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      member_id INTEGER UNIQUE NOT NULL,
      create_datetime DATETIME NOT NULL,
      expire_datetime DATETIME NOT NULL,
      FOREIGN KEY(member_id) REFERENCES member(id)
);

-- Table: trip_attending_code
CREATE TABLE IF NOT EXISTS trip_attending_code (
      id TEXT PRIMARY KEY UNIQUE NOT NULL,
      description TEXT UNIQUE NOT NULL
);

-- Table: trip_signup
CREATE TABLE IF NOT EXISTS trip_signup (
      id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL,
      trip_id INTEGER NOT NULL,
      member_id INTEGER NOT NULL,
      leader BOOLEAN NOT NULL,
      signup_datetime DATETIME NOT NULL,
      paid_member BOOLEAN NOT NULL,
      attending_code TEXT NOT NULL,
      boot_reason TEXT,
      short_notice BOOLEAN NOT NULL,
      driver BOOLEAN NOT NULL,
      carpool BOOLEAN NOT NULL,
      car_capacity_total INTEGER NOT NULL,
      notes TEXT NOT NULL,
      pet BOOLEAN NOT NULL,
      attended BOOLEAN NOT NULL,
      FOREIGN KEY(member_id) REFERENCES member(id),
      FOREIGN KEY(trip_id) REFERENCES trip(id),
      FOREIGN KEY(attending_code) REFERENCES trip_attending_code(id)
);
