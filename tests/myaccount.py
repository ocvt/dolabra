import json

import requests as req

from data import *

# Test methods relating to auth and accounts

def TestAuth(s, subject):
  url = HOST + '/auth/dev/' + subject
  r = s.get(url, allow_redirects=False)
  assert r.status_code == 307

def TestMyAccountNotRegistered(s):
  for path in ['', '/name', '/notifications']:
    url = HOST + '/myaccount' + path
    r = s.get(url)
    assert r.status_code == 404
    assert json.loads(r.text) == member_not_registered_json

def TestMyAccountRegister(s):
  url = HOST + '/myaccount'
  r = s.post(url, json=member_json)
  assert r.status_code == 201

def TestMyAccount(s):
  url = HOST + '/myaccount'
  r = s.get(url)
  assert r.status_code == 200
  data = json.loads(r.text)
  del data['id'], data['createDatetime']
  assert data == member_json

def TestMyAccountName(s):
  url = HOST + '/myaccount/name'
  r = s.get(url)
  assert r.status_code == 200
  assert json.loads(r.text) == {'firstName': member_json['firstName']}

def TestMyAccountNotifications(s):
  url = HOST + '/myaccount/notifications'
  r = s.get(url)
  assert r.status_code == 200
  assert json.loads(r.text)['notifications'] == notifications_json

def TestMyAccountNotificationsNone(s):
  url = HOST + '/myaccount/notifications'
  n_none = {i: False for i in notifications_json}
  r = s.get(url)
  assert r.status_code == 200
  assert json.loads(r.text)['notifications'] == n_none
