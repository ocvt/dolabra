package handler

import (
  "encoding/json"
  "net/http"
)

type notificationsStruct struct {
  GeneralEvents bool `json:"generalEvents"`
  GeneralItemsOfInterest bool `json:"generalItemsOfInterest"`
  GeneralItemsForSale bool `json:"generalItemsForSale"`
  GeneralMeetings bool `json:"generalMeetings"`
  GeneralNews bool `json:"generalNews"`
  GeneralOther bool `json:"generalOther"`
  TripAlert bool `json:"tripAlerts"`
  TripBackpacking bool `json:"tripBackpacking"`
  TripBiking bool `json:"tripBiking"`
  TripCamping bool `json:"tripCamping"`
  TripClimbing bool `json:"tripClimbing"`
  TripDayhike bool `json:"tripDayhike"`
  TripLaserTag bool `json:"tripLaserTag"`
  TripOfficialMeeting bool `json:"tripOfficialMeeting"`
  TripOther bool `json:"tripOther"`
  TripRaftingCanoeingKayaking bool `json:"tripRaftingCanoeingKayaking"`
  TripRoadTrip bool `json:"tripRoadTrip"`
  TripSkiingSnowboarding bool `json:"tripSkiingSnowboarding"`
  TripSnowOther bool `json:"tripSnowOther"`
  TripSocial bool `json:"tripSocial"`
  TripSpecialEvent bool `json:"tripSpecialEvent"`
  TripTeamSportsMisc bool `json:"tripTeamSportsMisc"`
  TripWaterOther bool `json:"tripWaterOther"`
  TripWorkTrip bool `json:"tripWorkTrip"`
}

func setAllPreferences() *notificationsStruct {
  return &notificationsStruct{
    GeneralEvents: true,
    GeneralItemsOfInterest: true,
    GeneralItemsForSale: true,
    GeneralMeetings: true,
    GeneralNews: true,
    GeneralOther: true,
    TripAlert: true,
    TripBackpacking: true,
    TripBiking: true,
    TripCamping: true,
    TripClimbing: true,
    TripDayhike: true,
    TripLaserTag: true,
    TripOfficialMeeting: true,
    TripOther: true,
    TripRaftingCanoeingKayaking: true,
    TripRoadTrip: true,
    TripSkiingSnowboarding: true,
    TripSnowOther: true,
    TripSocial: true,
    TripSpecialEvent: true,
    TripTeamSportsMisc: true,
    TripWaterOther: true,
    TripWorkTrip: true,
  }
}

func GetMyAccountNotifications(w http.ResponseWriter, r *http.Request) {
  _, subject, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Get memberId
  memberId, ok := dbGetActiveMemberId(w, subject)
  if !ok {
    return
  }

  notifications, ok := dbGetMemberNotifications(w, memberId)
  if !ok {
    return
  }

  respondJSON(w, http.StatusOK,
      map[string]notificationsStruct{"notifications": notifications})
}

func PatchMyAccountNotifications(w http.ResponseWriter, r *http.Request) {
  _, subject, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Get memberId
  memberId, ok := dbGetActiveMemberId(w, subject)
  if !ok {
    return
  }

  // Get request body, ensure correct formatted correctly
  decoder := json.NewDecoder(r.Body)
  decoder.DisallowUnknownFields()
  var notifications notificationsStruct
  err := decoder.Decode(&notifications)
  if err != nil {
    respondError(w, http.StatusBadRequest, err.Error())
    return
  }
  notifications.TripAlert = true

  notificationsArr, err := json.Marshal(notifications)
  if !checkError(w, err) {
    return
  }
  notificationsStr := string(notificationsArr)

  stmt := `
    UPDATE member
    SET notification_preference = ?
    WHERE id = ?`
  _, err = db.Exec(stmt, notificationsStr, memberId)
  if !checkError(w, err) {
    return
  }

  respondJSON(w, http.StatusNoContent, nil)
}
