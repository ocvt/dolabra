package handler

import (
  "context"
  "log"
  "net/http"
  "time"

  "golang.org/x/oauth2"
	"google.golang.org/api/option"
  oidcgoogle "google.golang.org/api/oauth2/v2"
)

func deleteAuthCookies(w http.ResponseWriter) {
  deleteCookie(w, "idp")
  deleteCookie(w, "token")
}

// TODO Idp specific operations
// Get token from cookie and put user id in context
func ProcessClientAuth(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    // Get token and refresh if expired
    // https://github.com/golang/oauth2/issues/84#issuecomment-520099526
    var token oauth2.Token
    err := getCookie(r, "token", &token)
    if err != nil {
      log.Printf("Failed to get \"token\" cookie: %s", err.Error())
      deleteAuthCookies(w)
    } else {
      // Refresh token if expired
      if token.Expiry.Before(time.Now()) {
        tokenSource := googleOAuthConfig.TokenSource(context.Background(), &token)

        newToken, err := tokenSource.Token()
        if err != nil {
          log.Printf("Error getting token from Google: ", err.Error())
        }

        // Update token if it was refreshed
        if newToken.AccessToken != token.AccessToken {
          setCookie(w, "token", *newToken)
          token = *newToken
        }
      }

      // Get user id
      service, err := oidcgoogle.NewService(context.Background(), option.WithTokenSource(googleOAuthConfig.TokenSource(context.Background(), &token)))
      if err != nil {
        deleteAuthCookies(w)
        respondError(w, http.StatusInternalServerError, err.Error())
      }

      // Get userinfo response
      response, err := service.Userinfo.Get().Do()
      if err != nil {
        deleteAuthCookies(w)
        respondError(w, http.StatusInternalServerError, err.Error())
      }

      // Store user id for later access
      ctx = context.WithValue(ctx, "idp", "GOOGLE")
      ctx = context.WithValue(ctx, "subject", response.Id)
    }

    next.ServeHTTP(w, r.WithContext(ctx))
  })
}
