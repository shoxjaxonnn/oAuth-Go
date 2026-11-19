package auth_test

import (
	"testing"

	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/auth"
	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/store"
)

func ptr(s string) *string { return &s }

func TestSignAndVerifyAccessToken(t *testing.T) {
	user := &store.User{ID: "user-1", Email: ptr("a@b.com"), Name: ptr("Alice")}

	token, err := auth.SignAccessToken("test-secret", user, store.ProviderGoogle)
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}

	claims, err := auth.VerifySession("test-secret", token)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if claims.Sub != "user-1" || claims.Provider != "google" || claims.Email != "a@b.com" || claims.Name != "Alice" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestVerifySession_RejectsTamperedToken(t *testing.T) {
	user := &store.User{ID: "user-1"}
	token, _ := auth.SignAccessToken("test-secret", user, store.ProviderGoogle)
	tampered := token[:len(token)-2] + "xx"

	if _, err := auth.VerifySession("test-secret", tampered); err == nil {
		t.Fatal("expected a tampered token to fail verification")
	}
}

func TestVerifySession_RejectsWrongSecret(t *testing.T) {
	user := &store.User{ID: "user-1"}
	token, _ := auth.SignAccessToken("secret-a", user, store.ProviderGoogle)

	if _, err := auth.VerifySession("secret-b", token); err == nil {
		t.Fatal("expected verification with the wrong secret to fail")
	}
}
