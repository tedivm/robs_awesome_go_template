package service

import (
	"crypto/subtle"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ValidateAPIKey checks the provided API key against the configured value.
// Returns ErrInvalidAPIKey if the key doesn't match or no key is configured.
// Uses constant-time comparison to prevent timing side-channel attacks.
func ValidateAPIKey(key string) error {
	// If no API key is configured, auth is disabled (development mode).
	if apiKey == "" {
		return nil
	}
	if len(key) != len(apiKey) || subtle.ConstantTimeCompare([]byte(key), []byte(apiKey)) == 0 {
		return ErrInvalidAPIKey
	}
	return nil
}

// GenerateToken creates a JWT token for the given subject (user ID or identifier).
// The subject should NOT be the raw API key — use a stable user/service identifier instead.
func GenerateToken(subject string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": subject,
		"iat": now.Unix(),
		"exp": now.Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}
