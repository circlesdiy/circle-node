package profile

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"circles.diy/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles profile business logic
type Service struct {
	repo    domain.ProfileRepository
	storage domain.FileStorage
	logger  *zap.Logger
}

// NewService creates a new profile service
func NewService(repo domain.ProfileRepository, storage domain.FileStorage, logger *zap.Logger) *Service {
	return &Service{
		repo:    repo,
		storage: storage,
		logger:  logger,
	}
}

// CreateDefault creates a default profile for a new user
// This is called during user registration to create the initial profile
func (s *Service) CreateDefault(ctx context.Context, userID, username string) (*domain.Profile, error) {
	now := time.Now().UTC()

	profile := &domain.Profile{
		ID:          uuid.New().String(),
		UserID:      userID,
		Handle:      username, // Use username as initial handle
		Name:        username,
		DisplayName: username,
		Bio:         "",
		AvatarURL:   "",
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	settings := &domain.ProfileSettings{
		ProfileID: profile.ID,
		IsPublic:  true,
		Location:  "",
		Website:   "",
		UpdatedAt: now,
	}

	s.logger.Debug("creating default profile",
		zap.String("user_id", userID),
		zap.String("profile_id", profile.ID),
		zap.String("handle", profile.Handle))

	// Create profile and settings atomically
	err := s.repo.CreateProfileWithSettings(ctx, profile, settings)
	if err != nil {
		s.logger.Error("failed to create default profile",
			zap.String("user_id", userID),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("default profile created successfully",
		zap.String("user_id", userID),
		zap.String("profile_id", profile.ID),
		zap.String("handle", profile.Handle))

	return profile, nil
}

// GetActiveForUser retrieves the user's active profile ID
// Returns nil if no active profile exists
func (s *Service) GetActiveForUser(ctx context.Context, userID string) (*string, error) {
	return s.repo.GetActiveProfileByUserID(ctx, userID)
}

// GetByID retrieves a profile by its ID
func (s *Service) GetByID(ctx context.Context, id string) (*domain.Profile, error) {
	return s.repo.GetProfileByID(ctx, id)
}

func (s *Service) GetByIDs(ctx context.Context, ids []string) ([]domain.Profile, error) {
	return s.repo.GetProfilesByIDs(ctx, ids)
}

// GetByHandle retrieves a profile by its handle
func (s *Service) GetByHandle(ctx context.Context, handle string) (*domain.Profile, error) {
	return s.repo.GetProfileByHandle(ctx, handle)
}

// GetByUserID retrieves all profiles for a user
func (s *Service) GetByUserID(ctx context.Context, userID string) ([]domain.Profile, error) {
	return s.repo.GetProfilesByUserID(ctx, userID)
}

// Update updates an existing profile
func (s *Service) Update(ctx context.Context, profile *domain.Profile) error {
	profile.UpdatedAt = time.Now().UTC()

	s.logger.Debug("updating profile",
		zap.String("profile_id", profile.ID),
		zap.String("handle", profile.Handle))

	err := s.repo.UpdateProfile(ctx, profile)
	if err != nil {
		s.logger.Error("failed to update profile",
			zap.String("profile_id", profile.ID),
			zap.Error(err))
		return err
	}

	s.logger.Info("profile updated successfully",
		zap.String("profile_id", profile.ID),
		zap.String("handle", profile.Handle))

	return nil
}

// Delete soft deletes a profile
func (s *Service) Delete(ctx context.Context, id string) error {
	s.logger.Debug("deleting profile", zap.String("profile_id", id))

	err := s.repo.DeleteProfile(ctx, id)
	if err != nil {
		s.logger.Error("failed to delete profile",
			zap.String("profile_id", id),
			zap.Error(err))
		return err
	}

	s.logger.Info("profile deleted successfully", zap.String("profile_id", id))
	return nil
}

// CheckHandleExists checks if a handle is already taken
func (s *Service) CheckHandleExists(ctx context.Context, handle string) (bool, error) {
	return s.repo.HandleExists(ctx, handle)
}

// GetSettings retrieves profile settings
func (s *Service) GetSettings(ctx context.Context, profileID string) (*domain.ProfileSettings, error) {
	return s.repo.GetProfileSettings(ctx, profileID)
}

// UpdateSettings updates profile settings
func (s *Service) UpdateSettings(ctx context.Context, settings *domain.ProfileSettings) error {
	settings.UpdatedAt = time.Now().UTC()

	s.logger.Debug("updating profile settings",
		zap.String("profile_id", settings.ProfileID))

	err := s.repo.UpdateProfileSettings(ctx, settings)
	if err != nil {
		s.logger.Error("failed to update profile settings",
			zap.String("profile_id", settings.ProfileID),
			zap.Error(err))
		return err
	}

	s.logger.Info("profile settings updated successfully",
		zap.String("profile_id", settings.ProfileID))

	return nil
}

// UpdateAvatar uploads a new avatar and updates the profile
func (s *Service) UpdateAvatar(ctx context.Context, profileID string, file multipart.File, header *multipart.FileHeader) error {
	s.logger.Debug("updating avatar",
		zap.String("profile_id", profileID),
		zap.String("filename", header.Filename))

	// Get current profile to retrieve old avatar URL
	profile, err := s.repo.GetProfileByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}
	if profile == nil {
		return fmt.Errorf("profile not found")
	}

	oldAvatarURL := profile.AvatarURL

	// Upload new avatar
	avatarURL, err := s.storage.SaveAvatar(ctx, profileID, file, header)
	if err != nil {
		s.logger.Error("failed to save avatar",
			zap.String("profile_id", profileID),
			zap.Error(err))
		return fmt.Errorf("failed to save avatar: %w", err)
	}

	// Update profile with new avatar URL
	profile.AvatarURL = avatarURL
	profile.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateProfile(ctx, profile); err != nil {
		// Rollback: delete newly uploaded avatar
		if deleteErr := s.storage.DeleteAvatar(ctx, profileID); deleteErr != nil {
			s.logger.Warn("failed to delete avatar after DB update failure",
				zap.String("profile_id", profileID),
				zap.Error(deleteErr))
		}
		s.logger.Error("failed to update profile with new avatar",
			zap.String("profile_id", profileID),
			zap.Error(err))
		return fmt.Errorf("failed to update profile: %w", err)
	}

	// Delete old avatar file if it exists and is different
	if oldAvatarURL != "" && oldAvatarURL != avatarURL {
		if err := s.storage.DeleteAvatar(ctx, profileID); err != nil {
			s.logger.Warn("failed to delete old avatar",
				zap.String("profile_id", profileID),
				zap.String("old_url", oldAvatarURL),
				zap.Error(err))
		}
	}

	s.logger.Info("avatar updated successfully",
		zap.String("profile_id", profileID),
		zap.String("new_url", avatarURL))

	return nil
}

// UpdateBanner uploads a new banner and updates the profile
func (s *Service) UpdateBanner(ctx context.Context, profileID string, file multipart.File, header *multipart.FileHeader) error {
	s.logger.Debug("updating banner",
		zap.String("profile_id", profileID),
		zap.String("filename", header.Filename))

	// Get current profile to retrieve old banner URL
	profile, err := s.repo.GetProfileByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}
	if profile == nil {
		return fmt.Errorf("profile not found")
	}

	oldBannerURL := profile.BannerURL

	// Upload new banner
	bannerURL, err := s.storage.SaveBanner(ctx, profileID, file, header)
	if err != nil {
		s.logger.Error("failed to save banner",
			zap.String("profile_id", profileID),
			zap.Error(err))
		return fmt.Errorf("failed to save banner: %w", err)
	}

	// Update profile with new banner URL
	profile.BannerURL = bannerURL
	profile.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateProfile(ctx, profile); err != nil {
		// Rollback: delete newly uploaded banner
		if deleteErr := s.storage.DeleteBanner(ctx, profileID); deleteErr != nil {
			s.logger.Warn("failed to delete banner after DB update failure",
				zap.String("profile_id", profileID),
				zap.Error(deleteErr))
		}
		s.logger.Error("failed to update profile with new banner",
			zap.String("profile_id", profileID),
			zap.Error(err))
		return fmt.Errorf("failed to update profile: %w", err)
	}

	// Delete old banner file if it exists and is different
	if oldBannerURL != "" && oldBannerURL != bannerURL {
		if err := s.storage.DeleteBanner(ctx, profileID); err != nil {
			s.logger.Warn("failed to delete old banner",
				zap.String("profile_id", profileID),
				zap.String("old_url", oldBannerURL),
				zap.Error(err))
		}
	}

	s.logger.Info("banner updated successfully",
		zap.String("profile_id", profileID),
		zap.String("new_url", bannerURL))

	return nil
}

// RemoveAvatar removes the avatar from a profile
func (s *Service) RemoveAvatar(ctx context.Context, profileID string) error {
	s.logger.Debug("removing avatar", zap.String("profile_id", profileID))

	// Get current profile
	profile, err := s.repo.GetProfileByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}
	if profile == nil {
		return fmt.Errorf("profile not found")
	}

	// Delete avatar file
	if err := s.storage.DeleteAvatar(ctx, profileID); err != nil {
		s.logger.Warn("failed to delete avatar file",
			zap.String("profile_id", profileID),
			zap.Error(err))
	}

	// Update profile to remove avatar URL
	profile.AvatarURL = ""
	profile.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateProfile(ctx, profile); err != nil {
		s.logger.Error("failed to update profile to remove avatar",
			zap.String("profile_id", profileID),
			zap.Error(err))
		return fmt.Errorf("failed to update profile: %w", err)
	}

	s.logger.Info("avatar removed successfully", zap.String("profile_id", profileID))
	return nil
}

// RemoveBanner removes the banner from a profile
func (s *Service) RemoveBanner(ctx context.Context, profileID string) error {
	s.logger.Debug("removing banner", zap.String("profile_id", profileID))

	// Get current profile
	profile, err := s.repo.GetProfileByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}
	if profile == nil {
		return fmt.Errorf("profile not found")
	}

	// Delete banner file
	if err := s.storage.DeleteBanner(ctx, profileID); err != nil {
		s.logger.Warn("failed to delete banner file",
			zap.String("profile_id", profileID),
			zap.Error(err))
	}

	// Update profile to remove banner URL
	profile.BannerURL = ""
	profile.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateProfile(ctx, profile); err != nil {
		s.logger.Error("failed to update profile to remove banner",
			zap.String("profile_id", profileID),
			zap.Error(err))
		return fmt.Errorf("failed to update profile: %w", err)
	}

	s.logger.Info("banner removed successfully", zap.String("profile_id", profileID))
	return nil
}
