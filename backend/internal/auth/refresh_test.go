package auth_test

import (
	"database/sql"
	"testing"

	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/auth"
	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/db"
	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/store"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}

func TestRefreshToken_RotatesAndInvalidatesOld(t *testing.T) {
	database := openTestDB(t)
	user, _ := store.FindOrCreateUser(database, store.Identity{Provider: store.ProviderGoogle, ProviderID: "g-rt-1", Email: "x@example.com"})

	token, err := auth.CreateRefreshToken(database, user.ID, store.ProviderGoogle)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	rotated, err := auth.RotateRefreshToken(database, token)
	if err != nil || rotated == nil {
		t.Fatalf("expected a successful rotation, got result=%v err=%v", rotated, err)
	}
	if rotated.UserID != user.ID || rotated.Provider != store.ProviderGoogle || rotated.Token == token {
		t.Fatalf("unexpected rotation result: %+v", rotated)
	}

	// Replaying the old token (e.g. it was stolen and the real client
	// already rotated it) must fail.
	again, err := auth.RotateRefreshToken(database, token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if again != nil {
		t.Fatal("expected the old, already-rotated token to be rejected")
	}

	// The newly issued token works.
	second, err := auth.RotateRefreshToken(database, rotated.Token)
	if err != nil || second == nil {
		t.Fatalf("expected the new token to rotate successfully, got %v, %v", second, err)
	}
}

func TestRefreshToken_UnknownTokenIsRejected(t *testing.T) {
	database := openTestDB(t)

	result, err := auth.RotateRefreshToken(database, "not-a-real-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected an unknown token to be rejected")
	}
}

func TestRefreshToken_RevokeMakesItUnusable(t *testing.T) {
	database := openTestDB(t)
	user, _ := store.FindOrCreateUser(database, store.Identity{Provider: store.ProviderTelegram, ProviderID: "tg-rt-1"})
	token, _ := auth.CreateRefreshToken(database, user.ID, store.ProviderTelegram)

	if err := auth.RevokeRefreshToken(database, token); err != nil {
		t.Fatalf("revoke failed: %v", err)
	}

	result, err := auth.RotateRefreshToken(database, token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected a revoked token to be rejected")
	}
}
