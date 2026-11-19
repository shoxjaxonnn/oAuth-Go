package store_test

import (
	"testing"
	"time"

	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/store"
)

func TestOTPStore_CorrectCodeSucceedsOnce(t *testing.T) {
	otps := store.NewOTPStore()
	otps.Save("chat-1", "123456", time.Minute)

	if !otps.Verify("chat-1", "123456") {
		t.Fatal("expected the correct code to verify")
	}
	if otps.Verify("chat-1", "123456") {
		t.Fatal("expected the code to be consumed after one successful verify")
	}
}

func TestOTPStore_ExpiredCodeFails(t *testing.T) {
	otps := store.NewOTPStore()
	otps.Save("chat-2", "111111", -time.Second) // already expired

	if otps.Verify("chat-2", "111111") {
		t.Fatal("expected an expired code to fail")
	}
}

func TestOTPStore_LocksOutAfterMaxAttempts(t *testing.T) {
	otps := store.NewOTPStore()
	otps.Save("chat-3", "999999", time.Minute)

	for i := 0; i < 5; i++ {
		if otps.Verify("chat-3", "wrong") {
			t.Fatal("wrong code should never verify")
		}
	}

	// Attempts are exhausted — even the correct code must fail now.
	if otps.Verify("chat-3", "999999") {
		t.Fatal("expected verification to be locked out after 5 failed attempts")
	}
}
