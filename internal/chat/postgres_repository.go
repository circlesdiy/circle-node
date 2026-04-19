package chat

import (
	"context"
	"errors"
	"fmt"
	"time"

	"circles.diy/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

const chatSelectColumns = `c.id, c.type, COALESCE(c.name, '') AS name, c.circle_id, c.last_message_at, c.created_at, c.updated_at, c.deleted_at`

const messageSelectColumns = `id, chat_id, sender_profile_id, content, message_type,
	COALESCE(encryption_scheme, '') AS encryption_scheme,
	COALESCE(nonce, '') AS nonce,
	reply_to_message_id, is_flagged, edited_at, created_at, updated_at, deleted_at`

func scanChat(row pgx.Row, c *domain.Chat) error {
	return row.Scan(
		&c.ID,
		&c.Type,
		&c.Name,
		&c.CircleID,
		&c.LastMessageAt,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.DeletedAt,
	)
}

func scanMessage(row pgx.Row, m *domain.Message) error {
	return row.Scan(
		&m.ID,
		&m.ChatID,
		&m.SenderProfileID,
		&m.Content,
		&m.MessageType,
		&m.EncryptionScheme,
		&m.Nonce,
		&m.ReplyToMessageID,
		&m.IsFlagged,
		&m.EditedAt,
		&m.CreatedAt,
		&m.UpdatedAt,
		&m.DeletedAt,
	)
}

func (r *PostgresRepository) CreateChat(ctx context.Context, chat *domain.Chat) error {
	query := `
		INSERT INTO chats (id, type, name, circle_id, last_message_at, created_at, updated_at)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, query,
		chat.ID,
		chat.Type,
		chat.Name,
		chat.CircleID,
		chat.LastMessageAt,
		chat.CreatedAt,
		chat.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create chat: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetChatByID(ctx context.Context, id string) (*domain.Chat, error) {
	query := `SELECT ` + chatSelectColumns + ` FROM chats c WHERE c.id = $1 AND c.deleted_at IS NULL`

	var c domain.Chat
	if err := scanChat(r.pool.QueryRow(ctx, query, id), &c); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get chat by id: %w", err)
	}
	return &c, nil
}

func (r *PostgresRepository) GetChatsByProfileID(ctx context.Context, profileID string) ([]domain.Chat, error) {
	query := `
		SELECT ` + chatSelectColumns + `
		FROM chats c
		JOIN chat_participants cp ON cp.chat_id = c.id
		WHERE cp.profile_id = $1 AND c.deleted_at IS NULL
		ORDER BY c.last_message_at DESC NULLS LAST, c.created_at DESC`

	rows, err := r.pool.Query(ctx, query, profileID)
	if err != nil {
		return nil, fmt.Errorf("list chats for profile: %w", err)
	}
	defer rows.Close()

	var chats []domain.Chat
	for rows.Next() {
		var c domain.Chat
		if err := scanChat(rows, &c); err != nil {
			return nil, fmt.Errorf("scan chat row: %w", err)
		}
		chats = append(chats, c)
	}
	return chats, rows.Err()
}

func (r *PostgresRepository) GetChatsByCircleID(ctx context.Context, circleID string) ([]domain.Chat, error) {
	query := `
		SELECT ` + chatSelectColumns + `
		FROM chats c
		WHERE c.circle_id = $1 AND c.deleted_at IS NULL
		ORDER BY c.last_message_at DESC NULLS LAST, c.created_at DESC`

	rows, err := r.pool.Query(ctx, query, circleID)
	if err != nil {
		return nil, fmt.Errorf("list chats for circle: %w", err)
	}
	defer rows.Close()

	var chats []domain.Chat
	for rows.Next() {
		var c domain.Chat
		if err := scanChat(rows, &c); err != nil {
			return nil, fmt.Errorf("scan chat row: %w", err)
		}
		chats = append(chats, c)
	}
	return chats, rows.Err()
}

func (r *PostgresRepository) GetDirectChatByProfiles(ctx context.Context, profileID1, profileID2 string) (*domain.Chat, error) {
	query := `
		SELECT ` + chatSelectColumns + `
		FROM chats c
		WHERE c.type = 'direct'
		  AND c.deleted_at IS NULL
		  AND c.circle_id IS NULL
		  AND EXISTS (SELECT 1 FROM chat_participants p WHERE p.chat_id = c.id AND p.profile_id = $1)
		  AND EXISTS (SELECT 1 FROM chat_participants p WHERE p.chat_id = c.id AND p.profile_id = $2)
		  AND (SELECT COUNT(*) FROM chat_participants p WHERE p.chat_id = c.id) = 2
		ORDER BY c.created_at ASC
		LIMIT 1`

	var c domain.Chat
	if err := scanChat(r.pool.QueryRow(ctx, query, profileID1, profileID2), &c); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("lookup direct chat: %w", err)
	}
	return &c, nil
}

func (r *PostgresRepository) UpdateChat(ctx context.Context, chat *domain.Chat) error {
	query := `
		UPDATE chats
		SET type = $2,
		    name = NULLIF($3, ''),
		    circle_id = $4,
		    last_message_at = $5
		WHERE id = $1 AND deleted_at IS NULL`

	tag, err := r.pool.Exec(ctx, query,
		chat.ID,
		chat.Type,
		chat.Name,
		chat.CircleID,
		chat.LastMessageAt,
	)
	if err != nil {
		return fmt.Errorf("update chat: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("chat %s not found", chat.ID)
	}
	return nil
}

func (r *PostgresRepository) DeleteChat(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE chats SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`,
		id)
	if err != nil {
		return fmt.Errorf("delete chat: %w", err)
	}
	return nil
}

func (r *PostgresRepository) AddParticipant(ctx context.Context, participant *domain.ChatParticipant) error {
	query := `
		INSERT INTO chat_participants (id, chat_id, profile_id, is_admin, joined_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (chat_id, profile_id) DO NOTHING`

	_, err := r.pool.Exec(ctx, query,
		participant.ID,
		participant.ChatID,
		participant.ProfileID,
		participant.IsAdmin,
		participant.JoinedAt,
	)
	if err != nil {
		return fmt.Errorf("add participant: %w", err)
	}
	return nil
}

func (r *PostgresRepository) RemoveParticipant(ctx context.Context, chatID, profileID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM chat_participants WHERE chat_id = $1 AND profile_id = $2`,
		chatID, profileID)
	if err != nil {
		return fmt.Errorf("remove participant: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetParticipantsByChatID(ctx context.Context, chatID string) ([]domain.ChatParticipant, error) {
	query := `
		SELECT id, chat_id, profile_id, is_admin, joined_at
		FROM chat_participants
		WHERE chat_id = $1
		ORDER BY joined_at ASC`

	rows, err := r.pool.Query(ctx, query, chatID)
	if err != nil {
		return nil, fmt.Errorf("list participants: %w", err)
	}
	defer rows.Close()

	var participants []domain.ChatParticipant
	for rows.Next() {
		var p domain.ChatParticipant
		if err := rows.Scan(&p.ID, &p.ChatID, &p.ProfileID, &p.IsAdmin, &p.JoinedAt); err != nil {
			return nil, fmt.Errorf("scan participant row: %w", err)
		}
		participants = append(participants, p)
	}
	return participants, rows.Err()
}

func (r *PostgresRepository) IsParticipant(ctx context.Context, chatID, profileID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM chat_participants WHERE chat_id = $1 AND profile_id = $2)`,
		chatID, profileID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check participant: %w", err)
	}
	return exists, nil
}

// CreateMessage inserts a message and updates the parent chat's last_message_at atomically.
// The chats.updated_at column is bumped by the existing update_chats_updated_at trigger.
func (r *PostgresRepository) CreateMessage(ctx context.Context, message *domain.Message) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	insert := `
		INSERT INTO messages (
			id, chat_id, sender_profile_id, content, message_type,
			encryption_scheme, nonce, reply_to_message_id, is_flagged,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''), $8, $9, $10, $11)`

	if _, err := tx.Exec(ctx, insert,
		message.ID,
		message.ChatID,
		message.SenderProfileID,
		message.Content,
		message.MessageType,
		message.EncryptionScheme,
		message.Nonce,
		message.ReplyToMessageID,
		message.IsFlagged,
		message.CreatedAt,
		message.UpdatedAt,
	); err != nil {
		return fmt.Errorf("insert message: %w", err)
	}

	if _, err := tx.Exec(ctx,
		`UPDATE chats SET last_message_at = $2 WHERE id = $1 AND deleted_at IS NULL`,
		message.ChatID, message.CreatedAt,
	); err != nil {
		return fmt.Errorf("bump last_message_at: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetMessageByID(ctx context.Context, id string) (*domain.Message, error) {
	query := `SELECT ` + messageSelectColumns + ` FROM messages WHERE id = $1 AND deleted_at IS NULL`

	var m domain.Message
	if err := scanMessage(r.pool.QueryRow(ctx, query, id), &m); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get message by id: %w", err)
	}
	return &m, nil
}

// GetMessagesByChatID returns the most recent messages first. The service layer reverses them
// for chronological display. limit and offset provide a simple paginator for the public interface.
func (r *PostgresRepository) GetMessagesByChatID(ctx context.Context, chatID string, limit, offset int) ([]domain.Message, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT ` + messageSelectColumns + `
		FROM messages
		WHERE chat_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, chatID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var m domain.Message
		if err := scanMessage(rows, &m); err != nil {
			return nil, fmt.Errorf("scan message row: %w", err)
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

// GetMessagesBefore returns messages older than the supplied keyset (created_at, id) cursor.
// When beforeTime is zero, it returns the most recent page. Results are ordered newest-first.
func (r *PostgresRepository) GetMessagesBefore(ctx context.Context, chatID string, beforeTime time.Time, beforeID string, limit int) ([]domain.Message, error) {
	if limit <= 0 {
		limit = 50
	}

	if beforeTime.IsZero() {
		return r.GetMessagesByChatID(ctx, chatID, limit, 0)
	}

	query := `
		SELECT ` + messageSelectColumns + `
		FROM messages
		WHERE chat_id = $1
		  AND deleted_at IS NULL
		  AND (created_at, id) < ($2, $3)
		ORDER BY created_at DESC, id DESC
		LIMIT $4`

	rows, err := r.pool.Query(ctx, query, chatID, beforeTime, beforeID, limit)
	if err != nil {
		return nil, fmt.Errorf("list messages before: %w", err)
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var m domain.Message
		if err := scanMessage(rows, &m); err != nil {
			return nil, fmt.Errorf("scan message row: %w", err)
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

// GetMessagesAfter replays messages the client missed while disconnected, returned in
// chronological order so they can be appended directly.
func (r *PostgresRepository) GetMessagesAfter(ctx context.Context, chatID, afterID string, limit int) ([]domain.Message, error) {
	if limit <= 0 || limit > 200 {
		limit = 200
	}

	var query string
	var args []any
	if afterID == "" {
		query = `
			SELECT ` + messageSelectColumns + `
			FROM messages
			WHERE chat_id = $1 AND deleted_at IS NULL
			ORDER BY created_at ASC, id ASC
			LIMIT $2`
		args = []any{chatID, limit}
	} else {
		query = `
			SELECT ` + messageSelectColumns + `
			FROM messages
			WHERE chat_id = $1
			  AND deleted_at IS NULL
			  AND (created_at, id) > (
			      SELECT created_at, id FROM messages WHERE id = $2
			  )
			ORDER BY created_at ASC, id ASC
			LIMIT $3`
		args = []any{chatID, afterID, limit}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list messages after: %w", err)
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var m domain.Message
		if err := scanMessage(rows, &m); err != nil {
			return nil, fmt.Errorf("scan message row: %w", err)
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

// LatestMessageByChat returns the most recent non-deleted message for a chat, or nil if none exist.
func (r *PostgresRepository) LatestMessageByChat(ctx context.Context, chatID string) (*domain.Message, error) {
	query := `
		SELECT ` + messageSelectColumns + `
		FROM messages
		WHERE chat_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC, id DESC
		LIMIT 1`

	var m domain.Message
	if err := scanMessage(r.pool.QueryRow(ctx, query, chatID), &m); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("latest message: %w", err)
	}
	return &m, nil
}

func (r *PostgresRepository) UpdateMessage(ctx context.Context, message *domain.Message) error {
	query := `
		UPDATE messages
		SET content = $2, is_flagged = $3, edited_at = $4
		WHERE id = $1 AND deleted_at IS NULL`

	tag, err := r.pool.Exec(ctx, query,
		message.ID,
		message.Content,
		message.IsFlagged,
		message.EditedAt,
	)
	if err != nil {
		return fmt.Errorf("update message: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("message %s not found", message.ID)
	}
	return nil
}

func (r *PostgresRepository) DeleteMessage(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE messages SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`,
		id)
	if err != nil {
		return fmt.Errorf("delete message: %w", err)
	}
	return nil
}

func (r *PostgresRepository) MarkMessageAsRead(ctx context.Context, messageRead *domain.MessageRead) error {
	query := `
		INSERT INTO message_reads (id, message_id, profile_id, read_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (message_id, profile_id) DO NOTHING`

	_, err := r.pool.Exec(ctx, query,
		messageRead.ID,
		messageRead.MessageID,
		messageRead.ProfileID,
		messageRead.ReadAt,
	)
	if err != nil {
		return fmt.Errorf("mark read: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetReadReceiptsByMessageID(ctx context.Context, messageID string) ([]domain.MessageRead, error) {
	query := `
		SELECT id, message_id, profile_id, read_at
		FROM message_reads
		WHERE message_id = $1
		ORDER BY read_at ASC`

	rows, err := r.pool.Query(ctx, query, messageID)
	if err != nil {
		return nil, fmt.Errorf("list read receipts: %w", err)
	}
	defer rows.Close()

	var receipts []domain.MessageRead
	for rows.Next() {
		var r domain.MessageRead
		if err := rows.Scan(&r.ID, &r.MessageID, &r.ProfileID, &r.ReadAt); err != nil {
			return nil, fmt.Errorf("scan read receipt: %w", err)
		}
		receipts = append(receipts, r)
	}
	return receipts, rows.Err()
}

// IsBlockedEitherWay reports whether either profile has blocked the other.
// Used by the chat service to reject DM creation between mutually-blocked users.
func (r *PostgresRepository) IsBlockedEitherWay(ctx context.Context, profileA, profileB string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM blocks
			WHERE (blocker_profile_id = $1 AND blocked_profile_id = $2)
			   OR (blocker_profile_id = $2 AND blocked_profile_id = $1)
		)`, profileA, profileB).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check block: %w", err)
	}
	return exists, nil
}

func (r *PostgresRepository) GetUnreadMessageCount(ctx context.Context, chatID, profileID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM messages m
		JOIN chat_participants cp
		  ON cp.chat_id = m.chat_id AND cp.profile_id = $2
		WHERE m.chat_id = $1
		  AND m.deleted_at IS NULL
		  AND m.sender_profile_id <> $2
		  AND m.created_at > COALESCE(cp.last_read_at, '1970-01-01'::timestamp)`

	var count int
	if err := r.pool.QueryRow(ctx, query, chatID, profileID).Scan(&count); err != nil {
		return 0, fmt.Errorf("unread count: %w", err)
	}
	return count, nil
}

// MarkChatRead advances the watermark for the viewer in the given chat.
// The comparison ensures the watermark only moves forward, never backward.
func (r *PostgresRepository) MarkChatRead(ctx context.Context, chatID, profileID string, readAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE chat_participants
		SET last_read_at = $3
		WHERE chat_id = $1 AND profile_id = $2
		  AND (last_read_at IS NULL OR last_read_at < $3)`,
		chatID, profileID, readAt)
	if err != nil {
		return fmt.Errorf("mark chat read: %w", err)
	}
	return nil
}

// InboxRow is the flat result of the batch inbox query. The service layer
// maps these into InboxEntry values.
type InboxRow struct {
	// Chat fields
	Chat domain.Chat
	// Latest message (nullable)
	LastMessageID          *string
	LastMessageSenderID    *string
	LastMessageContent     *string
	LastMessageType        *string
	LastMessageCreatedAt   *time.Time
	// Unread count
	UnreadCount int
}

// GetInboxEntries returns all chats the profile participates in together with
// the latest message and unread count, in a single query.
func (r *PostgresRepository) GetInboxEntries(ctx context.Context, profileID string) ([]InboxRow, error) {
	query := `
		SELECT
			c.id, c.type, COALESCE(c.name,'') AS name, c.circle_id,
			c.last_message_at, c.created_at, c.updated_at, c.deleted_at,
			lm.id, lm.sender_profile_id, lm.content, lm.message_type, lm.created_at,
			(SELECT COUNT(*)
			 FROM messages um
			 WHERE um.chat_id = c.id
			   AND um.deleted_at IS NULL
			   AND um.sender_profile_id <> $1
			   AND um.created_at > COALESCE(cp.last_read_at, '1970-01-01'::timestamp)
			) AS unread_count
		FROM chats c
		JOIN chat_participants cp ON cp.chat_id = c.id AND cp.profile_id = $1
		LEFT JOIN LATERAL (
			SELECT id, sender_profile_id, content, message_type, created_at
			FROM messages
			WHERE chat_id = c.id AND deleted_at IS NULL
			ORDER BY created_at DESC, id DESC
			LIMIT 1
		) lm ON true
		WHERE c.deleted_at IS NULL
		ORDER BY c.last_message_at DESC NULLS LAST, c.created_at DESC`

	rows, err := r.pool.Query(ctx, query, profileID)
	if err != nil {
		return nil, fmt.Errorf("inbox entries: %w", err)
	}
	defer rows.Close()

	var entries []InboxRow
	for rows.Next() {
		var e InboxRow
		if err := rows.Scan(
			&e.Chat.ID, &e.Chat.Type, &e.Chat.Name, &e.Chat.CircleID,
			&e.Chat.LastMessageAt, &e.Chat.CreatedAt, &e.Chat.UpdatedAt, &e.Chat.DeletedAt,
			&e.LastMessageID, &e.LastMessageSenderID, &e.LastMessageContent,
			&e.LastMessageType, &e.LastMessageCreatedAt,
			&e.UnreadCount,
		); err != nil {
			return nil, fmt.Errorf("scan inbox row: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

var _ domain.ChatRepository = (*PostgresRepository)(nil)
