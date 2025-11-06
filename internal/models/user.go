package models

import (
	"encoding/json"

	"circles.diy/internal/utils"
)

type User struct {
	ID     string `json:"id"`
	Handle string `json:"handle"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Bio    string `json:"bio"`
	Banner string `json:"banner"`
}

// GetAvatar returns the avatar URL or fallback if empty
func (u *User) GetAvatar() string {
	return utils.GetAvatarURL(u.Avatar)
}

// GetBanner returns the banner URL or fallback if empty
func (u *User) GetBanner() string {
	return utils.GetBannerURL(u.Banner)
}

type ProfileStats struct {
	Posts       int `json:"posts"`
	Connections int `json:"connections"`
	Circles     int `json:"circles"`
}

// Profile represents a profile for template rendering
// This aligns with domain.Profile but includes additional fields for UI
type Profile struct {
	ID          string          `json:"id"`
	UserID      string          `json:"user_id,omitempty"`      // Only included for owner views
	Handle      string          `json:"handle"`
	Name        string          `json:"name"`
	DisplayName string          `json:"display_name"`
	Bio         string          `json:"bio"`
	AvatarURL   string          `json:"avatar_url"`
	BannerURL   string          `json:"banner_url,omitempty"`   // Optional banner image
	IsActive    bool            `json:"is_active,omitempty"`    // Only included for owner views

	// Settings fields (from ProfileSettings)
	IsPublic    bool            `json:"is_public"`
	Location    string          `json:"location,omitempty"`
	Website     string          `json:"website,omitempty"`
	Interests   json.RawMessage `json:"interests,omitempty"`    // JSONB array of strings
	SocialLinks json.RawMessage `json:"social_links,omitempty"` // JSONB map of platform->url

	// Computed/UI fields
	Stats       ProfileStats    `json:"stats"`
	IsConnected bool            `json:"is_connected"`           // Whether viewer is connected to this profile
	IsOwner     bool            `json:"is_owner"`               // Whether viewer owns this profile
	IsVerified  bool            `json:"is_verified,omitempty"`  // Future: verification badge
}

// GetAvatar returns the avatar URL or fallback if empty
func (p *Profile) GetAvatar() string {
	return utils.GetAvatarURL(p.AvatarURL)
}

// GetBanner returns the banner URL or fallback if empty
func (p *Profile) GetBanner() string {
	return utils.GetBannerURL(p.BannerURL)
}