package user

import (
	"context"
	"database/sql"

	"circles.diy/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles all database operations for users
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new user repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreateUser creates a new user in the database
func (r *Repository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, username, email, account_status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.Exec(ctx, query,
		user.ID, user.Username, user.Email, user.AccountStatus,
		user.CreatedAt, user.UpdatedAt)

	return err
}

// GetUserByID retrieves a user by their ID
func (r *Repository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, username, email, account_status, created_at, updated_at, deleted_at
		FROM users WHERE id = $1 AND deleted_at IS NULL`

	var user domain.User
	var deletedAt sql.NullTime

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.AccountStatus,
		&user.CreatedAt, &user.UpdatedAt, &deletedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if deletedAt.Valid {
		user.DeletedAt = &deletedAt.Time
	}

	return &user, nil
}

// GetUserByUsername retrieves a user by their username
func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, email, account_status, created_at, updated_at, deleted_at
		FROM users WHERE username = $1 AND deleted_at IS NULL`

	var user domain.User
	var deletedAt sql.NullTime

	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.AccountStatus,
		&user.CreatedAt, &user.UpdatedAt, &deletedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if deletedAt.Valid {
		user.DeletedAt = &deletedAt.Time
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by their email
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, username, email, account_status, created_at, updated_at, deleted_at
		FROM users WHERE email = $1 AND deleted_at IS NULL`

	var user domain.User
	var deletedAt sql.NullTime

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.AccountStatus,
		&user.CreatedAt, &user.UpdatedAt, &deletedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if deletedAt.Valid {
		user.DeletedAt = &deletedAt.Time
	}

	return &user, nil
}

// UpdateUser updates an existing user
// Currently not used, but implements the domain.UserRepository interface
func (r *Repository) UpdateUser(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET username = $2, email = $3, account_status = $4, updated_at = $5
		WHERE id = $1`

	_, err := r.db.Exec(ctx, query,
		user.ID, user.Username, user.Email, user.AccountStatus, user.UpdatedAt)

	return err
}

// DeleteUser soft deletes a user by setting deleted_at timestamp
// Currently not used, but implements the domain.UserRepository interface
func (r *Repository) DeleteUser(ctx context.Context, id string) error {
	query := `
		UPDATE users
		SET deleted_at = NOW()
		WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id)
	return err
}

// UsernameExists checks if a username is already taken
func (r *Repository) UsernameExists(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`

	var exists bool
	err := r.db.QueryRow(ctx, query, username).Scan(&exists)
	return exists, err
}

// EmailExists checks if an email is already registered
func (r *Repository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	return exists, err
}

// UpdateUserTheme updates user theme preferences
// Note: This method is deprecated - theme preferences are now in user_preferences table
// Kept for interface compatibility but should not be used
func (r *Repository) UpdateUserTheme(ctx context.Context, userID, themeMode, themeRadius string) error {
	// This is a no-op since theme is now in user_preferences table
	// Return nil to maintain compatibility
	return nil
}
