package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/store"
)

const RefreshTokenTTL = 30 * 24 * time.Hour

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func issue(db *sql.DB, userID string, provider store.Provider) (string, error) {
	token, err := randomToken()
	if err != nil {
		return "", err
	}
	now := time.Now()
	_, err = db.Exec(
		"INSERT INTO refresh_tokens (id, user_id, provider, token_hash, created_at, expires_at) VALUES (?, ?, ?, ?, ?, ?)",
		uuid.NewString(), userID, provider, hashToken(token), now.UnixMilli(), now.Add(RefreshTokenTTL).UnixMilli(),
	)
	if err != nil {
		return "", err
	}
	return token, nil
}

func CreateRefreshToken(db *sql.DB, userID string, provider store.Provider) (string, error) {
	return issue(db, userID, provider)
}

type RotateResult struct {
	UserID   string
	Provider store.Provider
	Token    string
}

type refreshRow struct {
	id        string
	userID    string
	provider  store.Provider
	expiresAt int64
	revokedAt sql.NullInt64
}

func findByHash(db *sql.DB, hash string) (*refreshRow, error) {
	var row refreshRow
	err := db.QueryRow(
		"SELECT id, user_id, provider, expires_at, revoked_at FROM refresh_tokens WHERE token_hash = ?",
		hash,
	).Scan(&row.id, &row.userID, &row.provider, &row.expiresAt, &row.revokedAt)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// RotateRefreshToken validates a refresh token and rotates it: the old one
// is revoked and a brand new one is issued. Presenting a token twice (e.g.
// it was stolen and replayed after the legitimate client already rotated
// it) returns an error, so the whole chain can be treated as compromised.
func RotateRefreshToken(db *sql.DB, plainToken string) (*RotateResult, error) {
	row, err := findByHash(db, hashToken(plainToken))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if row.revokedAt.Valid {
		return nil, nil
	}
	if time.Now().UnixMilli() > row.expiresAt {
		return nil, nil
	}

	newToken, err := issue(db, row.userID, row.provider)
	if err != nil {
		return nil, err
	}
	newRow, err := findByHash(db, hashToken(newToken))
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(
		"UPDATE refresh_tokens SET revoked_at = ?, replaced_by = ? WHERE id = ?",
		time.Now().UnixMilli(), newRow.id, row.id,
	); err != nil {
		return nil, err
	}

	return &RotateResult{UserID: row.userID, Provider: row.provider, Token: newToken}, nil
}

func RevokeRefreshToken(db *sql.DB, plainToken string) error {
	row, err := findByHash(db, hashToken(plainToken))
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	_, err = db.Exec("UPDATE refresh_tokens SET revoked_at = ? WHERE id = ?", time.Now().UnixMilli(), row.id)
	return err
}
