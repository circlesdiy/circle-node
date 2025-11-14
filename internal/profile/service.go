package profile

import (
	"context"
	"time"

	"circles.diy/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles profile business logic
type Service struct {
	repo   domain.ProfileRepository
	logger *zap.Logger
}

// NewService creates a new profile service
func NewService(repo domain.ProfileRepository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
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
