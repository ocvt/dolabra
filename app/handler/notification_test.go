package handler

import "testing"

/*
 * Transactional mail must not carry unsubscribe headers: Gmail treats them as a
 * bulk signal and files the mail under Promotions, where approvers never see it.
 */
func TestIsBulkNotification(t *testing.T) {
	transactional := []string{
		"DIRECT",
		"TRIP_APPROVAL",
		"TRIP_ALERT_ATTEND",
		"TRIP_ALERT_BOOT",
		"TRIP_ALERT_CANCEL",
		"TRIP_ALERT_FORCE",
		"TRIP_ALERT_LEADER",
		"TRIP_ALERT_WAIT",
		"TRIP_MESSAGE_ATTEND",
		"TRIP_MESSAGE_DIRECT",
		"TRIP_MESSAGE_NOTIFY",
		"TRIP_MESSAGE_WAIT",
	}
	for _, id := range transactional {
		if isBulkNotification(id) {
			t.Errorf("%s is transactional, must not be treated as bulk", id)
		}
	}

	bulk := []string{
		"GENERAL_ANNOUNCEMENTS",
		"TRIP_BACKPACKING",
		"TRIP_BIKING",
		"TRIP_CAMPING",
		"TRIP_CLIMBING",
		"TRIP_DAYHIKE",
		"TRIP_LASER_TAG",
		"TRIP_OFFICIAL_MEETING",
		"TRIP_OTHER",
		"TRIP_RAFTING_CANOEING_KAYAKING",
		"TRIP_ROAD_TRIP",
		"TRIP_SKIING_SNOWBOARDING",
		"TRIP_SNOW_OTHER",
		"TRIP_SOCIAL",
		"TRIP_SPECIAL_EVENT",
		"TRIP_TEAM_SPORTS_MISC",
		"TRIP_WATER_OTHER",
		"TRIP_WORK_TRIP",
	}
	for _, id := range bulk {
		if !isBulkNotification(id) {
			t.Errorf("%s is a member-opt-in announcement, must be treated as bulk", id)
		}
	}
}
