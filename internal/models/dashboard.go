package models

import "html/template"

type DashboardData struct {
	BaseData
	WhatsNew       []WhatsNewItem
	UpcomingEvents []UpcomingEvent
	CirclesSummary []CircleSummary
	Discovery      *DiscoverySection
}

type WhatsNewItem struct {
	Type         string
	Icon         string
	IconBgColor  template.CSS
	Title        string
	Description  string
	SourceCircle string
	TimeAgo      string
}

type UpcomingEvent struct {
	ID               string
	DayOfWeek        string
	DayOfMonth       int
	Time             string
	Title            string
	Circle           CircleBadge
	AttendeeCount    int
	RSVPStatus       string
	HasCoordination  bool
	CoordinationText string
}

type CircleBadge struct {
	Name        string
	Icon        string
	IconBgColor template.CSS
	AvatarURL   string
}

type CircleSummary struct {
	ID           string
	Name         string
	Icon         string
	IconBgColor  template.CSS
	AvatarURL    string
	BannerURL    string
	LastActivity string
	NextEvent    string
}

type DiscoverySection struct {
	SerendipityCount int
	Domain           string
	Title            string
	Description      string
	EventTime        string
}
