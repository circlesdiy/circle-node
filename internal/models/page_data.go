package models

import "circles.diy/internal/domain"

type PageData struct {
	Success   bool
	CSRFToken string
}

type ThemeSettings struct {
	Mode   string `json:"mode"`   // light, dark, system
	Radius string `json:"radius"` // 0, 6, 12, 32
}

type BaseData struct {
	Title     string
	ActiveNav string
	Theme     ThemeSettings
	CSRFToken string
	User      *domain.User `json:"user,omitempty"`
}

type DashboardData struct {
	BaseData
	Feed             []FeedItem        `json:"feed"`
	FeedOffset       int               `json:"feed_offset"`
	Circles          []Circle          `json:"circles"`
	Discussions      []Discussion      `json:"discussions"`
	Events           []Event           `json:"events"`
	Ripples          []Ripple          `json:"ripples"`
	MarketplaceItems []MarketplaceItem `json:"marketplace_items"`
	Impact           []ImpactItem      `json:"impact"`
}

type ProfileData struct {
	BaseData
	Profile      Profile     `json:"profile"`
	Posts        []Post      `json:"posts"`
	PostOffset   int         `json:"post_offset"`
	HasMorePosts bool        `json:"has_more_posts"`
	IsOwner      bool        `json:"is_owner"`
	Extensions   []Extension `json:"extensions"`
	Analytics    Analytics   `json:"analytics"`
	Drafts       []DraftPost `json:"drafts"`
	DraftCount   int         `json:"draft_count"`
}

type CirclesPageData struct {
	BaseData
	Circles         []Circle         `json:"circles"`
	RecentActivity  []CircleActivity `json:"recent_activity"`
	Stats           CircleStats      `json:"stats"`
	FeaturedCircles []Circle         `json:"featured_circles"`
}

type ChatPageData struct {
	BaseData
	Conversations []Conversation `json:"conversations"`
	ActiveChat    *Conversation  `json:"active_chat,omitempty"`
	Messages      []Message      `json:"messages"`
	Contacts      []Contact      `json:"contacts"`
}

type Conversation struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Avatar       string  `json:"avatar"`
	LastMessage  string  `json:"last_message"`
	LastTime     string  `json:"last_time"`
	UnreadCount  int     `json:"unread_count"`
	IsOnline     bool    `json:"is_online"`
	IsGroup      bool    `json:"is_group"`
	Participants []User  `json:"participants,omitempty"`
}

type Message struct {
	ID        string     `json:"id"`
	Content   string     `json:"content"`
	Timestamp string     `json:"timestamp"`
	Sender    User       `json:"sender"`
	IsOwn     bool       `json:"is_own"`
	IsRead    bool       `json:"is_read"`
	Type      string     `json:"type"` // text, image, voice, video, call
	Media     *MediaItem `json:"media,omitempty"`
}

type Contact struct {
	User
	IsOnline     bool   `json:"is_online"`
	LastSeen     string `json:"last_seen,omitempty"`
	Relationship string `json:"relationship"` // friend, circle_member, etc.
}

type GatherPageData struct {
	BaseData
	CircleGatherings  []CircleGathering  `json:"circle_gatherings"`
	DomainDiscovery   []DomainSection    `json:"domain_discovery"`
	UserCircles       []CircleOption     `json:"user_circles"`
	ActiveCoordination []CoordinationItem `json:"active_coordination,omitempty"`
	UserHosting       []HostingEvent     `json:"user_hosting,omitempty"`
	// Legacy fields - kept for backward compatibility
	FeaturedEvents   []GatherEvent      `json:"featured_events,omitempty"`
	UpcomingEvents   []GatherEvent      `json:"upcoming_events,omitempty"`
	MyEvents         []GatherEvent      `json:"my_events,omitempty"`
	EventCategories  []EventCategory    `json:"event_categories,omitempty"`
	PopularLocations []EventLocation    `json:"popular_locations,omitempty"`
}

type GatherEvent struct {
	ID              string          `json:"id"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	Host            User            `json:"host"`
	Circle          string          `json:"circle,omitempty"`
	DateTime        string          `json:"date_time"`
	TimeAgo         string          `json:"time_ago"`
	Duration        string          `json:"duration"`
	Location        EventLocation   `json:"location"`
	Type            string          `json:"type"` // in-person, online, hybrid
	Category        string          `json:"category"`
	IsTicketed      bool            `json:"is_ticketed"`
	Price           string          `json:"price,omitempty"`
	Currency        string          `json:"currency,omitempty"`
	Capacity        int             `json:"capacity"`
	AttendeeCount   int             `json:"attendee_count"`
	RSVPStatus      string          `json:"rsvp_status"` // going, maybe, not_going, not_responded
	IsHost          bool            `json:"is_host"`
	Image           *MediaItem      `json:"image,omitempty"`
	Tags            []string        `json:"tags"`
	Announcements   []Announcement  `json:"announcements"`
	Attendees       []EventAttendee `json:"attendees"`
}

type EventLocation struct {
	Type        string  `json:"type"` // venue, online, address
	Name        string  `json:"name"`
	Address     string  `json:"address,omitempty"`
	City        string  `json:"city,omitempty"`
	OnlineLink  string  `json:"online_link,omitempty"`
	Coordinates string  `json:"coordinates,omitempty"`
}

type EventCategory struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Count int    `json:"count"`
}

type Announcement struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Author  User   `json:"author"`
	TimeAgo string `json:"time_ago"`
}

type EventAttendee struct {
	User
	RSVPStatus string `json:"rsvp_status"`
	JoinedAt   string `json:"joined_at"`
}

// New Gather Page Types
type CircleGathering struct {
	CircleID         string           `json:"circle_id"`
	CircleName       string           `json:"circle_name"`
	Icon             string           `json:"icon"`
	IconBgColor      string           `json:"icon_bg_color"`
	GatheringCount   int              `json:"gathering_count"`
	NeedsCoordination bool            `json:"needs_coordination"`
	Gatherings       []GatheringItem  `json:"gatherings"`
}

type GatheringItem struct {
	ID                 string              `json:"id"`
	Title              string              `json:"title"`
	TimeAgo            string              `json:"time_ago"`
	TimeRange          string              `json:"time_range"`
	RSVPStatus         string              `json:"rsvp_status"`
	Description        string              `json:"description,omitempty"`
	Location           string              `json:"location,omitempty"`
	AttendeeCount      int                 `json:"attendee_count,omitempty"`
	CoordinationNeeded *CoordinationNeed   `json:"coordination_needed,omitempty"`
}

type CoordinationNeed struct {
	Icon  string `json:"icon"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

type DomainSection struct {
	Name             string           `json:"name"`
	Icon             string           `json:"icon"`
	SerendipityCount int              `json:"serendipity_count,omitempty"`
	Gatherings       []DomainGathering `json:"gatherings"`
}

type DomainGathering struct {
	ID                 string `json:"id"`
	Title              string `json:"title"`
	TimeAgo            string `json:"time_ago"`
	TimeRange          string `json:"time_range"`
	Description        string `json:"description,omitempty"`
	Location           string `json:"location,omitempty"`
	SerendipityCircles int    `json:"serendipity_circles,omitempty"`
}

type CircleOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CoordinationItem struct {
	Label       string `json:"label"`
	Value       string `json:"value"`
	GatheringID string `json:"gathering_id"`
}

type HostingEvent struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	TimeAgo string `json:"time_ago"`
}

type MarketplacePageData struct {
	BaseData
	Items          []MarketplaceItem      `json:"items"`
	FeaturedItems  []MarketplaceItem      `json:"featured_items"`
	Categories     []MarketplaceCategory  `json:"categories"`
	PopularLocations []Location           `json:"popular_locations"`
	TotalItems     int                    `json:"total_items"`
	ItemsPerPage   int                    `json:"items_per_page"`
	CurrentPage    int                    `json:"current_page"`
	HasMore        bool                   `json:"has_more"`
	Filters        MarketplaceFilter      `json:"filters"`
	ActiveFilters  map[string]interface{} `json:"active_filters"`
}