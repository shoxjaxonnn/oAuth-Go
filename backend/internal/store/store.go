package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Provider string

const (
	ProviderGoogle   Provider = "google"
	ProviderFacebook Provider = "facebook"
	ProviderTelegram Provider = "telegram"
	ProviderFirebase Provider = "firebase"
)

type User struct {
	ID    string
	Email *string
	Name  *string
}

type Identity struct {
	Provider   Provider
	ProviderID string
	Email      string // optional; empty string means "not provided"
	Name       string // optional; empty string means "not provided"
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// FindOrCreateUser resolves the user for a given provider identity, creating
// one if needed. When the identity is new but its email matches an existing
// user (e.g. the same person signing in with Google after already
// registering with Firebase email/password), the new identity is linked to
// that existing user instead of creating a duplicate account.
func FindOrCreateUser(db *sql.DB, identity Identity) (*User, error) {
	var userID string
	err := db.QueryRow(
		"SELECT user_id FROM identities WHERE provider = ? AND provider_id = ?",
		identity.Provider, identity.ProviderID,
	).Scan(&userID)

	if err == nil {
		if identity.Email != "" || identity.Name != "" {
			if _, err := db.Exec(
				"UPDATE users SET email = COALESCE(NULLIF(?, ''), email), name = COALESCE(NULLIF(?, ''), name) WHERE id = ?",
				identity.Email, identity.Name, userID,
			); err != nil {
				return nil, err
			}
		}
		return GetUserByID(db, userID)
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// No existing identity. If the email matches an existing user, link to
	// it instead of creating a duplicate account.
	if identity.Email != "" {
		var existingID string
		err := db.QueryRow("SELECT id FROM users WHERE email = ?", identity.Email).Scan(&existingID)
		if err == nil {
			if _, err := db.Exec(
				"INSERT INTO identities (id, user_id, provider, provider_id, created_at) VALUES (?, ?, ?, ?, ?)",
				uuid.NewString(), existingID, identity.Provider, identity.ProviderID, time.Now().UnixMilli(),
			); err != nil {
				return nil, err
			}
			if identity.Name != "" {
				if _, err := db.Exec(
					"UPDATE users SET name = COALESCE(NULLIF(name, ''), ?) WHERE id = ?",
					identity.Name, existingID,
				); err != nil {
					return nil, err
				}
			}
			return GetUserByID(db, existingID)
		}
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	// Brand new user.
	userID = uuid.NewString()
	now := time.Now().UnixMilli()
	if _, err := db.Exec(
		"INSERT INTO users (id, email, name, created_at) VALUES (?, ?, ?, ?)",
		userID, nilIfEmpty(identity.Email), nilIfEmpty(identity.Name), now,
	); err != nil {
		return nil, err
	}
	if _, err := db.Exec(
		"INSERT INTO identities (id, user_id, provider, provider_id, created_at) VALUES (?, ?, ?, ?, ?)",
		uuid.NewString(), userID, identity.Provider, identity.ProviderID, now,
	); err != nil {
		return nil, err
	}
	return &User{ID: userID, Email: nilIfEmpty(identity.Email), Name: nilIfEmpty(identity.Name)}, nil
}

func GetUserByID(db *sql.DB, id string) (*User, error) {
	var u User
	err := db.QueryRow("SELECT id, email, name FROM users WHERE id = ?", id).Scan(&u.ID, &u.Email, &u.Name)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
