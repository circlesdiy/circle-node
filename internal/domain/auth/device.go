package auth

import "time"

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

// IsActive checks if the device is active (not revoked)
func (d *Device) IsActive() bool {
	return d.RevokedAt == nil
}

// IsRevoked checks if the device has been revoked
func (d *Device) IsRevoked() bool {
	return d.RevokedAt != nil
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

// IsCompleted checks if the verification has been completed
func (d *DeviceVerification) IsCompleted() bool {
	return d.IsVerified
}

// Verification method constants
const (
	VerificationMethodEmail = "email"
	VerificationMethodSMS   = "sms"
)
