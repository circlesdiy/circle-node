package auth

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

// AuthLevel represents different authentication levels
type AuthLevel int

// AuthLevel constants for different authentication levels
const (
	AuthLevelNone           AuthLevel = 0 // No authentication
	AuthLevelBasic          AuthLevel = 1 // Basic WebAuthn authentication
	AuthLevelTrustedDevice  AuthLevel = 2 // Authentication from trusted device
	AuthLevelMultiFactor    AuthLevel = 3 // Multi-factor authentication
	AuthLevelAdministrative AuthLevel = 4 // Administrative operations
)

// Session duration constants
const (
	DefaultSessionDuration     = 24 * time.Hour
	RefreshTokenDuration       = 7 * 24 * time.Hour
	ChallengeDuration          = 5 * time.Minute
	DeviceVerificationDuration = 10 * time.Minute
)

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}
