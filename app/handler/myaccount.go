package handler

import (
//  "encoding/json"
  "net/http"

//  "github.com/go-chi/chi"
)

func GetMyAccountSummary(w http.ResponseWriter, r *http.Request) {
  // get access token from header
  // get provider from query param
  uniqueID, err = getUniqueID(r)
  if err != nil {
    return err
  }

  userSummary, err = getUserSummary(uniqueID)
  //err includes unregistered
  if err != nil {
    return err
  }

  return userSummary

  w.Header().Set("Content-Type", "text")
  w.WriteHeader(http.StatusOK)
  w.Write([]byte("hello wOrld"))
}
