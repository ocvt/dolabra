package handler

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/go-chi/chi"
	"github.com/ocvt/dolabra/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	oidcgoogle "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

var googleOAuthConfig = &oauth2.Config{
	ClientID:     utils.GetConfig().GoogleClientId,
	ClientSecret: utils.GetConfig().GoogleClientSecret,
	RedirectURL:  utils.GetConfig().ApiUrl + "/auth/google/callback",
	Scopes:       []string{oidcgoogle.UserinfoProfileScope},
	Endpoint:     google.Endpoint,
}

const SUB_LENGTH = 16
const OAUTH_STATE_COOKIE = "DOLABRA_OAUTH_STATE"
const OAUTH_NONCE_LENGTH = 32

/* HELPERS */
// Only allow redirects back to the frontend to prevent open redirects
func safeReturnUrl(returnUrl string) string {
	frontendUrl := utils.GetConfig().FrontendUrl
	if strings.HasPrefix(returnUrl, frontendUrl+"/") || returnUrl == frontendUrl {
		return returnUrl
	}
	return frontendUrl
}
func processIdp(w http.ResponseWriter, idp string, idpSub string) bool {
	// Generate hash from idpSub
	idpHashBytes := sha256.Sum256([]byte(idpSub))
	idpHash := hex.EncodeToString(idpHashBytes[:])

	exists, ok := dbIsMemberWithIdp(w, idp, idpHash)
	if !ok {
		return false
	}

	var sub string
	if exists {
		sub, ok = dbGetMemberSubWithIdp(w, idp, idpHash)
		if !ok {
			return false
		}
	} else {
		// Generate new, unused ocvt sub
		exists := true
		for exists {
			sub = generateCode(SUB_LENGTH)
			exists, ok = dbIsMemberWithSub(w, sub)
			if !ok {
				return false
			}
		}

		// Insert new sub using system member id as placeholder
		//   member_id is changed once user completes registration
		stmt := `
			INSERT INTO auth(
				member_id,
				sub,
				idp,
				idp_hash)
			VALUES (8000000, ?, ?, ?)`
		_, err := db.Exec(stmt, sub, idp, idpHash)
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
	idpHash := chi.URLParam(r, "sub")

	ok := processIdp(w, "DEV", idpHash)
	if !ok {
		return
	}

	http.Redirect(w, r, safeReturnUrl(r.URL.Query().Get("state")), http.StatusTemporaryRedirect)
}

func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Verify oauth state against the nonce issued at login
	stateCookie, err := getCookie(r, OAUTH_STATE_COOKIE)
	deleteCookie(w, OAUTH_STATE_COOKIE)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Missing oauth state cookie")
		return
	}
	nonce, escapedReturnUrl, found := strings.Cut(stateCookie, "|")
	state := r.FormValue("state")
	if !found || len(state) != OAUTH_NONCE_LENGTH ||
		subtle.ConstantTimeCompare([]byte(nonce), []byte(state)) != 1 {
		respondError(w, http.StatusBadRequest, "Invalid oauth state")
		return
	}
	returnUrl, err := url.QueryUnescape(escapedReturnUrl)
	if err != nil {
		returnUrl = ""
	}

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

	http.Redirect(w, r, safeReturnUrl(returnUrl), http.StatusTemporaryRedirect)
}

func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	// Random nonce as the oauth state; the desired return url rides along in a
	// short-lived cookie and is verified against the nonce on callback
	nonce := generateCode(OAUTH_NONCE_LENGTH)
	returnUrl := safeReturnUrl(r.URL.Query().Get("state"))
	http.SetCookie(w, &http.Cookie{
		Domain:   utils.GetConfig().CookieDomain,
		HttpOnly: true,
		MaxAge:   600,
		Name:     OAUTH_STATE_COOKIE,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Secure:   os.Getenv("DEV") != "1",
		Value:    nonce + "|" + url.QueryEscape(returnUrl),
	})

	promptParam := oauth2.SetAuthURLParam("prompt", "consent select_account")
	authUrl := googleOAuthConfig.AuthCodeURL(nonce, promptParam)
	http.Redirect(w, r, authUrl, http.StatusTemporaryRedirect)
}
