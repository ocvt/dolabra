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
  TripBackpacking bool `json:"tripBackpacking"`
  TripBiking bool `json:"tripBiking"`
  TripCamping bool `json:"tripCamping"`
  TripClimbing bool `json:"tripClimbing"`
  TripDayhike bool `json:"tripDayhike"`
  TripLaserTag bool `json:"tripLaserTag"`
  TripMeeting bool `json:"tripMeeting"`
  TripOther bool `json:"tripOther"`
  TripRaftingCanoeingKayaking bool `json:"tripRaftingCanoeingKayaking"`
  TripRoadTrip bool `json:"tripRoadTrip"`
  TripSkiingSnowboarding bool `json:"tripSkiingSnowboarding"`
  TripSnowOther bool `json:"tripSnowOther"`
  TripSocial bool `json:"tripSocial"`
  TripSpecialEvent bool `json:"tripSpecialEvent"`
  TripTeamSportsMisc bool `json:"tripTeamSportsMisc"`
  TripWorkTrip bool `json:"tripWorkTrip"`
  TripWaterOther bool `json:"tripWaterOther"`
}

func setAllPreferences() notificationsStruct {
  return notificationsStruct{
    GeneralEvents: true,
    GeneralItemsOfInterest: true,
    GeneralItemsForSale: true,
    GeneralMeetings: true,
    GeneralNews: true,
    GeneralOther: true,
    TripBackpacking: true,
    TripBiking: true,
    TripCamping: true,
    TripClimbing: true,
    TripDayhike: true,
    TripLaserTag: true,
    TripMeeting: true,
    TripOther: true,
    TripRaftingCanoeingKayaking: true,
    TripRoadTrip: true,
    TripSkiingSnowboarding: true,
    TripSnowOther: true,
    TripSocial: true,
    TripSpecialEvent: true,
    TripTeamSportsMisc: true,
    TripWorkTrip: true,
    TripWaterOther: true,
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

  stmt := `
    SELECT notification_preference
    FROM member
    WHERE id = ?`
  var notificationsStr string
  err := db.QueryRow(stmt, memberId).Scan(&notificationsStr)
  if !checkError(w, err) {
    return
  }

  var notifications = notificationsStruct{}
  err = json.Unmarshal([]byte(notificationsStr), &notifications)
  if !checkError(w, err) {
    return
  }

  respondJSON(w, http.StatusOK, map[string]notificationsStruct{"notifications": notifications})
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

  notificationsStr, err := json.Marshal(notifications)
  if !checkError(w, err) {
    return
  }

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