package domain

import "time"

// UserModeration represents moderation state for a user
type UserModeration struct {
	UserID         string     `json:"user_id"`
	TrustScore     float64    `json:"trust_score"`
	WarningCount   int        `json:"warning_count"`
	ViolationCount int        `json:"violation_count"`
	BannedUntil    *time.Time `json:"banned_until,omitempty"`
	Notes          string     `json:"notes"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// IsBanned checks if the user is currently banned
func (m *UserModeration) IsBanned() bool {
	if m.BannedUntil == nil {
		return false
	}
	return time.Now().Before(*m.BannedUntil)
}

// IsPermanentlyBanned checks if the user is permanently banned (far future date)
func (m *UserModeration) IsPermanentlyBanned() bool {
	if m.BannedUntil == nil {
		return false
	}
	// Consider ban permanent if it's more than 100 years in the future
	farFuture := time.Now().AddDate(100, 0, 0)
	return m.BannedUntil.After(farFuture)
}

// HasWarnings checks if the user has any warnings
func (m *UserModeration) HasWarnings() bool {
	return m.WarningCount > 0
}

// HasViolations checks if the user has any violations
func (m *UserModeration) HasViolations() bool {
	return m.ViolationCount > 0
}

// Report represents a report filed by a profile about content or behavior
type Report struct {
	ID                 string    `json:"id"`
	ReporterProfileID  string    `json:"reporter_profile_id"`
	TargetType         string    `json:"target_type"` // post, comment, message, profile, etc.
	TargetID           string    `json:"target_id"`
	Reason             string    `json:"reason"`
	State              string    `json:"state"` // pending, reviewing, resolved, dismissed
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// IsPending checks if the report is still pending review
func (r *Report) IsPending() bool {
	return r.State == ReportStatePending
}

// IsResolved checks if the report has been resolved
func (r *Report) IsResolved() bool {
	return r.State == ReportStateResolved
}

// IsDismissed checks if the report was dismissed
func (r *Report) IsDismissed() bool {
	return r.State == ReportStateDismissed
}

// Report target type constants
const (
	ReportTargetPost       = "post"
	ReportTargetComment    = "comment"
	ReportTargetMessage    = "message"
	ReportTargetProfile    = "profile"
	ReportTargetCircle     = "circle"
	ReportTargetDiscussion = "discussion"
)

// Report state constants
const (
	ReportStatePending   = "pending"
	ReportStateReviewing = "reviewing"
	ReportStateResolved  = "resolved"
	ReportStateDismissed = "dismissed"
)

// ModerationAction represents an action taken by a moderator
type ModerationAction struct {
	ID                 string    `json:"id"`
	ModeratorProfileID string    `json:"moderator_profile_id"`
	TargetType         string    `json:"target_type"` // post, comment, profile, etc.
	TargetID           string    `json:"target_id"`
	Action             string    `json:"action"` // delete, hide, warn, ban, restore
	Reason             string    `json:"reason"`
	CreatedAt          time.Time `json:"created_at"`
}

// Moderation action type constants
const (
	ModerationActionDelete  = "delete"
	ModerationActionHide    = "hide"
	ModerationActionWarn    = "warn"
	ModerationActionBan     = "ban"
	ModerationActionRestore = "restore"
	ModerationActionFlag    = "flag"
)
