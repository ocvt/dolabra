package handler

import (
	"context"
	"encoding/json"
	"net/http"
)

func PostUnsubscribeAll(w http.ResponseWriter, r *http.Request) {
	// Get request body
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var email simpleEmailStruct
	err := decoder.Decode(&email)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	notificationsArr, err := json.Marshal(notificationsStruct{})
	if !checkError(w, err) {
		return
	}
	notificationsStr := string(notificationsArr)

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if !checkError(w, err) {
		return
	}

	stmt := `
		DELETE FROM quick_signup
		WHERE email = ?`
	_, err = tx.ExecContext(ctx, stmt, email.Email)
	if !checkError(w, err) {
		tx.Rollback()
		return
	}

	stmt = `
		UPDATE member
		SET notification_preference = ?
		WHERE email = ?`
	_, err = tx.ExecContext(ctx, stmt, notificationsStr, email.Email)
	if !checkError(w, err) {
		tx.Rollback()
		return
	}

	err = tx.Commit()
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}
