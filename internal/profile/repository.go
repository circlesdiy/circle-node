package profile

import (
	"context"
	"database/sql"

	"circles.diy/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles all database operations for profiles
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new profile repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreateProfile creates a new profile in the database
func (r *Repository) CreateProfile(ctx context.Context, profile *domain.Profile) error {
	query := `
		INSERT INTO profiles (id, user_id, handle, name, display_name, bio, avatar_url, banner_url, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.Exec(ctx, query,
		profile.ID, profile.UserID, profile.Handle, profile.Name, profile.DisplayName,
		profile.Bio, profile.AvatarURL, profile.BannerURL, profile.IsActive, profile.CreatedAt, profile.UpdatedAt)

	return err
}

// CreateProfileSettings creates profile settings
func (r *Repository) CreateProfileSettings(ctx context.Context, settings *domain.ProfileSettings) error {
	query := `
		INSERT INTO profile_settings (profile_id, is_public, location, website, interests, social_links, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.Exec(ctx, query,
		settings.ProfileID, settings.IsPublic, settings.Location, settings.Website,
		settings.Interests, settings.SocialLinks, settings.UpdatedAt)

	return err
}

// CreateProfileWithSettings creates a profile and its settings atomically in a transaction
func (r *Repository) CreateProfileWithSettings(ctx context.Context, profile *domain.Profile, settings *domain.ProfileSettings) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Create profile
	profileQuery := `
		INSERT INTO profiles (id, user_id, handle, name, display_name, bio, avatar_url, banner_url, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = tx.Exec(ctx, profileQuery,
		profile.ID, profile.UserID, profile.Handle, profile.Name, profile.DisplayName,
		profile.Bio, profile.AvatarURL, profile.BannerURL, profile.IsActive, profile.CreatedAt, profile.UpdatedAt)

	if err != nil {
		return err
	}

	// Create profile settings
	settingsQuery := `
		INSERT INTO profile_settings (profile_id, is_public, location, website, interests, social_links, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = tx.Exec(ctx, settingsQuery,
		settings.ProfileID, settings.IsPublic, settings.Location, settings.Website,
		settings.Interests, settings.SocialLinks, settings.UpdatedAt)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// GetProfileByID retrieves a profile by its ID
func (r *Repository) GetProfileByID(ctx context.Context, id string) (*domain.Profile, error) {
	query := `
		SELECT id, user_id, handle, name, display_name, bio, avatar_url, banner_url, is_active, created_at, updated_at, deleted_at
		FROM profiles WHERE id = $1 AND deleted_at IS NULL`

	var profile domain.Profile
	var name, displayName, bio, avatarURL, bannerURL sql.NullString
	var deletedAt sql.NullTime

	err := r.db.QueryRow(ctx, query, id).Scan(
		&profile.ID, &profile.UserID, &profile.Handle, &name, &displayName,
		&bio, &avatarURL, &bannerURL, &profile.IsActive, &profile.CreatedAt, &profile.UpdatedAt,
		&deletedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Assign nullable fields
	if name.Valid {
		profile.Name = name.String
	}
	if displayName.Valid {
		profile.DisplayName = displayName.String
	}
	if bio.Valid {
		profile.Bio = bio.String
	}
	if avatarURL.Valid {
		profile.AvatarURL = avatarURL.String
	}
	if bannerURL.Valid {
		profile.BannerURL = bannerURL.String
	}
	if deletedAt.Valid {
		profile.DeletedAt = &deletedAt.Time
	}

	return &profile, nil
}

// GetProfileByHandle retrieves a profile by its handle
func (r *Repository) GetProfileByHandle(ctx context.Context, handle string) (*domain.Profile, error) {
	query := `
		SELECT id, user_id, handle, name, display_name, bio, avatar_url, banner_url, is_active, created_at, updated_at, deleted_at
		FROM profiles WHERE handle = $1 AND deleted_at IS NULL`

	var profile domain.Profile
	var name, displayName, bio, avatarURL, bannerURL sql.NullString
	var deletedAt sql.NullTime

	err := r.db.QueryRow(ctx, query, handle).Scan(
		&profile.ID, &profile.UserID, &profile.Handle, &name, &displayName,
		&bio, &avatarURL, &bannerURL, &profile.IsActive, &profile.CreatedAt, &profile.UpdatedAt,
		&deletedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Assign nullable fields
	if name.Valid {
		profile.Name = name.String
	}
	if displayName.Valid {
		profile.DisplayName = displayName.String
	}
	if bio.Valid {
		profile.Bio = bio.String
	}
	if avatarURL.Valid {
		profile.AvatarURL = avatarURL.String
	}
	if bannerURL.Valid {
		profile.BannerURL = bannerURL.String
	}
	if deletedAt.Valid {
		profile.DeletedAt = &deletedAt.Time
	}

	return &profile, nil
}

// GetProfilesByUserID retrieves all profiles for a user
func (r *Repository) GetProfilesByUserID(ctx context.Context, userID string) ([]domain.Profile, error) {
	query := `
		SELECT id, user_id, handle, name, display_name, bio, avatar_url, banner_url, is_active, created_at, updated_at, deleted_at
		FROM profiles WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []domain.Profile
	for rows.Next() {
		var profile domain.Profile
		var name, displayName, bio, avatarURL, bannerURL sql.NullString
		var deletedAt sql.NullTime

		err := rows.Scan(
			&profile.ID, &profile.UserID, &profile.Handle, &name, &displayName,
			&bio, &avatarURL, &bannerURL, &profile.IsActive, &profile.CreatedAt, &profile.UpdatedAt,
			&deletedAt)

		if err != nil {
			return nil, err
		}

		// Assign nullable fields
		if name.Valid {
			profile.Name = name.String
		}
		if displayName.Valid {
			profile.DisplayName = displayName.String
		}
		if bio.Valid {
			profile.Bio = bio.String
		}
		if avatarURL.Valid {
			profile.AvatarURL = avatarURL.String
		}
		if bannerURL.Valid {
			profile.BannerURL = bannerURL.String
		}
		if deletedAt.Valid {
			profile.DeletedAt = &deletedAt.Time
		}

		profiles = append(profiles, profile)
	}

	return profiles, rows.Err()
}

// GetActiveProfileByUserID gets the user's active profile ID
// Returns nil if no active profile exists
func (r *Repository) GetActiveProfileByUserID(ctx context.Context, userID string) (*string, error) {
	query := `
		SELECT id FROM profiles
		WHERE user_id = $1 AND is_active = true AND deleted_at IS NULL
		ORDER BY created_at ASC
		LIMIT 1`

	var profileID string
	err := r.db.QueryRow(ctx, query, userID).Scan(&profileID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &profileID, nil
}

// UpdateProfile updates an existing profile
func (r *Repository) UpdateProfile(ctx context.Context, profile *domain.Profile) error {
	query := `
		UPDATE profiles
		SET handle = $2, name = $3, display_name = $4, bio = $5, avatar_url = $6, is_active = $7, updated_at = $8
		WHERE id = $1`

	_, err := r.db.Exec(ctx, query,
		profile.ID, profile.Handle, profile.Name, profile.DisplayName,
		profile.Bio, profile.AvatarURL, profile.IsActive, profile.UpdatedAt)

	return err
}

// DeleteProfile soft deletes a profile by setting deleted_at timestamp
func (r *Repository) DeleteProfile(ctx context.Context, id string) error {
	query := `
		UPDATE profiles
		SET deleted_at = NOW()
		WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id)
	return err
}

// HandleExists checks if a handle is already taken
func (r *Repository) HandleExists(ctx context.Context, handle string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM profiles WHERE handle = $1)`

	var exists bool
	err := r.db.QueryRow(ctx, query, handle).Scan(&exists)
	return exists, err
}

// GetProfileSettings retrieves profile settings by profile ID
func (r *Repository) GetProfileSettings(ctx context.Context, profileID string) (*domain.ProfileSettings, error) {
	query := `
		SELECT profile_id, is_public, location, website, interests, social_links, updated_at
		FROM profile_settings WHERE profile_id = $1`

	var settings domain.ProfileSettings
	var location, website sql.NullString
	var interests, socialLinks []byte

	err := r.db.QueryRow(ctx, query, profileID).Scan(
		&settings.ProfileID, &settings.IsPublic, &location, &website,
		&interests, &socialLinks, &settings.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Assign nullable fields
	if location.Valid {
		settings.Location = location.String
	}
	if website.Valid {
		settings.Website = website.String
	}
	if len(interests) > 0 {
		settings.Interests = interests
	}
	if len(socialLinks) > 0 {
		settings.SocialLinks = socialLinks
	}

	return &settings, nil
}

// UpdateProfileSettings updates existing profile settings
func (r *Repository) UpdateProfileSettings(ctx context.Context, settings *domain.ProfileSettings) error {
	query := `
		UPDATE profile_settings
		SET is_public = $2, location = $3, website = $4, interests = $5, social_links = $6, updated_at = $7
		WHERE profile_id = $1`

	_, err := r.db.Exec(ctx, query,
		settings.ProfileID, settings.IsPublic, settings.Location, settings.Website,
		settings.Interests, settings.SocialLinks, settings.UpdatedAt)

	return err
}

// generateID generates a new UUID
func generateID() string {
	return uuid.New().String()
}
