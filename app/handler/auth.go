package handler

import (
  "context"
  "net/http"

  "golang.org/x/oauth2"

  "gitlab.com/ocvt/dolabra/utils"
)

func GoogleLogin(w http.ResponseWriter, r *http.Request) {
  promptParam := oauth2.SetAuthURLParam("prompt", "consent select_account")
  url := googleOAuthConfig.AuthCodeURL("state", promptParam)
  http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallback(w http.ResponseWriter, r *http.Request) {
  // Get token and put in cookie
  token, err := googleOAuthConfig.Exchange(context.Background(), r.FormValue("code"))
  if err != nil {
    respondError(w, http.StatusBadRequest, err.Error())
  }

  setCookie(w, "idp", map[string]string{"idp": "GOOGLE"})
  setCookie(w, "token", token)

  config := utils.GetConfig()
  http.Redirect(w, r, config.ClientUrl + "/login", http.StatusTemporaryRedirect)
}
