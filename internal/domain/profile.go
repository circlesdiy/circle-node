package domain

import (
	"encoding/json"
	"time"
)

// Profile represents a user's social identity
// One user can have multiple profiles for different contexts
type Profile struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Handle      string     `json:"handle"`
	Name        string     `json:"name"`
	DisplayName string     `json:"display_name"`
	Bio         string     `json:"bio"`
	AvatarURL   string     `json:"avatar_url"`
	BannerURL   string     `json:"banner_url"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// IsDeleted checks if the profile has been soft deleted
func (p *Profile) IsDeleted() bool {
	return p.DeletedAt != nil
}

// IsAvailable checks if the profile is active and not deleted
func (p *Profile) IsAvailable() bool {
	return p.IsActive && p.DeletedAt == nil
}

// ProfileSettings represents settings and preferences for a profile
type ProfileSettings struct {
	ProfileID   string          `json:"profile_id"`
	IsPublic    bool            `json:"is_public"`
	Location    string          `json:"location"`
	Website     string          `json:"website"`
	Interests   json.RawMessage `json:"interests"`    // JSONB stored as JSON
	SocialLinks json.RawMessage `json:"social_links"` // JSONB stored as JSON
	UpdatedAt   time.Time       `json:"updated_at"`
}

// Interests represents a profile's interests (structured JSONB)
type Interests []string

// SocialLinks represents a profile's social media links (structured JSONB)
type SocialLinks map[string]string
