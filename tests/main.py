import requests as req

from data import *
from myaccount import *
from root import *
from trips import *

USER1_SUB = 'e6d3923cd6554d12861fb87dec458c13'
USER2_SUB = 'a877c314b56d4ec1a6cf4adf21b1927e'
USER3_SUB = 'c138d54f608c4144aa8d13a7cf3f6e17'

user1 = req.Session()
user2 = req.Session()
user3 = req.Session()


# Test authentication & member endpoints

TestAuth(user1, USER1_SUB)
TestMyAccountNotRegistered(user1)
TestMyAccountRegister(user1)
TestMyAccountPersonal(user1)
TestMyAccountUpdateEmergency(user1)
TestMyAccount(user1)
TestMyAccountName(user1)
TestMyAccountNotifications(user1)

TestAuth(user3, USER3_SUB)
TestMyAccountRegister(user3)

TestAuth(user2, USER2_SUB)
TestMyAccountRegister(user2)


# Test root level endpoints

TestHomePhotos()
TestNews()
TestNewsArchive()
#TestPayment(user1)
TestPaymentRedeem(user1)
TestQuicksignup()
TestUnsubscribe()
TestMyAccountNotificationsNone(user1)


# Test trips endpoints

TestGetTripsNone()
TestMyTrips(user1)
TestTripsPhotos()
TestTripsTypes()
TestTripsAdmin(user1)
TestTripPhotos()
TestTripsModify(user1)
TestTripsCreate(user1)
TestTripsPublish(user1, user2)
TestTripsPostSignup(user1, user2, user3)
TestTripsGetSignup(user1, user2)
TestTripsSignupCancel(user1, user2)
TestTripsCancel(user1, user2, user3)

## Manually test:
## - email/notify
##   - /trips/{tripId}/notify/signup/{signupId}
##   - /trips/{tripId}/notify/notify/{groupId}
## - photo upload
##   - /trips/{tripId}/mainphoto
##   - /trips/{tripId}/photos

