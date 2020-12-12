package handler

import (
	"container/list"
	"net/http"
)

/* Helper functions to migrate user data from the old site */

func lookupOldMember(w http.ResponseWriter, oldMember oldMemberStruct) (*memberStruct, *string, *list.List, bool) {
	var member memberStruct
	var notificationsStr string
	stmt := `
		SELECT *
		FROM oldsite_member
		WHERE email = ?
			AND first_name = ?
			AND last_name = ?
			AND birth_year = ?
			AND gender = ?`
	err := db.QueryRow(stmt,
		oldMember.Email,
		oldMember.FirstName,
		oldMember.LastName,
		oldMember.Birthyear,
		oldMember.Gender).Scan(
		&member.Id,
		&member.Email,
		&member.Name,
		&member.CreateDatetime,
		&member.CellNumber,
		&member.Pronouns,
		&member.Birthyear,
		&member.Active,
		&member.MedicalCond,
		&member.MedicalCondDesc,
		&member.PaidExpireDatetime,
		&member.ECName,
		&member.ECNumber,
		&member.ECRelationship,
		&notificationsStr)
	if err != nil && err.Error() == "sql: no rows in result set" {
		respondError(w, http.StatusBadRequest, "Account from old site cannot be found")
		return nil, nil, nil, false
	}

	// Convert gender to pronouns. Oldsite was limited to M/F only
	if member.Pronouns == "F" {
		member.Pronouns = "she/her"
	} else {
		member.Pronouns = "he/him"
	}

	// Uses oldsite member id (701XXXX)
	stmt = `
		SELECT *
		FROM oldsite_payment
		WHERE member_id = ?`
	rows, err := db.Query(stmt, member.Id)
	if !checkError(w, err) {
		return nil, nil, nil, false
	}
	defer rows.Close()

	payments := list.New()
	for rows.Next() {
		payment := paymentStruct{}
		err = rows.Scan(
			&payment.Id,
			&payment.CreateDatetime,
			&payment.EnteredById,
			&payment.Note,
			&payment.MemberId,
			&payment.StoreItemId,
			&payment.StoreItemCount,
			&payment.Amount,
			&payment.PaymentMethod,
			&payment.PaymentId,
			&payment.Completed)
		if !checkError(w, err) {
			return nil, nil, nil, false
		}
		payments.PushBack(payment)
	}

	return &member, &notificationsStr, payments, true
}
