package domain

import (
	"html/template"
	"time"
)

// Circle represents a community or group
type Circle struct {
	ID              string       `json:"id"`
	OwnerProfileID  string       `json:"owner_profile_id"`
	Name            string       `json:"name"`
	Description     string       `json:"description"`
	Visibility      string       `json:"visibility"` // public, private, unlisted
	AutoModEnabled  bool         `json:"auto_mod_enabled"`
	Icon            string       `json:"icon,omitempty"`
	IconBgColor     template.CSS `json:"icon_bg_color,omitempty"`
	AvatarURL       string       `json:"avatar_url,omitempty"`
	BannerURL       string       `json:"banner_url,omitempty"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
	DeletedAt       *time.Time   `json:"deleted_at,omitempty"`
}

// IsDeleted checks if the circle has been soft deleted
func (c *Circle) IsDeleted() bool {
	return c.DeletedAt != nil
}

// IsPublic checks if the circle is publicly visible
func (c *Circle) IsPublic() bool {
	return c.Visibility == CircleVisibilityPublic
}

// IsPrivate checks if the circle is private
func (c *Circle) IsPrivate() bool {
	return c.Visibility == CircleVisibilityPrivate
}

// Circle visibility constants
const (
	CircleVisibilityPublic   = "public"
	CircleVisibilityPrivate  = "private"
	CircleVisibilityUnlisted = "unlisted"
)

// CircleMembership represents a profile's membership in a circle
type CircleMembership struct {
	ID               string     `json:"id"`
	CircleID         string     `json:"circle_id"`
	ProfileID        string     `json:"profile_id"`
	InviterProfileID *string    `json:"inviter_profile_id,omitempty"` // Who invited this member
	State            string     `json:"state"`                         // active, invited, banned, left
	JoinedAt         time.Time  `json:"joined_at"`
	LeftAt           *time.Time `json:"left_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// IsActive checks if the membership is currently active
func (m *CircleMembership) IsActive() bool {
	return m.State == MembershipStateActive && m.LeftAt == nil
}

// IsBanned checks if the member has been banned
func (m *CircleMembership) IsBanned() bool {
	return m.State == MembershipStateBanned
}

// Membership state constants
const (
	MembershipStateActive  = "active"
	MembershipStateInvited = "invited"
	MembershipStateBanned  = "banned"
	MembershipStateLeft    = "left"
)

// CircleMembershipWithInviter includes membership data with inviter profile information
type CircleMembershipWithInviter struct {
	CircleMembership
	InviterName   string `json:"inviter_name,omitempty"`
	InviterHandle string `json:"inviter_handle,omitempty"`
}
