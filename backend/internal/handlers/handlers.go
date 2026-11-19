package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/config"
	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/store"
)

// Deps bundles everything a handler needs, so each handler stays a plain
// function with no package-level globals.
type Deps struct {
	DB     *sql.DB
	Cfg    *config.Config
	OTPs   *store.OTPStore
	Client *http.Client // HTTP client for calling Google/Facebook/Telegram/Firebase; overridden in tests.
}

func (d *Deps) httpClient() *http.Client {
	if d.Client != nil {
		return d.Client
	}
	return http.DefaultClient
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
