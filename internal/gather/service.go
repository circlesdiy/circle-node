package gather

import (
	"context"
	"fmt"
	"html/template"
	"sync"
	"time"

	"circles.diy/internal/models"
	"go.uber.org/zap"
)

type Service struct {
	repo   Repository
	logger *zap.Logger
}

func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) GetGatherPageData(ctx context.Context, profileID string) (*models.GatherPageData, error) {
	var (
		circleGatherings   []models.CircleGathering
		domainDiscovery    []models.DomainSection
		userCircles        []models.CircleOption
		activeCoordination []models.CoordinationItem
		userHosting        []models.HostingEvent

		wg sync.WaitGroup
		mu sync.Mutex
	)

	// Fetch data in parallel for performance
	wg.Add(5)

	go func() {
		defer wg.Done()
		cg, err := s.buildCircleGatherings(ctx, profileID)
		if err != nil {
			s.logger.Error("failed to build circle gatherings", zap.Error(err))
			return
		}
		mu.Lock()
		circleGatherings = cg
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		dd, err := s.buildDomainDiscovery(ctx, profileID)
		if err != nil {
			s.logger.Error("failed to build domain discovery", zap.Error(err))
			return
		}
		mu.Lock()
		domainDiscovery = dd
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		uc, err := s.buildUserCircles(ctx, profileID)
		if err != nil {
			s.logger.Error("failed to build user circles", zap.Error(err))
			return
		}
		mu.Lock()
		userCircles = uc
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		ac, err := s.buildActiveCoordination(ctx, profileID)
		if err != nil {
			s.logger.Error("failed to build active coordination", zap.Error(err))
			return
		}
		mu.Lock()
		activeCoordination = ac
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		uh, err := s.buildUserHosting(ctx, profileID)
		if err != nil {
			s.logger.Error("failed to build user hosting", zap.Error(err))
			return
		}
		mu.Lock()
		userHosting = uh
		mu.Unlock()
	}()

	wg.Wait()

	return &models.GatherPageData{
		CircleGatherings:   circleGatherings,
		DomainDiscovery:    domainDiscovery,
		UserCircles:        userCircles,
		ActiveCoordination: activeCoordination,
		UserHosting:        userHosting,
	}, nil
}

func (s *Service) buildCircleGatherings(ctx context.Context, profileID string) ([]models.CircleGathering, error) {
	circlesWithEvents, err := s.repo.GetUserCircleEvents(ctx, profileID)
	if err != nil {
		return nil, err
	}

	var gatherings []models.CircleGathering
	for _, cwe := range circlesWithEvents {
		var gatheringItems []models.GatheringItem
		for _, event := range cwe.Events {
			item := models.GatheringItem{
				ID:            event.ID,
				Title:         event.Title,
				TimeAgo:       formatRelativeTime(event.StartTime),
				TimeRange:     formatTimeRange(event.StartTime, event.EndTime),
				RSVPStatus:    s.mapRSVPStatus(event.UserRSVPStatus),
				Location:      event.Location,
				AttendeeCount: event.AttendeeCount,
			}

			// Add coordination info if needed
			if event.CoordinationType != "" && !event.CoordinationResolved {
				item.CoordinationNeeded = &models.CoordinationNeed{
					Icon:  s.getCoordinationIcon(event.CoordinationType),
					Title: s.getCoordinationTitle(event.CoordinationType),
					Text:  event.CoordinationMessage,
				}
			}

			gatheringItems = append(gatheringItems, item)
		}

		gatherings = append(gatherings, models.CircleGathering{
			CircleID:          cwe.CircleID,
			CircleName:        cwe.CircleName,
			Icon:              cwe.Icon,
			IconBgColor:       template.CSS(cwe.IconBgColor),
			AvatarURL:         cwe.AvatarURL,
			GatheringCount:    len(gatheringItems),
			NeedsCoordination: cwe.HasUnresolved,
			Gatherings:        gatheringItems,
		})
	}

	return gatherings, nil
}

func (s *Service) buildDomainDiscovery(ctx context.Context, profileID string) ([]models.DomainSection, error) {
	domainEvents, err := s.repo.GetDomainEvents(ctx, profileID)
	if err != nil {
		return nil, err
	}

	var sections []models.DomainSection
	for _, de := range domainEvents {
		var gatherings []models.DomainGathering
		for _, event := range de.Events {
			gatherings = append(gatherings, models.DomainGathering{
				ID:                 event.ID,
				Title:              event.Title,
				TimeAgo:            formatRelativeTime(event.StartTime),
				TimeRange:          formatTimeRange(event.StartTime, event.EndTime),
				Description:        event.Description,
				Location:           event.Location,
				SerendipityCircles: event.MatchingCircles,
			})
		}

		sections = append(sections, models.DomainSection{
			Name:             de.DomainName,
			Icon:             de.Icon,
			SerendipityCount: de.SerendipityCount,
			Gatherings:       gatherings,
		})
	}

	return sections, nil
}

func (s *Service) buildUserCircles(ctx context.Context, profileID string) ([]models.CircleOption, error) {
	circles, err := s.repo.GetUserCircles(ctx, profileID)
	if err != nil {
		return nil, err
	}

	var options []models.CircleOption
	for _, circle := range circles {
		options = append(options, models.CircleOption{
			ID:   circle.ID,
			Name: circle.Name,
		})
	}

	return options, nil
}

func (s *Service) buildActiveCoordination(ctx context.Context, profileID string) ([]models.CoordinationItem, error) {
	needs, err := s.repo.GetUserCoordinationNeeds(ctx, profileID)
	if err != nil {
		return nil, err
	}

	var items []models.CoordinationItem
	for _, need := range needs {
		items = append(items, models.CoordinationItem{
			Label:       need.Label,
			Value:       need.Value,
			GatheringID: need.EventID,
		})
	}

	return items, nil
}

func (s *Service) buildUserHosting(ctx context.Context, profileID string) ([]models.HostingEvent, error) {
	events, err := s.repo.GetUserHostedEvents(ctx, profileID)
	if err != nil {
		return nil, err
	}

	var hosting []models.HostingEvent
	for _, event := range events {
		hosting = append(hosting, models.HostingEvent{
			ID:      event.ID,
			Title:   event.Title,
			TimeAgo: formatRelativeTime(event.StartTime),
		})
	}

	return hosting, nil
}

// Helper functions

func (s *Service) mapRSVPStatus(status string) string {
	switch status {
	case "attending":
		return "going"
	case "maybe":
		return "maybe"
	case "not_attending":
		return "not_going"
	default:
		return "rsvp"
	}
}

func (s *Service) getCoordinationIcon(coordType string) string {
	switch coordType {
	case "transport":
		return "🚗"
	case "tickets":
		return "🎫"
	case "supplies":
		return "📦"
	default:
		return "❓"
	}
}

func (s *Service) getCoordinationTitle(coordType string) string {
	switch coordType {
	case "transport":
		return "Transport"
	case "tickets":
		return "Tickets"
	case "supplies":
		return "Supplies"
	default:
		return "Coordination"
	}
}

// UpdateRSVP updates a user's RSVP status for an event
func (s *Service) UpdateRSVP(ctx context.Context, eventID, profileID, status string) error {
	// Validate status
	if status != "attending" && status != "maybe" && status != "not_attending" {
		return fmt.Errorf("invalid RSVP status: %s", status)
	}

	return s.repo.UpdateRSVP(ctx, eventID, profileID, status)
}

// GetEventAttendeeCount returns the count of attendees for an event
func (s *Service) GetEventAttendeeCount(ctx context.Context, eventID string) (int, error) {
	return s.repo.GetEventAttendeeCount(ctx, eventID)
}

// GetUserCircles returns circles for the user to display in dropdowns
func (s *Service) GetUserCircles(ctx context.Context, profileID string) ([]models.CircleOption, error) {
	return s.buildUserCircles(ctx, profileID)
}

func formatRelativeTime(t time.Time) string {
	duration := time.Until(t)

	if duration < 0 {
		// Event has passed or is ongoing
		return "Now"
	}

	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 0 {
			return "Soon"
		}
		return fmt.Sprintf("In %d hours", hours)
	}

	days := int(duration.Hours() / 24)
	if days == 1 {
		return "Tomorrow"
	}
	if days < 7 {
		return fmt.Sprintf("In %d days", days)
	}

	// For events within 2 weeks, show day name
	if days < 14 {
		dayName := t.Format("Monday")
		if days < 7 {
			return fmt.Sprintf("Next %s", dayName)
		}
		return dayName
	}

	// For further out events, show date
	return t.Format("Jan 2")
}

func formatTimeRange(start, end time.Time) string {
	startFmt := start.Format("3:04 PM")
	endFmt := end.Format("3:04 PM")
	return fmt.Sprintf("%s - %s", startFmt, endFmt)
}
