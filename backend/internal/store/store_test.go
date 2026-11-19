package store_test

import (
	"database/sql"
	"testing"

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

func TestFindOrCreateUser_CreatesNewUser(t *testing.T) {
	database := openTestDB(t)

	user, err := store.FindOrCreateUser(database, store.Identity{
		Provider: store.ProviderGoogle, ProviderID: "g-1", Email: "alice@example.com", Name: "Alice",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email == nil || *user.Email != "alice@example.com" {
		t.Fatalf("expected email alice@example.com, got %v", user.Email)
	}
}

func TestFindOrCreateUser_SameIdentityReturnsSameUser(t *testing.T) {
	database := openTestDB(t)

	first, _ := store.FindOrCreateUser(database, store.Identity{
		Provider: store.ProviderFacebook, ProviderID: "fb-1", Email: "bob@example.com",
	})
	second, _ := store.FindOrCreateUser(database, store.Identity{
		Provider: store.ProviderFacebook, ProviderID: "fb-1", Email: "bob@example.com",
	})

	if first.ID != second.ID {
		t.Fatalf("expected same user id, got %s and %s", first.ID, second.ID)
	}
}

func TestFindOrCreateUser_LinksAccountsBySameEmail(t *testing.T) {
	database := openTestDB(t)

	viaFirebase, _ := store.FindOrCreateUser(database, store.Identity{
		Provider: store.ProviderFirebase, ProviderID: "fb-local-1", Email: "carol@example.com", Name: "Carol",
	})
	viaGoogle, _ := store.FindOrCreateUser(database, store.Identity{
		Provider: store.ProviderGoogle, ProviderID: "g-carol", Email: "carol@example.com",
	})

	if viaFirebase.ID != viaGoogle.ID {
		t.Fatalf("expected linked accounts to share a user id, got %s and %s", viaFirebase.ID, viaGoogle.ID)
	}
}

func TestFindOrCreateUser_DoesNotLinkDifferentEmails(t *testing.T) {
	database := openTestDB(t)

	a, _ := store.FindOrCreateUser(database, store.Identity{Provider: store.ProviderGoogle, ProviderID: "g-a", Email: "a1@example.com"})
	b, _ := store.FindOrCreateUser(database, store.Identity{Provider: store.ProviderGoogle, ProviderID: "g-b", Email: "a2@example.com"})

	if a.ID == b.ID {
		t.Fatalf("expected different users, got the same id %s", a.ID)
	}
}

func TestFindOrCreateUser_NoEmailForTelegram(t *testing.T) {
	database := openTestDB(t)

	user, err := store.FindOrCreateUser(database, store.Identity{Provider: store.ProviderTelegram, ProviderID: "12345"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != nil {
		t.Fatalf("expected nil email, got %v", *user.Email)
	}
}
