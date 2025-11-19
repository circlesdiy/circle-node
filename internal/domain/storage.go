package domain

import (
	"context"
	"mime/multipart"
)

// FileStorage defines the interface for file storage operations
// This abstraction allows swapping between local filesystem, S3, MinIO, etc.
type FileStorage interface {
	// SaveAvatar processes and stores an avatar image
	// Returns the URL for accessing the saved avatar
	SaveAvatar(ctx context.Context, profileID string, file multipart.File, header *multipart.FileHeader) (string, error)

	// SaveBanner processes and stores a banner image
	// Returns the URL for accessing the saved banner
	SaveBanner(ctx context.Context, profileID string, file multipart.File, header *multipart.FileHeader) (string, error)

	// DeleteAvatar removes the avatar file for a profile
	// Should be idempotent (no error if file doesn't exist)
	DeleteAvatar(ctx context.Context, profileID string) error

	// DeleteBanner removes the banner file for a profile
	// Should be idempotent (no error if file doesn't exist)
	DeleteBanner(ctx context.Context, profileID string) error

	// HealthCheck verifies the storage backend is accessible and writable
	HealthCheck(ctx context.Context) error
}
