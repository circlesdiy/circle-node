package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateCSRFToken generates a random CSRF token
func GenerateCSRFToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

// GenerateRandomBytes generates cryptographically secure random bytes
func GenerateRandomBytes(n int) ([]byte, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return bytes, nil
}