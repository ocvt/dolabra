HOST = 'http://api.cabinet.seaturtle.pw:3000'

member_is_trip_creator_json = {
  'error':'Cannot modify trip creator status.'
}

member_new_json = {
  'email': 'test@example.com',
  'name': 'No Thanks',
  'cellNumber': '5551234567',
  'pronouns': 'they/them',
  'birthyear': 1990,
  'active': True,
  'medicalCond': True,
  'medicalCondDesc': 'very allergic to tomatoes'
}

member_new_emergency_json = {
  'ECName': 'Elon Musk',
  'ECNumber': '9993729484',
  'ECRelationship': 'father'
}

member_json = {**member_new_json, **member_new_emergency_json}

member_not_authenticated_json = {
  'error': 'Member is not authenticated'
}

member_not_officer_tripleader_json = {
  'error': 'Must be officer or trip leader.'
}

member_not_registered_json = {
  'error': 'Member is not registered.'
}

member_not_tripleader_json = {
  'error': 'Not a trip leader.'
}

member_on_trip_json = {
  'error': 'Member is on trip (or has canceled or been booted).'
}

notifications_json = {
  'GENERAL_ANNOUNCEMENTS': True,
  'TRIP_BACKPACKING': True,
  'TRIP_BIKING': True,
  'TRIP_CAMPING': True,
  'TRIP_CLIMBING': True,
  'TRIP_DAYHIKE': True,
  'TRIP_LASER_TAG': True,
  'TRIP_OFFICIAL_MEETING': True,
  'TRIP_OTHER': True,
  'TRIP_RAFTING_CANOEING_KAYAKING': True,
  'TRIP_ROAD_TRIP': True,
  'TRIP_SKIING_SNOWBOARDING': True,
  'TRIP_SNOW_OTHER': True,
  'TRIP_SOCIAL': True,
  'TRIP_SPECIAL_EVENT': True,
  'TRIP_TEAM_SPORTS_MISC': True,
  'TRIP_WATER_OTHER': True,
  'TRIP_WORK_TRIP': True
}

payment_redeem_json = {
  'code': '3h0875f2903847h5f02938457fh0239'
}

signup_status_cancel_json = {
  'error':'Signup status is CANCEL.'
}

simple_email_json = {
  'email': 'test@example.com'
}

trip_json = {
  'membersOnly': False,
  'allowLateSignups': True,
  'drivingRequired': False,
  'hasCost': False,
  'costDescription': '',
  'maxPeople': 10,
  'name': 'test trip',
  'notificationTypeId': 'TRIP_BIKING',
  'startDatetime': '3020-12-24T15:52:01Z',
  'endDatetime': '3020-12-30T15:52:01Z',
  'summary': 'hello world',
  'description': 'this IS a DESCRIPTION',
  'location': 'i dunno',
  'locationDirections': 'go straight',
  'meetupLocation': 'idk surge',
  'distance': 2.6,
  'difficulty': 5,
  'difficultyDescription': 'very very hard',
  'instructions': 'bring all this shit',
  'petsAllowed': False,
  'petsDescription': ''
}

trip_does_not_exist_json = {
  'error': 'Trip does not exist.'
}

trip_canceled_json = {
  'error': 'Trip is canceled.'
}

trip_not_published_json = {
  'error': 'Trip is not published.'
}

trip_signup_json = {
  'shortNotice': True,
  'driver': True,
  'carpool': False,
  'capCapacityTotal': 100,
  'notes': 'signup notes',
  'pet': False,
  'attended': False
}

trips_types_json = {
  "GENERAL_ANNOUNCEMENTS": {
    "description": "Important Club Announcements",
    "id": "GENERAL_ANNOUNCEMENTS",
    "name": "Club Meetings / News / Events"
  },
  "TRIP_ALERT_ATTEND": {
    "description": "Signup status changed to attending",
    "id": "TRIP_ALERT_ATTEND",
    "name": "Signup status changed to attending"
  },
  "TRIP_ALERT_BOOT": {
    "description": "Signup status changed to booted",
    "id": "TRIP_ALERT_BOOT",
    "name": "Signup status changed to booted"
  },
  "TRIP_ALERT_CANCEL": {
    "description": "Signup status changed to cancel",
    "id": "TRIP_ALERT_CANCEL",
    "name": "Signup status changed to cancel"
  },
  "TRIP_ALERT_FORCE": {
    "description": "Signup status changed to force-added",
    "id": "TRIP_ALERT_FORCE",
    "name": "Signup status changed to force-added"
  },
  "TRIP_ALERT_LEADER": {
    "description": "Signup status changed to trip leader",
    "id": "TRIP_ALERT_LEADER",
    "name": "Signup status changed to trip leader"
  },
  "TRIP_ALERT_WAIT": {
    "description": "Signup status changed to waitlisted",
    "id": "TRIP_ALERT_WAIT",
    "name": "Signup status changed to waitlisted"
  },
  "TRIP_APPROVAL": {
    "description": "Alert asking member (probably officer) to approve a trip",
    "id": "TRIP_APPROVAL",
    "name": "Trip Approval Alert"
  },
  "TRIP_BACKPACKING": {
    "description": "Multi day hikes.",
    "id": "TRIP_BACKPACKING",
    "name": "Backpacking"
  },
  "TRIP_BIKING": {
    "description": "Road or mountain biking.",
    "id": "TRIP_BIKING",
    "name": "Biking"
  },
  "TRIP_CAMPING": {
    "description": "Single overnight trips.",
    "id": "TRIP_CAMPING",
    "name": "Camping"
  },
  "TRIP_CLIMBING": {
    "description": "Rock climbing or bouldering.",
    "id": "TRIP_CLIMBING",
    "name": "Climbing"
  },
  "TRIP_DAYHIKE": {
    "description": "In and out on the same day.",
    "id": "TRIP_DAYHIKE",
    "name": "Dayhike"
  },
  "TRIP_LASER_TAG": {
    "description": "Laser Tag with LCAT",
    "id": "TRIP_LASER_TAG",
    "name": "Laser Tag"
  },
  "TRIP_MESSAGE_ATTEND": {
    "description": "Trip message to force-added and attendees",
    "id": "TRIP_MESSAGE_ATTEND",
    "name": "Trip message to force-added and attendees"
  },
  "TRIP_MESSAGE_DIRECT": {
    "description": "Direct message related to trip",
    "id": "TRIP_MESSAGE_DIRECT",
    "name": "Direct message related to trip"
  },
  "TRIP_MESSAGE_NOTIFY": {
    "description": "Trip message for force-added, attendees, and waitlist",
    "id": "TRIP_MESSAGE_NOTIFY",
    "name": "Trip message for force-added, attendees, and waitlist"
  },
  "TRIP_MESSAGE_WAIT": {
    "description": "Trip message to waitlist",
    "id": "TRIP_MESSAGE_WAIT",
    "name": "Trip message to waitlist"
  },
  "TRIP_OFFICIAL_MEETING": {
    "description": "An official meeting",
    "id": "TRIP_OFFICIAL_MEETING",
    "name": "Official Meeting"
  },
  "TRIP_OTHER": {
    "description": "Anything else not covered. ",
    "id": "TRIP_OTHER",
    "name": "Other"
  },
  "TRIP_RAFTING_CANOEING_KAYAKING": {
    "description": "Rafting / Canoeing / Kayaking",
    "id": "TRIP_RAFTING_CANOEING_KAYAKING",
    "name": "Rafting / Canoeing / Kayaking"
  },
  "TRIP_ROAD_TRIP": {
    "description": "Just getting out and about, Ex a trip to Busch Gardens or DC etc",
    "id": "TRIP_ROAD_TRIP",
    "name": "Road Trip"
  },
  "TRIP_SKIING_SNOWBOARDING": {
    "description": "Skiing / Snowboarding",
    "id": "TRIP_SKIING_SNOWBOARDING",
    "name": "Skiing / Snowboarding"
  },
  "TRIP_SNOW_OTHER": {
    "description": "Sledding snowshoeing etc",
    "id": "TRIP_SNOW_OTHER",
    "name": "Snow / Other"
  },
  "TRIP_SOCIAL": {
    "description": "Strictly social, potluck, movie nights, games or other casual gatherings",
    "id": "TRIP_SOCIAL",
    "name": "Social"
  },
  "TRIP_SPECIAL_EVENT": {
    "description": "A special event.",
    "id": "TRIP_SPECIAL_EVENT",
    "name": "Special Event"
  },
  "TRIP_TEAM_SPORTS_MISC": {
    "description": "Football, basketball ultimate Frisbee etc.",
    "id": "TRIP_TEAM_SPORTS_MISC",
    "name": "Team Sports / Misc."
  },
  "TRIP_WATER_OTHER": {
    "description": "Swimming, tubing anything else in the water.",
    "id": "TRIP_WATER_OTHER",
    "name": "Water / Other"
  },
  "TRIP_WORK_TRIP": {
    "description": "Trail work or other maintenance.",
    "id": "TRIP_WORK_TRIP",
    "name": "Worktrip"
  }
}
