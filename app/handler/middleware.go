package handler

import (
	//	"bytes"
	"context"
	//	"encoding/json"
	"fmt"
	//	"io"
	//	"io/ioutil"
	"net/http"
	//	"strings"

	"github.com/golang-jwt/jwt"
)

func deleteAuthCookies(w http.ResponseWriter) {
	deleteCookie(w, "DOLABRA_SESSION")
}

func EnsureOfficer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sub, ok := checkLogin(w, r)
		if !ok {
			return
		}

		// Get memberId
		memberId, ok := dbGetActiveMemberId(w, sub)
		if !ok {
			return
		}

		if !dbEnsureOfficer(w, memberId) {
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Get token from cookie and put user id in context
func ProcessClientAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sub, err := func(w http.ResponseWriter, r *http.Request) (string, error) {
			// Get JWT
			tokenStr, err := getCookie(r, "DOLABRA_SESSION")
			// Error indicates cookie does not exist
			if err != nil {
				return "", nil
			}

			// Parse JWT
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return key, nil
			})
			// Issue parsing JWT
			// Assume will only happen if user intentionally alters JWT
			if err != nil {
				deleteAuthCookies(w)
				return "", nil
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			// Issue looking up claims or invalid signature
			// Assume will only happen if user intentionally alters JWT
			if !ok || !token.Valid {
				deleteAuthCookies(w)
				return "", nil
			}

			// Expired
			err = claims.Valid()
			if err != nil {
				deleteAuthCookies(w)
				return "", err
			}

			return claims["sub"].(string), nil
		}(w, r)

		// Assume error is due to expired token
		if err != nil {
			respondError(w, http.StatusUnauthorized, err.Error())
			return
		}

		if sub == "" {
			respondError(w, http.StatusUnauthorized, "Member is not authenticated")
			return
		}

		ctx := context.WithValue(r.Context(), "sub", sub)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

//func ValidateInput(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		ok := func(w http.ResponseWriter, r *http.Request) bool {
//			specialPath := strings.HasSuffix(r.URL.Path, "/photos") ||
//				strings.HasSuffix(r.URL.Path, "/admin/mainphoto") ||
//				strings.HasSuffix(r.URL.Path, "/webtools/emails") ||
//				strings.HasSuffix(r.URL.Path, "/webtools/news")
//			if specialPath {
//				return true
//			}
//
//			bodyBytes, err := ioutil.ReadAll(r.Body)
//			if err == io.EOF {
//				return true
//			} else if !checkError(w, err) {
//				return false
//			}
//			body := string(bodyBytes)
//
//			// Attempt to convert to JSON
//			var input map[string]interface{}
//			err = json.Unmarshal(bodyBytes, &input)
//
//			// Not JSON
//			if err != nil {
//				newBody := strictHTML.Sanitize(body)
//				if string(newBody) != body {
//					respondError(w, http.StatusBadRequest, "HTTP body is not valid: "+string(body))
//					return false
//				}
//			}
//
//			// JSON, check each value
//			for k := range input {
//				if v, ok := input[k].(string); ok {
//					newValue := strictHTML.Sanitize(v)
//					if newValue != v {
//						respondError(w, http.StatusBadRequest, "HTTP body is not valid: "+v)
//						return false
//					}
//				}
//			}
//
//			r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
//			return true
//		}(w, r)
//
//		if !ok {
//			return
//		}
//
//		next.ServeHTTP(w, r)
//	})
//}

/* Misc */
func GetHealthCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusNoContent, nil)
}
