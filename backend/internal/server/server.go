package server

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/config"
	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/docs"
	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/handlers"
	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/middleware"
	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/store"
)

// New wires every route and returns the final http.Handler. Using a
// function instead of package-level globals makes it straightforward to
// spin up an independent server per test (see server_test.go). client is
// the HTTP client used for outbound calls to Google/Facebook/Telegram/
// Firebase; pass nil to use http.DefaultClient (tests override it with a
// fake transport to avoid real network calls).
func New(cfg *config.Config, db *sql.DB, client *http.Client) http.Handler {
	deps := &handlers.Deps{DB: db, Cfg: cfg, OTPs: store.NewOTPStore(), Client: client}

	authLimiter := middleware.NewRateLimiter(10, 15*time.Minute)
	otpLimiter := middleware.NewRateLimiter(5, 10*time.Minute)
	requireAuth := middleware.RequireAuth(cfg.JWTSecret)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handlers.Health())

	mux.HandleFunc("GET /auth/google/url", handlers.GoogleURL(deps))
	mux.HandleFunc("GET /auth/google/callback", handlers.GoogleCallback(deps))

	mux.HandleFunc("GET /auth/facebook/url", handlers.FacebookURL(deps))
	mux.HandleFunc("GET /auth/facebook/callback", handlers.FacebookCallback(deps))

	mux.Handle("POST /auth/telegram/send-otp", otpLimiter.Middleware(handlers.TelegramSendOTP(deps)))
	mux.Handle("POST /auth/telegram/verify-otp", authLimiter.Middleware(handlers.TelegramVerifyOTP(deps)))

	mux.Handle("POST /auth/firebase/register", authLimiter.Middleware(handlers.FirebaseRegister(deps)))
	mux.Handle("POST /auth/firebase/login", authLimiter.Middleware(handlers.FirebaseLogin(deps)))
	mux.Handle("POST /auth/firebase/forgot-password", authLimiter.Middleware(handlers.FirebaseForgotPassword(deps)))

	mux.Handle("GET /auth/me", requireAuth(handlers.Me()))
	mux.Handle("POST /auth/refresh", authLimiter.Middleware(handlers.Refresh(deps)))
	mux.HandleFunc("POST /auth/logout", handlers.Logout(deps))

	docs.Mount(mux)

	return middleware.CORS(cfg.FrontendURL)(mux)
}
