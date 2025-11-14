package auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"circles.diy/internal/domain"
	domainauth "circles.diy/internal/domain/auth"
	"circles.diy/internal/preferences"
	"circles.diy/internal/profile"
	"circles.diy/internal/user"
	"github.com/fxamacker/webauthn"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles authentication business logic
type Service struct {
	repo           *Repository
	userService    *user.Service
	profileService *profile.Service
	prefsService   *preferences.Service
	config         *WebAuthnConfig
	logger         *zap.Logger
}

// WebAuthnConfig holds WebAuthn configuration
type WebAuthnConfig struct {
	*webauthn.Config
	RPOrigin string // Not part of webauthn.Config but needed for verification
}

// NewService creates a new authentication service
func NewService(
	repo *Repository,
	userService *user.Service,
	profileService *profile.Service,
	prefsService *preferences.Service,
	config *WebAuthnConfig,
	logger *zap.Logger,
) *Service {
	return &Service{
		repo:           repo,
		userService:    userService,
		profileService: profileService,
		prefsService:   prefsService,
		config:         config,
		logger:         logger,
	}
}

// Default configuration
func DefaultWebAuthnConfig() *WebAuthnConfig {
	return &WebAuthnConfig{
		Config: &webauthn.Config{
			ChallengeLength:         32,
			Timeout:                 uint64(5 * time.Minute / time.Millisecond),
			RPID:                    "localhost",
			RPName:                  "Circles.DIY",
			RPIcon:                  "",
			AuthenticatorAttachment: webauthn.AuthenticatorCrossPlatform, // Prefer external authenticators (1Password, security keys, etc.)
			ResidentKey:             webauthn.ResidentKeyDiscouraged,
			UserVerification:        webauthn.UserVerificationPreferred,
			Attestation:             webauthn.AttestationNone,
			CredentialAlgs:          []int{webauthn.COSEAlgES256, webauthn.COSEAlgRS256},
		},
		RPOrigin: "http://localhost:8080",
	}
}

// Registration flow

// BeginRegistration starts the WebAuthn registration process
func (s *Service) BeginRegistration(ctx context.Context, username, email string) (*webauthn.PublicKeyCredentialCreationOptions, error) {
	s.logger.Info("Beginning WebAuthn registration", zap.String("username", username))

	// Check if user already exists
	existingUser, err := s.userService.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("username already exists")
	}

	// Check if email already exists
	existingEmail, err := s.userService.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing email: %w", err)
	}
	if existingEmail != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Generate user ID as UUID
	userID := uuid.New().String()

	// Create domain user
	domainUser := &domain.User{
		ID:            userID,
		Username:      username,
		Email:         email,
		AccountStatus: domain.AccountStatusActive,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	// Wrap in auth.User for WebAuthn functionality
	authUser := NewUser(domainUser)

	// Convert to WebAuthn user (with empty credentials for registration)
	webauthnUser := authUser.ToWebAuthnUser([]domainauth.WebAuthnCredential{})

	// Create WebAuthn creation options using fxamacker API
	options, err := webauthn.NewAttestationOptions(s.config.Config, webauthnUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create attestation options: %w", err)
	}

	// Store challenge (no UserID yet since user doesn't exist in DB during registration)
	// Use RawURLEncoding (without padding) to match what the WebAuthn client sends
	challengeStr := base64.RawURLEncoding.EncodeToString(options.Challenge)
	challengeRecord := &domainauth.AuthenticationChallenge{
		ID:        GenerateID(),
		UserID:    nil, // User doesn't exist yet - will be created after challenge is completed
		Challenge: challengeStr,
		Type:      "registration",
		Options: map[string]interface{}{
			"user_id":  domainUser.ID,
			"username": domainUser.Username,
			"email":    domainUser.Email,
		},
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(time.Duration(s.config.Timeout) * time.Millisecond),
	}

	if err := s.repo.CreateAuthenticationChallenge(ctx, challengeRecord); err != nil {
		return nil, fmt.Errorf("failed to store challenge: %w", err)
	}

	s.logger.Info("Registration challenge created", zap.String("user_id", domainUser.ID), zap.String("challenge_id", challengeRecord.ID), zap.String("challenge_str", challengeStr))

	return options, nil
}

// FinishRegistration completes the WebAuthn registration process
func (s *Service) FinishRegistration(ctx context.Context, response *webauthn.PublicKeyCredentialAttestation, r *http.Request) (*domain.User, error) {
	s.logger.Info("Finishing WebAuthn registration")

	// Get client data (already parsed by fxamacker)
	clientData := response.ClientData

	s.logger.Info("Looking up challenge", zap.String("challenge", clientData.Challenge))

	challenge, err := s.repo.GetAuthenticationChallenge(ctx, clientData.Challenge)
	if err != nil {
		return nil, fmt.Errorf("failed to get challenge: %w", err)
	}
	if challenge == nil {
		return nil, fmt.Errorf("challenge not found")
	}
	if challenge.IsExpired() {
		return nil, fmt.Errorf("challenge expired")
	}
	if challenge.Type != "registration" {
		return nil, fmt.Errorf("invalid challenge type")
	}

	// Extract user info from challenge
	userID := challenge.Options["user_id"].(string)
	username := challenge.Options["username"].(string)
	email := challenge.Options["email"].(string)

	// Get authenticator data (already parsed by fxamacker)
	authnData := response.AuthnData

	// Verify client data
	if clientData.Type != "webauthn.create" {
		return nil, fmt.Errorf("invalid client data type")
	}
	if clientData.Origin != s.config.RPOrigin {
		return nil, fmt.Errorf("invalid origin")
	}

	// Verify RP ID hash
	expectedRPIDHash := sha256.Sum256([]byte(s.config.RPID))
	if !bytes.Equal(authnData.RPIDHash[:], expectedRPIDHash[:]) {
		return nil, fmt.Errorf("invalid RP ID hash")
	}

	// Create domain user
	domainUser := &domain.User{
		ID:            userID,
		Username:      username,
		Email:         email,
		AccountStatus: domain.AccountStatusActive,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	if err := s.userService.Create(ctx, domainUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create default profile for the user
	defaultProfile, err := s.profileService.CreateDefault(ctx, domainUser.ID, domainUser.Username)
	if err != nil {
		s.logger.Error("Failed to create default profile", zap.Error(err), zap.String("user_id", domainUser.ID))
		return nil, fmt.Errorf("failed to create default profile: %w", err)
	}
	s.logger.Info("Created default profile", zap.String("profile_id", defaultProfile.ID), zap.String("user_id", domainUser.ID))

	// Create default user preferences
	if err := s.prefsService.CreateDefaultUserPreferences(ctx, domainUser.ID); err != nil {
		s.logger.Warn("Failed to create default user preferences", zap.Error(err), zap.String("user_id", domainUser.ID))
		// Don't fail registration if preferences creation fails
	} else {
		s.logger.Info("Created default user preferences", zap.String("user_id", domainUser.ID))
	}

	// Create credential
	credential := &domainauth.WebAuthnCredential{
		ID:                GenerateID(),
		UserID:            domainUser.ID,
		CredentialID:      response.ID,
		PublicKey:         base64.StdEncoding.EncodeToString(response.RawID), // Store raw ID as public key placeholder
		CredentialType:    "public-key",
		SignCount:         int(authnData.Counter),
		AttestationFormat: "",             // Will be set by attestation verification
		AttestationObject: response.RawID, // Store raw credential data
		FriendlyName:      fmt.Sprintf("%s's authenticator", username),
		IsSynced:          false,
		IsBackup:          false,
		CreatedAt:         time.Now().UTC(),
	}

	if err := s.repo.CreateWebAuthnCredential(ctx, credential); err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	// Create device record
	device := s.createDeviceFromRequest(domainUser.ID, r)
	if err := s.repo.CreateDevice(ctx, device); err != nil {
		s.logger.Warn("Failed to create device record", zap.Error(err))
	} else {
		// Link device to credential
		deviceCred := &domainauth.DeviceCredential{
			ID:           GenerateID(),
			DeviceID:     device.ID,
			CredentialID: credential.ID,
			LinkedAt:     time.Now().UTC(),
			LastUsedAt:   nil, // Will be set on first authentication use
		}
		if err := s.repo.LinkDeviceCredential(ctx, deviceCred); err != nil {
			s.logger.Warn("Failed to link device to credential", zap.Error(err), zap.String("device_id", device.ID), zap.String("credential_id", credential.ID))
		} else {
			s.logger.Info("Linked device to credential", zap.String("device_id", device.ID), zap.String("credential_id", credential.ID))
		}
	}

	// Complete challenge
	if err := s.repo.CompleteAuthenticationChallenge(ctx, challenge.ID); err != nil {
		s.logger.Warn("Failed to complete challenge", zap.Error(err))
	}

	s.logger.Info("Registration completed successfully", zap.String("user_id", domainUser.ID), zap.String("credential_id", credential.ID))

	return domainUser, nil
}

// Authentication flow

// BeginAuthentication starts the WebAuthn authentication process
func (s *Service) BeginAuthentication(ctx context.Context, username string) (*webauthn.PublicKeyCredentialRequestOptions, error) {
	s.logger.Info("Beginning WebAuthn authentication", zap.String("username", username))

	// Get user
	user, err := s.userService.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Get user's credentials
	credentials, err := s.repo.GetWebAuthnCredentialsByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}
	if len(credentials) == 0 {
		return nil, fmt.Errorf("no credentials found for user")
	}

	// Wrap domain user in auth.User for WebAuthn functionality
	authUser := NewUser(user)

	// Convert to WebAuthn user
	webauthnUser := authUser.ToWebAuthnUser(credentials)

	// Create WebAuthn assertion options using fxamacker API
	options, err := webauthn.NewAssertionOptions(s.config.Config, webauthnUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create assertion options: %w", err)
	}

	// Store challenge
	challengeRecord := &domainauth.AuthenticationChallenge{
		ID:        GenerateID(),
		UserID:    &user.ID,
		Challenge: base64.RawURLEncoding.EncodeToString(options.Challenge),
		Type:      "authentication",
		Options: map[string]interface{}{
			"user_id": user.ID,
		},
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(time.Duration(s.config.Timeout) * time.Millisecond),
	}

	if err := s.repo.CreateAuthenticationChallenge(ctx, challengeRecord); err != nil {
		return nil, fmt.Errorf("failed to store challenge: %w", err)
	}

	s.logger.Info("Authentication challenge created", zap.String("user_id", user.ID), zap.String("challenge_id", challengeRecord.ID))

	return options, nil
}

// FinishAuthentication completes the WebAuthn authentication process
func (s *Service) FinishAuthentication(ctx context.Context, response *webauthn.PublicKeyCredentialAssertion, r *http.Request) (*domainauth.Session, error) {
	s.logger.Info("Finishing WebAuthn authentication")

	// Get client data (already parsed by fxamacker)
	clientData := response.ClientData

	challenge, err := s.repo.GetAuthenticationChallenge(ctx, clientData.Challenge)
	if err != nil {
		return nil, fmt.Errorf("failed to get challenge: %w", err)
	}
	if challenge == nil || challenge.IsExpired() || challenge.Type != "authentication" {
		return nil, fmt.Errorf("invalid or expired challenge")
	}

	// Get credential
	credential, err := s.repo.GetWebAuthnCredentialByID(ctx, response.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}
	if credential == nil {
		return nil, fmt.Errorf("credential not found")
	}

	// Get user
	user, err := s.userService.GetByID(ctx, credential.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Verify the assertion (simplified - in production you'd do full verification)
	if clientData.Type != "webauthn.get" {
		return nil, fmt.Errorf("invalid client data type")
	}
	if clientData.Origin != s.config.RPOrigin {
		return nil, fmt.Errorf("invalid origin")
	}

	// Update sign count (replay attack protection)
	newSignCount := int(response.AuthnData.Counter)
	s.logger.Debug("Sign count check", zap.Int("new", newSignCount), zap.Int("stored", credential.SignCount))

	// Only enforce sign count increment if the stored count is > 0
	// Some authenticators (like 1Password) may not support sign counts or start at 0
	if credential.SignCount > 0 && newSignCount <= credential.SignCount {
		return nil, fmt.Errorf("potential replay attack detected: sign count did not increment (stored: %d, received: %d)", credential.SignCount, newSignCount)
	}

	if err := s.repo.UpdateWebAuthnCredentialSignCount(ctx, credential.CredentialID, newSignCount); err != nil {
		return nil, fmt.Errorf("failed to update sign count: %w", err)
	}

	// Create session
	session, err := s.CreateSession(ctx, user, r)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Link device to credential (if we have a device)
	if session.DeviceID != nil {
		// Try to link or update device-credential association
		deviceCred := &domainauth.DeviceCredential{
			ID:           GenerateID(),
			DeviceID:     *session.DeviceID,
			CredentialID: credential.ID,
			LinkedAt:     time.Now().UTC(),
			LastUsedAt:   &session.CreatedAt,
		}

		// ON CONFLICT DO NOTHING in the query will handle if link already exists
		if err := s.repo.LinkDeviceCredential(ctx, deviceCred); err != nil {
			s.logger.Debug("Device-credential link exists or failed", zap.Error(err))
		}

		// Update last used timestamp
		if err := s.repo.UpdateDeviceCredentialLastUsed(ctx, *session.DeviceID, credential.ID, time.Now().UTC()); err != nil {
			s.logger.Debug("Failed to update device-credential last used", zap.Error(err))
		}
	}

	// Complete challenge
	if err := s.repo.CompleteAuthenticationChallenge(ctx, challenge.ID); err != nil {
		s.logger.Warn("Failed to complete challenge", zap.Error(err))
	}

	return session, nil
}

// Session management

// CreateSession creates a new session for a user
func (s *Service) CreateSession(ctx context.Context, user *domain.User, r *http.Request) (*domainauth.Session, error) {
	sessionToken, err := domainauth.GenerateSecureToken(64)
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	refreshToken, err := domainauth.GenerateSecureToken(64)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Create or update device
	fingerprint := s.generateDeviceFingerprint(r)
	device, err := s.repo.GetDeviceByFingerprint(ctx, user.ID, fingerprint)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	var deviceID *string
	if device == nil {
		device = s.createDeviceFromRequest(user.ID, r)
		if err := s.repo.CreateDevice(ctx, device); err != nil {
			// If device creation fails (likely duplicate), try to fetch it again
			s.logger.Debug("Device creation failed, attempting to fetch existing device", zap.Error(err))
			device, fetchErr := s.repo.GetDeviceByFingerprint(ctx, user.ID, fingerprint)
			if fetchErr != nil {
				s.logger.Warn("Failed to create or fetch device record", zap.Error(err), zap.Error(fetchErr))
			} else if device != nil {
				deviceID = &device.ID
				// Update last seen since we just used it
				if err := s.repo.UpdateDeviceLastSeen(ctx, device.ID, time.Now().UTC()); err != nil {
					s.logger.Warn("Failed to update device last seen", zap.Error(err))
				}
			}
		} else {
			deviceID = &device.ID
		}
	} else {
		deviceID = &device.ID
		if err := s.repo.UpdateDeviceLastSeen(ctx, device.ID, time.Now().UTC()); err != nil {
			s.logger.Warn("Failed to update device last seen", zap.Error(err))
		}
	}

	// Get user's active profile
	activeProfileID, err := s.profileService.GetActiveForUser(ctx, user.ID)
	if err != nil {
		s.logger.Warn("Failed to get active profile", zap.Error(err), zap.String("user_id", user.ID))
	}

	session := &domainauth.Session{
		ID:                         GenerateID(),
		UserID:                     user.ID,
		DeviceID:                   deviceID,
		ActiveProfileID:            activeProfileID,
		AuthenticatedCredentialIDs: "[]", // JSON array of credential IDs used
		SessionToken:               sessionToken,
		RefreshToken:               refreshToken,
		IPAddress:                  getClientIP(r),
		UserAgent:                  r.UserAgent(),
		AuthLevel:                  int(domainauth.AuthLevelBasic),
		CreatedAt:                  time.Now().UTC(),
		LastActivityAt:             time.Now().UTC(),
		ExpiresAt:                  time.Now().UTC().Add(domainauth.DefaultSessionDuration),
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// ValidateSession validates a session token and returns the session if valid
func (s *Service) ValidateSession(ctx context.Context, token string) (*domainauth.Session, *domain.User, error) {
	session, err := s.repo.GetSessionByToken(ctx, token)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get session: %w", err)
	}

	if session == nil || !session.IsActive() {
		return nil, nil, fmt.Errorf("invalid or expired session")
	}

	// Update last activity
	if err := s.repo.UpdateSessionActivity(ctx, session.ID, time.Now().UTC()); err != nil {
		s.logger.Warn("Failed to update session activity", zap.Error(err))
	}

	// Get user
	user, err := s.userService.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user: %w", err)
	}

	return session, user, nil
}

// RevokeSession revokes a session
func (s *Service) RevokeSession(ctx context.Context, sessionID string) error {
	return s.repo.RevokeSession(ctx, sessionID)
}

// Helper functions

// generateDeviceFingerprint creates a device fingerprint from request
func (s *Service) generateDeviceFingerprint(r *http.Request) string {
	hash := sha256.New()
	hash.Write([]byte(r.UserAgent()))
	hash.Write([]byte(getClientIP(r)))
	// Could add more fingerprinting data here
	return base64.URLEncoding.EncodeToString(hash.Sum(nil))
}

// createDeviceFromRequest creates a device record from HTTP request
func (s *Service) createDeviceFromRequest(userID string, r *http.Request) *domainauth.Device {
	return &domainauth.Device{
		ID:                GenerateID(),
		UserID:            userID,
		DeviceFingerprint: s.generateDeviceFingerprint(r),
		FriendlyName:      parseUserAgent(r.UserAgent()),
		DeviceType:        "browser",
		OS:                parseOS(r.UserAgent()),
		Browser:           parseBrowser(r.UserAgent()),
		IsTrusted:         false,
		FirstSeenAt:       time.Now().UTC(),
		LastSeenAt:        time.Now().UTC(),
	}
}

// getClientIP extracts client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return strings.Split(xff, ",")[0]
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Use remote address
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// parseUserAgent extracts friendly name from user agent
func parseUserAgent(ua string) string {
	if strings.Contains(ua, "Chrome") {
		return "Chrome Browser"
	}
	if strings.Contains(ua, "Firefox") {
		return "Firefox Browser"
	}
	if strings.Contains(ua, "Safari") {
		return "Safari Browser"
	}
	return "Unknown Browser"
}

// parseOS extracts OS from user agent
func parseOS(ua string) string {
	if strings.Contains(ua, "Windows") {
		return "Windows"
	}
	if strings.Contains(ua, "Macintosh") {
		return "macOS"
	}
	if strings.Contains(ua, "Linux") {
		return "Linux"
	}
	if strings.Contains(ua, "Android") {
		return "Android"
	}
	if strings.Contains(ua, "iOS") {
		return "iOS"
	}
	return "Unknown"
}

// parseBrowser extracts browser from user agent
func parseBrowser(ua string) string {
	if strings.Contains(ua, "Chrome") {
		return "Chrome"
	}
	if strings.Contains(ua, "Firefox") {
		return "Firefox"
	}
	if strings.Contains(ua, "Safari") {
		return "Safari"
	}
	if strings.Contains(ua, "Edge") {
		return "Edge"
	}
	return "Unknown"
}

// GenerateID generates a new UUID-like ID
func GenerateID() string {
	// Generate a proper UUID v4
	return uuid.New().String()
}

// Delegation methods for handlers

// CheckUsernameExists checks if a username is already taken
// This delegates to the user service for cleaner separation of concerns
func (s *Service) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	return s.userService.CheckUsernameExists(ctx, username)
}

// CheckEmailExists checks if an email is already registered
// This delegates to the user service for cleaner separation of concerns
func (s *Service) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	return s.userService.CheckEmailExists(ctx, email)
}

// Cleanup operations

// RunCleanup performs cleanup of expired records
func (s *Service) RunCleanup(ctx context.Context) error {
	s.logger.Debug("running auth cleanup")

	if err := s.repo.CleanupExpiredChallenges(ctx); err != nil {
		s.logger.Warn("cleanup challenges failed", zap.Error(err))
	}

	if err := s.repo.CleanupExpiredSessions(ctx); err != nil {
		s.logger.Warn("cleanup sessions failed", zap.Error(err))
	}

	return nil
}
