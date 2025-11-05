package domain

import "time"

// Attachment represents a file attachment (image, video, document, etc.)
type Attachment struct {
	ID             string     `json:"id"`
	OwnerProfileID string     `json:"owner_profile_id"`
	MimeType       string     `json:"mime_type"`
	Size           int        `json:"size"` // Size in bytes
	StorageURL     string     `json:"storage_url"`
	CID            string     `json:"cid"` // Content-addressed identifier
	CreatedAt      time.Time  `json:"created_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

// IsDeleted checks if the attachment has been soft deleted
func (a *Attachment) IsDeleted() bool {
	return a.DeletedAt != nil
}

// IsImage checks if the attachment is an image
func (a *Attachment) IsImage() bool {
	return len(a.MimeType) >= 6 && a.MimeType[:6] == "image/"
}

// IsVideo checks if the attachment is a video
func (a *Attachment) IsVideo() bool {
	return len(a.MimeType) >= 6 && a.MimeType[:6] == "video/"
}

// IsAudio checks if the attachment is audio
func (a *Attachment) IsAudio() bool {
	return len(a.MimeType) >= 6 && a.MimeType[:6] == "audio/"
}

// SizeInMB returns the size in megabytes
func (a *Attachment) SizeInMB() float64 {
	return float64(a.Size) / 1024.0 / 1024.0
}

// PostAttachment links attachments to posts (junction table)
type PostAttachment struct {
	ID           string `json:"id"`
	PostID       string `json:"post_id"`
	AttachmentID string `json:"attachment_id"`
}

// MessageAttachment links attachments to messages (junction table)
type MessageAttachment struct {
	ID           string `json:"id"`
	MessageID    string `json:"message_id"`
	AttachmentID string `json:"attachment_id"`
}
