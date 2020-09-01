import json

import requests as req

from data import *

# Test home level endpoints

def TestHomePhotos():
  url = HOST + '/homephotos'
  r = req.get(url)
  assert r.status_code == 200
  assert len(json.loads(r.text)['images']) > 0

# TODO Test after post
def TestNews():
  url = HOST + '/news'
  r = req.get(url)
  assert r.status_code == 200
  assert len(json.loads(r.text)['news']) == 0

def TestNewsArchive():
  url = HOST + '/news/archive'
  r = req.get(url)
  assert r.status_code == 200
  assert len(json.loads(r.text)['news']) == 0

def TestPayment(s):
  url = HOST + '/payment/invalid'
  r = s.get(url)
  assert r.status_code == 404

  for path in ['/dues', '/duesShirt', '/freshmanSpecial']:
    url = HOST + '/payment' + path
    r = req.get(url)
    assert r.status_code == 401
    assert json.loads(r.text) == member_not_authenticated_json
    
    r = s.get(url)
    assert r.status_code == 200
    assert 'stripeClientSecret' in json.loads(r.text)

def TestPaymentRedeem(s):
  url = HOST + '/payment/redeem'
  r = req.post(url)
  assert r.status_code == 401
  assert json.loads(r.text) == member_not_authenticated_json

  r = s.post(url)
  assert r.status_code == 400
  
  r = s.post(url, json=payment_redeem_json)
  assert r.status_code == 403

def TestQuicksignup():
  url = HOST + '/quicksignup'
  r = req.post(url)
  assert r.status_code == 400

  r = req.post(url, json=simple_email_json)
  assert r.status_code == 204

def TestUnsubscribe():
  url = HOST + '/unsubscribe'
  r = req.post(url)
  assert r.status_code == 404

  url += '/all'
  r = req.post(url)
  assert r.status_code == 400

  r = req.post(url, json=simple_email_json)
  assert r.status_code == 204
