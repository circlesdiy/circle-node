package auth

import "time"

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

// IsActive checks if the challenge is still active (not expired and not completed)
func (c *AuthenticationChallenge) IsActive() bool {
	return !c.IsExpired() && !c.IsCompleted()
}

// Challenge type constants
const (
	ChallengeTypeRegistration   = "registration"
	ChallengeTypeAuthentication = "authentication"
)
