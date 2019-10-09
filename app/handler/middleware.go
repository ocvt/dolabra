package handler

import (
//  "encoding/json"
  "net/http"

//  "github.com/go-chi/chi"
)

func recoverHandler(next http.Handler) http.Handler {
  fn := func(w http.ResponseWriter, r *http.Request) {
    defer func() {
      if err := recover(); err != nil {
        log.Printf("panic: %+v", err)
        http.Error(w, http.StatusText(500), 500)
      }
    }()

    next.ServeHTTP(w, r)
  }

  return http.HandlerFunc(fn)
}

func ProcessClientAuth(next http.Handler) http.Handler
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
