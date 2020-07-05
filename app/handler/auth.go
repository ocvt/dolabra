package handler

import (
	"context"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	oidcgoogle "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

	"gitlab.com/ocvt/dolabra/utils"
)

var googleOAuthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	RedirectURL:  utils.GetConfig().ApiUrl + "/auth/google/callback",
	Scopes:       []string{oidcgoogle.UserinfoProfileScope},
	Endpoint:     google.Endpoint,
}

const SUB_LENGTH = 16

/* HELPERS */
func processIdp(w http.ResponseWriter, idp string, idpSub string) bool {
	exists, err := dbIsMemberWithIdp(w, idp, idpSub)
	if err != nil {
		return false
	}

	var sub string
	var ok bool
	if exists {
		sub, ok = dbGetMemberSubWithIdp(w, idp, idpSub)
		if !ok {
			return false
		}
	} else {
		exists := true
		for exists {
			sub = generateCode(SUB_LENGTH)
			exists, err = dbIsMemberWithSub(w, sub)
			if err != nil {
				return false
			}
		}

		stmt := `
      INSERT INTO auth(
        member_id,
        sub,
        idp,
        idp_sub)
		  VALUES (0, ?, ?, ?)`
		_, err = db.Exec(stmt, sub, idp, idpSub)
		if !checkError(w, err) {
			return false
		}
	}

	token, err := createJWT(w, sub)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return false
	}

	setCookie(w, "DOLABRA_SESSION", token)
	return true
}

/* MAIN FUNCTIONS */
func DevLogin(w http.ResponseWriter, r *http.Request) {
	idpSub := chi.URLParam(r, "sub")

	ok := processIdp(w, "DEV", idpSub)
	if !ok {
		return
	}

	http.Redirect(w, r, r.URL.Query().Get("state"), http.StatusTemporaryRedirect)
}

func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Get access token
	accessToken, err := googleOAuthConfig.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create oauth2 service from access token
	service, err := oidcgoogle.NewService(context.Background(), option.WithTokenSource(googleOAuthConfig.TokenSource(context.Background(), accessToken)))
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get userinfo sub claim
	response, err := service.Userinfo.Get().Do()
	if err != nil {
		deleteAuthCookies(w)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Process sub
	ok := processIdp(w, "GOOGLE", response.Id)
	if !ok {
		return
	}

	http.Redirect(w, r, r.URL.Query().Get("state"), http.StatusTemporaryRedirect)
}

func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	promptParam := oauth2.SetAuthURLParam("prompt", "consent select_account")
	url := googleOAuthConfig.AuthCodeURL(r.URL.Query().Get("state"), promptParam)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
