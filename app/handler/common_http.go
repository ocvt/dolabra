package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/crypto/nacl/secretbox"

	"gitlab.com/ocvt/dolabra/utils"
)

/************************ COOKIES ************************/
// Key for encrypting cookies
var key [32]byte

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
func getCookie(r *http.Request, name string, payload interface{}) error {
	payloadCookie, err := r.Cookie(name)
	if err != nil {
		return err
	}

	encodedB64Payload := payloadCookie.Value
	encryptedPayload, err := base64.StdEncoding.DecodeString(encodedB64Payload)
	if err != nil {
		return err
	}
	if len(encryptedPayload) < 24 {
		return &errInvalidPayload{"Payload is invalid length"}
	}

	// Get nonce and decrypt
	var nonce [24]byte
	copy(nonce[:], encryptedPayload[:24])
	encodedJSONPayload, ok := secretbox.Open(nil, encryptedPayload[24:], &nonce, &key)
	if !ok {
		return &errInvalidPayload{"Payload failed to decrypt"}
	}

	err = json.Unmarshal(encodedJSONPayload, payload)
	if err != nil {
		log.Fatal("Failed to unmarshal decrypted payload", err)
	}

	return nil
}

// Set encrypted cookie
func setCookie(w http.ResponseWriter, name string, payload interface{}) {
	encodedJSONPayload, err := json.Marshal(payload)
	if err != nil {
		log.Fatal("Failed to marshal payload", err)
	}

	// Create nonce, append to front, and encrypt
	var nonce [24]byte
	_, err = rand.Read(nonce[:])
	if err != nil {
		log.Fatal("Failed to generate nonce", err)
	}
	encryptedPayload := secretbox.Seal(nonce[:], encodedJSONPayload, &nonce, &key)
	encodedB64Payload := base64.StdEncoding.EncodeToString(encryptedPayload)

	// Set cookie
	cookieDomain := utils.GetConfig().CookieDomain
	cookie := http.Cookie{
		Domain: cookieDomain,
		Name:   name,
		Path:   "/",
		Value:  encodedB64Payload,
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
