package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/store"
)

const AccessTokenTTL = 15 * time.Minute

type Claims struct {
	Sub      string `json:"sub"`
	Provider string `json:"provider"`
	Email    string `json:"email,omitempty"`
	Name     string `json:"name,omitempty"`
	jwt.RegisteredClaims
}

func SignAccessToken(secret string, user *store.User, provider store.Provider) (string, error) {
	claims := Claims{
		Sub:      user.ID,
		Provider: string(provider),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	if user.Email != nil {
		claims.Email = *user.Email
	}
	if user.Name != nil {
		claims.Name = *user.Name
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func VerifySession(secret, tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrTokenSignatureInvalid
	}
	return claims, nil
}
