package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/ocvt/dolabra/utils"
)

/************************ COOKIES ************************/
// Key for signing JWTs
var key []byte

// Create JWT with given sub
func createJWT(w http.ResponseWriter, sub string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().UTC().Add(time.Hour * 3).Unix(),
		Subject:   sub,
	})

	tokenStr, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

// Delete Cookie
func deleteCookie(w http.ResponseWriter, name string) {
	cookieDomain := utils.GetConfig().CookieDomain
	cookie := http.Cookie{
		Domain: cookieDomain,
		MaxAge: -1,
		Name:   name,
		Path:   "/",
		Value:  "",
	}
	http.SetCookie(w, &cookie)
}

// Get cookie and decrypt
func getCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func setCookie(w http.ResponseWriter, name string, payload string) {
	cookieDomain := utils.GetConfig().CookieDomain
	cookie := http.Cookie{
		Domain: cookieDomain,
		Name:   name,
		Path:   "/",
		Value:  payload,
	}
	http.SetCookie(w, &cookie)
}

/************************ COOKIES ************************/

// Return error message as JSON
func respondError(w http.ResponseWriter, code int, message string) {
	respondJSON(w, code, map[string]string{"error": message})
}

// Properly return JSON response
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	if payload == nil {
		w.WriteHeader(status)
		return
	}

	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte("Error marshalling JSON payload: " + err.Error()))
		if err != nil {
			log.Fatal("Failed writing response: ", err.Error())
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write([]byte(response))
	if err != nil {
		log.Fatal("Failed writing response: ", err.Error())
	}
}
