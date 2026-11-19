package handlers

import (
	"net/http"

	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/auth"
	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/middleware"
	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/store"
)

func Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

func Me() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, _ := middleware.ClaimsFromContext(r)
		writeJSON(w, http.StatusOK, map[string]any{"user": claims})
	}
}

// Refresh rotates the refresh token found in the httpOnly cookie and
// returns a fresh access token. The frontend never reads this cookie
// directly — the browser sends it automatically.
func Refresh(d *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(auth.RefreshCookieName)
		if err != nil || cookie.Value == "" {
			writeError(w, http.StatusUnauthorized, "no refresh token")
			return
		}

		rotated, err := auth.RotateRefreshToken(d.DB, cookie.Value)
		if err != nil || rotated == nil {
			auth.ClearRefreshCookie(w)
			writeError(w, http.StatusUnauthorized, "invalid or expired refresh token")
			return
		}

		user, err := store.GetUserByID(d.DB, rotated.UserID)
		if err != nil {
			auth.ClearRefreshCookie(w)
			writeError(w, http.StatusUnauthorized, "user not found")
			return
		}

		auth.SetRefreshCookie(w, rotated.Token)
		accessToken, err := auth.SignAccessToken(d.Cfg.JWTSecret, user, rotated.Provider)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not sign token")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"token": accessToken})
	}
}

func Logout(d *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie(auth.RefreshCookieName); err == nil && cookie.Value != "" {
			_ = auth.RevokeRefreshToken(d.DB, cookie.Value)
		}
		auth.ClearRefreshCookie(w)
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}
