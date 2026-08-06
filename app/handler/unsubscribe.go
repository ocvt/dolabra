package handler

import (
	"context"
	"crypto/hmac"
	"encoding/json"
	"net/http"
)

type unsubscribeStruct struct {
	Email string `json:"email"`
	Sig   string `json:"sig"`
}

func PostUnsubscribeAll(w http.ResponseWriter, r *http.Request) {
	// Get request body
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var email unsubscribeStruct
	err := decoder.Decode(&email)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Require the HMAC signature from the emailed unsubscribe link so only
	// the recipient can unsubscribe their address
	if !hmac.Equal([]byte(email.Sig), []byte(unsubscribeSig(email.Email))) {
		respondError(w, http.StatusForbidden, "Invalid unsubscribe link")
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
