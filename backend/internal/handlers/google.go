package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/auth"
	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/store"
)

type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type googleUserInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// GoogleURL: frontend asks the backend for the Google consent-screen URL.
func GoogleURL(d *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := url.Values{
			"client_id":     {d.Cfg.Google.ClientID},
			"redirect_uri":  {d.Cfg.Google.RedirectURI},
			"response_type": {"code"},
			"scope":         {"openid email profile"},
			"access_type":   {"offline"},
			"prompt":        {"consent"},
		}
		writeJSON(w, http.StatusOK, map[string]string{
			"url": "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode(),
		})
	}
}

// GoogleCallback: Google redirects the browser here with ?code=...
func GoogleCallback(d *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}

		form := url.Values{
			"code":          {code},
			"client_id":     {d.Cfg.Google.ClientID},
			"client_secret": {d.Cfg.Google.ClientSecret},
			"redirect_uri":  {d.Cfg.Google.RedirectURI},
			"grant_type":    {"authorization_code"},
		}
		tokenReq, _ := http.NewRequest("POST", "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
		tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := d.httpClient().Do(tokenReq)
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Println("[google] token exchange failed", err)
			http.Error(w, "google auth failed", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var tokenRes googleTokenResponse
		if err := json.NewDecoder(resp.Body).Decode(&tokenRes); err != nil {
			http.Error(w, "google auth failed", http.StatusInternalServerError)
			return
		}

		req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
		req.Header.Set("Authorization", "Bearer "+tokenRes.AccessToken)
		profileRes, err := d.httpClient().Do(req)
		if err != nil || profileRes.StatusCode != http.StatusOK {
			log.Println("[google] profile fetch failed", err)
			http.Error(w, "google auth failed", http.StatusInternalServerError)
			return
		}
		defer profileRes.Body.Close()

		var profile googleUserInfo
		if err := json.NewDecoder(profileRes.Body).Decode(&profile); err != nil {
			http.Error(w, "google auth failed", http.StatusInternalServerError)
			return
		}

		result, err := auth.CompleteLogin(w, d.DB, d.Cfg.JWTSecret, store.Identity{
			Provider:   store.ProviderGoogle,
			ProviderID: profile.Sub,
			Email:      profile.Email,
			Name:       profile.Name,
		})
		if err != nil {
			log.Println("[google] login failed", err)
			http.Error(w, "google auth failed", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, d.Cfg.FrontendURL+"/auth/callback?token="+url.QueryEscape(result.AccessToken), http.StatusFound)
	}
}
