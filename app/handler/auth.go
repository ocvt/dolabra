package handler

import (
  "fmt"
  "context"
  "net/http"
  "os"

  "github.com/go-chi/chi"
  "golang.org/x/oauth2"
  "golang.org/x/oauth2/google"
  oidcgoogle "google.golang.org/api/oauth2/v2"

  "gitlab.com/ocvt/dolabra/utils"
)

// Google OAuth config
var googleOAuthConfig = &oauth2.Config{
    ClientID: os.Getenv("GOOGLE_CLIENT_ID"),
    ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
    RedirectURL: "https://ocvt.club/auth/google/callback",
    Scopes: []string{oidcgoogle.UserinfoProfileScope},
    Endpoint: google.Endpoint,
}

func DevLogin(w http.ResponseWriter, r *http.Request) {
  subject := chi.URLParam(r, "subject")

  setCookie(w, "idp", map[string]string{"idp": "DEV"})
  setCookie(w, "token", map[string]string{"token": subject})

  config := utils.GetConfig()
  http.Redirect(w, r, config.ClientUrl + "/login", http.StatusTemporaryRedirect)
}

func GoogleLogin(w http.ResponseWriter, r *http.Request) {
  fmt.Printf("URL: %s\n", r.URL.Scheme)
  promptParam := oauth2.SetAuthURLParam("prompt", "consent select_account")
  url := googleOAuthConfig.AuthCodeURL("state", promptParam)
  http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallback(w http.ResponseWriter, r *http.Request) {
  // Get token and put in cookie
  token, err := googleOAuthConfig.Exchange(context.Background(), r.FormValue("code"))
  if err != nil {
    respondError(w, http.StatusBadRequest, err.Error())
    return
  }

  setCookie(w, "idp", map[string]string{"idp": "GOOGLE"})
  setCookie(w, "token", token)

  config := utils.GetConfig()
  http.Redirect(w, r, config.ClientUrl + "/login", http.StatusTemporaryRedirect)
}
