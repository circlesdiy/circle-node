package events

import (
	"context"

	"circles.diy/internal/domain"
)

// Repository defines the interface for event data access
type Repository interface {
	// CreateEvent creates a new event
	CreateEvent(ctx context.Context, event *domain.Event) error

	// GetEventByID retrieves an event by ID
	GetEventByID(ctx context.Context, eventID string) (*domain.Event, error)

	// GetEventDetails retrieves detailed event information including attendees and coordination
	GetEventDetails(ctx context.Context, eventID, viewerProfileID string) (*EventDetails, error)

	// UpdateEvent updates an event
	UpdateEvent(ctx context.Context, event *domain.Event) error

	// DeleteEvent soft deletes an event
	DeleteEvent(ctx context.Context, eventID string) error

	// GetEventsByCircle retrieves all events for a circle
	GetEventsByCircle(ctx context.Context, circleID string) ([]domain.Event, error)

	// CreateCoordinationNeed adds a coordination need for an event
	CreateCoordinationNeed(ctx context.Context, need *CoordinationNeed) error

	// ResolveCoordinationNeed marks a coordination need as resolved
	ResolveCoordinationNeed(ctx context.Context, needID string) error

	// GetCoordinationNeeds retrieves coordination needs for an event
	GetCoordinationNeeds(ctx context.Context, eventID string) ([]CoordinationNeed, error)
}

// EventDetails represents detailed event information
type EventDetails struct {
	Event              domain.Event
	CircleName         string
	OrganizerName      string
	OrganizerAvatarURL string
	ViewerRSVPStatus   string
	Attendees          []Attendee
	CoordinationNeeds  []CoordinationNeed
}

// Attendee represents someone who has RSVP'd to an event
type Attendee struct {
	ProfileID  string
	Name       string
	AvatarURL  string
	RSVPStatus string
}

// CoordinationNeed represents a coordination item for an event
type CoordinationNeed struct {
	ID              string
	EventID         string
	Type            string // transport, tickets, supplies
	Message         string
	AuthorProfileID string
	AuthorName      string
	CreatedAt       string
	ResolvedAt      *string
}
