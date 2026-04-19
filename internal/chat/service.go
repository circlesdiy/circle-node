package chat

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"circles.diy/internal/circle"
	"circles.diy/internal/content"
	"circles.diy/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Sentinel errors returned by the chat service. Handlers map these to HTTP
// status codes rather than inspecting string contents.
var (
	ErrChatNotFound      = errors.New("chat not found")
	ErrNotAuthorized     = errors.New("not authorized for chat")
	ErrBlocked           = errors.New("cannot start DM with blocked profile")
	ErrEmptyMessage      = errors.New("message content is empty")
	ErrInvalidParticipants = errors.New("invalid participant set")
	ErrCircleAccessDenied = errors.New("must be an active circle member")
	ErrCircleRequired     = errors.New("circle must be selected for this action")
)

// Repository captures the persistence surface the service actually needs.
// We extend the narrow domain.ChatRepository with the optional helpers used
// for keyset history and block checks.
type Repository interface {
	domain.ChatRepository

	GetMessagesBefore(ctx context.Context, chatID string, beforeTime time.Time, beforeID string, limit int) ([]domain.Message, error)
	GetMessagesAfter(ctx context.Context, chatID, afterID string, limit int) ([]domain.Message, error)
	LatestMessageByChat(ctx context.Context, chatID string) (*domain.Message, error)
	IsBlockedEitherWay(ctx context.Context, profileA, profileB string) (bool, error)
	MarkChatRead(ctx context.Context, chatID, profileID string, readAt time.Time) error
	GetInboxEntries(ctx context.Context, profileID string) ([]InboxRow, error)
}

// Service orchestrates chat domain logic: authorization, DM/group/circle
// creation, lazy participant sync for circle-scoped chats, and message
// persistence. It leaves fan-out to the HTTP layer so templates never
// appear in the service package.
type Service struct {
	repo    Repository
	circles *circle.Service
	logger  *zap.Logger
}

func NewService(repo Repository, circles *circle.Service, logger *zap.Logger) *Service {
	return &Service{
		repo:    repo,
		circles: circles,
		logger:  logger,
	}
}

// InboxEntry bundles a chat with its latest message and unread count for
// rendering the sidebar.
type InboxEntry struct {
	Chat        domain.Chat
	LastMessage *domain.Message
	UnreadCount int
}

// ListInbox returns all chats the profile participates in, most recently
// active first. Uses a single batch query to avoid N+1.
func (s *Service) ListInbox(ctx context.Context, profileID string) ([]InboxEntry, error) {
	rows, err := s.repo.GetInboxEntries(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("list inbox: %w", err)
	}

	entries := make([]InboxEntry, 0, len(rows))
	for _, r := range rows {
		entry := InboxEntry{
			Chat:        r.Chat,
			UnreadCount: r.UnreadCount,
		}
		if r.LastMessageID != nil {
			entry.LastMessage = &domain.Message{
				ID:              *r.LastMessageID,
				ChatID:          r.Chat.ID,
				SenderProfileID: stringVal(r.LastMessageSenderID),
				Content:         stringVal(r.LastMessageContent),
				MessageType:     stringVal(r.LastMessageType),
				CreatedAt:       timeVal(r.LastMessageCreatedAt),
			}
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func stringVal(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func timeVal(p *time.Time) time.Time {
	if p == nil {
		return time.Time{}
	}
	return *p
}

// MarkChatAsRead advances the read watermark for the viewer in the given chat.
func (s *Service) MarkChatAsRead(ctx context.Context, chatID, viewerProfileID string) error {
	if _, err := s.GetChat(ctx, chatID, viewerProfileID); err != nil {
		return err
	}
	return s.repo.MarkChatRead(ctx, chatID, viewerProfileID, time.Now().UTC())
}

// GetChat verifies the viewer can access the chat, running lazy participant
// sync for circle-scoped chats when the viewer is a member but not yet a
// recorded participant.
func (s *Service) GetChat(ctx context.Context, chatID, viewerProfileID string) (*domain.Chat, error) {
	chat, err := s.repo.GetChatByID(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("get chat: %w", err)
	}
	if chat == nil {
		return nil, ErrChatNotFound
	}

	if err := s.ensureAccess(ctx, chat, viewerProfileID); err != nil {
		return nil, err
	}
	return chat, nil
}

// LoadRecentMessages returns the most recent page of messages for a chat
// in chronological order (oldest first) so they can be appended directly.
func (s *Service) LoadRecentMessages(ctx context.Context, chatID, viewerProfileID string, limit int) ([]domain.Message, error) {
	if _, err := s.GetChat(ctx, chatID, viewerProfileID); err != nil {
		return nil, err
	}
	return s.loadMessages(ctx, chatID, time.Time{}, "", limit)
}

// LoadHistoryBefore returns a page of messages older than the supplied
// cursor, ordered chronologically oldest-first for prepend-style rendering.
func (s *Service) LoadHistoryBefore(ctx context.Context, chatID, viewerProfileID string, beforeTime time.Time, beforeID string, limit int) ([]domain.Message, error) {
	if _, err := s.GetChat(ctx, chatID, viewerProfileID); err != nil {
		return nil, err
	}
	return s.loadMessages(ctx, chatID, beforeTime, beforeID, limit)
}

func (s *Service) loadMessages(ctx context.Context, chatID string, beforeTime time.Time, beforeID string, limit int) ([]domain.Message, error) {
	messages, err := s.repo.GetMessagesBefore(ctx, chatID, beforeTime, beforeID, limit)
	if err != nil {
		return nil, fmt.Errorf("load history: %w", err)
	}
	// Reverse to chronological order for natural append rendering.
	sort.SliceStable(messages, func(i, j int) bool {
		if messages[i].CreatedAt.Equal(messages[j].CreatedAt) {
			return messages[i].ID < messages[j].ID
		}
		return messages[i].CreatedAt.Before(messages[j].CreatedAt)
	})
	return messages, nil
}

// LoadMessagesSince replays messages created after afterID so a reconnecting
// SSE client can fill in anything it missed while offline.
func (s *Service) LoadMessagesSince(ctx context.Context, chatID, viewerProfileID, afterID string, limit int) ([]domain.Message, error) {
	if _, err := s.GetChat(ctx, chatID, viewerProfileID); err != nil {
		return nil, err
	}
	return s.repo.GetMessagesAfter(ctx, chatID, afterID, limit)
}

// FindOrCreateDM returns the existing 1:1 chat between two profiles or
// creates a new one. Rejects with ErrBlocked if either side blocked the other.
func (s *Service) FindOrCreateDM(ctx context.Context, creatorProfileID, otherProfileID string) (*domain.Chat, error) {
	if creatorProfileID == "" || otherProfileID == "" || creatorProfileID == otherProfileID {
		return nil, ErrInvalidParticipants
	}

	blocked, err := s.repo.IsBlockedEitherWay(ctx, creatorProfileID, otherProfileID)
	if err != nil {
		return nil, fmt.Errorf("block check: %w", err)
	}
	if blocked {
		return nil, ErrBlocked
	}

	existing, err := s.repo.GetDirectChatByProfiles(ctx, creatorProfileID, otherProfileID)
	if err != nil {
		return nil, fmt.Errorf("lookup dm: %w", err)
	}
	if existing != nil {
		return existing, nil
	}

	now := time.Now().UTC()
	chat := &domain.Chat{
		ID:        uuid.New().String(),
		Type:      domain.ChatTypeDirect,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.CreateChat(ctx, chat); err != nil {
		return nil, fmt.Errorf("create dm: %w", err)
	}

	for _, pid := range []string{creatorProfileID, otherProfileID} {
		if err := s.repo.AddParticipant(ctx, &domain.ChatParticipant{
			ID:        uuid.New().String(),
			ChatID:    chat.ID,
			ProfileID: pid,
			IsAdmin:   pid == creatorProfileID,
			JoinedAt:  now,
		}); err != nil {
			return nil, fmt.Errorf("add dm participant: %w", err)
		}
	}

	s.logger.Info("dm created",
		zap.String("chat_id", chat.ID),
		zap.String("creator", creatorProfileID),
		zap.String("other", otherProfileID),
	)
	return chat, nil
}

// FindOrCreateDMInCircle verifies both profiles are active members of
// circleID and then delegates to FindOrCreateDM. The circle serves as a
// gating policy: the DM itself is not stored as circle-scoped so that a
// conversation survives either party leaving the circle later.
func (s *Service) FindOrCreateDMInCircle(ctx context.Context, creatorProfileID, otherProfileID, circleID string) (*domain.Chat, error) {
	if circleID == "" {
		return nil, ErrCircleRequired
	}
	if creatorProfileID == "" || otherProfileID == "" || creatorProfileID == otherProfileID {
		return nil, ErrInvalidParticipants
	}
	for _, pid := range []string{creatorProfileID, otherProfileID} {
		isMember, err := s.circles.IsMember(ctx, circleID, pid)
		if err != nil {
			return nil, fmt.Errorf("circle membership check: %w", err)
		}
		if !isMember {
			return nil, ErrCircleAccessDenied
		}
	}
	return s.FindOrCreateDM(ctx, creatorProfileID, otherProfileID)
}

// CreateGroupChatInCircle asserts every participant is an active member of
// circleID before delegating to CreateGroupChat. The circle is used as a
// gating policy only; the group chat is NOT stored as circle-scoped so it
// persists independently of future membership changes in the circle.
func (s *Service) CreateGroupChatInCircle(ctx context.Context, creatorProfileID, name string, participantProfileIDs []string, circleID string) (*domain.Chat, error) {
	if circleID == "" {
		return nil, ErrCircleRequired
	}
	if creatorProfileID == "" {
		return nil, ErrInvalidParticipants
	}
	seen := map[string]struct{}{}
	for _, pid := range append([]string{creatorProfileID}, participantProfileIDs...) {
		if pid == "" {
			continue
		}
		if _, dup := seen[pid]; dup {
			continue
		}
		seen[pid] = struct{}{}
		isMember, err := s.circles.IsMember(ctx, circleID, pid)
		if err != nil {
			return nil, fmt.Errorf("circle membership check: %w", err)
		}
		if !isMember {
			return nil, ErrCircleAccessDenied
		}
	}
	return s.CreateGroupChat(ctx, creatorProfileID, name, participantProfileIDs)
}

// CreateGroupChat creates a new group chat with the creator plus supplied participants.
func (s *Service) CreateGroupChat(ctx context.Context, creatorProfileID, name string, participantProfileIDs []string) (*domain.Chat, error) {
	if creatorProfileID == "" {
		return nil, ErrInvalidParticipants
	}

	uniqueIDs := map[string]struct{}{creatorProfileID: {}}
	for _, pid := range participantProfileIDs {
		if pid == "" {
			continue
		}
		uniqueIDs[pid] = struct{}{}
	}
	if len(uniqueIDs) < 2 {
		return nil, ErrInvalidParticipants
	}

	trimmedName := content.SanitizePlaintext(name)
	if trimmedName == "" {
		trimmedName = "New group"
	}

	now := time.Now().UTC()
	chat := &domain.Chat{
		ID:        uuid.New().String(),
		Type:      domain.ChatTypeGroup,
		Name:      trimmedName,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.CreateChat(ctx, chat); err != nil {
		return nil, fmt.Errorf("create group chat: %w", err)
	}

	for pid := range uniqueIDs {
		if err := s.repo.AddParticipant(ctx, &domain.ChatParticipant{
			ID:        uuid.New().String(),
			ChatID:    chat.ID,
			ProfileID: pid,
			IsAdmin:   pid == creatorProfileID,
			JoinedAt:  now,
		}); err != nil {
			return nil, fmt.Errorf("add group participant: %w", err)
		}
	}

	s.logger.Info("group chat created",
		zap.String("chat_id", chat.ID),
		zap.String("creator", creatorProfileID),
		zap.Int("participants", len(uniqueIDs)),
	)
	return chat, nil
}

// CreateCircleChat creates a circle-scoped chat. Only active members of the
// circle may create one. Other members are lazily added as participants the
// first time they open the chat.
func (s *Service) CreateCircleChat(ctx context.Context, creatorProfileID, circleID, name string) (*domain.Chat, error) {
	if circleID == "" {
		return nil, ErrInvalidParticipants
	}

	isMember, err := s.circles.IsMember(ctx, circleID, creatorProfileID)
	if err != nil {
		return nil, fmt.Errorf("circle membership check: %w", err)
	}
	if !isMember {
		return nil, ErrCircleAccessDenied
	}

	trimmedName := content.SanitizePlaintext(name)
	if trimmedName == "" {
		trimmedName = "General"
	}

	now := time.Now().UTC()
	circleIDCopy := circleID
	chat := &domain.Chat{
		ID:        uuid.New().String(),
		Type:      domain.ChatTypeGroup,
		Name:      trimmedName,
		CircleID:  &circleIDCopy,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.CreateChat(ctx, chat); err != nil {
		return nil, fmt.Errorf("create circle chat: %w", err)
	}

	if err := s.repo.AddParticipant(ctx, &domain.ChatParticipant{
		ID:        uuid.New().String(),
		ChatID:    chat.ID,
		ProfileID: creatorProfileID,
		IsAdmin:   true,
		JoinedAt:  now,
	}); err != nil {
		return nil, fmt.Errorf("add circle chat creator: %w", err)
	}

	s.logger.Info("circle chat created",
		zap.String("chat_id", chat.ID),
		zap.String("circle_id", circleID),
		zap.String("creator", creatorProfileID),
	)
	return chat, nil
}

// GetOrCreateDefaultCircleChat returns the most recently active circle-scoped
// chat for the circle, creating a "General" chat if none exist. Membership
// is verified before any write.
func (s *Service) GetOrCreateDefaultCircleChat(ctx context.Context, circleID, viewerProfileID string) (*domain.Chat, error) {
	isMember, err := s.circles.IsMember(ctx, circleID, viewerProfileID)
	if err != nil {
		return nil, fmt.Errorf("circle membership check: %w", err)
	}
	if !isMember {
		return nil, ErrCircleAccessDenied
	}

	chats, err := s.repo.GetChatsByCircleID(ctx, circleID)
	if err != nil {
		return nil, fmt.Errorf("list circle chats: %w", err)
	}
	if len(chats) > 0 {
		chat := chats[0]
		if err := s.ensureCircleParticipant(ctx, &chat, viewerProfileID); err != nil {
			return nil, err
		}
		return &chat, nil
	}

	return s.CreateCircleChat(ctx, viewerProfileID, circleID, "General")
}

// SendTextMessage validates access, sanitizes the body, persists the message
// and returns both the created message and the set of profile IDs that should
// receive a live push event. The caller is responsible for rendering the HTML
// fragment and invoking Hub.PublishAll for the recipients.
func (s *Service) SendTextMessage(ctx context.Context, chatID, senderProfileID, body string) (*domain.Message, []string, error) {
	chat, err := s.repo.GetChatByID(ctx, chatID)
	if err != nil {
		return nil, nil, fmt.Errorf("get chat: %w", err)
	}
	if chat == nil {
		return nil, nil, ErrChatNotFound
	}
	if err := s.ensureAccess(ctx, chat, senderProfileID); err != nil {
		return nil, nil, err
	}

	sanitized := content.SanitizePlaintext(body)
	if content.IsEmptyContent(sanitized) {
		return nil, nil, ErrEmptyMessage
	}

	if chat.IsDirect() {
		participants, err := s.repo.GetParticipantsByChatID(ctx, chat.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("load dm participants: %w", err)
		}
		for _, p := range participants {
			if p.ProfileID == senderProfileID {
				continue
			}
			blocked, err := s.repo.IsBlockedEitherWay(ctx, senderProfileID, p.ProfileID)
			if err != nil {
				return nil, nil, fmt.Errorf("block check: %w", err)
			}
			if blocked {
				return nil, nil, ErrBlocked
			}
		}
	}

	now := time.Now().UTC()
	message := &domain.Message{
		ID:              uuid.New().String(),
		ChatID:          chat.ID,
		SenderProfileID: senderProfileID,
		Content:         sanitized,
		MessageType:     domain.MessageTypeText,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repo.CreateMessage(ctx, message); err != nil {
		return nil, nil, fmt.Errorf("persist message: %w", err)
	}

	participants, err := s.repo.GetParticipantsByChatID(ctx, chat.ID)
	if err != nil {
		s.logger.Warn("failed to load participants for fan-out",
			zap.String("chat_id", chat.ID),
			zap.Error(err),
		)
		return message, nil, nil
	}
	recipients := make([]string, 0, len(participants))
	for _, p := range participants {
		if p.ProfileID == senderProfileID {
			continue
		}
		recipients = append(recipients, p.ProfileID)
	}
	return message, recipients, nil
}

// ensureAccess centralizes the DM/group/circle authorization rules.
func (s *Service) ensureAccess(ctx context.Context, chat *domain.Chat, viewerProfileID string) error {
	if viewerProfileID == "" {
		return ErrNotAuthorized
	}

	if chat.IsCircleScoped() {
		isMember, err := s.circles.IsMember(ctx, *chat.CircleID, viewerProfileID)
		if err != nil {
			return fmt.Errorf("circle membership check: %w", err)
		}
		if !isMember {
			return ErrCircleAccessDenied
		}
		return s.ensureCircleParticipant(ctx, chat, viewerProfileID)
	}

	isParticipant, err := s.repo.IsParticipant(ctx, chat.ID, viewerProfileID)
	if err != nil {
		return fmt.Errorf("participant check: %w", err)
	}
	if !isParticipant {
		return ErrNotAuthorized
	}
	return nil
}

// ensureCircleParticipant lazily inserts the viewer into chat_participants
// for circle-scoped chats, idempotent across concurrent opens.
func (s *Service) ensureCircleParticipant(ctx context.Context, chat *domain.Chat, profileID string) error {
	already, err := s.repo.IsParticipant(ctx, chat.ID, profileID)
	if err != nil {
		return fmt.Errorf("participant check: %w", err)
	}
	if already {
		return nil
	}
	return s.repo.AddParticipant(ctx, &domain.ChatParticipant{
		ID:        uuid.New().String(),
		ChatID:    chat.ID,
		ProfileID: profileID,
		IsAdmin:   false,
		JoinedAt:  time.Now().UTC(),
	})
}

// Participants returns the full participant set for a chat. Useful for the
// handler to render chat headers / avatar collages.
func (s *Service) Participants(ctx context.Context, chatID string) ([]domain.ChatParticipant, error) {
	return s.repo.GetParticipantsByChatID(ctx, chatID)
}
