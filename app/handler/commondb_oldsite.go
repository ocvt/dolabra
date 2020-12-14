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
			AND name = ?
			AND birth_year = ?
			AND pronouns = ?`
	err := db.QueryRow(stmt,
		oldMember.Email,
		oldMember.Name,
		oldMember.Birthyear,
		oldMember.Pronouns).Scan(
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
