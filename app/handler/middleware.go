package handler

import (
	"log"
	//  "encoding/json"
	"net/http"
	"strings"
	//  "github.com/go-chi/chi"
)

func ProcessClientAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		authHeaderList := strings.Split(authHeader, "Bearer ")
		jwt := authHeaderList[1]
		log.Printf(jwt)
		log.Printf("middleware called!")
		next.ServeHTTP(w, r)
	})
}
