package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type OAuthProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type Config struct {
	Port        string
	JWTSecret   string
	FrontendURL string

	Google   OAuthProvider
	Facebook OAuthProvider

	TelegramBotToken string
	FirebaseAPIKey   string
}

func required(name string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Printf("[env] %s is not set — the related auth method will fail until configured", name)
	}
	return v
}

func withDefault(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}

// Load reads backend/.env (if present) and environment variables into a
// Config. Missing provider credentials only log a warning — the server
// still starts, and that one auth method simply fails until configured.
func Load() *Config {
	_ = godotenv.Load()

	jwtSecret := required("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-only-insecure-secret"
	}

	return &Config{
		Port:        withDefault("PORT", "4000"),
		JWTSecret:   jwtSecret,
		FrontendURL: withDefault("FRONTEND_URL", "http://localhost:5173"),

		Google: OAuthProvider{
			ClientID:     required("GOOGLE_CLIENT_ID"),
			ClientSecret: required("GOOGLE_CLIENT_SECRET"),
			RedirectURI:  withDefault("GOOGLE_REDIRECT_URI", "http://localhost:4000/auth/google/callback"),
		},
		Facebook: OAuthProvider{
			ClientID:     required("FACEBOOK_CLIENT_ID"),
			ClientSecret: required("FACEBOOK_CLIENT_SECRET"),
			RedirectURI:  withDefault("FACEBOOK_REDIRECT_URI", "http://localhost:4000/auth/facebook/callback"),
		},

		TelegramBotToken: required("TELEGRAM_BOT_TOKEN"),
		FirebaseAPIKey:   required("FIREBASE_API_KEY"),
	}
}
