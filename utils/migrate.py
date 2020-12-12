#!/usr/bin/env python3

# This script processes MySQL data from the old site into sqlite for the new site
# What this does NOT do:
# - Migrate trip approvers (should be done manually)
# - Migrate officers (should be done manually)
# Assumptions:
# - local mysql db is running with old site data loaded
# - local sqlite db exists titled 'dolabra-sqlite.sqlite3'
# Migrations:
# - news
#   - ocvt_news.create_date -> news.create_datetime
#   - ocvt_news.title       -> news.title, news.summary
#   - ocvt_news.content     -> news.content
#   - SET news.publish TRUE
# - users -> 'oldsite_member'
#   - ocvt_users.email                        -> oldsite_member.email
#   - ocvt_users.name_first                   -> oldsite_member.first_name
#   - ocvt_users.name_last                    -> oldsite_member.last_name
#   - ocvt_users.date_created                 -> oldsite_member.create_datetime
#   - ocvt_users.cell_number                  -> oldsite_member.cell_number
#   - ocvt_users.gender                       -> oldsite_member.gender
#   - ocvt_users.birth_year                   -> oldsite_member.birth_year
#   - ocvt_users.active                       -> oldsite_member.active
#   - ocvt_users.medical_cond                 -> oldsite_member.medical_cond
#   - ocvt_users.medical_cond_desc            -> oldsite_member.medical_cond_desc
#   - ocvt_members.expiration_date            -> oldsite_member.paid_expire_datetime
#   - emergency_contacts.contact_name         -> oldsite_member.ec_name
#   - emergency_contacts.contact_number       -> oldsite_member.ec_number
#   - emergency_contacts.contact_relationship -> oldsite_member.ec_relationship
# - notification preferences
#   - for each row in notification_settings:
#     - if member_id in oldsite_member:
#       - set preference for notification preference to false
# - orders -> 'oldsite_payment' (online orders), 'oldsite_payment_manual' (manual orders)
#   - ocvt_orders + order_items + order_items(ONLINE ORDERS)
#   - ocvt_manual_payments (MANUAL ORDERS)
#     - ASK DOUG
#     - Create webtools page to view all incomplete manual orders
#     - Email all incomplete people >= 2019
#     - Add note on myocvt page to email webmaster if any issues

import datetime
import sqlite3
from MySQLdb import _mysql

mdb = _mysql.connect('127.0.0.1', 'root', 'ocvt', 'ocvt')
scon = sqlite3.connect('dolabra-sqlite.sqlite3')
sc = scon.cursor()

## NEWS
mdb.query("SELECT * FROM ocvt_news")
rows = mdb.use_result()

while True:
    row = rows.fetch_row()
    if len(row) == 0:
        break

    # parse & decode
    create_datetime = row[0][1].decode('utf-8')
    title           = row[0][2].decode('utf-8')
    summary         = row[0][2].decode('utf-8')
    content         = row[0][3].decode('utf-8')

    stmt = """
        INSERT INTO news (member_id, create_datetime, publish, title, summary, content)
        VALUES (8000000, ?, true, ?, ?, ?)
    """
    sc.execute(stmt, (create_datetime, title, summary, content))


## USERS
mdb.query("""
    SELECT
        ocvt_users.member_id, ocvt_users.email, ocvt_users.name_first, ocvt_users.name_last,
        ocvt_users.date_created, ocvt_users.cell_number, ocvt_users.gender, ocvt_users.birth_year,
        ocvt_users.active, ocvt_users.medical_cond, ocvt_users.medical_cond_desc,
        emergency_contacts.contact_name, emergency_contacts.contact_number,
        emergency_contacts.contact_relationship, ocvt_members.expiration_date
    FROM ocvt_users
    LEFT JOIN emergency_contacts ON emergency_contacts.member_id = ocvt_users.member_id
    LEFT JOIN ocvt_members ON ocvt_members.member_id = ocvt_users.member_id
""")
rows = mdb.use_result()

default_n = {"GENERAL_ANNOUNCEMENTS":True,"TRIP_BACKPACKING":True,"TRIP_BIKING":True,"TRIP_CAMPING":True,"TRIP_CLIMBING":True,"TRIP_DAYHIKE":True,"TRIP_LASER_TAG":True,"TRIP_OFFICIAL_MEETING":True,"TRIP_OTHER":True,"TRIP_RAFTING_CANOEING_KAYAKING":True,"TRIP_ROAD_TRIP":True,"TRIP_SKIING_SNOWBOARDING":True,"TRIP_SNOW_OTHER":True,"TRIP_SOCIAL":True,"TRIP_SPECIAL_EVENT":True,"TRIP_TEAM_SPORTS_MISC":True,"TRIP_WATER_OTHER":True,"TRIP_WORK_TRIP":True}
default_n_str = str(default_n).replace('True', 'true').replace('False', 'false').replace("'", '"')
while True:
    row = rows.fetch_row()
    if len(row) == 0:
        break

    # parse & decode
    id              = int(row[0][0])
    email           = row[0][1].decode('utf-8')
    name            = row[0][2].decode('utf-8') + ' ' + row[0][3].decode('utf-8')
    create_datetime = row[0][4].decode('utf-8')
    cell_number     = row[0][5].decode('utf-8')
    gender          = row[0][6].decode('utf-8')
    birth_year      = int(row[0][7])
    active          = int(row[0][8]) == 1
    mc              = int(row[0][9]) == 1
    mcd             = "" if row[0][10] is None else row[0][10].decode('utf-8')
    # emergency contacts
    ec_name         = "" if row[0][11] is None else row[0][11].decode('utf-8')
    ec_number       = "" if row[0][12] is None else row[0][12].decode('utf-8')
    ec_relationship = "" if row[0][13] is None else row[0][13].decode('utf-8')
    # membership
    paid_expire_datetime = None if row[0][14] is None else row[0][14].decode('utf-8')

    stmt = """
        INSERT INTO oldsite_member (
            id, email, name, create_datetime, cell_number, gender, birth_year, active,
            medical_cond, medical_cond_desc, paid_expire_datetime, ec_name, ec_number, ec_relationship, notification_preference
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), ?, ?, ?, ?)
    """
    sc.execute(stmt, (
        id, email, first_name, last_name, create_datetime, cell_number, gender, birth_year, active,
        mc, mcd, ec_name, ec_number, ec_relationship, default_n_str)
    )

    if paid_expire_datetime is not None:
        stmt = """
            UPDATE oldsite_member
            SET paid_expire_datetime = ?
            WHERE id = ?
        """
        sc.execute(stmt, (paid_expire_datetime, id))


## NOTIFICATION SETTINGS
mdb.query("""
    SELECT
        ocvt_users.member_id, notification_settings.type, notification_settings.subtype
    FROM ocvt_users
    INNER JOIN notification_settings ON notification_settings.member_id = ocvt_users.member_id
""")
rows = mdb.use_result()

member_n = {}
oldsite_n = {
    "TRIP": {
        "TR01": "TRIP_DAYHIKE",
        "TR02": "TRIP_WORK_HIKE",
        "TR03": "TRIP_BACKPACKING",
        "TR04": "TRIP_CAMPING",
        "TR05": "TRIP_OFFICIAL_MEETING",
        "TR06": "TRIP_SOCIAL",
        "TR07": "TRIP_RAFTING_CANOEING_KAYAKING",
        "TR08": "TRIP_WATER_OTHER",
        "TR09": "TRIP_BIKING",
        "TR10": "TRIP_TEAM_SPORTS_MISC",
        "TR11": "TRIP_CLIMBING",
        "TR12": "TRIP_SKIING_SNOWBOARDING",
        "TR13": "TRIP_SNOW_OTHER",
        "TR14": "TRIP_ROAD_TRIP",
        "TR15": "TRIP_SPECIAL_EVENT",
        "TR16": "TRIP_OTHER",
        "TR17": "TRIP_LASER_TAG"
    }
}
while True:
    row = rows.fetch_row()
    if len(row) == 0:
        break

    member_id = row[0][0].decode('utf-8')
    n_type = row[0][1].decode('utf-8')
    n_subtype = row[0][2].decode('utf-8')

    if member_id not in member_n:
        member_n[member_id] = default_n

    if n_type == "GENERAL":
        member_n[member_id]["GENERAL_ANNOUNCEMENTS"] = False
    else:
        member_n[member_id][oldsite_n[n_type][n_subtype]] = False

for m_id in member_n:
    stmt = """
        UPDATE member
        SET notification_preference = ?
        WHERE id = ?
    """
    n = str(member_n[m_id]).replace('True', 'true').replace('False', 'false').replace("'", '"')
    sc.execute(stmt, (n, int(m_id)))


## ORDERS
mdb.query("""
    SELECT ocvt_orders.member_id, ocvt_orders.order_date, ocvt_orders.order_number, order_items.item_number
    FROM ocvt_orders
    INNER JOIN order_items ON order_items.order_number = ocvt_orders.order_number
""")
rows = mdb.use_result()


oldsite_items = {
    "2010026": {"MEMBERSHIP": 1, "SHIRT": 0, "cost": 2000},
    "2010027": {"MEMBERSHIP": 1, "SHIRT": 1, "cost": 3000},
    "2010028": {"MEMBERSHIP": 4, "SHIRT": 0, "cost": 6500},
    "2010029": {"MEMBERSHIP": 0, "SHIRT": 0}
}
while True:
    row = rows.fetch_row()
    if len(row) == 0:
        break

    # parse & decode
    member_id       = int(row[0][0])
    create_datetime = row[0][1].decode('utf-8')
    payment_id      = row[0][2].decode('utf-8')
    item_number     = row[0][3].decode('utf-8')

    if oldsite_items[item_number]["MEMBERSHIP"] > 0:
        stmt = """
            INSERT INTO oldsite_payment (
                create_datetime, entered_by_id, note, member_id, store_item_id, store_item_count,
                amount, payment_method, payment_id, completed
            ) VALUES (?, 8000000, '', ?, ?, ?, ?, 'OLDSITE', ?, true)
        """
        sc.execute(stmt, (create_datetime, member_id, 'MEMBERSHIP', oldsite_items[item_number]["MEMBERSHIP"], oldsite_items[item_number]["cost"], payment_id))
    if oldsite_items[item_number]["SHIRT"] > 0:
        stmt = """
            INSERT INTO oldsite_payment (
                create_datetime, entered_by_id, note, member_id, store_item_id, store_item_count,
                amount, payment_method, payment_id, completed
            ) VALUES (?, 8000000, '', ?, ?, ?, ?, 'OLDSITE', ?, true)
        """
        sc.execute(stmt, (create_datetime, member_id, 'SHIRT', oldsite_items[item_number]["SHIRT"], oldsite_items[item_number]["cost"], payment_id))
    

## MANUAL PAYMENTS
mdb.query("SELECT * FROM ocvt_manual_payments")
rows = mdb.use_result()
while True:
    row = rows.fetch_row()
    if len(row) == 0:
        break

    member_id = None if row[0][11] is None else int(row[0][11])
    stmt = """
        INSERT INTO oldsite_manual_payments (
            id, email, name, create_date, membership_days, entered_by, notes,
            completed, completed_date, member_completed, member_completed_date,
            member_id, shirt_completed, shirt_completed_date
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    """
    sc.execute(stmt, (int(row[0][0]), row[0][1], row[0][2], row[0][3], int(row[0][4]),
        int(row[0][5]), row[0][6], int(row[0][7]) == 1, row[0][8], int(row[0][9]) == 1,
        row[0][10], m_id, int(row[0][12]) == 1, row[0][13]))


scon.commit()
scon.close()
