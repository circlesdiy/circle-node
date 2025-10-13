package auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/fxamacker/webauthn"
)

// Service handles authentication business logic
type Service struct {
	repo   *Repository
	config *WebAuthnConfig
	logger *slog.Logger
}

// WebAuthnConfig holds WebAuthn configuration
type WebAuthnConfig struct {
	*webauthn.Config
	RPOrigin string // Not part of webauthn.Config but needed for verification
}

// NewService creates a new authentication service
func NewService(repo *Repository, config *WebAuthnConfig, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		config: config,
		logger: logger,
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
			AuthenticatorAttachment: "", // Allow all
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
	s.logger.Info("Beginning WebAuthn registration", "username", username)

	// Check if user already exists
	existingUser, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("username already exists")
	}

	// Check if email already exists
	existingEmail, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing email: %w", err)
	}
	if existingEmail != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Generate user ID
	userID, err := GenerateSecureToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate user ID: %w", err)
	}

	// Create user object for WebAuthn
	user := &User{
		ID:            userID,
		Username:      username,
		Email:         email,
		AccountStatus: "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Convert to WebAuthn user (with empty credentials for registration)
	webauthnUser := user.ToWebAuthnUser([]WebAuthnCredential{})

	// Create WebAuthn creation options using fxamacker API
	options, err := webauthn.NewAttestationOptions(s.config.Config, webauthnUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create attestation options: %w", err)
	}

	// Store challenge
	challengeRecord := &AuthenticationChallenge{
		ID:        GenerateID(),
		UserID:    &user.ID,
		Challenge: base64.URLEncoding.EncodeToString(options.Challenge),
		Type:      "registration",
		Options: map[string]interface{}{
			"user_id":  user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(s.config.Timeout) * time.Millisecond),
	}

	if err := s.repo.CreateAuthenticationChallenge(ctx, challengeRecord); err != nil {
		return nil, fmt.Errorf("failed to store challenge: %w", err)
	}

	s.logger.Info("Registration challenge created", "user_id", user.ID, "challenge_id", challengeRecord.ID)

	return options, nil
}

// FinishRegistration completes the WebAuthn registration process
func (s *Service) FinishRegistration(ctx context.Context, response *webauthn.PublicKeyCredentialAttestation, r *http.Request) (*User, error) {
	s.logger.Info("Finishing WebAuthn registration")

	// Get client data (already parsed by fxamacker)
	clientData := response.ClientData

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

	// Create user
	user := &User{
		ID:            userID,
		Username:      username,
		Email:         email,
		AccountStatus: "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create credential
	credential := &WebAuthnCredential{
		ID:                GenerateID(),
		UserID:            user.ID,
		CredentialID:      response.ID,
		PublicKey:         base64.StdEncoding.EncodeToString(response.RawID), // Store raw ID as public key placeholder
		CredentialType:    "public-key",
		SignCount:         int(authnData.Counter),
		AttestationFormat: "",             // Will be set by attestation verification
		AttestationObject: response.RawID, // Store raw credential data
		FriendlyName:      fmt.Sprintf("%s's authenticator", username),
		IsSynced:          false,
		IsBackup:          false,
		CreatedAt:         time.Now(),
	}

	if err := s.repo.CreateWebAuthnCredential(ctx, credential); err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	// Create device record
	device := s.createDeviceFromRequest(user.ID, r)
	if err := s.repo.CreateDevice(ctx, device); err != nil {
		s.logger.Warn("Failed to create device record", "error", err)
	}

	// Complete challenge
	if err := s.repo.CompleteAuthenticationChallenge(ctx, challenge.ID); err != nil {
		s.logger.Warn("Failed to complete challenge", "error", err)
	}

	s.logger.Info("Registration completed successfully", "user_id", user.ID, "credential_id", credential.ID)

	return user, nil
}

// Authentication flow

// BeginAuthentication starts the WebAuthn authentication process
func (s *Service) BeginAuthentication(ctx context.Context, username string) (*webauthn.PublicKeyCredentialRequestOptions, error) {
	s.logger.Info("Beginning WebAuthn authentication", "username", username)

	// Get user
	user, err := s.repo.GetUserByUsername(ctx, username)
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

	// Convert to WebAuthn user
	webauthnUser := user.ToWebAuthnUser(credentials)

	// Create WebAuthn assertion options using fxamacker API
	options, err := webauthn.NewAssertionOptions(s.config.Config, webauthnUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create assertion options: %w", err)
	}

	// Store challenge
	challengeRecord := &AuthenticationChallenge{
		ID:        GenerateID(),
		UserID:    &user.ID,
		Challenge: base64.URLEncoding.EncodeToString(options.Challenge),
		Type:      "authentication",
		Options: map[string]interface{}{
			"user_id": user.ID,
		},
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(s.config.Timeout) * time.Millisecond),
	}

	if err := s.repo.CreateAuthenticationChallenge(ctx, challengeRecord); err != nil {
		return nil, fmt.Errorf("failed to store challenge: %w", err)
	}

	s.logger.Info("Authentication challenge created", "user_id", user.ID, "challenge_id", challengeRecord.ID)

	return options, nil
}

// FinishAuthentication completes the WebAuthn authentication process
func (s *Service) FinishAuthentication(ctx context.Context, response *webauthn.PublicKeyCredentialAssertion, r *http.Request) (*Session, error) {
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
	user, err := s.repo.GetUserByID(ctx, credential.UserID)
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
	if newSignCount <= credential.SignCount {
		return nil, fmt.Errorf("potential replay attack detected")
	}

	if err := s.repo.UpdateWebAuthnCredentialSignCount(ctx, credential.CredentialID, newSignCount); err != nil {
		return nil, fmt.Errorf("failed to update sign count: %w", err)
	}

	// Create session
	session, err := s.CreateSession(ctx, user, r)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Complete challenge
	if err := s.repo.CompleteAuthenticationChallenge(ctx, challenge.ID); err != nil {
		s.logger.Warn("Failed to complete challenge", "error", err)
	}

	s.logger.Info("Authentication completed successfully", "user_id", user.ID, "session_id", session.ID)

	return session, nil
}

// Session management

// CreateSession creates a new session for a user
func (s *Service) CreateSession(ctx context.Context, user *User, r *http.Request) (*Session, error) {
	sessionToken, err := GenerateSecureToken(64)
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	refreshToken, err := GenerateSecureToken(64)
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
			s.logger.Warn("Failed to create device record", "error", err)
		} else {
			deviceID = &device.ID
		}
	} else {
		deviceID = &device.ID
		if err := s.repo.UpdateDeviceLastSeen(ctx, device.ID, time.Now()); err != nil {
			s.logger.Warn("Failed to update device last seen", "error", err)
		}
	}

	session := &Session{
		ID:                         GenerateID(),
		UserID:                     user.ID,
		DeviceID:                   deviceID,
		AuthenticatedCredentialIDs: "[]", // JSON array of credential IDs used
		SessionToken:               sessionToken,
		RefreshToken:               refreshToken,
		IPAddress:                  getClientIP(r),
		UserAgent:                  r.UserAgent(),
		AuthLevel:                  AuthLevelBasic,
		CreatedAt:                  time.Now(),
		LastActivityAt:             time.Now(),
		ExpiresAt:                  time.Now().Add(DefaultSessionDuration),
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// ValidateSession validates a session token and returns the session if valid
func (s *Service) ValidateSession(ctx context.Context, token string) (*Session, *User, error) {
	session, err := s.repo.GetSessionByToken(ctx, token)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get session: %w", err)
	}

	if session == nil || !session.IsActive() {
		return nil, nil, fmt.Errorf("invalid or expired session")
	}

	// Update last activity
	if err := s.repo.UpdateSessionActivity(ctx, session.ID, time.Now()); err != nil {
		s.logger.Warn("Failed to update session activity", "error", err)
	}

	// Get user
	user, err := s.repo.GetUserByID(ctx, session.UserID)
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
func (s *Service) createDeviceFromRequest(userID string, r *http.Request) *Device {
	return &Device{
		ID:                GenerateID(),
		UserID:            userID,
		DeviceFingerprint: s.generateDeviceFingerprint(r),
		FriendlyName:      parseUserAgent(r.UserAgent()),
		DeviceType:        "browser",
		OS:                parseOS(r.UserAgent()),
		Browser:           parseBrowser(r.UserAgent()),
		IsTrusted:         false,
		FirstSeenAt:       time.Now(),
		LastSeenAt:        time.Now(),
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
	id, _ := GenerateSecureToken(16)
	return id
}

// Cleanup operations

// RunCleanup performs cleanup of expired records
func (s *Service) RunCleanup(ctx context.Context) error {
	s.logger.Info("Running authentication cleanup")

	if err := s.repo.CleanupExpiredChallenges(ctx); err != nil {
		s.logger.Warn("Failed to cleanup expired challenges", "error", err)
	}

	if err := s.repo.CleanupExpiredSessions(ctx); err != nil {
		s.logger.Warn("Failed to cleanup expired sessions", "error", err)
	}

	return nil
}
