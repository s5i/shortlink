package auth

import (
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Auth provides authentication via OAuth.
type Auth struct {
	oAuthCfg      *oauth2.Config
	oAuthInFlight map[string]*oauthState
	jwtSecret     []byte
	jwtTTL        time.Duration
	jwtCookieName string
}

// New instantiates an Auth object and registers it against provided http.ServeMux.
func New(clientID, clientSecret, jwtSecret string, jwtTTL time.Duration, hostname string, mux *http.ServeMux) *Auth {
	redirectPath := "/auth/callback"
	redirectURL := fmt.Sprintf("http://%s%s", hostname, redirectPath)
	auth := &Auth{
		oAuthCfg: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"email"},
			Endpoint:     google.Endpoint,
		},
		oAuthInFlight: map[string]*oauthState{},
		jwtSecret:     []byte(jwtSecret),
		jwtTTL:        jwtTTL,
		jwtCookieName: "token",
	}
	mux.Handle(redirectPath, auth)
	return auth
}

// ServeHTTP serves OAuth callback page.
func (a *Auth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.oAuthCallback(w, r)
}

// RequireUser is a middleware function that ensures that the user is authenticated.
// Use User to get the username.
func (a *Auth) RequireUser(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del(userHeader)
		for _, c := range r.Cookies() {
			if c.Name != a.jwtCookieName {
				continue
			}
			user, ok := a.jwtVerify(c.Value)
			if !ok {
				a.oAuthDialog(w, r)
				return
			}
			r.Header.Add(userHeader, user)
			next.ServeHTTP(w, r)
			return
		}
		a.oAuthDialog(w, r)
		return
	}
}

// User extracts username from request's headers.
// Requires Auth.RequireUser middleware (otherwise users can spoof the header).
func User(r *http.Request) string {
	return r.Header.Get(userHeader)
}

const userHeader = "X-Shortlink-Authenticated-User"
