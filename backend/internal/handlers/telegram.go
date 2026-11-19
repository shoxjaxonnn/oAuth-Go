package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"time"

	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/auth"
	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/store"
)

const otpTTL = 5 * time.Minute

func randomOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()+100000), nil
}

type sendOTPRequest struct {
	ChatID string `json:"chatId"`
}

// TelegramSendOTP messages a 6-digit code to the given Telegram chat. The
// user must have already pressed /start on the bot, so the bot has a
// chatId it's allowed to message.
func TelegramSendOTP(d *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body sendOTPRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ChatID == "" {
			writeError(w, http.StatusBadRequest, "chatId required")
			return
		}

		code, err := randomOTP()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to generate otp")
			return
		}
		d.OTPs.Save(body.ChatID, code, otpTTL)

		payload, _ := json.Marshal(map[string]string{
			"chat_id": body.ChatID,
			"text":    fmt.Sprintf("Tasdiqlash kodi: %s (5 daqiqa amal qiladi)", code),
		})
		apiURL := "https://api.telegram.org/bot" + d.Cfg.TelegramBotToken + "/sendMessage"
		req, _ := http.NewRequest("POST", apiURL, bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		resp, err := d.httpClient().Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Println("[telegram] send-otp failed", err)
			writeError(w, http.StatusInternalServerError, "failed to send otp")
			return
		}
		defer resp.Body.Close()

		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

type verifyOTPRequest struct {
	ChatID string `json:"chatId"`
	Code   string `json:"code"`
}

func TelegramVerifyOTP(d *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body verifyOTPRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ChatID == "" || body.Code == "" {
			writeError(w, http.StatusBadRequest, "chatId and code required")
			return
		}

		if !d.OTPs.Verify(body.ChatID, body.Code) {
			writeError(w, http.StatusUnauthorized, "invalid or expired code")
			return
		}

		result, err := auth.CompleteLogin(w, d.DB, d.Cfg.JWTSecret, store.Identity{
			Provider:   store.ProviderTelegram,
			ProviderID: body.ChatID,
		})
		if err != nil {
			log.Println("[telegram] login failed", err)
			writeError(w, http.StatusInternalServerError, "login failed")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"token": result.AccessToken})
	}
}
