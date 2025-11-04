package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles all database operations for authentication
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new authentication repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// User operations

// CreateUser creates a new user in the database
func (r *Repository) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (id, username, email, account_status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.Exec(ctx, query,
		user.ID, user.Username, user.Email, user.AccountStatus,
		user.CreatedAt, user.UpdatedAt)

	return err
}

// GetUserByID retrieves a user by their ID
func (r *Repository) GetUserByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, username, email, account_status, created_at, updated_at, deleted_at
		FROM users WHERE id = $1 AND deleted_at IS NULL`

	var user User
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
func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT id, username, email, account_status, created_at, updated_at, deleted_at
		FROM users WHERE username = $1 AND deleted_at IS NULL`

	var user User
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
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, username, email, account_status, created_at, updated_at, deleted_at
		FROM users WHERE email = $1 AND deleted_at IS NULL`

	var user User
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

// Profile operations (temporary until profile domain is created)

// CreateDefaultProfile creates a default profile for a new user
func (r *Repository) CreateDefaultProfile(ctx context.Context, userID, username string) (string, error) {
	profileID := GenerateID()
	handle := username // Use username as initial handle
	now := time.Now()

	query := `
		INSERT INTO profiles (id, user_id, handle, name, display_name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	err := r.db.QueryRow(ctx, query,
		profileID, userID, handle, username, username, true, now, now).Scan(&profileID)

	if err != nil {
		return "", err
	}

	// Create default profile settings
	settingsQuery := `
		INSERT INTO profile_settings (profile_id, is_public, updated_at)
		VALUES ($1, $2, $3)`

	_, err = r.db.Exec(ctx, settingsQuery, profileID, true, now)
	if err != nil {
		return "", err
	}

	return profileID, nil
}

// GetUserActiveProfile gets the user's active profile ID
func (r *Repository) GetUserActiveProfile(ctx context.Context, userID string) (*string, error) {
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

// WebAuthn Credential operations

// CreateWebAuthnCredential stores a new WebAuthn credential
func (r *Repository) CreateWebAuthnCredential(ctx context.Context, cred *WebAuthnCredential) error {
	query := `
		INSERT INTO webauthn_credentials (
			id, user_id, credential_id, public_key, credential_type, sign_count,
			transports, aaguid, attestation_format, attestation_object,
			friendly_name, is_synced, is_backup, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := r.db.Exec(ctx, query,
		cred.ID, cred.UserID, cred.CredentialID, cred.PublicKey,
		cred.CredentialType, cred.SignCount, cred.Transports,
		cred.AAGUID, cred.AttestationFormat, cred.AttestationObject,
		cred.FriendlyName, cred.IsSynced, cred.IsBackup, cred.CreatedAt)

	return err
}

// GetWebAuthnCredentialsByUserID retrieves all credentials for a user
func (r *Repository) GetWebAuthnCredentialsByUserID(ctx context.Context, userID string) ([]WebAuthnCredential, error) {
	query := `
		SELECT id, user_id, credential_id, public_key, credential_type, sign_count,
			   transports, aaguid, attestation_format, attestation_object,
			   friendly_name, is_synced, is_backup, last_used_at, created_at, revoked_at
		FROM webauthn_credentials
		WHERE user_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credentials []WebAuthnCredential
	for rows.Next() {
		var cred WebAuthnCredential
		var lastUsedAt, revokedAt sql.NullTime

		err := rows.Scan(
			&cred.ID, &cred.UserID, &cred.CredentialID, &cred.PublicKey,
			&cred.CredentialType, &cred.SignCount, &cred.Transports,
			&cred.AAGUID, &cred.AttestationFormat, &cred.AttestationObject,
			&cred.FriendlyName, &cred.IsSynced, &cred.IsBackup,
			&lastUsedAt, &cred.CreatedAt, &revokedAt)

		if err != nil {
			return nil, err
		}

		if lastUsedAt.Valid {
			cred.LastUsedAt = &lastUsedAt.Time
		}
		if revokedAt.Valid {
			cred.RevokedAt = &revokedAt.Time
		}

		credentials = append(credentials, cred)
	}

	return credentials, rows.Err()
}

// GetWebAuthnCredentialByID retrieves a specific credential by ID
func (r *Repository) GetWebAuthnCredentialByID(ctx context.Context, credentialID string) (*WebAuthnCredential, error) {
	query := `
		SELECT id, user_id, credential_id, public_key, credential_type, sign_count,
			   transports, aaguid, attestation_format, attestation_object,
			   friendly_name, is_synced, is_backup, last_used_at, created_at, revoked_at
		FROM webauthn_credentials
		WHERE credential_id = $1 AND revoked_at IS NULL`

	var cred WebAuthnCredential
	var lastUsedAt, revokedAt sql.NullTime

	err := r.db.QueryRow(ctx, query, credentialID).Scan(
		&cred.ID, &cred.UserID, &cred.CredentialID, &cred.PublicKey,
		&cred.CredentialType, &cred.SignCount, &cred.Transports,
		&cred.AAGUID, &cred.AttestationFormat, &cred.AttestationObject,
		&cred.FriendlyName, &cred.IsSynced, &cred.IsBackup,
		&lastUsedAt, &cred.CreatedAt, &revokedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if lastUsedAt.Valid {
		cred.LastUsedAt = &lastUsedAt.Time
	}
	if revokedAt.Valid {
		cred.RevokedAt = &revokedAt.Time
	}

	return &cred, nil
}

// UpdateWebAuthnCredentialSignCount updates the sign count for a credential
func (r *Repository) UpdateWebAuthnCredentialSignCount(ctx context.Context, credentialID string, signCount int) error {
	query := `
		UPDATE webauthn_credentials
		SET sign_count = $1, last_used_at = $2
		WHERE credential_id = $3`

	_, err := r.db.Exec(ctx, query, signCount, time.Now(), credentialID)
	return err
}

// Device operations

// CreateDevice creates a new device record
func (r *Repository) CreateDevice(ctx context.Context, device *Device) error {
	query := `
		INSERT INTO devices (
			id, user_id, device_fingerprint, friendly_name, device_type,
			os, browser, is_trusted, first_seen_at, last_seen_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.Exec(ctx, query,
		device.ID, device.UserID, device.DeviceFingerprint,
		device.FriendlyName, device.DeviceType, device.OS,
		device.Browser, device.IsTrusted, device.FirstSeenAt, device.LastSeenAt)

	return err
}

// GetDeviceByFingerprint retrieves a device by its fingerprint
func (r *Repository) GetDeviceByFingerprint(ctx context.Context, userID, fingerprint string) (*Device, error) {
	query := `
		SELECT id, user_id, device_fingerprint, friendly_name, device_type,
			   os, browser, is_trusted, last_verified_at, first_seen_at, last_seen_at, revoked_at
		FROM devices
		WHERE user_id = $1 AND device_fingerprint = $2 AND revoked_at IS NULL`

	var device Device
	var lastVerifiedAt, revokedAt sql.NullTime

	err := r.db.QueryRow(ctx, query, userID, fingerprint).Scan(
		&device.ID, &device.UserID, &device.DeviceFingerprint,
		&device.FriendlyName, &device.DeviceType, &device.OS,
		&device.Browser, &device.IsTrusted, &lastVerifiedAt,
		&device.FirstSeenAt, &device.LastSeenAt, &revokedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if lastVerifiedAt.Valid {
		device.LastVerifiedAt = &lastVerifiedAt.Time
	}
	if revokedAt.Valid {
		device.RevokedAt = &revokedAt.Time
	}

	return &device, nil
}

// UpdateDeviceLastSeen updates the last seen timestamp for a device
func (r *Repository) UpdateDeviceLastSeen(ctx context.Context, deviceID string, lastSeen time.Time) error {
	query := `UPDATE devices SET last_seen_at = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, lastSeen, deviceID)
	return err
}

// Session operations

// CreateSession creates a new user session
func (r *Repository) CreateSession(ctx context.Context, session *Session) error {
	query := `
		INSERT INTO sessions (
			id, user_id, device_id, active_profile_id, authenticated_credential_ids,
			session_token, refresh_token, ip_address, user_agent, auth_level,
			created_at, last_activity_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.db.Exec(ctx, query,
		session.ID, session.UserID, session.DeviceID, session.ActiveProfileID,
		session.AuthenticatedCredentialIDs, session.SessionToken, session.RefreshToken,
		session.IPAddress, session.UserAgent, session.AuthLevel,
		session.CreatedAt, session.LastActivityAt, session.ExpiresAt)

	return err
}

// GetSessionByToken retrieves a session by its token
func (r *Repository) GetSessionByToken(ctx context.Context, token string) (*Session, error) {
	query := `
		SELECT id, user_id, device_id, active_profile_id, authenticated_credential_ids,
			   session_token, refresh_token, ip_address::text, user_agent, auth_level,
			   created_at, last_activity_at, expires_at, revoked_at
		FROM sessions
		WHERE session_token = $1 AND revoked_at IS NULL`

	var session Session
	var deviceID, activeProfileID sql.NullString
	var revokedAt sql.NullTime

	err := r.db.QueryRow(ctx, query, token).Scan(
		&session.ID, &session.UserID, &deviceID, &activeProfileID,
		&session.AuthenticatedCredentialIDs, &session.SessionToken,
		&session.RefreshToken, &session.IPAddress, &session.UserAgent,
		&session.AuthLevel, &session.CreatedAt, &session.LastActivityAt,
		&session.ExpiresAt, &revokedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if deviceID.Valid {
		session.DeviceID = &deviceID.String
	}
	if activeProfileID.Valid {
		session.ActiveProfileID = &activeProfileID.String
	}
	if revokedAt.Valid {
		session.RevokedAt = &revokedAt.Time
	}

	return &session, nil
}

// UpdateSessionActivity updates the last activity timestamp for a session
func (r *Repository) UpdateSessionActivity(ctx context.Context, sessionID string, lastActivity time.Time) error {
	query := `UPDATE sessions SET last_activity_at = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, lastActivity, sessionID)
	return err
}

// RevokeSession revokes a session
func (r *Repository) RevokeSession(ctx context.Context, sessionID string) error {
	query := `UPDATE sessions SET revoked_at = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, time.Now(), sessionID)
	return err
}

// Authentication Challenge operations

// CreateAuthenticationChallenge creates a new authentication challenge
func (r *Repository) CreateAuthenticationChallenge(ctx context.Context, challenge *AuthenticationChallenge) error {
	optionsJSON, err := json.Marshal(challenge.Options)
	if err != nil {
		return fmt.Errorf("failed to marshal challenge options: %w", err)
	}

	query := `
		INSERT INTO authentication_challenges (
			id, user_id, challenge, type, required_credential_id, options,
			created_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = r.db.Exec(ctx, query,
		challenge.ID, challenge.UserID, challenge.Challenge, challenge.Type,
		challenge.RequiredCredentialID, optionsJSON, challenge.CreatedAt, challenge.ExpiresAt)

	return err
}

// GetAuthenticationChallenge retrieves a challenge by its challenge string
func (r *Repository) GetAuthenticationChallenge(ctx context.Context, challengeStr string) (*AuthenticationChallenge, error) {
	query := `
		SELECT id, user_id, challenge, type, required_credential_id, options,
			   created_at, expires_at, completed_at
		FROM authentication_challenges
		WHERE challenge = $1`

	var challenge AuthenticationChallenge
	var userID, requiredCredentialID sql.NullString
	var completedAt sql.NullTime
	var optionsJSON []byte

	err := r.db.QueryRow(ctx, query, challengeStr).Scan(
		&challenge.ID, &userID, &challenge.Challenge, &challenge.Type,
		&requiredCredentialID, &optionsJSON, &challenge.CreatedAt,
		&challenge.ExpiresAt, &completedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if userID.Valid {
		challenge.UserID = &userID.String
	}
	if requiredCredentialID.Valid {
		challenge.RequiredCredentialID = &requiredCredentialID.String
	}
	if completedAt.Valid {
		challenge.CompletedAt = &completedAt.Time
	}

	if err := json.Unmarshal(optionsJSON, &challenge.Options); err != nil {
		return nil, fmt.Errorf("failed to unmarshal challenge options: %w", err)
	}

	return &challenge, nil
}

// CompleteAuthenticationChallenge marks a challenge as completed
func (r *Repository) CompleteAuthenticationChallenge(ctx context.Context, challengeID string) error {
	query := `UPDATE authentication_challenges SET completed_at = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, time.Now(), challengeID)
	return err
}

// Recovery Method operations

// CreateRecoveryMethod creates a new recovery method
func (r *Repository) CreateRecoveryMethod(ctx context.Context, method *RecoveryMethod) error {
	query := `
		INSERT INTO recovery_methods (
			id, user_id, type, encrypted_secret, salt, is_verified,
			remaining_uses, expires_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.Exec(ctx, query,
		method.ID, method.UserID, method.Type, method.EncryptedSecret,
		method.Salt, method.IsVerified, method.RemainingUses,
		method.ExpiresAt, method.CreatedAt)

	return err
}

// GetRecoveryMethodsByUserID retrieves all recovery methods for a user
func (r *Repository) GetRecoveryMethodsByUserID(ctx context.Context, userID string) ([]RecoveryMethod, error) {
	query := `
		SELECT id, user_id, type, encrypted_secret, salt, is_verified,
			   remaining_uses, verified_at, last_used_at, expires_at, created_at, revoked_at
		FROM recovery_methods
		WHERE user_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methods []RecoveryMethod
	for rows.Next() {
		var method RecoveryMethod
		var remainingUses sql.NullInt32
		var verifiedAt, lastUsedAt, expiresAt, revokedAt sql.NullTime

		err := rows.Scan(
			&method.ID, &method.UserID, &method.Type, &method.EncryptedSecret,
			&method.Salt, &method.IsVerified, &remainingUses, &verifiedAt,
			&lastUsedAt, &expiresAt, &method.CreatedAt, &revokedAt)

		if err != nil {
			return nil, err
		}

		if remainingUses.Valid {
			uses := int(remainingUses.Int32)
			method.RemainingUses = &uses
		}
		if verifiedAt.Valid {
			method.VerifiedAt = &verifiedAt.Time
		}
		if lastUsedAt.Valid {
			method.LastUsedAt = &lastUsedAt.Time
		}
		if expiresAt.Valid {
			method.ExpiresAt = &expiresAt.Time
		}
		if revokedAt.Valid {
			method.RevokedAt = &revokedAt.Time
		}

		methods = append(methods, method)
	}

	return methods, rows.Err()
}

// UseRecoveryMethod decrements the remaining uses for a recovery method
func (r *Repository) UseRecoveryMethod(ctx context.Context, methodID string) error {
	query := `
		UPDATE recovery_methods
		SET remaining_uses = CASE
			WHEN remaining_uses IS NOT NULL THEN remaining_uses - 1
			ELSE remaining_uses
		END,
		last_used_at = $1
		WHERE id = $2`

	_, err := r.db.Exec(ctx, query, time.Now(), methodID)
	return err
}

// Validation methods

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

// Cleanup methods for expired records

// CleanupExpiredChallenges removes expired authentication challenges
func (r *Repository) CleanupExpiredChallenges(ctx context.Context) error {
	query := `DELETE FROM authentication_challenges WHERE expires_at < $1`
	_, err := r.db.Exec(ctx, query, time.Now())
	return err
}

// CleanupExpiredSessions removes expired sessions
func (r *Repository) CleanupExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < $1 OR revoked_at IS NOT NULL`
	_, err := r.db.Exec(ctx, query, time.Now())
	return err
}