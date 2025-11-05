package domain

import "time"

// Block represents a profile blocking another profile
type Block struct {
	ID                string    `json:"id"`
	BlockerProfileID  string    `json:"blocker_profile_id"`
	BlockedProfileID  string    `json:"blocked_profile_id"`
	CreatedAt         time.Time `json:"created_at"`
}

// Mute represents a profile muting another profile
type Mute struct {
	ID               string    `json:"id"`
	MuterProfileID   string    `json:"muter_profile_id"`
	MutedProfileID   string    `json:"muted_profile_id"`
	CreatedAt        time.Time `json:"created_at"`
}

// Reaction represents a reaction (like, love, etc.) to content
type Reaction struct {
	ID         string    `json:"id"`
	TargetType string    `json:"target_type"` // post, comment, message
	TargetID   string    `json:"target_id"`
	ProfileID  string    `json:"profile_id"`
	Key        string    `json:"key"` // like, love, laugh, sad, angry, etc.
	CreatedAt  time.Time `json:"created_at"`
}

// Reaction key constants
const (
	ReactionKeyLike  = "like"
	ReactionKeyLove  = "love"
	ReactionKeyLaugh = "laugh"
	ReactionKeySad   = "sad"
	ReactionKeyAngry = "angry"
	ReactionKeyWow   = "wow"
)

// Reaction target type constants
const (
	ReactionTargetPost    = "post"
	ReactionTargetComment = "comment"
	ReactionTargetMessage = "message"
)
