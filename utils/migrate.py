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
#   - ocvt_news.title -> news.title, news.
#   - ocvt_news.content -> news.content
#   - ocvt_news.content first X characters -> news.summary
#   - ocvt_news.entered_by -> news.member_id ----> MANUALLY UPDATE AFTER USERS ARE ADDED
#   - SET news.publish TRUE
# - users -> 'oldsite_member' START member_id at 12000
#   - ocvt_members.expiration_date -> oldsite_member.paid_expire_datetime
#   - ocvt_users.email -> oldsite_member.email
#   - ocvt_users.name_first -> oldsite_member.first_name
#   - ocvt_users.name_last -> oldsite_member.last_name
#   - ocvt_users.date_created -> oldsite_member.create_datetime
#   - ocvt_users.cell_number -> oldsite_member.cell_number
#   - ocvt_users.gender -> oldsite_member.gender
#   - ocvt_users.birth_year -> oldsite_member.birth_year
#   - ocvt_users.active -> oldsite_member.active
#   - ocvt_users.medical_cond -> oldsite_member.medical_cond
#   - ocvt_users.medical_cond_desc -> oldsite_member.medical_cond_desc

# - emergency contacts -> 'oldsite_emergency_contact' JOIN ON oldsite_member.member_id
#   - emergency_contacts.contact_name -> oldsite_emergency_contact.name
#   - emergency_contacts.contact_number -> oldsite_emergency_contact.number
#   - emergency_contacts.contact_relationship -> oldsite_emergency_contact.relationship
# - notification preferences
#   - for each row in notification_settings:
#     - if member_id in oldsite_member:
#       - set preference for notification preference to false
# - orders -> 'oldsite_payment' (online orders), 'oldsite_payment_manual' (manual orders)
#   - ocvt_items: unused
#   - ocvt_orders + order_items (ONLINE ORDERS)
#   - ocvt_manual_payments (MANUAL ORDERS)
#     - ASK DOUG
#     - Create webtools page to view all incomplete manual orders
#     - Email all incomplete people >= 2019
#     - Add note on myocvt page to email webmaster if any issues
# Dependencies:
# - mysqlclient

import datetime
import sqlite3
from MySQLdb import _mysql

mdb = _mysql.connect('127.0.0.1', 'root', 'ocvt', 'ocvt')
scon = sqlite3.connect('dolabra-sqlite.sqlite3')
sc = scon.cursor()


mdb.query("SELECT * FROM ocvt_news")
rows = mdb.use_result()
while True:
    row = rows.fetch_row()
    if len(row) == 0:
        break
    print(row)
    sc.execute(f"INSERT INTO news (member_id, create_datetime, publish, title, summary, content) VALUES (0, ?, true, ?, ?, ?)", (row[0][1].decode('utf-8'), row[0][2].decode('utf-8'), row[0][2].decode('utf-8'), row[0][3].decode('utf-8')))

scon.commit()
scon.close()
#con = mdb.connect('127.0.0.1', 'root', 'ocvt', 'ocvt')

#with con:
#    cur = con.cursor()
#    cur.execute("SELECT * FROM ocvt_news")
#
#    rows = cur.fetchall()
#
#    for row in rows:
#        print(row.encode('utf8'))
