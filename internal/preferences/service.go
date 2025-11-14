package preferences

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"circles.diy/internal/domain"
	"go.uber.org/zap"
)

// Service handles preferences business logic
type Service struct {
	repo   *Repository
	logger *zap.Logger
}

// NewService creates a new preferences service
func NewService(repo *Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// GetEffectiveTheme retrieves the effective theme for a user session
// This merges the user's base theme with profile-specific overrides
func (s *Service) GetEffectiveTheme(ctx context.Context, userID string, profileID *string) (*domain.EffectiveTheme, error) {
	// Get user base theme
	baseThemeJSON, err := s.repo.GetBaseTheme(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get base theme", zap.String("user_id", userID), zap.Error(err))
		return domain.DefaultEffectiveTheme(), nil
	}

	// If no base theme exists, return defaults
	if baseThemeJSON == nil {
		return domain.DefaultEffectiveTheme(), nil
	}

	var baseTheme domain.BaseThemeSettings
	if err := json.Unmarshal(baseThemeJSON, &baseTheme); err != nil {
		s.logger.Error("failed to unmarshal base theme", zap.String("user_id", userID), zap.Error(err))
		return domain.DefaultEffectiveTheme(), nil
	}

	effective := &domain.EffectiveTheme{
		BaseThemeSettings: baseTheme,
	}

	// Apply profile overrides if profile is specified
	if profileID != nil && *profileID != "" {
		overridesJSON, err := s.repo.GetProfileThemeOverrides(ctx, *profileID)
		if err != nil {
			s.logger.Warn("failed to get profile theme overrides",
				zap.String("profile_id", *profileID),
				zap.Error(err))
			// Continue with base theme only
		} else if overridesJSON != nil {
			var overrides domain.ProfileThemeOverrides
			if err := json.Unmarshal(overridesJSON, &overrides); err != nil {
				s.logger.Warn("failed to unmarshal profile theme overrides",
					zap.String("profile_id", *profileID),
					zap.Error(err))
				// Continue with base theme only
			} else {
				effective.ProfileThemeOverrides = overrides
			}
		}
	}

	return effective, nil
}

// GetBaseTheme retrieves the user's base theme settings
func (s *Service) GetBaseTheme(ctx context.Context, userID string) (*domain.BaseThemeSettings, error) {
	baseThemeJSON, err := s.repo.GetBaseTheme(ctx, userID)
	if err != nil {
		return nil, err
	}

	if baseThemeJSON == nil {
		return domain.DefaultBaseTheme(), nil
	}

	var baseTheme domain.BaseThemeSettings
	if err := json.Unmarshal(baseThemeJSON, &baseTheme); err != nil {
		return nil, err
	}

	return &baseTheme, nil
}

// UpdateBaseTheme updates the user's base theme settings
func (s *Service) UpdateBaseTheme(ctx context.Context, userID string, baseTheme *domain.BaseThemeSettings) error {
	s.logger.Debug("updating base theme",
		zap.String("user_id", userID),
		zap.String("mode", baseTheme.Mode))

	themeJSON, err := json.Marshal(baseTheme)
	if err != nil {
		return err
	}

	updatedAt := sql.NullTime{Time: time.Now().UTC(), Valid: true}
	err = s.repo.UpsertBaseTheme(ctx, userID, themeJSON, updatedAt)
	if err != nil {
		s.logger.Error("failed to update base theme",
			zap.String("user_id", userID),
			zap.Error(err))
		return err
	}

	s.logger.Info("base theme updated successfully",
		zap.String("user_id", userID),
		zap.String("mode", baseTheme.Mode))

	return nil
}

// GetProfileThemeOverrides retrieves profile-specific theme overrides
func (s *Service) GetProfileThemeOverrides(ctx context.Context, profileID string) (*domain.ProfileThemeOverrides, error) {
	overridesJSON, err := s.repo.GetProfileThemeOverrides(ctx, profileID)
	if err != nil {
		return nil, err
	}

	if overridesJSON == nil {
		// No overrides, return empty
		return &domain.ProfileThemeOverrides{}, nil
	}

	var overrides domain.ProfileThemeOverrides
	if err := json.Unmarshal(overridesJSON, &overrides); err != nil {
		return nil, err
	}

	return &overrides, nil
}

// UpdateProfileThemeOverrides updates profile-specific theme overrides
func (s *Service) UpdateProfileThemeOverrides(ctx context.Context, profileID string, overrides *domain.ProfileThemeOverrides) error {
	s.logger.Debug("updating profile theme overrides",
		zap.String("profile_id", profileID))

	overridesJSON, err := json.Marshal(overrides)
	if err != nil {
		return err
	}

	updatedAt := sql.NullTime{Time: time.Now().UTC(), Valid: true}
	err = s.repo.UpsertProfileThemeOverrides(ctx, profileID, overridesJSON, updatedAt)
	if err != nil {
		s.logger.Error("failed to update profile theme overrides",
			zap.String("profile_id", profileID),
			zap.Error(err))
		return err
	}

	s.logger.Info("profile theme overrides updated successfully",
		zap.String("profile_id", profileID))

	return nil
}

// CreateDefaultUserPreferences creates default preferences for a new user
// This is called during user registration
func (s *Service) CreateDefaultUserPreferences(ctx context.Context, userID string) error {
	s.logger.Debug("creating default user preferences", zap.String("user_id", userID))

	defaults := domain.DefaultBaseTheme()
	themeJSON, err := json.Marshal(defaults)
	if err != nil {
		return err
	}

	prefs := &domain.UserPreferences{
		UserID:    userID,
		BaseTheme: themeJSON,
		UpdatedAt: time.Now().UTC(),
	}

	err = s.repo.UpsertUserPreferences(ctx, prefs)
	if err != nil {
		s.logger.Error("failed to create default user preferences",
			zap.String("user_id", userID),
			zap.Error(err))
		return err
	}

	s.logger.Info("default user preferences created successfully",
		zap.String("user_id", userID))

	return nil
}
