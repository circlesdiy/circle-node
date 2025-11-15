package domain

import "time"

// Event represents an event organized within a circle
type Event struct {
	ID                 string     `json:"id"`
	CircleID           string     `json:"circle_id"`
	OrganizerProfileID string     `json:"organizer_profile_id"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	Location           string     `json:"location"`
	Timezone           string     `json:"timezone"`
	StartTime          time.Time  `json:"start_time"`
	EndTime            time.Time  `json:"end_time"`
	RequiresTicket     bool       `json:"requires_ticket"`
	Capacity           int        `json:"capacity"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
}

// IsDeleted checks if the event has been soft deleted
func (e *Event) IsDeleted() bool {
	return e.DeletedAt != nil
}

// HasStarted checks if the event has already started
func (e *Event) HasStarted() bool {
	return time.Now().UTC().After(e.StartTime)
}

// HasEnded checks if the event has ended
func (e *Event) HasEnded() bool {
	return time.Now().UTC().After(e.EndTime)
}

// IsOngoing checks if the event is currently happening
func (e *Event) IsOngoing() bool {
	now := time.Now().UTC()
	return now.After(e.StartTime) && now.Before(e.EndTime)
}

// IsUpcoming checks if the event is in the future
func (e *Event) IsUpcoming() bool {
	return time.Now().UTC().Before(e.StartTime)
}

// HasCapacity checks if there are no capacity limits
func (e *Event) HasCapacity() bool {
	return e.Capacity == 0 // 0 means unlimited
}

// EventRSVP represents a profile's RSVP to an event
type EventRSVP struct {
	ID        string    `json:"id"`
	EventID   string    `json:"event_id"`
	ProfileID string    `json:"profile_id"`
	Status    string    `json:"status"` // attending, maybe, not_attending
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IsAttending checks if the profile is marked as attending
func (r *EventRSVP) IsAttending() bool {
	return r.Status == RSVPStatusAttending
}

// RSVP status constants
const (
	RSVPStatusAttending    = "attending"
	RSVPStatusMaybe        = "maybe"
	RSVPStatusNotAttending = "not_attending"
)
