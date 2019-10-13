package handler

import (
  //  "encoding/json"
  "log"
  "net/http"
  //  "github.com/go-chi/chi"
)

func GetMyAccountSummary(w http.ResponseWriter, r *http.Request) {
  // get access token from header
  // get provider from query param
  //  uniqueID, err = getUniqueID(r)
  //  if err != nil {
  //    return err
  //  }
  //
  //  userSummary, err = getUserSummary(uniqueID)
  //  //err includes unregistered
  //  if err != nil {
  //    return err
  //  }
  //
  //  return userSummary
  //
  IdPUserId := r.Context().Value("IdPUserId")
  w.Header().Set("Content-Type", "text")
  w.WriteHeader(http.StatusOK)
  if IdPUserId != nil {
    log.Printf("IdPUserId: " + IdPUserId.(string))
  }
}
