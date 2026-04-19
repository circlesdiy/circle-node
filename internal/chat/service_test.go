package chat

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"circles.diy/internal/domain"
	"go.uber.org/zap"
)

// fakeRepo is an in-memory implementation of the chat Repository used for
// unit testing the service without a database.
type fakeRepo struct {
	chats        map[string]*domain.Chat
	participants map[string][]domain.ChatParticipant
	messages     map[string][]domain.Message
	reads        []domain.MessageRead
	blocks       map[string]bool
	lastReadAt   map[string]map[string]time.Time // chatID -> profileID -> time
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		chats:        make(map[string]*domain.Chat),
		participants: make(map[string][]domain.ChatParticipant),
		messages:     make(map[string][]domain.Message),
		blocks:       make(map[string]bool),
		lastReadAt:   make(map[string]map[string]time.Time),
	}
}

func blockKey(a, b string) string {
	if a < b {
		return a + "|" + b
	}
	return b + "|" + a
}

func (r *fakeRepo) CreateChat(_ context.Context, chat *domain.Chat) error {
	if _, exists := r.chats[chat.ID]; exists {
		return errors.New("chat exists")
	}
	clone := *chat
	r.chats[chat.ID] = &clone
	return nil
}
func (r *fakeRepo) GetChatByID(_ context.Context, id string) (*domain.Chat, error) {
	if c, ok := r.chats[id]; ok {
		clone := *c
		return &clone, nil
	}
	return nil, nil
}
func (r *fakeRepo) GetChatsByProfileID(_ context.Context, profileID string) ([]domain.Chat, error) {
	var out []domain.Chat
	for chatID, parts := range r.participants {
		for _, p := range parts {
			if p.ProfileID == profileID {
				out = append(out, *r.chats[chatID])
				break
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}
func (r *fakeRepo) GetChatsByCircleID(_ context.Context, circleID string) ([]domain.Chat, error) {
	var out []domain.Chat
	for _, c := range r.chats {
		if c.CircleID != nil && *c.CircleID == circleID {
			out = append(out, *c)
		}
	}
	return out, nil
}
func (r *fakeRepo) GetDirectChatByProfiles(_ context.Context, a, b string) (*domain.Chat, error) {
	for chatID, parts := range r.participants {
		c := r.chats[chatID]
		if c == nil || c.Type != domain.ChatTypeDirect {
			continue
		}
		set := map[string]struct{}{}
		for _, p := range parts {
			set[p.ProfileID] = struct{}{}
		}
		if _, ok1 := set[a]; ok1 {
			if _, ok2 := set[b]; ok2 {
				clone := *c
				return &clone, nil
			}
		}
	}
	return nil, nil
}
func (r *fakeRepo) UpdateChat(_ context.Context, chat *domain.Chat) error {
	clone := *chat
	r.chats[chat.ID] = &clone
	return nil
}
func (r *fakeRepo) DeleteChat(_ context.Context, id string) error {
	delete(r.chats, id)
	delete(r.participants, id)
	delete(r.messages, id)
	return nil
}
func (r *fakeRepo) AddParticipant(_ context.Context, p *domain.ChatParticipant) error {
	r.participants[p.ChatID] = append(r.participants[p.ChatID], *p)
	return nil
}
func (r *fakeRepo) RemoveParticipant(_ context.Context, chatID, profileID string) error {
	parts := r.participants[chatID]
	for i, p := range parts {
		if p.ProfileID == profileID {
			r.participants[chatID] = append(parts[:i], parts[i+1:]...)
			return nil
		}
	}
	return nil
}
func (r *fakeRepo) GetParticipantsByChatID(_ context.Context, chatID string) ([]domain.ChatParticipant, error) {
	return append([]domain.ChatParticipant(nil), r.participants[chatID]...), nil
}
func (r *fakeRepo) IsParticipant(_ context.Context, chatID, profileID string) (bool, error) {
	for _, p := range r.participants[chatID] {
		if p.ProfileID == profileID {
			return true, nil
		}
	}
	return false, nil
}
func (r *fakeRepo) CreateMessage(_ context.Context, m *domain.Message) error {
	clone := *m
	r.messages[m.ChatID] = append(r.messages[m.ChatID], clone)
	if c, ok := r.chats[m.ChatID]; ok {
		t := m.CreatedAt
		c.LastMessageAt = &t
		c.UpdatedAt = t
	}
	return nil
}
func (r *fakeRepo) GetMessageByID(_ context.Context, id string) (*domain.Message, error) {
	for _, msgs := range r.messages {
		for _, m := range msgs {
			if m.ID == id {
				clone := m
				return &clone, nil
			}
		}
	}
	return nil, nil
}
func (r *fakeRepo) GetMessagesByChatID(_ context.Context, chatID string, limit, offset int) ([]domain.Message, error) {
	msgs := r.messages[chatID]
	if offset >= len(msgs) {
		return nil, nil
	}
	end := offset + limit
	if end > len(msgs) {
		end = len(msgs)
	}
	return append([]domain.Message(nil), msgs[offset:end]...), nil
}
func (r *fakeRepo) UpdateMessage(_ context.Context, m *domain.Message) error {
	for i, existing := range r.messages[m.ChatID] {
		if existing.ID == m.ID {
			r.messages[m.ChatID][i] = *m
			return nil
		}
	}
	return errors.New("not found")
}
func (r *fakeRepo) DeleteMessage(_ context.Context, id string) error {
	for chatID, msgs := range r.messages {
		for i, m := range msgs {
			if m.ID == id {
				r.messages[chatID] = append(msgs[:i], msgs[i+1:]...)
				return nil
			}
		}
	}
	return nil
}
func (r *fakeRepo) MarkMessageAsRead(_ context.Context, mr *domain.MessageRead) error {
	r.reads = append(r.reads, *mr)
	return nil
}
func (r *fakeRepo) GetReadReceiptsByMessageID(_ context.Context, messageID string) ([]domain.MessageRead, error) {
	var out []domain.MessageRead
	for _, mr := range r.reads {
		if mr.MessageID == messageID {
			out = append(out, mr)
		}
	}
	return out, nil
}
func (r *fakeRepo) GetUnreadMessageCount(_ context.Context, chatID, profileID string) (int, error) {
	var watermark time.Time
	if byProfile, ok := r.lastReadAt[chatID]; ok {
		watermark = byProfile[profileID]
	}
	count := 0
	for _, m := range r.messages[chatID] {
		if m.SenderProfileID == profileID {
			continue
		}
		if m.CreatedAt.After(watermark) {
			count++
		}
	}
	return count, nil
}

func (r *fakeRepo) MarkChatRead(_ context.Context, chatID, profileID string, readAt time.Time) error {
	if r.lastReadAt[chatID] == nil {
		r.lastReadAt[chatID] = make(map[string]time.Time)
	}
	if existing, ok := r.lastReadAt[chatID][profileID]; ok && !readAt.After(existing) {
		return nil // monotonic: never go backward
	}
	r.lastReadAt[chatID][profileID] = readAt
	return nil
}

func (r *fakeRepo) GetInboxEntries(_ context.Context, profileID string) ([]InboxRow, error) {
	var rows []InboxRow
	for chatID, parts := range r.participants {
		for _, p := range parts {
			if p.ProfileID != profileID {
				continue
			}
			c := r.chats[chatID]
			if c == nil {
				break
			}
			row := InboxRow{Chat: *c}

			// Latest message
			msgs := r.messages[chatID]
			if len(msgs) > 0 {
				latest := msgs[0]
				for _, m := range msgs[1:] {
					if m.CreatedAt.After(latest.CreatedAt) {
						latest = m
					}
				}
				row.LastMessageID = &latest.ID
				row.LastMessageSenderID = &latest.SenderProfileID
				row.LastMessageContent = &latest.Content
				row.LastMessageType = &latest.MessageType
				t := latest.CreatedAt
				row.LastMessageCreatedAt = &t
			}

			// Unread count via watermark
			var watermark time.Time
			if byProfile, ok := r.lastReadAt[chatID]; ok {
				watermark = byProfile[profileID]
			}
			for _, m := range msgs {
				if m.SenderProfileID != profileID && m.CreatedAt.After(watermark) {
					row.UnreadCount++
				}
			}

			rows = append(rows, row)
			break
		}
	}
	// Sort by last_message_at DESC, created_at DESC
	sort.Slice(rows, func(i, j int) bool {
		a, b := rows[i].Chat, rows[j].Chat
		if a.LastMessageAt != nil && b.LastMessageAt != nil {
			if !a.LastMessageAt.Equal(*b.LastMessageAt) {
				return a.LastMessageAt.After(*b.LastMessageAt)
			}
		}
		if a.LastMessageAt != nil && b.LastMessageAt == nil {
			return true
		}
		if a.LastMessageAt == nil && b.LastMessageAt != nil {
			return false
		}
		return a.CreatedAt.After(b.CreatedAt)
	})
	return rows, nil
}

func (r *fakeRepo) GetMessagesBefore(_ context.Context, chatID string, beforeTime time.Time, beforeID string, limit int) ([]domain.Message, error) {
	msgs := append([]domain.Message(nil), r.messages[chatID]...)
	sort.Slice(msgs, func(i, j int) bool {
		if msgs[i].CreatedAt.Equal(msgs[j].CreatedAt) {
			return msgs[i].ID > msgs[j].ID
		}
		return msgs[i].CreatedAt.After(msgs[j].CreatedAt)
	})
	var filtered []domain.Message
	for _, m := range msgs {
		if !beforeTime.IsZero() {
			if m.CreatedAt.After(beforeTime) {
				continue
			}
			if m.CreatedAt.Equal(beforeTime) && m.ID >= beforeID {
				continue
			}
		}
		filtered = append(filtered, m)
		if len(filtered) >= limit {
			break
		}
	}
	return filtered, nil
}
func (r *fakeRepo) GetMessagesAfter(_ context.Context, chatID, afterID string, limit int) ([]domain.Message, error) {
	msgs := append([]domain.Message(nil), r.messages[chatID]...)
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].CreatedAt.Before(msgs[j].CreatedAt)
	})
	if afterID == "" {
		if len(msgs) > limit {
			msgs = msgs[:limit]
		}
		return msgs, nil
	}
	var out []domain.Message
	seen := false
	for _, m := range msgs {
		if !seen {
			if m.ID == afterID {
				seen = true
			}
			continue
		}
		out = append(out, m)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}
func (r *fakeRepo) LatestMessageByChat(_ context.Context, chatID string) (*domain.Message, error) {
	msgs := r.messages[chatID]
	if len(msgs) == 0 {
		return nil, nil
	}
	latest := msgs[0]
	for _, m := range msgs[1:] {
		if m.CreatedAt.After(latest.CreatedAt) {
			latest = m
		}
	}
	return &latest, nil
}
func (r *fakeRepo) IsBlockedEitherWay(_ context.Context, a, b string) (bool, error) {
	return r.blocks[blockKey(a, b)], nil
}

func newTestService(repo Repository) *Service {
	return &Service{repo: repo, logger: zap.NewNop()}
}

func TestFindOrCreateDMCreatesNewChat(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)

	chat, err := svc.FindOrCreateDM(context.Background(), "alice", "bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chat == nil || chat.Type != domain.ChatTypeDirect {
		t.Fatalf("expected DM chat, got %+v", chat)
	}
	parts, _ := repo.GetParticipantsByChatID(context.Background(), chat.ID)
	if len(parts) != 2 {
		t.Fatalf("expected 2 DM participants, got %d", len(parts))
	}
}

func TestFindOrCreateDMIsIdempotent(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)

	first, err := svc.FindOrCreateDM(context.Background(), "alice", "bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := svc.FindOrCreateDM(context.Background(), "bob", "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("expected same DM to be reused, got %s and %s", first.ID, second.ID)
	}
}

func TestFindOrCreateDMRejectsBlocked(t *testing.T) {
	repo := newFakeRepo()
	repo.blocks[blockKey("alice", "bob")] = true
	svc := newTestService(repo)

	_, err := svc.FindOrCreateDM(context.Background(), "alice", "bob")
	if !errors.Is(err, ErrBlocked) {
		t.Fatalf("expected ErrBlocked, got %v", err)
	}
}

func TestFindOrCreateDMRejectsInvalidArgs(t *testing.T) {
	svc := newTestService(newFakeRepo())
	_, err := svc.FindOrCreateDM(context.Background(), "alice", "alice")
	if !errors.Is(err, ErrInvalidParticipants) {
		t.Fatalf("expected ErrInvalidParticipants, got %v", err)
	}
	_, err = svc.FindOrCreateDM(context.Background(), "alice", "")
	if !errors.Is(err, ErrInvalidParticipants) {
		t.Fatalf("expected ErrInvalidParticipants for empty id, got %v", err)
	}
}

func TestCreateGroupChatRequiresAtLeastTwoUniqueParticipants(t *testing.T) {
	svc := newTestService(newFakeRepo())
	_, err := svc.CreateGroupChat(context.Background(), "alice", "name", []string{"alice"})
	if !errors.Is(err, ErrInvalidParticipants) {
		t.Fatalf("expected ErrInvalidParticipants, got %v", err)
	}
}

func TestCreateGroupChatDefaultsName(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)

	chat, err := svc.CreateGroupChat(context.Background(), "alice", "   ", []string{"bob", "carol"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chat.Name == "" {
		t.Fatal("expected default group name to be applied")
	}
	parts, _ := repo.GetParticipantsByChatID(context.Background(), chat.ID)
	if len(parts) != 3 {
		t.Fatalf("expected 3 participants, got %d", len(parts))
	}
}

func TestSendTextMessagePersistsAndComputesRecipients(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)

	chat, err := svc.FindOrCreateDM(context.Background(), "alice", "bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg, recipients, err := svc.SendTextMessage(context.Background(), chat.ID, "alice", "hello bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Content != "hello bob" {
		t.Fatalf("unexpected content %q", msg.Content)
	}
	if len(recipients) != 1 || recipients[0] != "bob" {
		t.Fatalf("expected recipient list [bob], got %v", recipients)
	}
	if len(repo.messages[chat.ID]) != 1 {
		t.Fatalf("expected message to be persisted")
	}
}

func TestSendTextMessageRejectsEmpty(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)

	chat, _ := svc.FindOrCreateDM(context.Background(), "alice", "bob")
	_, _, err := svc.SendTextMessage(context.Background(), chat.ID, "alice", "   ")
	if !errors.Is(err, ErrEmptyMessage) {
		t.Fatalf("expected ErrEmptyMessage, got %v", err)
	}
}

func TestSendTextMessageRejectsNonParticipant(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)

	chat, _ := svc.FindOrCreateDM(context.Background(), "alice", "bob")
	_, _, err := svc.SendTextMessage(context.Background(), chat.ID, "eve", "intrusion")
	if !errors.Is(err, ErrNotAuthorized) {
		t.Fatalf("expected ErrNotAuthorized, got %v", err)
	}
}

func TestSendTextMessageRejectsBlockedDM(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)

	chat, _ := svc.FindOrCreateDM(context.Background(), "alice", "bob")
	repo.blocks[blockKey("alice", "bob")] = true

	_, _, err := svc.SendTextMessage(context.Background(), chat.ID, "alice", "hello")
	if !errors.Is(err, ErrBlocked) {
		t.Fatalf("expected ErrBlocked, got %v", err)
	}
}

func TestSendTextMessageChatNotFound(t *testing.T) {
	svc := newTestService(newFakeRepo())
	_, _, err := svc.SendTextMessage(context.Background(), "missing", "alice", "hi")
	if !errors.Is(err, ErrChatNotFound) {
		t.Fatalf("expected ErrChatNotFound, got %v", err)
	}
}

func TestLoadRecentMessagesReturnsChronologicalOrder(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)

	chat, _ := svc.FindOrCreateDM(context.Background(), "alice", "bob")
	_, _, _ = svc.SendTextMessage(context.Background(), chat.ID, "alice", "first")
	time.Sleep(time.Millisecond)
	_, _, _ = svc.SendTextMessage(context.Background(), chat.ID, "bob", "second")
	time.Sleep(time.Millisecond)
	_, _, _ = svc.SendTextMessage(context.Background(), chat.ID, "alice", "third")

	msgs, err := svc.LoadRecentMessages(context.Background(), chat.ID, "alice", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}
	if msgs[0].Content != "first" || msgs[2].Content != "third" {
		t.Fatalf("messages not chronological: %+v", msgs)
	}
}

func TestLoadMessagesSinceReturnsOnlyNewer(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)

	chat, _ := svc.FindOrCreateDM(context.Background(), "alice", "bob")
	first, _, _ := svc.SendTextMessage(context.Background(), chat.ID, "alice", "first")
	time.Sleep(time.Millisecond)
	_, _, _ = svc.SendTextMessage(context.Background(), chat.ID, "bob", "second")
	time.Sleep(time.Millisecond)
	_, _, _ = svc.SendTextMessage(context.Background(), chat.ID, "alice", "third")

	msgs, err := svc.LoadMessagesSince(context.Background(), chat.ID, "alice", first.ID, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages since first, got %d", len(msgs))
	}
}

func TestListInboxReturnsChatsForViewer(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)

	chat, _ := svc.FindOrCreateDM(context.Background(), "alice", "bob")
	_, _, _ = svc.SendTextMessage(context.Background(), chat.ID, "bob", "hello alice")

	entries, err := svc.ListInbox(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 inbox entry, got %d", len(entries))
	}
	if entries[0].LastMessage == nil || entries[0].LastMessage.Content != "hello alice" {
		t.Fatalf("unexpected last message: %+v", entries[0].LastMessage)
	}
	if entries[0].UnreadCount != 1 {
		t.Fatalf("expected 1 unread, got %d", entries[0].UnreadCount)
	}
}

func TestMarkChatAsReadResetsUnreadCount(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)

	chat, _ := svc.FindOrCreateDM(context.Background(), "alice", "bob")
	_, _, _ = svc.SendTextMessage(context.Background(), chat.ID, "bob", "msg1")
	time.Sleep(time.Millisecond)
	_, _, _ = svc.SendTextMessage(context.Background(), chat.ID, "bob", "msg2")

	// Before marking read, alice should have 2 unread.
	entries, _ := svc.ListInbox(context.Background(), "alice")
	if len(entries) != 1 || entries[0].UnreadCount != 2 {
		t.Fatalf("expected 2 unread before mark, got %d", entries[0].UnreadCount)
	}

	// Mark read.
	if err := svc.MarkChatAsRead(context.Background(), chat.ID, "alice"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After marking read, alice should have 0 unread.
	entries, _ = svc.ListInbox(context.Background(), "alice")
	if len(entries) != 1 || entries[0].UnreadCount != 0 {
		t.Fatalf("expected 0 unread after mark, got %d", entries[0].UnreadCount)
	}
}

func TestMarkChatAsReadIsMonotonic(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)

	chat, _ := svc.FindOrCreateDM(context.Background(), "alice", "bob")
	_, _, _ = svc.SendTextMessage(context.Background(), chat.ID, "bob", "msg1")
	time.Sleep(time.Millisecond)

	// Mark read now.
	if err := svc.MarkChatAsRead(context.Background(), chat.ID, "alice"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Send another message.
	time.Sleep(time.Millisecond)
	_, _, _ = svc.SendTextMessage(context.Background(), chat.ID, "bob", "msg2")

	// Manually push an older watermark — should not regress.
	oldTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	_ = repo.MarkChatRead(context.Background(), chat.ID, "alice", oldTime)

	// Should still only see 1 unread (msg2), not 2.
	entries, _ := svc.ListInbox(context.Background(), "alice")
	if len(entries) != 1 || entries[0].UnreadCount != 1 {
		t.Fatalf("expected 1 unread after monotonic guard, got %d", entries[0].UnreadCount)
	}
}

func TestListInboxEmptyChatsHaveNoLastMessage(t *testing.T) {
	repo := newFakeRepo()
	svc := newTestService(repo)

	_, _ = svc.FindOrCreateDM(context.Background(), "alice", "bob")

	entries, err := svc.ListInbox(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 inbox entry, got %d", len(entries))
	}
	if entries[0].LastMessage != nil {
		t.Fatalf("expected nil LastMessage for empty chat, got %+v", entries[0].LastMessage)
	}
	if entries[0].UnreadCount != 0 {
		t.Fatalf("expected 0 unread for empty chat, got %d", entries[0].UnreadCount)
	}
}
