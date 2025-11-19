package models

import (
	"html/template"

	"circles.diy/internal/domain"
)

// ReactionResponse represents the response for reaction endpoints
type ReactionResponse struct {
	Counts       map[string]int     `json:"counts"`
	UserReaction *domain.Reaction   `json:"user_reaction,omitempty"`
}

type PageData struct {
	Success   bool
	CSRFToken string
}

type ThemeSettings struct {
	Mode   string `json:"mode"`   // light, dark, system
	Radius string `json:"radius"` // 0, 6, 12, 32
}

type BaseData struct {
	Title        string
	ActiveNav    string
	Theme        ThemeSettings
	CSRFToken    string
	User         *domain.User `json:"user,omitempty"`
	ProfileName  string       `json:"profile_name,omitempty"`
	AssetVersion string       `json:"asset_version"` // For cache busting static assets
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
	Circles           []Circle           `json:"circles"`
	PendingInvitations []CircleInvitation `json:"pending_invitations"`
	RecentActivity    []CircleActivity   `json:"recent_activity"`
	Stats             CircleStats        `json:"stats"`
	FeaturedCircles   []Circle           `json:"featured_circles"`
}

type CircleDetailPageData struct {
	BaseData
	Circle              domain.Circle       `json:"circle"`
	Members             []CircleMember      `json:"members"`
	MemberCount         int                 `json:"member_count"`
	RecentPosts         []CirclePost        `json:"recent_posts"`
	UpcomingGatherings  []GatheringItem     `json:"upcoming_gatherings"`
	SharedFiles         []CircleFile        `json:"shared_files"`
	IsOwner             bool                `json:"is_owner"`
	IsAdmin             bool                `json:"is_admin"`
	IsMember            bool                `json:"is_member"`
	CanInvite           bool                `json:"can_invite"`
	CanEditInfo         bool                `json:"can_edit_info"`
	CanEditVisibility   bool                `json:"can_edit_visibility"`
	CanEditPermissions  bool                `json:"can_edit_permissions"`
	UserRole            string              `json:"user_role"` // owner, admin, member, or empty
	CircleStats         CircleDetailStats   `json:"circle_stats"`
	ActiveTab           string              `json:"active_tab"` // chat, gatherings, files, members, settings
}

type CircleDetailStats struct {
	TotalPosts      int    `json:"total_posts"`
	TotalFiles      int    `json:"total_files"`
	TotalGatherings int    `json:"total_gatherings"`
	CreatedAt       string `json:"created_at"`
	LastActivity    string `json:"last_activity"`
}

type CircleMember struct {
	ProfileID string `json:"profile_id"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	Avatar    string `json:"avatar"`
	Role      string `json:"role"`       // owner, admin, member
	State     string `json:"state"`      // active, invited, banned, left
	JoinedAt  string `json:"joined_at"`
	LastActive string `json:"last_active,omitempty"`
}

type CirclePost struct {
	ID                string       `json:"id"`
	AuthorProfileID   string       `json:"author_profile_id"`
	AuthorName        string       `json:"author_name"`
	AuthorAvatar      string       `json:"author_avatar"`
	Body              string       `json:"body"`
	BodyFormat        string       `json:"body_format"`
	ContentWarning    string       `json:"content_warning"`
	Visibility        string       `json:"visibility"`
	CreatedAt         string       `json:"created_at"`
	EditedAt          *string      `json:"edited_at,omitempty"`
	FormattedTime     string       `json:"formatted_time"`
	IsEdited          bool         `json:"is_edited"`
	ReplyCount        int          `json:"reply_count"`
	LikeCount         int          `json:"like_count"`
	UserHasLiked      bool         `json:"user_has_liked"`
	CanEdit           bool         `json:"can_edit"`
	ShowComments      bool         `json:"show_comments"`
	Attachments       []Attachment `json:"attachments,omitempty"`
}

type CircleFile struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	SizeDisplay string `json:"size_display"`
	Type        string `json:"type"`
	Icon        string `json:"icon"`
	UploaderID  string `json:"uploader_id"`
	Uploader    string `json:"uploader"`
	UploadedAt  string `json:"uploaded_at"`
	TimeAgo     string `json:"time_ago"`
	DownloadURL string `json:"download_url"`
}

type Attachment struct {
	ID   string `json:"id"`
	Type string `json:"type"` // image, file, audio, video
	Name string `json:"name"`
	URL  string `json:"url"`
	Size int64  `json:"size,omitempty"`
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
	IconBgColor      template.CSS     `json:"icon_bg_color"`
	AvatarURL        string           `json:"avatar_url,omitempty"`
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

// Event Management Page Data
type EventCreatePageData struct {
	BaseData
	UserCircles          []CircleOption `json:"user_circles"`
	PreselectedCircleID  string         `json:"preselected_circle_id,omitempty"`
}

type EventDetailPageData struct {
	BaseData
	Event              domain.Event   `json:"event"`
	CircleName         string         `json:"circle_name"`
	OrganizerName      string         `json:"organizer_name"`
	OrganizerAvatar    string         `json:"organizer_avatar"`
	FormattedStartTime string         `json:"formatted_start_time"`
	FormattedEndTime   string         `json:"formatted_end_time,omitempty"`
	ViewerRSVPStatus   string         `json:"viewer_rsvp_status"`
	IsOrganizer        bool           `json:"is_organizer"`
	AttendeeCount      int            `json:"attendee_count"`
	Attendees          []EventAttendee `json:"attendees"`
	CoordinationNeeds  []EventCoordinationNeed `json:"coordination_needs"`
}

type EventCoordinationNeed struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	Message    string  `json:"message"`
	AuthorName string  `json:"author_name"`
	CreatedAt  string  `json:"created_at"`
	ResolvedAt *string `json:"resolved_at,omitempty"`
}

type MarketplacePageData struct{
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