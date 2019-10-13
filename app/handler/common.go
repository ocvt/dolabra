package handler

import (
  "encoding/json"
  "net/http"
)

// Properly return JSON response
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
  response, err := json.Marshal(payload)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte("Error encountered marshalling JSON payload: " + err.Error()))
    return
  }
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)
  w.Write([]byte(response))
}

// Return error message as JSON
func respondError(w http.ResponseWriter, code int, message string) {
  respondJSON(w, code, map[string]string{"error": message})
}
