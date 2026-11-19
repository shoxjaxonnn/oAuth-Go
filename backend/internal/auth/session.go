package auth

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/store"
)

const RefreshCookieName = "refresh_token"

func SetRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    token,
		Path:     "/auth",
		HttpOnly: true,
		Secure:   os.Getenv("ENV") == "production",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(RefreshTokenTTL.Seconds()),
	})
}

func ClearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    "",
		Path:     "/auth",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

type LoginResult struct {
	AccessToken string
	User        *store.User
}

// CompleteLogin finds/creates the user for this provider identity, issues a
// short-lived access token (returned to the caller) and a rotating refresh
// token (set as an httpOnly cookie, never exposed to frontend JS).
func CompleteLogin(w http.ResponseWriter, db *sql.DB, jwtSecret string, identity store.Identity) (*LoginResult, error) {
	user, err := store.FindOrCreateUser(db, identity)
	if err != nil {
		return nil, err
	}

	accessToken, err := SignAccessToken(jwtSecret, user, identity.Provider)
	if err != nil {
		return nil, err
	}

	refreshToken, err := CreateRefreshToken(db, user.ID, identity.Provider)
	if err != nil {
		return nil, err
	}
	SetRefreshCookie(w, refreshToken)

	return &LoginResult{AccessToken: accessToken, User: user}, nil
}
