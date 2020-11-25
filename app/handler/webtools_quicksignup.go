package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
)

// Only used for viewing quicksignups. New individual emails are added via the /quicksignup endpoint.
type quicksignupStruct struct {
	CreateDatetime string `json:"createDatetime"`
	ExpireDatetime string `json:"expireDatetime"`
	Email          string `json:"email"`
}

// Only used to bulk add/remove quicksignups
type quicksignupBulkStruct struct {
	Emails []string `json:"emails"`
}

func GetWebtoolsQuicksignups(w http.ResponseWriter, r *http.Request) {
	stmt := `
		SELECT
			create_datetime,
			expire_datetime,
			email
		FROM quick_signup
		ORDER BY datetime(create_datetime) DESC`
	rows, err := db.Query(stmt)
	if !checkError(w, err) {
		return
	}
	defer rows.Close()

	var quicksignups = []*quicksignupStruct{}
	i := 0
	for rows.Next() {
		quicksignups = append(quicksignups, &quicksignupStruct{})
		err = rows.Scan(
			&quicksignups[i].CreateDatetime,
			&quicksignups[i].ExpireDatetime,
			&quicksignups[i].Email)
		if !checkError(w, err) {
			return
		}
		i++
	}

	err = rows.Err()
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, map[string][]*quicksignupStruct{"quicksignups": quicksignups})
}

func PostWebtoolsQuicksignups(w http.ResponseWriter, r *http.Request) {
	action := chi.URLParam(r, "action")

	// Get request body
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var emails quicksignupBulkStruct
	err := decoder.Decode(&emails)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if action == "add" {
		for i := 0; i < len(emails.Emails); i++ {
			stmt := `
				INSERT OR REPLACE INTO quick_signup (
					create_datetime,
					expire_datetime,
					email)
				VALUES (datetime('now'), datetime('now', '+6 months'), ?)`
			_, err := db.Exec(stmt, emails.Emails[i]) // sqlvet: ignore
			if !checkError(w, err) {
				return
			}
		}
	} else if action == "remove" {
		for i := 0; i < len(emails.Emails); i++ {
			stmt := `
				DELETE FROM quick_signup
				WHERE email = ?`
			_, err := db.Exec(stmt, emails.Emails[i])
			if !checkError(w, err) {
				return
			}
		}
	} else {
		respondError(w, http.StatusBadRequest, "Invalid path action")
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}
