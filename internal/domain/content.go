package domain

import "time"

// Post represents a post within a circle
type Post struct {
	ID                string     `json:"id"`
	CircleID          string     `json:"circle_id"`
	AuthorProfileID   string     `json:"author_profile_id"`
	Body              string     `json:"body"`
	BodyFormat        string     `json:"body_format"` // markdown, plaintext, html
	ContentWarning    string     `json:"content_warning"`
	Visibility        string     `json:"visibility"` // public, members_only, private
	ReplyCount        int        `json:"reply_count"`
	AttachmentsCount  int        `json:"attachments_count"`
	ModerationStatus  string     `json:"moderation_status"` // pending, approved, rejected, flagged
	EditedAt          *time.Time `json:"edited_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty"`
	CID               string     `json:"cid"`       // Content-addressed identifier for federation
	Signature         string     `json:"signature"` // Cryptographic signature
}

// IsDeleted checks if the post has been soft deleted
func (p *Post) IsDeleted() bool {
	return p.DeletedAt != nil
}

// IsEdited checks if the post has been edited
func (p *Post) IsEdited() bool {
	return p.EditedAt != nil
}

// IsApproved checks if the post has been approved by moderation
func (p *Post) IsApproved() bool {
	return p.ModerationStatus == ModerationStatusApproved
}

// IsFlagged checks if the post has been flagged for moderation
func (p *Post) IsFlagged() bool {
	return p.ModerationStatus == ModerationStatusFlagged
}

// Post visibility constants
const (
	PostVisibilityPublic      = "public"
	PostVisibilityMembersOnly = "members_only"
	PostVisibilityPrivate     = "private"
)

// Body format constants
const (
	BodyFormatMarkdown  = "markdown"
	BodyFormatPlaintext = "plaintext"
	BodyFormatHTML      = "html"
)

// Moderation status constants
const (
	ModerationStatusPending  = "pending"
	ModerationStatusApproved = "approved"
	ModerationStatusRejected = "rejected"
	ModerationStatusFlagged  = "flagged"
)

// Discussion represents a threaded discussion within a circle
type Discussion struct {
	ID              string     `json:"id"`
	CircleID        string     `json:"circle_id"`
	AuthorProfileID string     `json:"author_profile_id"`
	Title           string     `json:"title"`
	Body            string     `json:"body"`
	IsPinned        bool       `json:"is_pinned"`
	IsLocked        bool       `json:"is_locked"`
	ReplyCount      int        `json:"reply_count"`
	EditedAt        *time.Time `json:"edited_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
	CID             string     `json:"cid"`
	Signature       string     `json:"signature"`
}

// IsDeleted checks if the discussion has been soft deleted
func (d *Discussion) IsDeleted() bool {
	return d.DeletedAt != nil
}

// IsEdited checks if the discussion has been edited
func (d *Discussion) IsEdited() bool {
	return d.EditedAt != nil
}

// CanReply checks if users can reply to this discussion
func (d *Discussion) CanReply() bool {
	return !d.IsLocked && d.DeletedAt == nil
}

// Comment represents a comment on a post or discussion
type Comment struct {
	ID                 string     `json:"id"`
	AuthorProfileID    string     `json:"author_profile_id"`
	Body               string     `json:"body"`
	BodyFormat         string     `json:"body_format"`
	ReplyToCommentID   *string    `json:"reply_to_comment_id,omitempty"` // For nested replies
	EditedAt           *time.Time `json:"edited_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
	CID                string     `json:"cid"`
	Signature          string     `json:"signature"`
}

// IsDeleted checks if the comment has been soft deleted
func (c *Comment) IsDeleted() bool {
	return c.DeletedAt != nil
}

// IsEdited checks if the comment has been edited
func (c *Comment) IsEdited() bool {
	return c.EditedAt != nil
}

// IsReply checks if this comment is a reply to another comment
func (c *Comment) IsReply() bool {
	return c.ReplyToCommentID != nil
}

// PostComment links comments to posts (junction table)
type PostComment struct {
	ID        string `json:"id"`
	PostID    string `json:"post_id"`
	CommentID string `json:"comment_id"`
}

// DiscussionComment links comments to discussions (junction table)
type DiscussionComment struct {
	ID           string `json:"id"`
	DiscussionID string `json:"discussion_id"`
	CommentID    string `json:"comment_id"`
}
