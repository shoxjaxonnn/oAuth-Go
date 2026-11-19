package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// RateLimiter is a simple fixed-window limiter, keyed per client IP: at most
// Limit requests within Window. It intentionally avoids any external
// dependency — this is all the semantics express-rate-limit gave the
// Node version of this project.
type RateLimiter struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	limit  int
	window time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{hits: make(map[string][]time.Time), limit: limit, window: window}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (l *RateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	kept := l.hits[key][:0]
	for _, t := range l.hits[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}

	if len(kept) >= l.limit {
		l.hits[key] = kept
		return false
	}

	l.hits[key] = append(kept, now)
	return true
}

func (l *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.allow(clientIP(r)) {
			writeJSONError(w, http.StatusTooManyRequests, "Juda ko'p urinish. Birozdan keyin qayta urinib ko'ring.")
			return
		}
		next.ServeHTTP(w, r)
	})
}
