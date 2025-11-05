package domain

import "time"

// Role represents a role within a circle
type Role struct {
	ID        string    `json:"id"`
	CircleID  string    `json:"circle_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Permission represents a permission that can be granted to roles
type Permission struct {
	ID          string `json:"id"`
	Key         string `json:"key"` // Unique permission key (e.g., "circle.post.create")
	Description string `json:"description"`
}

// RolePermission maps permissions to roles (many-to-many junction)
type RolePermission struct {
	ID           string `json:"id"`
	RoleID       string `json:"role_id"`
	PermissionID string `json:"permission_id"`
}

// ProfileRole assigns roles to profiles within a circle
type ProfileRole struct {
	ID        string    `json:"id"`
	CircleID  string    `json:"circle_id"`
	ProfileID string    `json:"profile_id"`
	RoleID    string    `json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Common permission keys as constants
const (
	// Circle permissions
	PermKeyCircleUpdate = "circle.update"
	PermKeyCircleDelete = "circle.delete"

	// Post permissions
	PermKeyPostCreate = "circle.post.create"
	PermKeyPostUpdate = "circle.post.update"
	PermKeyPostDelete = "circle.post.delete"

	// Discussion permissions
	PermKeyDiscussionCreate = "circle.discussion.create"
	PermKeyDiscussionUpdate = "circle.discussion.update"
	PermKeyDiscussionDelete = "circle.discussion.delete"
	PermKeyDiscussionPin    = "circle.discussion.pin"
	PermKeyDiscussionLock   = "circle.discussion.lock"

	// Comment permissions
	PermKeyCommentCreate = "circle.comment.create"
	PermKeyCommentUpdate = "circle.comment.update"
	PermKeyCommentDelete = "circle.comment.delete"

	// Member permissions
	PermKeyMemberInvite = "circle.member.invite"
	PermKeyMemberRemove = "circle.member.remove"
	PermKeyMemberBan    = "circle.member.ban"

	// Event permissions
	PermKeyEventCreate = "circle.event.create"
	PermKeyEventUpdate = "circle.event.update"
	PermKeyEventDelete = "circle.event.delete"

	// Moderation permissions
	PermKeyModerate = "circle.moderate"
)
