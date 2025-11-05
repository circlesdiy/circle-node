package auth

import "time"

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

// IsRevoked checks if the recovery method has been revoked
func (r *RecoveryMethod) IsRevoked() bool {
	return r.RevokedAt != nil
}

// Recovery method type constants
const (
	RecoveryTypeBackupCodes   = "backup_codes"
	RecoveryTypeRecoveryEmail = "recovery_email"
	RecoveryTypeRecoveryPhone = "recovery_phone"
)
