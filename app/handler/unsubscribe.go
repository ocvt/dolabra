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

/*
 * Remove email from quick signups and clear the member's notification
 * preferences
 */
func unsubscribeAll(email string) error {
	notificationsArr, err := json.Marshal(notificationsStruct{})
	if err != nil {
		return err
	}
	notificationsStr := string(notificationsArr)

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	stmt := `
		DELETE FROM quick_signup
		WHERE email = ?`
	_, err = tx.ExecContext(ctx, stmt, email)
	if err != nil {
		tx.Rollback()
		return err
	}

	stmt = `
		UPDATE member
		SET notification_preference = ?
		WHERE email = ?`
	_, err = tx.ExecContext(ctx, stmt, notificationsStr, email)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
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

	if !checkError(w, unsubscribeAll(email.Email)) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}

/*
 * RFC 8058 one-click unsubscribe target for the List-Unsubscribe header;
 * mailbox providers POST here directly with no body we care about
 */
func PostUnsubscribeOneClick(w http.ResponseWriter, r *http.Request) {
	// RFC 8058 requires this exact form body on real user-initiated
	// unsubscribes; mail security scanners blindly POSTing links from
	// scanned emails don't send it
	if r.FormValue("List-Unsubscribe") != "One-Click" {
		respondError(w, http.StatusBadRequest, "Missing one-click body")
		return
	}

	email := r.URL.Query().Get("email")
	sig := r.URL.Query().Get("sig")

	if !hmac.Equal([]byte(sig), []byte(unsubscribeSig(email))) {
		respondError(w, http.StatusForbidden, "Invalid unsubscribe link")
		return
	}

	if !checkError(w, unsubscribeAll(email)) {
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}
