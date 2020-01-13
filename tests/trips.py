import json

import requests as req

from data import *

# Test methods relating to trips

ENDPOINT = HOST + '/trips'

def TestGetTripsNone():
  url = ENDPOINT

  r = req.get(url)
  assert r.status_code == 200
  assert len(json.loads(r.text)['trips']) == 0

  for path in ['/archive', '/archive/1', '/archive/1/2']:
    url = ENDPOINT + path
    r = req.get(url)
    assert r.status_code == 200
    assert len(json.loads(r.text)['trips']) == 1

# TODO trip creation
def TestMyTrips(s):
  url = ENDPOINT + '/mytrips'
  
  r = req.get(url)
  assert r.status_code == 401

  r = s.get(url)
  assert r.status_code == 200
  assert len(json.loads(r.text)['trips']) == 0

def TestTripsPhotos():
  url = ENDPOINT + '/photos'
  
  r = req.get(url)
  assert r.status_code == 200
  assert 'images' in json.loads(r.text)

def TestTripsTypes():
  url = ENDPOINT + '/types'
  
  r = req.get(url)
  assert r.status_code == 200
  assert json.loads(r.text) == trips_types_json

# TODO officer and trip leader
def TestTripsAdmin(s):
  url = ENDPOINT + '/1/admin'
  
  r = req.get(url)
  assert r.status_code == 401

  r = s.get(url)
  assert r.status_code == 400
  assert json.loads(r.text) == member_not_officer_tripleader_json

# TODO test with photos
def TestTripPhotos():
  url = ENDPOINT + '/10000/photos'
  
  r = req.get(url)
  assert r.status_code == 200
  assert len(json.loads(r.text)['images']) == 0

# TODO officer and trip leader
def TestTripsModify(s):
  paths = {
    '/cancel': member_not_officer_tripleader_json,
    '/publish': member_not_tripleader_json
  }
  for path in paths:
    url = ENDPOINT + '/10000' + path
    
    r = req.patch(url)
    assert r.status_code == 401
    assert json.loads(r.text) == member_not_authenticated_json

    r = s.patch(url)
    assert r.status_code == 400
    assert json.loads(r.text) == paths[path]

def TestTripsCreate(s):
  url = ENDPOINT
  
  r = req.post(url)
  assert r.status_code == 401
  assert json.loads(r.text) == member_not_authenticated_json

  r = s.post(url)
  assert r.status_code == 400

  r = s.post(url, json=trip_json)
  assert r.status_code == 201
  assert json.loads(r.text)['tripId'] == 1

  r = s.post(url, json=trip_json)
  assert r.status_code == 201
  assert json.loads(r.text)['tripId'] == 2

def TestTripsPublish(s1, s2):
  url = ENDPOINT
  url_publish = ENDPOINT + '/1/publish'
  url_signup = ENDPOINT + '/1/signup'
  
  r = req.get(url)
  assert r.status_code == 200
  assert len(json.loads(r.text)['trips']) == 0

  r = s1.patch(url_publish)
  assert r.status_code == 400
  assert json.loads(r.text) == member_not_tripleader_json
 
  r = s1.post(url_signup, json=trip_signup_json)
  assert r.status_code == 204
  
  r = s2.post(url_signup, json=trip_signup_json)
  assert r.status_code == 400
  assert json.loads(r.text) == trip_not_published_json

  r = s1.patch(url_publish)
  assert r.status_code == 204
  
def TestTripsPostSignup(s1, s2, s3):
  url = ENDPOINT + '/1/signup'

  r = req.post(url)
  assert r.status_code == 401
  assert json.loads(r.text) == member_not_authenticated_json

  r = s1.post(url)
  assert r.status_code == 400

  r = s1.post(url, json=trip_signup_json)
  assert r.status_code == 400
  assert json.loads(r.text) == member_on_trip_json

  r = s2.post(url, json=trip_signup_json)
  assert r.status_code == 204

  r = s2.post(url, json=trip_signup_json)
  assert r.status_code == 400
  assert json.loads(r.text) == member_on_trip_json

  r = s3.post(url, json=trip_signup_json)
  assert r.status_code == 204

def TestTripsGetSignup(s1, s2):
  url = ENDPOINT + '/1/signup'

  r = req.get(url)
  assert r.status_code == 401
  assert json.loads(r.text) == member_not_authenticated_json

  r = s1.get(url)
  assert r.status_code == 200
  assert json.loads(r.text)['id'] == 1

  r = s2.get(url)
  assert r.status_code == 200
  assert json.loads(r.text)['id'] == 2

def TestTripsSignupCancel(s1, s2):
  url = ENDPOINT + '/1/signup/cancel'
  url_signup = ENDPOINT + '/1/signup'

  r = req.patch(url)
  assert r.status_code == 401
  assert json.loads(r.text) == member_not_authenticated_json

  r = s1.patch(url)
  assert r.status_code == 400
  assert json.loads(r.text) == member_is_trip_creator_json

  r = s2.patch(url)
  assert r.status_code == 204

  r = s2.patch(url)
  assert r.status_code == 400
  assert json.loads(r.text) == signup_status_cancel_json

  r = s2.post(url_signup, json=trip_signup_json)
  assert r.status_code == 400
  assert json.loads(r.text) == member_on_trip_json

# TODO Test admin

def TestTripsCancel(s1, s2, s3):
  url_trips = ENDPOINT
  url = ENDPOINT + '/1/cancel'

  r = req.patch(url)
  assert r.status_code == 401
  assert json.loads(r.text) == member_not_authenticated_json

  r = s2.patch(url)
  assert r.status_code == 400
  assert json.loads(r.text) == member_not_officer_tripleader_json

  r = req.get(ENDPOINT)
  assert r.status_code == 200
  assert len(json.loads(r.text)['trips']) == 1

  r = s1.patch(url)
  assert r.status_code == 204

  r = req.get(ENDPOINT)
  assert r.status_code == 200
  assert len(json.loads(r.text)['trips']) == 0
  
  r = s1.patch(url)
  assert r.status_code == 400
  assert json.loads(r.text) == trip_canceled_json

  r = s2.patch(url)
  assert r.status_code == 400
  assert json.loads(r.text) == member_not_officer_tripleader_json

  r = s3.patch(url)
  assert r.status_code == 400
  assert json.loads(r.text) == member_not_officer_tripleader_json

# TODO TEST absent/boot/forceadd/tripleader
