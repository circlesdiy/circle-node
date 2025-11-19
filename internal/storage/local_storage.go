package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"circles.diy/internal/config"
	"circles.diy/internal/uploads"
	"go.uber.org/zap"
)

// LocalStorage implements FileStorage interface using local filesystem
type LocalStorage struct {
	storageDir       string
	baseURL          string
	avatarValidator  *uploads.Validator
	bannerValidator  *uploads.Validator
	processor        *uploads.ImageProcessor
	logger           *zap.Logger
}

// NewLocalStorage creates a new local filesystem storage instance
func NewLocalStorage(cfg config.UploadConfig, logger *zap.Logger) (*LocalStorage, error) {
	// Create storage directories if they don't exist
	avatarDir := filepath.Join(cfg.StorageDir, "avatars")
	bannerDir := filepath.Join(cfg.StorageDir, "banners")

	for _, dir := range []string{avatarDir, bannerDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create validators with specific size limits
	avatarValidator := uploads.NewValidator(cfg.MaxAvatarSizeMB, cfg.AllowedTypes)
	bannerValidator := uploads.NewValidator(cfg.MaxBannerSizeMB, cfg.AllowedTypes)

	// Create image processor
	processor := uploads.NewImageProcessor()

	logger.Info("local storage initialized",
		zap.String("storage_dir", cfg.StorageDir),
		zap.String("base_url", cfg.BaseURL))

	return &LocalStorage{
		storageDir:      cfg.StorageDir,
		baseURL:         cfg.BaseURL,
		avatarValidator: avatarValidator,
		bannerValidator: bannerValidator,
		processor:       processor,
		logger:          logger,
	}, nil
}

// SaveAvatar processes and stores an avatar image
func (s *LocalStorage) SaveAvatar(ctx context.Context, profileID string, file multipart.File, header *multipart.FileHeader) (string, error) {
	// Validate file
	if err := s.avatarValidator.ValidateFile(file, header); err != nil {
		return "", fmt.Errorf("avatar validation failed: %w", err)
	}

	// Reset file pointer after validation
	if _, err := file.Seek(0, 0); err != nil {
		return "", fmt.Errorf("failed to reset file pointer: %w", err)
	}

	// Process image (resize, re-encode)
	processed, err := s.processor.ProcessAvatar(file)
	if err != nil {
		return "", fmt.Errorf("avatar processing failed: %w", err)
	}

	// Determine file extension based on output format
	extension := ".jpg"

	// Generate filename: avatars/{profile-id}.jpg
	filename := profileID + extension
	filePath := filepath.Join(s.storageDir, "avatars", filename)

	// Write file atomically (temp file -> rename)
	if err := s.writeFileAtomic(filePath, processed); err != nil {
		return "", fmt.Errorf("failed to write avatar: %w", err)
	}

	// Generate URL: /uploads/avatars/{profile-id}.jpg
	url := fmt.Sprintf("%s/avatars/%s", s.baseURL, filename)

	s.logger.Info("avatar saved",
		zap.String("profile_id", profileID),
		zap.String("url", url),
		zap.String("original_filename", header.Filename))

	return url, nil
}

// SaveBanner processes and stores a banner image
func (s *LocalStorage) SaveBanner(ctx context.Context, profileID string, file multipart.File, header *multipart.FileHeader) (string, error) {
	// Validate file
	if err := s.bannerValidator.ValidateFile(file, header); err != nil {
		return "", fmt.Errorf("banner validation failed: %w", err)
	}

	// Reset file pointer after validation
	if _, err := file.Seek(0, 0); err != nil {
		return "", fmt.Errorf("failed to reset file pointer: %w", err)
	}

	// Process image (resize, re-encode)
	processed, err := s.processor.ProcessBanner(file)
	if err != nil {
		return "", fmt.Errorf("banner processing failed: %w", err)
	}

	// Determine file extension based on output format
	extension := ".jpg"

	// Generate filename: banners/{profile-id}.jpg
	filename := profileID + extension
	filePath := filepath.Join(s.storageDir, "banners", filename)

	// Write file atomically (temp file -> rename)
	if err := s.writeFileAtomic(filePath, processed); err != nil {
		return "", fmt.Errorf("failed to write banner: %w", err)
	}

	// Generate URL: /uploads/banners/{profile-id}.jpg
	url := fmt.Sprintf("%s/banners/%s", s.baseURL, filename)

	s.logger.Info("banner saved",
		zap.String("profile_id", profileID),
		zap.String("url", url),
		zap.String("original_filename", header.Filename))

	return url, nil
}

// DeleteAvatar removes an avatar file
func (s *LocalStorage) DeleteAvatar(ctx context.Context, profileID string) error {
	// Try both .jpg and .png extensions for backwards compatibility
	for _, ext := range []string{".jpg", ".png"} {
		filename := profileID + ext
		filePath := filepath.Join(s.storageDir, "avatars", filename)

		if err := os.Remove(filePath); err != nil {
			if !os.IsNotExist(err) {
				s.logger.Warn("failed to delete avatar",
					zap.String("profile_id", profileID),
					zap.String("path", filePath),
					zap.Error(err))
			}
		} else {
			s.logger.Info("avatar deleted",
				zap.String("profile_id", profileID),
				zap.String("path", filePath))
			return nil
		}
	}

	// Not an error if file doesn't exist (idempotent)
	return nil
}

// DeleteBanner removes a banner file
func (s *LocalStorage) DeleteBanner(ctx context.Context, profileID string) error {
	// Try both .jpg and .png extensions for backwards compatibility
	for _, ext := range []string{".jpg", ".png"} {
		filename := profileID + ext
		filePath := filepath.Join(s.storageDir, "banners", filename)

		if err := os.Remove(filePath); err != nil {
			if !os.IsNotExist(err) {
				s.logger.Warn("failed to delete banner",
					zap.String("profile_id", profileID),
					zap.String("path", filePath),
					zap.Error(err))
			}
		} else {
			s.logger.Info("banner deleted",
				zap.String("profile_id", profileID),
				zap.String("path", filePath))
			return nil
		}
	}

	// Not an error if file doesn't exist (idempotent)
	return nil
}

// HealthCheck verifies storage directory is accessible and writable
func (s *LocalStorage) HealthCheck(ctx context.Context) error {
	// Check if storage directory exists and is writable
	testFile := filepath.Join(s.storageDir, ".healthcheck")

	// Try to create a test file
	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("storage directory not writable: %w", err)
	}
	f.Close()

	// Clean up test file
	if err := os.Remove(testFile); err != nil {
		s.logger.Warn("failed to remove health check file", zap.Error(err))
	}

	return nil
}

// writeFileAtomic writes data to a file atomically using temp file + rename
func (s *LocalStorage) writeFileAtomic(path string, data io.Reader) error {
	// Create temp file in same directory (ensures same filesystem for rename)
	dir := filepath.Dir(path)
	tempFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()

	// Clean up temp file on error
	defer func() {
		if tempFile != nil {
			tempFile.Close()
			os.Remove(tempPath)
		}
	}()

	// Write data to temp file
	if _, err := io.Copy(tempFile, data); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Sync to disk
	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	// Close temp file
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	tempFile = nil // Prevent cleanup in defer

	// Atomic rename
	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath) // Clean up on rename failure
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
