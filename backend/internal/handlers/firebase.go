package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/auth"
	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/store"
)

type firebaseAuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type firebaseAuthResponse struct {
	LocalID string `json:"localId"`
	Email   string `json:"email"`
	IDToken string `json:"idToken"`
}

type firebaseErrorResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func firebasePost(client *http.Client, apiKey, endpoint string, payload map[string]any) (*http.Response, error) {
	body, _ := json.Marshal(payload)
	url := "https://identitytoolkit.googleapis.com/v1/accounts:" + endpoint + "?key=" + apiKey
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return client.Do(req)
}

func firebaseErrorCode(resp *http.Response) string {
	var errRes firebaseErrorResponse
	_ = json.NewDecoder(resp.Body).Decode(&errRes)
	return errRes.Error.Message
}

func sendVerificationEmail(client *http.Client, apiKey, idToken string) {
	resp, err := firebasePost(client, apiKey, "sendOobCode", map[string]any{
		"requestType": "VERIFY_EMAIL",
		"idToken":     idToken,
	})
	if err != nil {
		log.Println("[firebase] failed to send verification email", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println("[firebase] failed to send verification email", firebaseErrorCode(resp))
	}
}

func FirebaseRegister(d *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body firebaseAuthRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" || body.Password == "" {
			writeError(w, http.StatusBadRequest, "email and password required")
			return
		}

		resp, err := firebasePost(d.httpClient(), d.Cfg.FirebaseAPIKey, "signUp", map[string]any{
			"email": body.Email, "password": body.Password, "returnSecureToken": true,
		})
		if err != nil {
			writeError(w, http.StatusBadGateway, "Ro'yxatdan o'tishda xatolik")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			code := firebaseErrorCode(resp)
			message := "Ro'yxatdan o'tishda xatolik"
			if code == "EMAIL_EXISTS" {
				message = "Bu email allaqachon ro'yxatdan o'tgan"
			}
			writeError(w, http.StatusBadRequest, message)
			return
		}

		var result firebaseAuthResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			writeError(w, http.StatusBadGateway, "Ro'yxatdan o'tishda xatolik")
			return
		}

		sendVerificationEmail(d.httpClient(), d.Cfg.FirebaseAPIKey, result.IDToken)

		login, err := auth.CompleteLogin(w, d.DB, d.Cfg.JWTSecret, store.Identity{
			Provider:   store.ProviderFirebase,
			ProviderID: result.LocalID,
			Email:      result.Email,
		})
		if err != nil {
			log.Println("[firebase] register login failed", err)
			writeError(w, http.StatusInternalServerError, "Ro'yxatdan o'tishda xatolik")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"token": login.AccessToken, "emailVerificationSent": true})
	}
}

func FirebaseLogin(d *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body firebaseAuthRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" || body.Password == "" {
			writeError(w, http.StatusBadRequest, "email and password required")
			return
		}

		resp, err := firebasePost(d.httpClient(), d.Cfg.FirebaseAPIKey, "signInWithPassword", map[string]any{
			"email": body.Email, "password": body.Password, "returnSecureToken": true,
		})
		// Same generic message whether the request fails, the email doesn't
		// exist, or the password is wrong — a caller can't use this endpoint
		// to discover which emails are registered.
		if err != nil || resp.StatusCode != http.StatusOK {
			if resp != nil {
				defer resp.Body.Close()
			}
			writeError(w, http.StatusUnauthorized, "Email yoki parol noto'g'ri")
			return
		}
		defer resp.Body.Close()

		var result firebaseAuthResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			writeError(w, http.StatusUnauthorized, "Email yoki parol noto'g'ri")
			return
		}

		login, err := auth.CompleteLogin(w, d.DB, d.Cfg.JWTSecret, store.Identity{
			Provider:   store.ProviderFirebase,
			ProviderID: result.LocalID,
			Email:      result.Email,
		})
		if err != nil {
			log.Println("[firebase] login failed", err)
			writeError(w, http.StatusInternalServerError, "login failed")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"token": login.AccessToken})
	}
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

func FirebaseForgotPassword(d *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body forgotPasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" {
			writeError(w, http.StatusBadRequest, "email required")
			return
		}

		resp, err := firebasePost(d.httpClient(), d.Cfg.FirebaseAPIKey, "sendOobCode", map[string]any{
			"requestType": "PASSWORD_RESET", "email": body.Email,
		})
		if err != nil {
			log.Println("[firebase] forgot-password lookup failed", err)
		} else {
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				log.Println("[firebase] forgot-password lookup failed", firebaseErrorCode(resp))
			}
		}

		// Always the same response, whether or not the email is registered.
		writeJSON(w, http.StatusOK, map[string]string{
			"message": "Agar bu email ro'yxatdan o'tgan bo'lsa, parol tiklash havolasi yuborildi",
		})
	}
}
