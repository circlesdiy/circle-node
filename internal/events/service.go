package events

import (
	"context"
	"fmt"
	"strings"
	"time"

	"circles.diy/internal/circle"
	"circles.diy/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Service struct {
	repo          Repository
	circleService *circle.Service
	logger        *zap.Logger
}

func NewService(repo Repository, circleService *circle.Service, logger *zap.Logger) *Service {
	return &Service{
		repo:          repo,
		circleService: circleService,
		logger:        logger,
	}
}

// CreateEvent creates a new event
func (s *Service) CreateEvent(
	ctx context.Context,
	title, description, location, timezone string,
	startTime, endTime time.Time,
	circleID, organizerProfileID string,
	capacity int,
) (*domain.Event, error) {
	// Validate inputs
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("event title is required")
	}
	if len(title) > 500 {
		return nil, fmt.Errorf("event title cannot exceed 500 characters")
	}

	description = strings.TrimSpace(description)

	if circleID == "" {
		return nil, fmt.Errorf("circle is required")
	}

	if startTime.IsZero() {
		return nil, fmt.Errorf("start time is required")
	}

	// Validate start time is in the future
	if startTime.Before(time.Now().UTC()) {
		return nil, fmt.Errorf("start time must be in the future")
	}

	// Validate end time if provided
	if !endTime.IsZero() && endTime.Before(startTime) {
		return nil, fmt.Errorf("end time must be after start time")
	}

	// Verify organizer is a member of the circle
	isMember, err := s.circleService.IsMember(ctx, circleID, organizerProfileID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify circle membership: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("you must be a member of the circle to create events")
	}

	// Create event
	event := &domain.Event{
		ID:                 uuid.New().String(),
		CircleID:           circleID,
		OrganizerProfileID: organizerProfileID,
		Title:              title,
		Description:        description,
		Location:           strings.TrimSpace(location),
		Timezone:           timezone,
		StartTime:          startTime,
		EndTime:            endTime,
		Capacity:           capacity,
		CreatedAt:          time.Now().UTC(),
		UpdatedAt:          time.Now().UTC(),
	}

	if err := s.repo.CreateEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	s.logger.Info("event created",
		zap.String("event_id", event.ID),
		zap.String("circle_id", circleID),
		zap.String("organizer_profile_id", organizerProfileID),
	)

	return event, nil
}

// GetEventDetails retrieves detailed event information
func (s *Service) GetEventDetails(ctx context.Context, eventID, viewerProfileID string) (*EventDetails, error) {
	details, err := s.repo.GetEventDetails(ctx, eventID, viewerProfileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event details: %w", err)
	}

	if details == nil {
		return nil, fmt.Errorf("event not found")
	}

	return details, nil
}

// UpdateEvent updates an event
func (s *Service) UpdateEvent(
	ctx context.Context,
	eventID string,
	title, description, location, timezone string,
	startTime, endTime time.Time,
	capacity int,
	updaterProfileID string,
) error {
	// Get existing event
	event, err := s.repo.GetEventByID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("failed to get event: %w", err)
	}
	if event == nil {
		return fmt.Errorf("event not found")
	}

	// Check if already deleted
	if event.IsDeleted() {
		return fmt.Errorf("cannot update deleted event")
	}

	// Verify updater is the organizer
	if event.OrganizerProfileID != updaterProfileID {
		return fmt.Errorf("only the organizer can update the event")
	}

	// Validate inputs
	title = strings.TrimSpace(title)
	if title == "" {
		return fmt.Errorf("event title is required")
	}
	if len(title) > 500 {
		return fmt.Errorf("event title cannot exceed 500 characters")
	}

	description = strings.TrimSpace(description)

	if startTime.IsZero() {
		return fmt.Errorf("start time is required")
	}

	if !endTime.IsZero() && endTime.Before(startTime) {
		return fmt.Errorf("end time must be after start time")
	}

	// Update event fields
	event.Title = title
	event.Description = description
	event.Location = strings.TrimSpace(location)
	event.Timezone = timezone
	event.StartTime = startTime
	event.EndTime = endTime
	event.Capacity = capacity
	event.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	s.logger.Info("event updated",
		zap.String("event_id", eventID),
		zap.String("updater_profile_id", updaterProfileID),
	)

	return nil
}

// DeleteEvent soft deletes an event
func (s *Service) DeleteEvent(ctx context.Context, eventID, deleterProfileID string) error {
	// Get existing event
	event, err := s.repo.GetEventByID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("failed to get event: %w", err)
	}
	if event == nil {
		return fmt.Errorf("event not found")
	}

	// Check if already deleted
	if event.IsDeleted() {
		return fmt.Errorf("event already deleted")
	}

	// Verify deleter is the organizer
	if event.OrganizerProfileID != deleterProfileID {
		return fmt.Errorf("only the organizer can delete the event")
	}

	if err := s.repo.DeleteEvent(ctx, eventID); err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	s.logger.Info("event deleted",
		zap.String("event_id", eventID),
		zap.String("deleter_profile_id", deleterProfileID),
	)

	return nil
}

// CreateCoordinationNeed adds a coordination need for an event
func (s *Service) CreateCoordinationNeed(
	ctx context.Context,
	eventID, needType, message, authorProfileID string,
) (*CoordinationNeed, error) {
	// Validate inputs
	message = strings.TrimSpace(message)
	if message == "" {
		return nil, fmt.Errorf("coordination message is required")
	}

	// Validate need type
	validTypes := map[string]bool{
		"transport": true,
		"tickets":   true,
		"supplies":  true,
	}
	if !validTypes[needType] {
		return nil, fmt.Errorf("invalid coordination type: %s", needType)
	}

	// Verify event exists
	event, err := s.repo.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	if event == nil {
		return nil, fmt.Errorf("event not found")
	}

	need := &CoordinationNeed{
		EventID:         eventID,
		Type:            needType,
		Message:         message,
		AuthorProfileID: authorProfileID,
	}

	if err := s.repo.CreateCoordinationNeed(ctx, need); err != nil {
		return nil, fmt.Errorf("failed to create coordination need: %w", err)
	}

	s.logger.Info("coordination need created",
		zap.String("event_id", eventID),
		zap.String("type", needType),
		zap.String("author_profile_id", authorProfileID),
	)

	return need, nil
}

// ResolveCoordinationNeed marks a coordination need as resolved
func (s *Service) ResolveCoordinationNeed(ctx context.Context, needID, resolverProfileID string) error {
	// TODO: Add permission check - only event organizer or need author should resolve

	if err := s.repo.ResolveCoordinationNeed(ctx, needID); err != nil {
		return fmt.Errorf("failed to resolve coordination need: %w", err)
	}

	s.logger.Info("coordination need resolved",
		zap.String("need_id", needID),
		zap.String("resolver_profile_id", resolverProfileID),
	)

	return nil
}
