package handler

import (
  "context"
  "encoding/base64"
  "encoding/json"
  "io/ioutil"
  "net/http"
  "strings"
  "time"
)

func ProcessClientAuth(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
      next.ServeHTTP(w, r)
    } else {
      // Extract OCVT JWT if present and valid
      authHeaderList := strings.Split(authHeader, "Bearer ")
      if len(authHeaderList) != 2 {
        respondError(w, http.StatusBadRequest, "Authorization header formatted incorrectly")
        return
      }

      ocvtJwt := authHeaderList[1]
      sessionDataStr, err := base64.StdEncoding.DecodeString(ocvtJwt)
      if err != nil {
        respondError(w, http.StatusBadRequest, "Authorization token failed to decode: " + err.Error())
        return
      }

      // Convert Session Data to proper map
      var sessionData map[string]string
      err = json.Unmarshal([]byte(sessionDataStr), &sessionData)
      if err != nil {
        respondError(w, http.StatusBadRequest, "Session data formatted incorrectly: " + err.Error())
        return
      }
      if sessionData["idp"] == "" {
        respondError(w, http.StatusBadRequest, "Session data missing \"idp\" field")
        return
      }
      if sessionData["access_token"] == "" {
        respondError(w, http.StatusBadRequest, "Session data missing \"access_token\" field")
        return
      }

      // Use access token to get user id from idp
      // TODO make this more modular when we add more IDPs
      var userinfoEndpoint string
      if sessionData["idp"] == "GOOGLE" {
        // Endpoint to get user data. All we care about is the sub field
        userinfoEndpoint = "https://openidconnect.googleapis.com/v1/userinfo"

        // Setup http client
        OIDCClient := http.Client{
          Timeout: time.Second * 2,
        }

        OIDCRequest, err := http.NewRequest(http.MethodGet, userinfoEndpoint, nil)
        if err != nil {
          respondError(w, http.StatusInternalServerError, "OIDCRequest failed to create: " + err.Error())
          return
        }

        // Send request, specifying access token in header
        OIDCRequest.Header.Set("User-Agent", "OCVT Api")
        OIDCRequest.Header.Set("Authorization", "Bearer " + sessionData["access_token"])
        OIDCResponse, err := OIDCClient.Do(OIDCRequest)
        if err != nil {
          respondError(w, http.StatusInternalServerError, "OIDCResponse failed to receive data: " + err.Error())
          return
        }

        // Process response
        OIDCBody, err := ioutil.ReadAll(OIDCResponse.Body)
        if err != nil {
          respondError(w, http.StatusInternalServerError, "OIDCBody failed to read data: " + err.Error())
          return
        }

        var OIDCUserinfo map[string]string
        err = json.Unmarshal(OIDCBody, &OIDCUserinfo)
        if err != nil {
          respondError(w, http.StatusInternalServerError, "OIDCUserinfo failed to convert response to JSON: " + err.Error())
          return
        }

        // Check for Invalid Credenti
        if OIDCResponse.StatusCode != 200 {
          if OIDCUserinfo["error_description"] == "" {
            respondError(w, OIDCResponse.StatusCode, "Unknown error requesting user Id from Google")
          }
          respondError(w, OIDCResponse.StatusCode, "Error requesting user Id from Google: " + OIDCUserinfo["error_description"])
          return
        }

        // Store idp and user id for additional processing
        ctx := context.WithValue(r.Context(), "IdP", sessionData["idp"])
        ctx = context.WithValue(r.Context(), "IdPUserId", OIDCUserinfo["sub"])
        next.ServeHTTP(w, r.WithContext(ctx))
      } else {
        respondError(w, http.StatusBadRequest, "Invalid IDP: " + sessionData["idp"])
      }
    }
  })
}
