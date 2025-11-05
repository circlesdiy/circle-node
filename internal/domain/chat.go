package domain

import "time"

// Chat represents a chat conversation (1-on-1 or group)
type Chat struct {
	ID            string     `json:"id"`
	Type          string     `json:"type"` // direct, group
	Name          string     `json:"name"` // For group chats
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}

// IsDeleted checks if the chat has been soft deleted
func (c *Chat) IsDeleted() bool {
	return c.DeletedAt != nil
}

// IsGroup checks if this is a group chat
func (c *Chat) IsGroup() bool {
	return c.Type == ChatTypeGroup
}

// IsDirect checks if this is a direct/1-on-1 chat
func (c *Chat) IsDirect() bool {
	return c.Type == ChatTypeDirect
}

// Chat type constants
const (
	ChatTypeDirect = "direct"
	ChatTypeGroup  = "group"
)

// ChatParticipant represents a profile's participation in a chat
type ChatParticipant struct {
	ID        string    `json:"id"`
	ChatID    string    `json:"chat_id"`
	ProfileID string    `json:"profile_id"`
	IsAdmin   bool      `json:"is_admin"`
	JoinedAt  time.Time `json:"joined_at"`
}

// Message represents a message in a chat
type Message struct {
	ID               string     `json:"id"`
	ChatID           string     `json:"chat_id"`
	SenderProfileID  string     `json:"sender_profile_id"`
	Content          string     `json:"content"`
	MessageType      string     `json:"message_type"` // text, image, file, system
	EncryptionScheme string     `json:"encryption_scheme"`
	Nonce            string     `json:"nonce"` // For encryption
	ReplyToMessageID *string    `json:"reply_to_message_id,omitempty"`
	IsFlagged        bool       `json:"is_flagged"`
	EditedAt         *time.Time `json:"edited_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}

// IsDeleted checks if the message has been soft deleted
func (m *Message) IsDeleted() bool {
	return m.DeletedAt != nil
}

// IsEdited checks if the message has been edited
func (m *Message) IsEdited() bool {
	return m.EditedAt != nil
}

// IsReply checks if this message is a reply to another message
func (m *Message) IsReply() bool {
	return m.ReplyToMessageID != nil
}

// IsEncrypted checks if the message is encrypted
func (m *Message) IsEncrypted() bool {
	return m.EncryptionScheme != "" && m.EncryptionScheme != "none"
}

// Message type constants
const (
	MessageTypeText   = "text"
	MessageTypeImage  = "image"
	MessageTypeFile   = "file"
	MessageTypeSystem = "system"
)

// MessageRead tracks read receipts for messages
type MessageRead struct {
	ID        string    `json:"id"`
	MessageID string    `json:"message_id"`
	ProfileID string    `json:"profile_id"`
	ReadAt    time.Time `json:"read_at"`
}
