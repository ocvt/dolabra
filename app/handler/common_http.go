package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"net/mail"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/ocvt/dolabra/utils"
)

// HMAC signature binding an unsubscribe link to an email address
func unsubscribeSig(email string) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte("unsubscribe:" + email))
	return hex.EncodeToString(mac.Sum(nil))
}

// Returns true if email is a plain valid address (no display name)
func validEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}

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
		Domain:   cookieDomain,
		HttpOnly: true,
		MaxAge:   -1,
		Name:     name,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Secure:   os.Getenv("DEV") != "1",
		Value:    "",
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
		Domain:   cookieDomain,
		HttpOnly: true,
		Name:     name,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Secure:   os.Getenv("DEV") != "1",
		Value:    payload,
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
