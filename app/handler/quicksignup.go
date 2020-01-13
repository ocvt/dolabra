package handler

import (
	"encoding/json"
	"net/http"
)

type simpleEmailStruct struct {
	Email string `json:"email"`
}

func PostQuicksignup(w http.ResponseWriter, r *http.Request) {
	// Get request body
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var email simpleEmailStruct
	err := decoder.Decode(&email)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	stmt := `
    INSERT OR REPLACE INTO quick_signup (
      create_datetime,
      expire_datetime,
      email)
    VALUES (datetime('now'), datetime('now', '+6 months'), ?)`
  _, err = db.Exec(stmt, email.Email) // sqlvet: ignore
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}
