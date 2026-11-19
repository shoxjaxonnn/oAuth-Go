package store

import (
	"sync"
	"time"
)

const maxOTPAttempts = 5

type otpEntry struct {
	code     string
	expires  time.Time
	attempts int
}

// OTPStore is an in-memory, per-process OTP registry for Telegram login.
// Losing it on restart is acceptable — the user just requests a new code.
type OTPStore struct {
	mu      sync.Mutex
	entries map[string]*otpEntry
}

func NewOTPStore() *OTPStore {
	return &OTPStore{entries: make(map[string]*otpEntry)}
}

func (s *OTPStore) Save(chatID, code string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[chatID] = &otpEntry{code: code, expires: time.Now().Add(ttl)}
}

// Verify consumes the code on success, and counts failed attempts so a
// 6-digit code can't be brute-forced within its TTL.
func (s *OTPStore) Verify(chatID, code string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.entries[chatID]
	if !ok {
		return false
	}
	if time.Now().After(entry.expires) {
		delete(s.entries, chatID)
		return false
	}
	if entry.attempts >= maxOTPAttempts {
		delete(s.entries, chatID)
		return false
	}
	entry.attempts++

	if entry.code == code {
		delete(s.entries, chatID)
		return true
	}
	return false
}
