package auth

import "time"

// WebAuthnCredential represents a WebAuthn credential stored in the database
type WebAuthnCredential struct {
	ID                string     `json:"id"`
	UserID            string     `json:"user_id"`
	CredentialID      string     `json:"credential_id"` // Base64URL encoded
	PublicKey         string     `json:"public_key"`    // Base64 encoded
	CredentialType    string     `json:"credential_type"`
	SignCount         int        `json:"sign_count"`
	Transports        string     `json:"transports"`     // JSON array as string
	AAGUID            string     `json:"aaguid"`
	AttestationFormat string     `json:"attestation_format"`
	AttestationObject []byte     `json:"attestation_object"`
	FriendlyName      string     `json:"friendly_name"`
	IsSynced          bool       `json:"is_synced"`
	IsBackup          bool       `json:"is_backup"`
	LastUsedAt        *time.Time `json:"last_used_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	RevokedAt         *time.Time `json:"revoked_at,omitempty"`
}

// IsActive checks if the credential is active (not revoked)
func (c *WebAuthnCredential) IsActive() bool {
	return c.RevokedAt == nil
}

// IsRevoked checks if the credential has been revoked
func (c *WebAuthnCredential) IsRevoked() bool {
	return c.RevokedAt != nil
}

// DeviceCredential links devices to WebAuthn credentials (many-to-many)
type DeviceCredential struct {
	ID           string     `json:"id"`
	DeviceID     string     `json:"device_id"`
	CredentialID string     `json:"credential_id"`
	LinkedAt     time.Time  `json:"linked_at"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
}
