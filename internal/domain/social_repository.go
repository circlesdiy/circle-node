package domain

import "context"

// SocialRepository defines the interface for social interaction operations
type SocialRepository interface {
	// Block operations
	CreateBlock(ctx context.Context, block *Block) error
	GetBlockByID(ctx context.Context, id string) (*Block, error)
	GetBlockByProfiles(ctx context.Context, blockerID, blockedID string) (*Block, error)
	GetBlocksByBlockerID(ctx context.Context, blockerID string) ([]Block, error)
	DeleteBlock(ctx context.Context, id string) error
	IsBlocked(ctx context.Context, blockerID, blockedID string) (bool, error)

	// Mute operations
	CreateMute(ctx context.Context, mute *Mute) error
	GetMuteByID(ctx context.Context, id string) (*Mute, error)
	GetMuteByProfiles(ctx context.Context, muterID, mutedID string) (*Mute, error)
	GetMutesByMuterID(ctx context.Context, muterID string) ([]Mute, error)
	DeleteMute(ctx context.Context, id string) error
	IsMuted(ctx context.Context, muterID, mutedID string) (bool, error)

	// Export operations
	CreateExportBundle(ctx context.Context, bundle *ExportBundle) error
	GetExportBundleByID(ctx context.Context, id string) (*ExportBundle, error)
	GetExportBundlesByProfileID(ctx context.Context, profileID string) ([]ExportBundle, error)
}

// ModerationRepository defines the interface for moderation operations
type ModerationRepository interface {
	// Report operations
	CreateReport(ctx context.Context, report *Report) error
	GetReportByID(ctx context.Context, id string) (*Report, error)
	GetReportsByReporterID(ctx context.Context, reporterID string) ([]Report, error)
	GetReportsByTarget(ctx context.Context, targetType, targetID string) ([]Report, error)
	GetPendingReports(ctx context.Context, limit, offset int) ([]Report, error)
	UpdateReport(ctx context.Context, report *Report) error

	// Moderation action operations
	CreateModerationAction(ctx context.Context, action *ModerationAction) error
	GetModerationActionByID(ctx context.Context, id string) (*ModerationAction, error)
	GetModerationActionsByTarget(ctx context.Context, targetType, targetID string) ([]ModerationAction, error)
	GetModerationActionsByModerator(ctx context.Context, moderatorID string) ([]ModerationAction, error)
}

// ChatRepository defines the interface for chat and messaging operations
type ChatRepository interface {
	// Chat operations
	CreateChat(ctx context.Context, chat *Chat) error
	GetChatByID(ctx context.Context, id string) (*Chat, error)
	GetChatsByProfileID(ctx context.Context, profileID string) ([]Chat, error)
	GetDirectChatByProfiles(ctx context.Context, profileID1, profileID2 string) (*Chat, error)
	UpdateChat(ctx context.Context, chat *Chat) error
	DeleteChat(ctx context.Context, id string) error

	// ChatParticipant operations
	AddParticipant(ctx context.Context, participant *ChatParticipant) error
	RemoveParticipant(ctx context.Context, chatID, profileID string) error
	GetParticipantsByChatID(ctx context.Context, chatID string) ([]ChatParticipant, error)
	IsParticipant(ctx context.Context, chatID, profileID string) (bool, error)

	// Message operations
	CreateMessage(ctx context.Context, message *Message) error
	GetMessageByID(ctx context.Context, id string) (*Message, error)
	GetMessagesByChatID(ctx context.Context, chatID string, limit, offset int) ([]Message, error)
	UpdateMessage(ctx context.Context, message *Message) error
	DeleteMessage(ctx context.Context, id string) error

	// MessageRead operations
	MarkMessageAsRead(ctx context.Context, messageRead *MessageRead) error
	GetReadReceiptsByMessageID(ctx context.Context, messageID string) ([]MessageRead, error)
	GetUnreadMessageCount(ctx context.Context, chatID, profileID string) (int, error)
}

// NotificationRepository defines the interface for notification operations
type NotificationRepository interface {
	CreateNotification(ctx context.Context, notification *Notification) error
	GetNotificationByID(ctx context.Context, id string) (*Notification, error)
	GetNotificationsByProfileID(ctx context.Context, profileID string, limit, offset int) ([]Notification, error)
	GetUnreadNotificationsByProfileID(ctx context.Context, profileID string) ([]Notification, error)
	MarkNotificationAsRead(ctx context.Context, notificationID string) error
	MarkNotificationAsDelivered(ctx context.Context, notificationID string) error
	MarkAllAsRead(ctx context.Context, profileID string) error
	GetUnreadCount(ctx context.Context, profileID string) (int, error)
	DeleteNotification(ctx context.Context, id string) error
}
