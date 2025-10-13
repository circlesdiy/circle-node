package auth

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/fxamacker/webauthn"
)

// User represents a user in the authentication system
// Implements the webauthn.User interface for fxamacker/webauthn
type User struct {
	ID              string    `json:"id"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	AccountStatus   string    `json:"account_status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`

}

// ToWebAuthnUser converts to fxamacker webauthn.User struct
func (u *User) ToWebAuthnUser(credentials []WebAuthnCredential) *webauthn.User {
	credentialIDs := make([][]byte, len(credentials))
	for i, cred := range credentials {
		if credID, err := base64.URLEncoding.DecodeString(cred.CredentialID); err == nil {
			credentialIDs[i] = credID
		}
	}

	return &webauthn.User{
		ID:            []byte(u.ID),
		Name:          u.Username,
		DisplayName:   u.Username,
		Icon:          "", // Optional
		CredentialIDs: credentialIDs,
	}
}

// WebAuthnCredential represents a WebAuthn credential stored in the database
type WebAuthnCredential struct {
	ID                  string     `json:"id"`
	UserID              string     `json:"user_id"`
	CredentialID        string     `json:"credential_id"` // Base64URL encoded
	PublicKey           string     `json:"public_key"`    // Base64 encoded
	CredentialType      string     `json:"credential_type"`
	SignCount           int        `json:"sign_count"`
	Transports          string     `json:"transports"`     // JSON array as string
	AAGUID              string     `json:"aaguid"`
	AttestationFormat   string     `json:"attestation_format"`
	AttestationObject   []byte     `json:"attestation_object"`
	FriendlyName        string     `json:"friendly_name"`
	IsSynced            bool       `json:"is_synced"`
	IsBackup            bool       `json:"is_backup"`
	LastUsedAt          *time.Time `json:"last_used_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	RevokedAt           *time.Time `json:"revoked_at,omitempty"`
}

// ToWebAuthnCredential converts to fxamacker/webauthn.Credential
func (c *WebAuthnCredential) ToWebAuthnCredential() *webauthn.Credential {
	pubKey, _ := base64.StdEncoding.DecodeString(c.PublicKey)

	// Create credential with the proper fxamacker API
	cred, _, _ := webauthn.ParseCredential(pubKey)
	if cred == nil {
		// Fallback to empty credential
		cred = &webauthn.Credential{}
	}

	return cred
}

// Device represents a device/browser that a user has authenticated from
type Device struct {
	ID                string     `json:"id"`
	UserID            string     `json:"user_id"`
	DeviceFingerprint string     `json:"device_fingerprint"`
	FriendlyName      string     `json:"friendly_name"`
	DeviceType        string     `json:"device_type"`
	OS                string     `json:"os"`
	Browser           string     `json:"browser"`
	IsTrusted         bool       `json:"is_trusted"`
	LastVerifiedAt    *time.Time `json:"last_verified_at,omitempty"`
	FirstSeenAt       time.Time  `json:"first_seen_at"`
	LastSeenAt        time.Time  `json:"last_seen_at"`
	RevokedAt         *time.Time `json:"revoked_at,omitempty"`
}

// DeviceCredential links devices to WebAuthn credentials (many-to-many)
type DeviceCredential struct {
	ID           string     `json:"id"`
	DeviceID     string     `json:"device_id"`
	CredentialID string     `json:"credential_id"`
	LinkedAt     time.Time  `json:"linked_at"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
}

// Session represents a user session tied to a specific device
type Session struct {
	ID                         string     `json:"id"`
	UserID                     string     `json:"user_id"`
	DeviceID                   *string    `json:"device_id,omitempty"`
	ActiveProfileID            *string    `json:"active_profile_id,omitempty"`
	AuthenticatedCredentialIDs string     `json:"authenticated_credential_ids"` // JSON array as string
	SessionToken               string     `json:"session_token"`
	RefreshToken               string     `json:"refresh_token"`
	IPAddress                  string     `json:"ip_address"`
	UserAgent                  string     `json:"user_agent"`
	AuthLevel                  int        `json:"auth_level"`
	CreatedAt                  time.Time  `json:"created_at"`
	LastActivityAt             time.Time  `json:"last_activity_at"`
	ExpiresAt                  time.Time  `json:"expires_at"`
	RevokedAt                  *time.Time `json:"revoked_at,omitempty"`
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt) || s.RevokedAt != nil
}

// IsActive checks if the session is valid and active
func (s *Session) IsActive() bool {
	return !s.IsExpired() && s.RevokedAt == nil
}

// AuthenticationChallenge represents a time-bound authentication challenge
type AuthenticationChallenge struct {
	ID                   string                 `json:"id"`
	UserID               *string                `json:"user_id,omitempty"`
	Challenge            string                 `json:"challenge"`
	Type                 string                 `json:"type"` // "registration" or "authentication"
	RequiredCredentialID *string                `json:"required_credential_id,omitempty"`
	Options              map[string]interface{} `json:"options"`
	CreatedAt            time.Time              `json:"created_at"`
	ExpiresAt            time.Time              `json:"expires_at"`
	CompletedAt          *time.Time             `json:"completed_at,omitempty"`
}

// IsExpired checks if the challenge has expired
func (c *AuthenticationChallenge) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// IsCompleted checks if the challenge has been completed
func (c *AuthenticationChallenge) IsCompleted() bool {
	return c.CompletedAt != nil
}

// RecoveryMethod represents backup recovery options for a user
type RecoveryMethod struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	Type            string     `json:"type"` // "backup_codes", "recovery_email", etc.
	EncryptedSecret string     `json:"encrypted_secret"`
	Salt            string     `json:"salt"`
	IsVerified      bool       `json:"is_verified"`
	RemainingUses   *int       `json:"remaining_uses,omitempty"`
	VerifiedAt      *time.Time `json:"verified_at,omitempty"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	RevokedAt       *time.Time `json:"revoked_at,omitempty"`
}

// IsExpired checks if the recovery method has expired
func (r *RecoveryMethod) IsExpired() bool {
	return r.ExpiresAt != nil && time.Now().After(*r.ExpiresAt)
}

// IsUsable checks if the recovery method can still be used
func (r *RecoveryMethod) IsUsable() bool {
	if !r.IsVerified || r.RevokedAt != nil || r.IsExpired() {
		return false
	}
	if r.RemainingUses != nil && *r.RemainingUses <= 0 {
		return false
	}
	return true
}

// DeviceVerification represents device verification challenges
type DeviceVerification struct {
	ID               string     `json:"id"`
	DeviceID         string     `json:"device_id"`
	Method           string     `json:"method"` // "email", "sms", etc.
	VerificationCode string     `json:"verification_code"`
	IsVerified       bool       `json:"is_verified"`
	VerifiedAt       *time.Time `json:"verified_at,omitempty"`
	ExpiresAt        time.Time  `json:"expires_at"`
	CreatedAt        time.Time  `json:"created_at"`
}

// IsExpired checks if the verification has expired
func (d *DeviceVerification) IsExpired() bool {
	return time.Now().After(d.ExpiresAt)
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// AuthLevel constants for different authentication levels
const (
	AuthLevelNone           = 0 // No authentication
	AuthLevelBasic          = 1 // Basic WebAuthn authentication
	AuthLevelTrustedDevice  = 2 // Authentication from trusted device
	AuthLevelMultiFactor    = 3 // Multi-factor authentication
	AuthLevelAdministrative = 4 // Administrative operations
)

// Session duration constants
const (
	DefaultSessionDuration = 24 * time.Hour
	RefreshTokenDuration   = 7 * 24 * time.Hour
	ChallengeDuration      = 5 * time.Minute
	DeviceVerificationDuration = 10 * time.Minute
)