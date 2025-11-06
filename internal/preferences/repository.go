package preferences

import (
	"context"
	"database/sql"
	"encoding/json"

	"circles.diy/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles all database operations for preferences
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new preferences repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// GetUserPreferences retrieves user preferences by user ID
// Returns nil if preferences don't exist (user should get defaults)
func (r *Repository) GetUserPreferences(ctx context.Context, userID string) (*domain.UserPreferences, error) {
	query := `
		SELECT user_id, base_theme, updated_at
		FROM user_preferences WHERE user_id = $1`

	var prefs domain.UserPreferences
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&prefs.UserID, &prefs.BaseTheme, &prefs.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &prefs, nil
}

// UpsertUserPreferences creates or updates user preferences
func (r *Repository) UpsertUserPreferences(ctx context.Context, prefs *domain.UserPreferences) error {
	query := `
		INSERT INTO user_preferences (user_id, base_theme, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE
		SET base_theme = $2, updated_at = $3`

	_, err := r.db.Exec(ctx, query, prefs.UserID, prefs.BaseTheme, prefs.UpdatedAt)
	return err
}

// GetProfilePreferences retrieves profile preferences by profile ID
// Returns nil if preferences don't exist (profile should inherit user base theme)
func (r *Repository) GetProfilePreferences(ctx context.Context, profileID string) (*domain.ProfilePreferences, error) {
	query := `
		SELECT profile_id, theme_overrides, updated_at
		FROM profile_preferences WHERE profile_id = $1`

	var prefs domain.ProfilePreferences
	var themeOverrides sql.NullString

	err := r.db.QueryRow(ctx, query, profileID).Scan(
		&prefs.ProfileID, &themeOverrides, &prefs.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if themeOverrides.Valid {
		prefs.ThemeOverrides = json.RawMessage(themeOverrides.String)
	}

	return &prefs, nil
}

// UpsertProfilePreferences creates or updates profile preferences
func (r *Repository) UpsertProfilePreferences(ctx context.Context, prefs *domain.ProfilePreferences) error {
	query := `
		INSERT INTO profile_preferences (profile_id, theme_overrides, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (profile_id) DO UPDATE
		SET theme_overrides = $2, updated_at = $3`

	_, err := r.db.Exec(ctx, query, prefs.ProfileID, prefs.ThemeOverrides, prefs.UpdatedAt)
	return err
}

// GetBaseTheme retrieves the user's base theme as raw JSON
// Returns nil if no preferences exist
func (r *Repository) GetBaseTheme(ctx context.Context, userID string) (json.RawMessage, error) {
	query := `SELECT base_theme FROM user_preferences WHERE user_id = $1`

	var baseTheme json.RawMessage
	err := r.db.QueryRow(ctx, query, userID).Scan(&baseTheme)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return baseTheme, nil
}

// UpsertBaseTheme creates or updates the user's base theme
func (r *Repository) UpsertBaseTheme(ctx context.Context, userID string, themeJSON json.RawMessage, updatedAt sql.NullTime) error {
	query := `
		INSERT INTO user_preferences (user_id, base_theme, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE
		SET base_theme = $2, updated_at = $3`

	_, err := r.db.Exec(ctx, query, userID, themeJSON, updatedAt)
	return err
}

// GetProfileThemeOverrides retrieves the profile's theme overrides as raw JSON
// Returns nil if no overrides exist (inherit user base theme)
func (r *Repository) GetProfileThemeOverrides(ctx context.Context, profileID string) (json.RawMessage, error) {
	query := `SELECT theme_overrides FROM profile_preferences WHERE profile_id = $1`

	var themeOverrides sql.NullString
	err := r.db.QueryRow(ctx, query, profileID).Scan(&themeOverrides)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if !themeOverrides.Valid {
		return nil, nil
	}

	return json.RawMessage(themeOverrides.String), nil
}

// UpsertProfileThemeOverrides creates or updates the profile's theme overrides
func (r *Repository) UpsertProfileThemeOverrides(ctx context.Context, profileID string, overridesJSON json.RawMessage, updatedAt sql.NullTime) error {
	query := `
		INSERT INTO profile_preferences (profile_id, theme_overrides, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (profile_id) DO UPDATE
		SET theme_overrides = $2, updated_at = $3`

	_, err := r.db.Exec(ctx, query, profileID, overridesJSON, updatedAt)
	return err
}
