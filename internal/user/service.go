package user

import (
	"context"

	"circles.diy/internal/domain"
	"go.uber.org/zap"
)

// Service handles user business logic
type Service struct {
	repo   domain.UserRepository
	logger *zap.Logger
}

// NewService creates a new user service
func NewService(repo domain.UserRepository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// Create creates a new user
func (s *Service) Create(ctx context.Context, user *domain.User) error {
	s.logger.Debug("creating user",
		zap.String("user_id", user.ID),
		zap.String("username", user.Username))

	err := s.repo.CreateUser(ctx, user)
	if err != nil {
		s.logger.Error("failed to create user",
			zap.String("user_id", user.ID),
			zap.Error(err))
		return err
	}

	s.logger.Info("user created successfully",
		zap.String("user_id", user.ID),
		zap.String("username", user.Username))

	return nil
}

// GetByID retrieves a user by their ID
func (s *Service) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.GetUserByID(ctx, id)
}

// GetByUsername retrieves a user by their username
func (s *Service) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return s.repo.GetUserByUsername(ctx, username)
}

// GetByEmail retrieves a user by their email
func (s *Service) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.repo.GetUserByEmail(ctx, email)
}

// CheckUsernameExists checks if a username is already taken
func (s *Service) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	return s.repo.UsernameExists(ctx, username)
}

// CheckEmailExists checks if an email is already registered
func (s *Service) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	return s.repo.EmailExists(ctx, email)
}

// Update updates an existing user
// Currently not actively used, but available for future functionality
func (s *Service) Update(ctx context.Context, user *domain.User) error {
	s.logger.Debug("updating user",
		zap.String("user_id", user.ID),
		zap.String("username", user.Username))

	err := s.repo.UpdateUser(ctx, user)
	if err != nil {
		s.logger.Error("failed to update user",
			zap.String("user_id", user.ID),
			zap.Error(err))
		return err
	}

	s.logger.Info("user updated successfully",
		zap.String("user_id", user.ID),
		zap.String("username", user.Username))

	return nil
}

// Delete soft deletes a user
// Currently not actively used, but available for future functionality
func (s *Service) Delete(ctx context.Context, id string) error {
	s.logger.Debug("deleting user", zap.String("user_id", id))

	err := s.repo.DeleteUser(ctx, id)
	if err != nil {
		s.logger.Error("failed to delete user",
			zap.String("user_id", id),
			zap.Error(err))
		return err
	}

	s.logger.Info("user deleted successfully", zap.String("user_id", id))
	return nil
}
