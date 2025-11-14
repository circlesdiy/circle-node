package gather

import (
	"context"
	"html/template"
	"time"
)

// Repository defines the interface for gather data access
type Repository interface {
	// GetUserCircleEvents retrieves all events for circles the user is a member of
	GetUserCircleEvents(ctx context.Context, profileID string) ([]CircleWithEvents, error)

	// GetDomainEvents retrieves public events from domains based on user's interests
	GetDomainEvents(ctx context.Context, profileID string) ([]DomainEvents, error)

	// GetUserCoordinationNeeds retrieves active coordination needs for events the user is involved in
	GetUserCoordinationNeeds(ctx context.Context, profileID string) ([]CoordinationNeedItem, error)

	// GetUserHostedEvents retrieves events where the user is the organizer
	GetUserHostedEvents(ctx context.Context, profileID string) ([]HostedEvent, error)

	// GetUserCircles retrieves all circles the user is a member of (for the "organize" dropdown)
	GetUserCircles(ctx context.Context, profileID string) ([]Circle, error)

	// GetEventDetails retrieves detailed information about a specific event
	GetEventDetails(ctx context.Context, eventID, profileID string) (*EventDetails, error)

	// UpdateRSVP updates or creates an RSVP for an event
	UpdateRSVP(ctx context.Context, eventID, profileID, status string) error

	// GetEventAttendeeCount returns the count of attending/maybe attendees
	GetEventAttendeeCount(ctx context.Context, eventID string) (int, error)
}

// CircleWithEvents represents a circle and its associated events
type CircleWithEvents struct {
	CircleID      string
	CircleName    string
	Icon          string
	IconBgColor   template.CSS
	AvatarURL     string
	Events        []EventItem
	HasUnresolved bool // true if any events have unresolved coordination needs
}

// EventItem represents a single event in a circle's gathering list
type EventItem struct {
	ID                    string
	Title                 string
	StartTime             time.Time
	EndTime               time.Time
	Location              string
	Description           string
	AttendeeCount         int
	UserRSVPStatus        string // attending, maybe, not_attending, or empty
	CoordinationType      string // transport, tickets, supplies, or empty
	CoordinationMessage   string
	CoordinationResolved  bool
	OrganizerProfileID    string
	OrganizerName         string
}

// DomainEvents represents events grouped by domain
type DomainEvents struct {
	DomainName       string
	Icon             string
	SerendipityCount int // number of circles with matching interests
	Events           []DomainEvent
}

// DomainEvent represents a public event from a domain
type DomainEvent struct {
	ID              string
	Title           string
	StartTime       time.Time
	EndTime         time.Time
	Location        string
	Description     string
	CircleName      string
	MatchingCircles int // how many user circles have compatible interests
}

// CoordinationNeedItem represents an active coordination need
type CoordinationNeedItem struct {
	EventID     string
	EventTitle  string
	Type        string // transport, tickets, supplies
	Label       string // human-readable label
	Value       string // human-readable value
}

// HostedEvent represents an event the user is organizing
type HostedEvent struct {
	ID        string
	Title     string
	StartTime time.Time
}

// Circle represents a circle the user is a member of
type Circle struct {
	ID   string
	Name string
	Icon string
}

// EventDetails represents detailed information about an event
type EventDetails struct {
	Event         EventItem
	CircleName    string
	OrganizerName string
	Attendees     []Attendee
	Announcements []Announcement
}

// Attendee represents someone who has RSVP'd to an event
type Attendee struct {
	ProfileID  string
	Name       string
	AvatarURL  string
	RSVPStatus string
	JoinedAt   time.Time
}

// Announcement represents an announcement for an event
type Announcement struct {
	ID         string
	Title      string
	Content    string
	AuthorName string
	CreatedAt  time.Time
}
