package domain

import (
	"encoding/json"
	"time"
)

// Notification represents a notification sent to a profile
type Notification struct {
	ID              string     `json:"id"`
	ProfileID       string     `json:"profile_id"`
	ActorProfileID  *string    `json:"actor_profile_id,omitempty"` // Who triggered the notification
	Type            string     `json:"type"`                       // mention, reply, like, follow, etc.
	EntityType      string     `json:"entity_type"`                // post, comment, message, etc.
	EntityID        string     `json:"entity_id"`
	DeliveredAt     *time.Time `json:"delivered_at,omitempty"`
	ReadAt          *time.Time `json:"read_at,omitempty"`
	Channel         string     `json:"channel"` // in_app, email, push
	CreatedAt       time.Time  `json:"created_at"`
}

// IsRead checks if the notification has been read
func (n *Notification) IsRead() bool {
	return n.ReadAt != nil
}

// IsDelivered checks if the notification has been delivered
func (n *Notification) IsDelivered() bool {
	return n.DeliveredAt != nil
}

// IsUnread checks if the notification is unread
func (n *Notification) IsUnread() bool {
	return n.ReadAt == nil
}

// Notification type constants
const (
	NotificationTypeMention     = "mention"
	NotificationTypeReply       = "reply"
	NotificationTypeLike        = "like"
	NotificationTypeFollow      = "follow"
	NotificationTypeComment     = "comment"
	NotificationTypeInvite      = "invite"
	NotificationTypeEventRSVP   = "event_rsvp"
	NotificationTypeModeration  = "moderation"
	NotificationTypeSystem      = "system"
)

// Notification channel constants
const (
	NotificationChannelInApp = "in_app"
	NotificationChannelEmail = "email"
	NotificationChannelPush  = "push"
)

// Activity represents an activity log entry for a circle
type Activity struct {
	ID              string          `json:"id"`
	CircleID        string          `json:"circle_id"`
	ActorProfileID  string          `json:"actor_profile_id"`
	EventType       string          `json:"event_type"` // post_created, member_joined, etc.
	TargetType      string          `json:"target_type"`
	TargetID        string          `json:"target_id"`
	OccurredAt      time.Time       `json:"occurred_at"`
	Payload         json.RawMessage `json:"payload"` // JSONB for additional context
}

// Activity event type constants
const (
	ActivityEventPostCreated       = "post_created"
	ActivityEventPostUpdated       = "post_updated"
	ActivityEventPostDeleted       = "post_deleted"
	ActivityEventDiscussionCreated = "discussion_created"
	ActivityEventMemberJoined      = "member_joined"
	ActivityEventMemberLeft        = "member_left"
	ActivityEventMemberBanned      = "member_banned"
	ActivityEventEventCreated      = "event_created"
	ActivityEventCircleUpdated     = "circle_updated"
)
