HOST = 'http://api.cabinet.seaturtle.pw:3000'

member_is_trip_creator_json = {
  'error':'Cannot modify trip creator status.'
}

member_new_json = {
  'email': 'test@example.com',
  'firstName': 'No',
  'lastName': 'Thanks',
  'cellNumber': '5551234567',
  'gender': 'Apache Attack Helicopter',
  'birthyear': 1990,
  'active': True,
  'medicalCond': True,
  'medicalCondDesc': 'very allergic to tomatoes'
}

member_new_emergency_json = {
  'emergencyContactName': 'Elon Musk',
  'emergencyContactNumber': '9993729484',
  'emergencyContactRelationship': 'father'
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
  'generalEvents': True,
  'generalItemsOfInterest': True,
  'generalItemsForSale': True,
  'generalMeetings': True,
  'generalNews': True,
  'generalOther': True,
  'tripAlerts': True,
  'tripBackpacking': True,
  'tripBiking': True,
  'tripCamping': True,
  'tripClimbing': True,
  'tripDayhike': True,
  'tripLaserTag': True,
  'tripOfficialMeeting': True,
  'tripOther': True,
  'tripRaftingCanoeingKayaking': True,
  'tripRoadTrip': True,
  'tripSkiingSnowboarding': True,
  'tripSnowOther': True,
  'tripSocial': True,
  'tripSpecialEvent': True,
  'tripTeamSportsMisc': True,
  'tripWaterOther': True,
  'tripWorkTrip': True
}

payment_redeem_json = {
  'code': '3h0875f2903847h5f02938457fh0239'
}

signup_status_cancel_json = {
  "error":"Signup status is CANCEL."
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
  'tripTypeId': 'TR02',
  'startDatetime': '2020-12-24T15:52:01Z',
  'endDatetime': '2020-12-30T15:52:01Z',
  'summary': 'hello world',
  'description': 'this IS a DESCRIPTION',
  'location': 'i dunno',
  'locationDirections': 'go straight',
  'meetupLocation': 'idk surge',
  'distance': '2.6',
  'difficulty': 5,
  'difficultyDescription': 'very very hard',
  'instructions': 'bring all this shit',
  'petsAllowed': False,
  'petsDescription': ''
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
  'GENERAL_EVENTS': {
    'typeDescription': 'Gobblerfest, parade, etc',
    'typeName': 'General Events'
  },
  'GENERAL_IMPORTANT': {
    'typeDescription': 'Important Club Announcements',
    'typeName': 'General Important Items'
  },
  'GENERAL_ITEMS_FOR_SALE': {
    'typeDescription': 'Items for sale through the club',
    'typeName': 'General Items for Sale'
  },
  'GENERAL_MEETINGS': {
    'typeDescription': 'Announcements about Club Meetings',
    'typeName': 'General Meeting'
  },
  'GENERAL_NEWS': {
    'typeDescription': 'News from the Club',
    'typeName': 'General News'
  },
  'GENERAL_OTHER': {
    'typeDescription': 'Miscellaneous Club Announcements',
    'typeName': 'General Other'
  },
  'TRIP_BACKPACKING': {
    'typeDescription': 'Multi day hikes.',
    'typeName': 'Backpacking'
  },
  'TRIP_BIKING': {
    'typeDescription': 'Road or mountain biking.',
    'typeName': 'Biking'
  },
  'TRIP_CAMPING': {
    'typeDescription': 'Single overnight trips.',
    'typeName': 'Camping'
  },
  'TRIP_CLIMBING': {
    'typeDescription': 'Rock climbing or bouldering.',
    'typeName': 'Climbing'
  },
  'TRIP_DAYHIKE': {
    'typeDescription': 'In and out on the same day.',
    'typeName': 'Dayhike'
  },
  'TRIP_LASER_TAG': {
    'typeDescription': 'Laser Tag with LCAT',
    'typeName': 'Laser Tag'
  },
  'TRIP_OFFICIAL_MEETING': {
    'typeDescription': 'An official OCVT meeting',
    'typeName': 'Official Meeting'
  },
  'TRIP_OTHER': {
    'typeDescription': 'Anything else not covered. ',
    'typeName': 'Other'
  },
  'TRIP_RAFTING_CANOEING_KAYAKING': {
    'typeDescription': 'Rafting / Canoeing / Kayaking',
    'typeName': 'Rafting / Canoeing / Kayaking'
  },
  'TRIP_ROAD_TRIP': {
    'typeDescription': 'Just getting out and about, Ex a trip to Busch Gardens or DC etc',
    'typeName': 'Road Trip'
  },
  'TRIP_SKIING_SNOWBOARDING': {
    'typeDescription': 'Skiing / Snowboarding',
    'typeName': 'Skiing / Snowboarding'
  },
  'TRIP_SNOW_OTHER': {
    'typeDescription': 'Sledding snowshoeing etc',
    'typeName': 'Snow / Other'
  },
  'TRIP_SOCIAL': {
    'typeDescription': 'Strictly social, potluck, movie nights, games or other casual gatherings',
    'typeName': 'Social'
  },
  'TRIP_SPECIAL_EVENT': {
    'typeDescription': 'A special event.',
    'typeName': 'Special Event'
  },
  'TRIP_TEAM_SPORTS_MISC': {
    'typeDescription': 'Football, basketball ultimate Frisbee etc.',
    'typeName': 'Team Sports / Misc.'
  },
  'TRIP_WATER_OTHER': {
    'typeDescription': 'Swimming, tubing anything else in the water.',
    'typeName': 'Water / Other'
  },
  'TRIP_WORK_TRIP': {
    'typeDescription': 'Trail work or other maintenance.',
    'typeName': 'Worktrip'
  }
}
