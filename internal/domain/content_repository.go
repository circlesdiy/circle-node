package domain

import "context"

// ContentRepository defines the interface for content-related data operations
type ContentRepository interface {
	// Post operations
	CreatePost(ctx context.Context, post *Post) error
	GetPostByID(ctx context.Context, id string) (*Post, error)
	GetPostsByCircleID(ctx context.Context, circleID string, limit, offset int) ([]Post, error)
	GetPostsByProfileID(ctx context.Context, profileID string, limit, offset int) ([]Post, error)
	UpdatePost(ctx context.Context, post *Post) error
	DeletePost(ctx context.Context, id string) error
	IncrementPostReplyCount(ctx context.Context, postID string) error
	DecrementPostReplyCount(ctx context.Context, postID string) error

	// Discussion operations
	CreateDiscussion(ctx context.Context, discussion *Discussion) error
	GetDiscussionByID(ctx context.Context, id string) (*Discussion, error)
	GetDiscussionsByCircleID(ctx context.Context, circleID string, limit, offset int) ([]Discussion, error)
	UpdateDiscussion(ctx context.Context, discussion *Discussion) error
	DeleteDiscussion(ctx context.Context, id string) error
	PinDiscussion(ctx context.Context, discussionID string) error
	UnpinDiscussion(ctx context.Context, discussionID string) error
	LockDiscussion(ctx context.Context, discussionID string) error
	UnlockDiscussion(ctx context.Context, discussionID string) error

	// Comment operations
	CreateComment(ctx context.Context, comment *Comment) error
	GetCommentByID(ctx context.Context, id string) (*Comment, error)
	GetCommentsByPostID(ctx context.Context, postID string) ([]Comment, error)
	GetCommentsByDiscussionID(ctx context.Context, discussionID string) ([]Comment, error)
	GetRepliesByCommentID(ctx context.Context, commentID string) ([]Comment, error)
	UpdateComment(ctx context.Context, comment *Comment) error
	DeleteComment(ctx context.Context, id string) error

	// PostComment junction operations
	LinkCommentToPost(ctx context.Context, postComment *PostComment) error
	UnlinkCommentFromPost(ctx context.Context, postID, commentID string) error

	// DiscussionComment junction operations
	LinkCommentToDiscussion(ctx context.Context, discussionComment *DiscussionComment) error
	UnlinkCommentFromDiscussion(ctx context.Context, discussionID, commentID string) error

	// Attachment operations
	CreateAttachment(ctx context.Context, attachment *Attachment) error
	GetAttachmentByID(ctx context.Context, id string) (*Attachment, error)
	GetAttachmentsByProfileID(ctx context.Context, profileID string) ([]Attachment, error)
	DeleteAttachment(ctx context.Context, id string) error

	// PostAttachment operations
	LinkAttachmentToPost(ctx context.Context, postAttachment *PostAttachment) error
	GetAttachmentsByPostID(ctx context.Context, postID string) ([]Attachment, error)
	UnlinkAttachmentFromPost(ctx context.Context, postID, attachmentID string) error

	// MessageAttachment operations
	LinkAttachmentToMessage(ctx context.Context, messageAttachment *MessageAttachment) error
	GetAttachmentsByMessageID(ctx context.Context, messageID string) ([]Attachment, error)
	UnlinkAttachmentFromMessage(ctx context.Context, messageID, attachmentID string) error
}

// ReactionRepository defines the interface for reaction operations
type ReactionRepository interface {
	CreateReaction(ctx context.Context, reaction *Reaction) error
	GetReactionByID(ctx context.Context, id string) (*Reaction, error)
	GetReactionsByTarget(ctx context.Context, targetType, targetID string) ([]Reaction, error)
	GetReactionsByProfile(ctx context.Context, profileID string) ([]Reaction, error)
	DeleteReaction(ctx context.Context, id string) error
	DeleteReactionByProfileAndTarget(ctx context.Context, profileID, targetType, targetID string) error
	CountReactionsByTarget(ctx context.Context, targetType, targetID string) (map[string]int, error)
}
