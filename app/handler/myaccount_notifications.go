package handler

import (
	"encoding/json"
	"net/http"
)

type notificationsStruct struct {
	GeneralAnnouncements        bool `json:"GENERAL_ANNOUNCEMENTS"`
	TripBackpacking             bool `json:"TRIP_BACKPACKING"`
	TripBiking                  bool `json:"TRIP_BIKING"`
	TripCamping                 bool `json:"TRIP_CAMPING"`
	TripClimbing                bool `json:"TRIP_CLIMBING"`
	TripDayhike                 bool `json:"TRIP_DAYHIKE"`
	TripLaserTag                bool `json:"TRIP_LASER_TAG"`
	TripOfficialMeeting         bool `json:"TRIP_OFFICIAL_MEETING"`
	TripOther                   bool `json:"TRIP_OTHER"`
	TripRaftingCanoeingKayaking bool `json:"TRIP_RAFTING_CANOEING_KAYAKING"`
	TripRoadTrip                bool `json:"TRIP_ROAD_TRIP"`
	TripSkiingSnowboarding      bool `json:"TRIP_SKIING_SNOWBOARDING"`
	TripSnowOther               bool `json:"TRIP_SNOW_OTHER"`
	TripSocial                  bool `json:"TRIP_SOCIAL"`
	TripSpecialEvent            bool `json:"TRIP_SPECIAL_EVENT"`
	TripTeamSportsMisc          bool `json:"TRIP_TEAM_SPORTS_MISC"`
	TripWaterOther              bool `json:"TRIP_WATER_OTHER"`
	TripWorkTrip                bool `json:"TRIP_WORK_TRIP"`
}

func setAllPreferences() *notificationsStruct {
	return &notificationsStruct{
		GeneralAnnouncements:        true,
		TripBackpacking:             true,
		TripBiking:                  true,
		TripCamping:                 true,
		TripClimbing:                true,
		TripDayhike:                 true,
		TripLaserTag:                true,
		TripOfficialMeeting:         true,
		TripOther:                   true,
		TripRaftingCanoeingKayaking: true,
		TripRoadTrip:                true,
		TripSkiingSnowboarding:      true,
		TripSnowOther:               true,
		TripSocial:                  true,
		TripSpecialEvent:            true,
		TripTeamSportsMisc:          true,
		TripWaterOther:              true,
		TripWorkTrip:                true,
	}
}

func GetMyAccountNotifications(w http.ResponseWriter, r *http.Request) {
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId
	memberId, ok := dbGetActiveMemberId(w, sub)
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
	sub, ok := checkLogin(w, r)
	if !ok {
		return
	}

	// Get memberId
	memberId, ok := dbGetActiveMemberId(w, sub)
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
