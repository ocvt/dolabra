package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
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
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
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
