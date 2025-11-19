package dashboard

import (
	"context"
	"html/template"
	"time"
)

type Repository interface {
	GetCoordinationNeeds(ctx context.Context, profileID string) ([]CoordinationNeed, error)
	GetUnreadMessagesSummary(ctx context.Context, profileID string) (*MessagesSummary, error)
	GetRecentUpdates(ctx context.Context, profileID string, since time.Duration) ([]Activity, error)
	GetUpcomingEvents(ctx context.Context, profileID string, days int) ([]EventWithDetails, error)
	GetUserCirclesWithActivity(ctx context.Context, profileID string) ([]CircleActivity, error)
	GetSerendipityRecommendations(ctx context.Context, profileID string) (*Recommendation, error)
	GetPendingInvitations(ctx context.Context, profileID string) ([]PendingInvitation, error)
}

type CoordinationNeed struct {
	EventID    string
	Type       string
	Message    string
	AuthorName string
	CircleName string
	CreatedAt  time.Time
}

type MessagesSummary struct {
	UnreadCount int
	Senders     []MessageSender
}

type MessageSender struct {
	Name          string
	SharedCircles []string
}

type Activity struct {
	Type       string
	ActorName  string
	CircleName string
	Content    string
	OccurredAt time.Time
}

type EventWithDetails struct {
	ID            string
	Title         string
	StartTime     time.Time
	CircleID      string
	CircleName    string
	CircleIcon    string
	CircleBgColor template.CSS
	CircleAvatar  string
	AttendeeCount int
	RSVPStatus    string
	Coordination  *CoordinationNeed
}

type CircleActivity struct {
	ID            string
	Name          string
	Icon          string
	IconBgColor   template.CSS
	AvatarURL     string
	BannerURL     string
	LastActivity  time.Time
	NextEventTime *time.Time
}

type Recommendation struct {
	SerendipityCount int
	Domain           string
	EventTitle       string
	EventDescription string
	EventTime        string
}

type PendingInvitation struct {
	MembershipID  string
	CircleID      string
	CircleName    string
	CircleIcon    string
	CircleBgColor template.CSS
	CircleAvatar  string
	InviterID     string
	InviterName   string
	InviterHandle string
	InvitedAt     time.Time
}
