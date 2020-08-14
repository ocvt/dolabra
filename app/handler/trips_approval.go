package handler

import (
	"container/list"
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

const GUID_LENGTH = 64

type approvalStruct struct {
	TripId int    `json:"tripId"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type approverStruct struct {
	/* Managed server side */
	// from member table
	CellNumber string `json:"cellNumber,omitempty"`
	Email      string `json:"email,omitempty"`
	FirstName  string `json:"firstName,omitempty"`
	LastName   string `json:"lastName,omitempty"`
	/* Required fields for creating a trip */
	MemberId       int    `json:"memberId"`
	ExpireDatetime string `json:"expireDatetime"`
}

/* HELPERS */
func approveNewTrip(w http.ResponseWriter, tripId int) bool {
	/* Get trip approver member ids */
	stmt := `
		SELECT member_id
		FROM trip_approver`
	rows, err := db.Query(stmt)
	if !checkError(w, err) {
		return false
	}
	defer rows.Close()

	memberIds := list.New()
	for rows.Next() {
		var memberId int
		err = rows.Scan(&memberId)
		if !checkError(w, err) {
			return false
		}
		memberIds.PushBack(memberId)
	}

	err = rows.Err()
	if !checkError(w, err) {
		return false
	}

	/********************************/

	/* Create email for each trip approver */
	for m := memberIds.Front(); m != nil; m = m.Next() {
		memberId := m.Value.(int)
		guidCode := generateCode(GUID_LENGTH)

		// Create new guid entry
		stmt = `
			INSERT INTO trip_approval_guid (
				code,
				member_id,
				trip_id,
				status)
			VALUES (?, ?, ?, '')`
		_, err = db.Exec(stmt, guidCode, memberId, tripId)
		if !checkError(w, err) {
			return false
		}

		if !stageEmailTripApproval(w, tripId, memberId, guidCode) {
			return false
		}
	}
	/***************************************/

	return true
}

/* MAIN FUNCTIONS */
func PatchTripApproval(w http.ResponseWriter, r *http.Request) {
	guidCode := chi.URLParam(r, "guidCode")
	action := chi.URLParam(r, "action")

	// Get relevant trip id
	stmt := `
		SELECT trip_id
		FROM trip_approval_guid
		WHERE code = ?`
	var tripId int
	err := db.QueryRow(stmt, guidCode).Scan(&tripId)
	if !checkError(w, err) {
		return
	}

	// Check if canceled
	stmt = `
		SELECT cancel
		FROM trip
		WHERE id = ?`
	var canceled bool
	err = db.QueryRow(stmt, tripId).Scan(&canceled)
	if !checkError(w, err) {
		return
	}

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if !checkError(w, err) {
		return
	}

	if canceled {
		stmt = `
			INSERT INTO trip_approval_guid (
				code,
				member_id,
				trip_id,
				status)
			VALUES ('0', '0', ?, 'cancel')`
		_, err = tx.ExecContext(ctx, stmt, tripId)
		if !checkError(w, err) {
			tx.Rollback()
			return
		}
	}

	// Check for approval already exists
	stmt = `
		SELECT EXISTS (
			SELECT 1
			FROM trip_approval_guid
			WHERE trip_id = ? AND status != '')`
	var approvalExists bool
	err = tx.QueryRowContext(ctx, stmt, tripId).Scan(&approvalExists)
	if !checkError(w, err) {
		tx.Rollback()
		return
	}

	// Build approval response
	approval := approvalStruct{
		TripId: tripId,
	}

	if approvalExists {
		// Lookup member + status if already approved
		stmt = `
			SELECT
				member_id,
				status
			FROM trip_approval_guid
			WHERE trip_id = ? AND status != ''`
		var memberId int
		err = tx.QueryRowContext(ctx, stmt, tripId).Scan(&memberId, &approval.Status)
		if !checkError(w, err) {
			tx.Rollback()
			return
		}

		// Set response
		memberName, ok := dbGetMemberName(w, memberId)
		if !ok {
			tx.Rollback()
			return
		}
		approval.Reason = fmt.Sprintf("Trip already has status %s by %s",
			approval.Status, memberName)
	} else {
		// Set response
		if action == "approve" {
			if !stageEmailNewTrip(w, tripId) {
				tx.Rollback()
				return
			}
			approval.Reason = "Successfully approved trip"
			approval.Status = action
		} else if action == "deny" {
			approval.Reason = "Successfully denied trip"
			approval.Status = action
		}

		// Update db
		stmt = `
			UPDATE trip_approval_guid
			SET status = ?
			WHERE code = ?`
		_, err := tx.ExecContext(ctx, stmt, action, guidCode)
		if !checkError(w, err) {
			tx.Rollback()
			return
		}
	}

	err = tx.Commit()
	if !checkError(w, err) {
		return
	}

	respondJSON(w, http.StatusOK, approval)
}
