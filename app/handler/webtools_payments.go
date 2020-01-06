package handler

import (
  "net/http"
)


type paymentStruct struct {
  /* Managed Server side */
  Id int `json:"id,omitempty"`
  CreateDatetime string `json:"createDatetime,omitempty"`
  EnteredById int `json:"enteredById,omitempty"`
  PaymentMethod string `json:"paymentMethod,omitempty"`
  PaymentId string `json:"paymentId,omitempty"`
  // Not present in payment of store_code db, only for client view
  EnteredByName string `json:"enteredByName,omitempty"`
  MemberName string `json:"memberName,omitempty"`
  // Only used in payment db
  MemberId int `json:"memberId,omitempty"`
  /* Required fields for any payment or code generation */
  Note string `json:"note"`
  StoreItemId string `json:"storeItemId"`
  StoreItemCount int `json:"storeItemCount"`
  Amount int `json:"amount"`
  Completed bool `json:"completed"`
}

func GetWebtoolsPayments(w http.ResponseWriter, r *http.Request) {
  stmt := `
    SELECT *
    FROM payment
    ORDER BY datetime(create_datetim) DESC`
  rows, err := db.Query(stmt)
  if !checkError(w, err) {
    return
  }
  defer rows.Close()

  var payments = []*paymentStruct{}
  i := 0
  for rows.Next() {
    payments = append(payments, &paymentStruct{})
    err = rows.Scan(
      &payments[i].Id,
      &payments[i].CreateDatetime,
      &payments[i].EnteredById,
      &payments[i].Note,
      &payments[i].MemberId,
      &payments[i].StoreItemId,
      &payments[i].StoreItemCount,
      &payments[i].Amount,
      &payments[i].PaymentMethod,
      &payments[i].PaymentId,
      &payments[i].Completed)
    if !checkError(w, err) {
      return
    }

    enteredByName, ok := dbGetMemberName(w, payments[i].EnteredById)
    if !ok {
      return
    }
    memberName, ok := dbGetMemberName(w, payments[i].MemberId)
    if !ok {
      return
    }

    payments[i].EnteredByName = enteredByName
    payments[i].MemberName = memberName

    i++
  }

  err = rows.Err()
  if !checkError(w, err) {
    return
  }

  respondJSON(w, http.StatusOK, map[string][]*paymentStruct{"payments": payments})
}
