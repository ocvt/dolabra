package handler

import (
  "encoding/json"
  "net/http"
)

type quicksignupStruct struct {
  Email string `json:"email"`
}

func PostQuicksignup(w http.ResponseWriter, r *http.Request) {
  // Get request body
  decoder := json.NewDecoder(r.Body)
  decoder.DisallowUnknownFields()
  var quicksignup quicksignupStruct
  err := decoder.Decode(&quicksignup)
  if err != nil {
    respondError(w, http.StatusBadRequest, err.Error())
    return
  }

  stmt := `
    INSERT INTO quick_signup (
      create_datetime,
      email)
    VALUES (datetime('now'), ?)`
  _, err = db.Exec(stmt, quicksignup.Email)
  if !checkError(w, err) {
    return
  }

  respondJSON(w, http.StatusNoContent, nil)
}
