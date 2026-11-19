package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"

	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/auth"
	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/store"
)

type facebookTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type facebookProfile struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func FacebookURL(d *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := url.Values{
			"client_id":     {d.Cfg.Facebook.ClientID},
			"redirect_uri":  {d.Cfg.Facebook.RedirectURI},
			"scope":         {"email,public_profile"},
			"response_type": {"code"},
		}
		writeJSON(w, http.StatusOK, map[string]string{
			"url": "https://www.facebook.com/v21.0/dialog/oauth?" + params.Encode(),
		})
	}
}

func FacebookCallback(d *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}

		tokenURL := "https://graph.facebook.com/v21.0/oauth/access_token?" + url.Values{
			"client_id":     {d.Cfg.Facebook.ClientID},
			"client_secret": {d.Cfg.Facebook.ClientSecret},
			"redirect_uri":  {d.Cfg.Facebook.RedirectURI},
			"code":          {code},
		}.Encode()

		tokenReq, _ := http.NewRequest("GET", tokenURL, nil)
		resp, err := d.httpClient().Do(tokenReq)
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Println("[facebook] token exchange failed", err)
			http.Error(w, "facebook auth failed", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var tokenRes facebookTokenResponse
		if err := json.NewDecoder(resp.Body).Decode(&tokenRes); err != nil {
			http.Error(w, "facebook auth failed", http.StatusInternalServerError)
			return
		}

		profileURL := "https://graph.facebook.com/me?" + url.Values{
			"fields":       {"id,name,email"},
			"access_token": {tokenRes.AccessToken},
		}.Encode()

		profileReq, _ := http.NewRequest("GET", profileURL, nil)
		profileRes, err := d.httpClient().Do(profileReq)
		if err != nil || profileRes.StatusCode != http.StatusOK {
			log.Println("[facebook] profile fetch failed", err)
			http.Error(w, "facebook auth failed", http.StatusInternalServerError)
			return
		}
		defer profileRes.Body.Close()

		var profile facebookProfile
		if err := json.NewDecoder(profileRes.Body).Decode(&profile); err != nil {
			http.Error(w, "facebook auth failed", http.StatusInternalServerError)
			return
		}

		result, err := auth.CompleteLogin(w, d.DB, d.Cfg.JWTSecret, store.Identity{
			Provider:   store.ProviderFacebook,
			ProviderID: profile.ID,
			Email:      profile.Email,
			Name:       profile.Name,
		})
		if err != nil {
			log.Println("[facebook] login failed", err)
			http.Error(w, "facebook auth failed", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, d.Cfg.FrontendURL+"/auth/callback?token="+url.QueryEscape(result.AccessToken), http.StatusFound)
	}
}
