package handler

import (
  "net/http"
)

func GetWebtoolsOfficers(w http.ResponseWriter, r *http.Request) {
  _, subject, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Get memberId, tripId, signupId
  memberId, ok := dbGetActiveMemberId(w, subject)
  if !ok {
    return
  }

  // Permissions
  if !dbEnsureOfficer(w, memberId) {
    return
  }

  stmt := `
    SELECT
      member.id,
      member.email,
      member.first_name,
      member.last_name,
      member.cell_number,
      officer.expire_datetime,
      officer.position
    FROM member
    INNER JOIN officer ON officer.member_id = member.id`
  rows, err := db.Query(stmt)
  if !checkError(w, err) {
    return
  }
  defer rows.Close()

  var officers = []*map[string]string{}
  for rows.Next() {
    var officerId, email, firstName, lastName,
        cellNumber, expireDatetime, position string
    err = rows.Scan(
      &officerId,
      &email,
      &firstName,
      &lastName,
      &cellNumber,
      &expireDatetime,
      &position)

    officers = append(officers, &map[string]string{
        "memberId": officerId, "email": email, "firstName": firstName,
        "lastName": lastName, "cellNumber": cellNumber,
        "expireDatetime": expireDatetime, "position": position})
    if !checkError(w, err) {
      return
    }
  }

  err = rows.Err()
  if !checkError(w, err) {
    return
  }

  respondJSON(w, http.StatusOK, map[string][]*map[string]string{"officers": officers})
}

func DeleteWebtoolsOfficers(w http.ResponseWriter, r *http.Request) {
  _, subject, ok := checkLogin(w, r)
  if !ok {
    return
  }

  // Get memberId, tripId, signupId
  memberId, ok := dbGetActiveMemberId(w, subject)
  if !ok {
    return
  }
  officerId, ok := checkURLParam(w, r, "memberId")
  if !ok {
    return
  }

  // Permissions
  if memberId == officerId {
    respondError(w, http.StatusBadRequest, "Cannot remove yourself from officers.")
    return
  }
  // TODO prevent deletion of officers with more permissions (ie treasurer/maint officer can't delete admin officer)

  stmt := `
   DELETE FROM officer
   WHERE member_id = ?`
  _, err := db.Exec(stmt, officerId)
  if !checkError(w, err) {
    return
  }

  respondJSON(w, http.StatusNoContent, nil)
}
