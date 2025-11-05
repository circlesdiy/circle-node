package auth

import (
	"context"
	"time"

	"circles.diy/internal/domain"
)

// Repository defines the interface for authentication-related data operations
type Repository interface {
	// User operations
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	UsernameExists(ctx context.Context, username string) (bool, error)
	EmailExists(ctx context.Context, email string) (bool, error)

	// Credential operations
	CreateCredential(ctx context.Context, cred *WebAuthnCredential) error
	GetCredentialsByUserID(ctx context.Context, userID string) ([]WebAuthnCredential, error)
	GetCredentialByCredentialID(ctx context.Context, credentialID string) (*WebAuthnCredential, error)
	UpdateCredentialSignCount(ctx context.Context, credentialID string, signCount int) error
	UpdateCredentialLastUsed(ctx context.Context, credentialID string, lastUsed time.Time) error
	RevokeCredential(ctx context.Context, credentialID string) error

	// Session operations
	CreateSession(ctx context.Context, session *Session) error
	GetSessionByToken(ctx context.Context, token string) (*Session, error)
	GetSessionsByUserID(ctx context.Context, userID string) ([]Session, error)
	UpdateSessionActivity(ctx context.Context, sessionID string, lastActivity time.Time) error
	UpdateSessionProfile(ctx context.Context, sessionID string, profileID string) error
	RevokeSession(ctx context.Context, sessionID string) error
	RevokeAllUserSessions(ctx context.Context, userID string) error

	// Device operations
	CreateDevice(ctx context.Context, device *Device) error
	GetDeviceByID(ctx context.Context, id string) (*Device, error)
	GetDeviceByFingerprint(ctx context.Context, userID, fingerprint string) (*Device, error)
	GetDevicesByUserID(ctx context.Context, userID string) ([]Device, error)
	UpdateDeviceLastSeen(ctx context.Context, deviceID string, lastSeen time.Time) error
	UpdateDeviceTrustStatus(ctx context.Context, deviceID string, isTrusted bool) error
	RevokeDevice(ctx context.Context, deviceID string) error

	// DeviceCredential operations
	LinkDeviceCredential(ctx context.Context, deviceCred *DeviceCredential) error
	GetDeviceCredentials(ctx context.Context, deviceID string) ([]DeviceCredential, error)
	UpdateDeviceCredentialLastUsed(ctx context.Context, id string, lastUsed time.Time) error

	// Challenge operations
	CreateChallenge(ctx context.Context, challenge *AuthenticationChallenge) error
	GetChallengeByChallenge(ctx context.Context, challengeStr string) (*AuthenticationChallenge, error)
	CompleteChallenge(ctx context.Context, challengeID string, completedAt time.Time) error
	CleanupExpiredChallenges(ctx context.Context) error

	// Recovery operations
	CreateRecoveryMethod(ctx context.Context, method *RecoveryMethod) error
	GetRecoveryMethodsByUserID(ctx context.Context, userID string) ([]RecoveryMethod, error)
	GetRecoveryMethodByID(ctx context.Context, id string) (*RecoveryMethod, error)
	UseRecoveryMethod(ctx context.Context, methodID string) error
	RevokeRecoveryMethod(ctx context.Context, methodID string) error

	// DeviceVerification operations
	CreateDeviceVerification(ctx context.Context, verification *DeviceVerification) error
	GetDeviceVerificationByCode(ctx context.Context, deviceID, code string) (*DeviceVerification, error)
	CompleteDeviceVerification(ctx context.Context, verificationID string, verifiedAt time.Time) error
	CleanupExpiredVerifications(ctx context.Context) error

	// Cleanup operations
	CleanupExpiredSessions(ctx context.Context) error
}
