package domain

import (
	"context"
	"time"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// User CRUD operations
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id string) error

	// User validation
	UsernameExists(ctx context.Context, username string) (bool, error)
	EmailExists(ctx context.Context, email string) (bool, error)

	// Theme operations
	UpdateUserTheme(ctx context.Context, userID, themeMode, themeRadius string) error
}

// ProfileRepository defines the interface for profile data operations
type ProfileRepository interface {
	// Profile CRUD operations
	CreateProfile(ctx context.Context, profile *Profile) error
	GetProfileByID(ctx context.Context, id string) (*Profile, error)
	GetProfileByHandle(ctx context.Context, handle string) (*Profile, error)
	GetProfilesByUserID(ctx context.Context, userID string) ([]Profile, error)
	UpdateProfile(ctx context.Context, profile *Profile) error
	DeleteProfile(ctx context.Context, id string) error

	// Profile validation
	HandleExists(ctx context.Context, handle string) (bool, error)

	// Profile settings
	CreateProfileSettings(ctx context.Context, settings *ProfileSettings) error
	GetProfileSettings(ctx context.Context, profileID string) (*ProfileSettings, error)
	UpdateProfileSettings(ctx context.Context, settings *ProfileSettings) error

	// Atomic operations
	CreateProfileWithSettings(ctx context.Context, profile *Profile, settings *ProfileSettings) error

	// Active profile management
	GetActiveProfileByUserID(ctx context.Context, userID string) (*string, error)
}

// UserModerationRepository defines the interface for user moderation operations
type UserModerationRepository interface {
	CreateUserModeration(ctx context.Context, moderation *UserModeration) error
	GetUserModeration(ctx context.Context, userID string) (*UserModeration, error)
	UpdateUserModeration(ctx context.Context, moderation *UserModeration) error
	BanUser(ctx context.Context, userID string, until time.Time, reason string) error
	UnbanUser(ctx context.Context, userID string) error
	IncrementWarnings(ctx context.Context, userID string) error
	IncrementViolations(ctx context.Context, userID string) error
}
