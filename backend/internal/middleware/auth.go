package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/auth"
)

type contextKey string

const claimsContextKey contextKey = "claims"

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// RequireAuth checks the Authorization: Bearer <token> header and, on
// success, stores the parsed claims in the request context for handlers to
// read via ClaimsFromContext.
func RequireAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || token == "" {
				writeJSONError(w, http.StatusUnauthorized, "missing bearer token")
				return
			}

			claims, err := auth.VerifySession(jwtSecret, token)
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), claimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ClaimsFromContext(r *http.Request) (*auth.Claims, bool) {
	claims, ok := r.Context().Value(claimsContextKey).(*auth.Claims)
	return claims, ok
}
