package auth

import "time"

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

// IsRevoked checks if the session has been revoked
func (s *Session) IsRevoked() bool {
	return s.RevokedAt != nil
}
