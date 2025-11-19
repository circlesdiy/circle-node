package dashboard

import (
	"context"
	"fmt"
	"html/template"
	"strings"
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

func (s *Service) GetDashboardData(ctx context.Context, profileID string) (*models.DashboardData, error) {
	var (
		greeting       string
		whatsNew       []models.WhatsNewItem
		upcomingEvents []models.UpcomingEvent
		circlesSummary []models.CircleSummary
		discovery      *models.DiscoverySection

		wg sync.WaitGroup
		mu sync.Mutex
	)

	wg.Add(5)

	go func() {
		defer wg.Done()

		// TODO: Timezone aware greetings
		now := time.Now()
		hour := now.Hour()

		var greetingForCurrentTime string
		switch {
		case hour >= 5 && hour < 12:
			greetingForCurrentTime = "Good morning"
		case hour >= 12 && hour < 17:
			greetingForCurrentTime = "Good afternoon"
		default:
			greetingForCurrentTime = "Good evening"
		}

		mu.Lock()
		greeting = greetingForCurrentTime
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		items, err := s.buildWhatsNew(ctx, profileID)
		if err != nil {
			s.logger.Error("failed to build whats new", zap.Error(err))
			return
		}
		mu.Lock()
		whatsNew = items
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		events, err := s.buildUpcomingEvents(ctx, profileID)
		if err != nil {
			s.logger.Error("failed to build upcoming events", zap.Error(err))
			return
		}
		mu.Lock()
		upcomingEvents = events
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		circles, err := s.buildCirclesSummary(ctx, profileID)
		if err != nil {
			s.logger.Error("failed to build circles summary", zap.Error(err))
			return
		}
		mu.Lock()
		circlesSummary = circles
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		disc, err := s.buildDiscovery(ctx, profileID)
		if err != nil {
			s.logger.Error("failed to build discovery", zap.Error(err))
			return
		}
		mu.Lock()
		discovery = disc
		mu.Unlock()
	}()

	wg.Wait()

	return &models.DashboardData{
		Greeting:       greeting,
		WhatsNew:       whatsNew,
		UpcomingEvents: upcomingEvents,
		CirclesSummary: circlesSummary,
		Discovery:      discovery,
	}, nil
}

func (s *Service) buildWhatsNew(ctx context.Context, profileID string) ([]models.WhatsNewItem, error) {
	var items []models.WhatsNewItem

	// Get pending circle invitations
	invitations, err := s.repo.GetPendingInvitations(ctx, profileID)
	if err != nil {
		s.logger.Error("failed to get pending invitations", zap.Error(err))
	} else {
		for _, inv := range invitations {
			items = append(items, models.WhatsNewItem{
				Type:         "invitation",
				Icon:         "👋",
				IconBgColor:  template.CSS("var(--primary-accent)"),
				Title:        fmt.Sprintf("Invitation to %s", inv.CircleName),
				Description:  fmt.Sprintf("%s invited you to join", inv.InviterName),
				SourceCircle: inv.CircleName,
				TimeAgo:      formatTimeAgo(inv.InvitedAt),
				CircleID:     inv.CircleID,
				EntityID:     inv.MembershipID,
			})
		}
	}

	coordination, err := s.repo.GetCoordinationNeeds(ctx, profileID)
	if err != nil {
		return items, nil
	}

	for _, c := range coordination {
		iconType := "🚗"
		if c.Type == "tickets" {
			iconType = "🎫"
		} else if c.Type == "supplies" {
			iconType = "📦"
		}

		items = append(items, models.WhatsNewItem{
			Type:         "coordination",
			Icon:         iconType,
			IconBgColor:  template.CSS("var(--warm-accent)"),
			Title:        fmt.Sprintf("%s coordination needed", strings.Title(c.Type)),
			Description:  fmt.Sprintf("%s asked \"%s\"", c.AuthorName, c.Message),
			SourceCircle: c.CircleName,
			TimeAgo:      formatTimeAgo(c.CreatedAt),
		})
	}

	messages, err := s.repo.GetUnreadMessagesSummary(ctx, profileID)
	if err != nil {
		return items, nil
	}

	if messages.UnreadCount > 0 && len(messages.Senders) > 0 {
		var senderLines []string
		for _, sender := range messages.Senders {
			if len(sender.SharedCircles) > 0 {
				senderLines = append(senderLines, fmt.Sprintf("%s (%s)", sender.Name, sender.SharedCircles[0]))
			} else {
				senderLines = append(senderLines, sender.Name)
			}
		}

		items = append(items, models.WhatsNewItem{
			Type:        "messages",
			Icon:        "💬",
			IconBgColor: template.CSS("var(--cool-accent)"),
			Title:       "New messages",
			Description: strings.Join(senderLines, "\n"),
		})
	}

	return items, nil
}

func (s *Service) buildUpcomingEvents(ctx context.Context, profileID string) ([]models.UpcomingEvent, error) {
	events, err := s.repo.GetUpcomingEvents(ctx, profileID, 7)
	if err != nil {
		return nil, err
	}

	var upcomingEvents []models.UpcomingEvent
	for _, e := range events {
		timeStr := e.StartTime.Format("3pm")
		if e.StartTime.Minute() != 0 {
			timeStr = e.StartTime.Format("3:04pm")
		}

		ue := models.UpcomingEvent{
			ID:         e.ID,
			DayOfWeek:  e.StartTime.Format("Mon"),
			DayOfMonth: e.StartTime.Day(),
			Time:       timeStr,
			Title:      e.Title,
			Circle: models.CircleBadge{
				Name:        e.CircleName,
				Icon:        e.CircleIcon,
				IconBgColor: template.CSS(e.CircleBgColor),
				AvatarURL:   e.CircleAvatar,
			},
			AttendeeCount: e.AttendeeCount,
			RSVPStatus:    e.RSVPStatus,
		}

		if e.Coordination != nil {
			ue.HasCoordination = true
			ue.CoordinationText = fmt.Sprintf("%s's coordinating %s", e.Coordination.AuthorName, e.Coordination.Type)
		}

		upcomingEvents = append(upcomingEvents, ue)
	}

	return upcomingEvents, nil
}

func (s *Service) buildCirclesSummary(ctx context.Context, profileID string) ([]models.CircleSummary, error) {
	circles, err := s.repo.GetUserCirclesWithActivity(ctx, profileID)
	if err != nil {
		return nil, err
	}

	var summary []models.CircleSummary
	for _, c := range circles {
		cs := models.CircleSummary{
			ID:           c.ID,
			Name:         c.Name,
			Icon:         c.Icon,
			IconBgColor:  template.CSS(c.IconBgColor),
			AvatarURL:    c.AvatarURL,
			BannerURL:    c.BannerURL,
			LastActivity: fmt.Sprintf("Active %s", formatTimeAgo(c.LastActivity)),
		}

		if c.NextEventTime != nil {
			dayName := c.NextEventTime.Format("Monday")
			cs.NextEvent = fmt.Sprintf("Gathering on %s", dayName)
		}

		summary = append(summary, cs)
	}

	return summary, nil
}

func (s *Service) buildDiscovery(ctx context.Context, profileID string) (*models.DiscoverySection, error) {
	rec, err := s.repo.GetSerendipityRecommendations(ctx, profileID)
	if err != nil || rec == nil {
		return nil, nil
	}

	return &models.DiscoverySection{
		SerendipityCount: rec.SerendipityCount,
		Domain:           rec.Domain,
		Title:            rec.EventTitle,
		Description:      rec.EventDescription,
		EventTime:        rec.EventTime,
	}, nil
}

func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	}
	if duration < time.Hour {
		mins := int(duration.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	}
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		return fmt.Sprintf("%dh ago", hours)
	}
	days := int(duration.Hours() / 24)
	return fmt.Sprintf("%dd ago", days)
}
