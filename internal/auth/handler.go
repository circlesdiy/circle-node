package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"

	httphelpers "circles.diy/internal/http"
	"github.com/fxamacker/webauthn"
)

// Handler handles HTTP requests for authentication
type Handler struct {
	service *Service
	logger  *slog.Logger
}

// NewHandler creates a new authentication handler
func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers all authentication routes with the mux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// WebAuthn registration routes
	mux.HandleFunc("POST /auth/register/begin", h.handleRegistrationBegin)
	mux.HandleFunc("POST /auth/register/finish", h.handleRegistrationFinish)

	// WebAuthn authentication routes
	mux.HandleFunc("POST /auth/login/begin", h.handleAuthenticationBegin)
	mux.HandleFunc("POST /auth/login/finish", h.handleAuthenticationFinish)

	// Session management
	mux.HandleFunc("POST /auth/logout", h.handleLogout)
	mux.HandleFunc("GET /auth/session", h.handleGetSession)

	// Device management
	mux.HandleFunc("GET /auth/devices", h.handleGetDevices)
}

// Registration handlers

// RegistrationBeginRequest represents the request to begin registration
type RegistrationBeginRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

// handleRegistrationBegin starts the WebAuthn registration process
func (h *Handler) handleRegistrationBegin(w http.ResponseWriter, r *http.Request) {
	var req RegistrationBeginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httphelpers.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Username == "" || req.Email == "" {
		httphelpers.JSONError(w, http.StatusBadRequest, "Username and email are required")
		return
	}

	// Begin registration
	options, err := h.service.BeginRegistration(r.Context(), req.Username, req.Email)
	if err != nil {
		h.logger.Error("Failed to begin registration", "error", err, "username", req.Username)
		httphelpers.JSONError(w, http.StatusInternalServerError, "Registration failed")
		return
	}

	h.logger.Info("Registration begun", "username", req.Username)
	httphelpers.JSONResponse(w, http.StatusOK, options)
}

// handleRegistrationFinish completes the WebAuthn registration process
func (h *Handler) handleRegistrationFinish(w http.ResponseWriter, r *http.Request) {
	var response webauthn.PublicKeyCredentialAttestation
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		httphelpers.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Complete registration
	user, err := h.service.FinishRegistration(r.Context(), &response, r)
	if err != nil {
		h.logger.Error("Failed to finish registration", "error", err)
		httphelpers.JSONError(w, http.StatusInternalServerError, "Registration failed")
		return
	}

	// Create session
	session, err := h.service.CreateSession(r.Context(), user, r)
	if err != nil {
		h.logger.Error("Failed to create session after registration", "error", err)
		httphelpers.JSONError(w, http.StatusInternalServerError, "Session creation failed")
		return
	}

	// Set session cookie
	h.setSessionCookie(w, session.SessionToken)

	h.logger.Info("Registration completed", "user_id", user.ID, "username", user.Username)

	// Return user info (without sensitive data)
	userResponse := map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
	}

	httphelpers.JSONResponse(w, http.StatusCreated, userResponse)
}

// Authentication handlers

// AuthenticationBeginRequest represents the request to begin authentication
type AuthenticationBeginRequest struct {
	Username string `json:"username"`
}

// handleAuthenticationBegin starts the WebAuthn authentication process
func (h *Handler) handleAuthenticationBegin(w http.ResponseWriter, r *http.Request) {
	var req AuthenticationBeginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httphelpers.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Username == "" {
		httphelpers.JSONError(w, http.StatusBadRequest, "Username is required")
		return
	}

	// Begin authentication
	options, err := h.service.BeginAuthentication(r.Context(), req.Username)
	if err != nil {
		h.logger.Error("Failed to begin authentication", "error", err, "username", req.Username)
		httphelpers.JSONError(w, http.StatusUnauthorized, "Authentication failed")
		return
	}

	h.logger.Info("Authentication begun", "username", req.Username)
	httphelpers.JSONResponse(w, http.StatusOK, options)
}

// handleAuthenticationFinish completes the WebAuthn authentication process
func (h *Handler) handleAuthenticationFinish(w http.ResponseWriter, r *http.Request) {
	var response webauthn.PublicKeyCredentialAssertion
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		httphelpers.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Complete authentication
	session, err := h.service.FinishAuthentication(r.Context(), &response, r)
	if err != nil {
		h.logger.Error("Failed to finish authentication", "error", err)
		httphelpers.JSONError(w, http.StatusUnauthorized, "Authentication failed")
		return
	}

	// Set session cookie
	h.setSessionCookie(w, session.SessionToken)

	h.logger.Info("Authentication completed", "user_id", session.UserID, "session_id", session.ID)

	// Return session info (without sensitive tokens)
	sessionResponse := map[string]interface{}{
		"user_id":    session.UserID,
		"auth_level": session.AuthLevel,
		"expires_at": session.ExpiresAt,
	}

	httphelpers.JSONResponse(w, http.StatusOK, sessionResponse)
}

// Session management handlers

// handleLogout logs out the current user
func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Get session token from cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		httphelpers.JSONError(w, http.StatusBadRequest, "No active session")
		return
	}

	// Validate and get session
	session, _, err := h.service.ValidateSession(r.Context(), cookie.Value)
	if err != nil {
		httphelpers.JSONError(w, http.StatusUnauthorized, "Invalid session")
		return
	}

	// Revoke session
	if err := h.service.RevokeSession(r.Context(), session.ID); err != nil {
		h.logger.Error("Failed to revoke session", "error", err, "session_id", session.ID)
		httphelpers.JSONError(w, http.StatusInternalServerError, "Logout failed")
		return
	}

	// Clear session cookie
	h.clearSessionCookie(w)

	h.logger.Info("User logged out", "user_id", session.UserID, "session_id", session.ID)
	httphelpers.JSONResponse(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// handleGetSession returns current session info
func (h *Handler) handleGetSession(w http.ResponseWriter, r *http.Request) {
	// Get session token from cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		httphelpers.JSONError(w, http.StatusUnauthorized, "No active session")
		return
	}

	// Validate and get session
	session, user, err := h.service.ValidateSession(r.Context(), cookie.Value)
	if err != nil {
		httphelpers.JSONError(w, http.StatusUnauthorized, "Invalid session")
		return
	}

	// Return session and user info
	response := map[string]interface{}{
		"session": map[string]interface{}{
			"id":            session.ID,
			"auth_level":    session.AuthLevel,
			"created_at":    session.CreatedAt,
			"last_activity": session.LastActivityAt,
			"expires_at":    session.ExpiresAt,
		},
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	}

	httphelpers.JSONResponse(w, http.StatusOK, response)
}

// Device management handlers

// handleGetDevices returns the user's registered devices
func (h *Handler) handleGetDevices(w http.ResponseWriter, r *http.Request) {
	// Get current session
	_, user, err := h.getCurrentUser(r)
	if err != nil {
		httphelpers.JSONError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Get user's WebAuthn credentials (representing devices/authenticators)
	credentials, err := h.service.repo.GetWebAuthnCredentialsByUserID(r.Context(), user.ID)
	if err != nil {
		h.logger.Error("Failed to get user credentials", "error", err, "user_id", user.ID)
		httphelpers.JSONError(w, http.StatusInternalServerError, "Failed to get devices")
		return
	}

	// Convert to device list
	devices := make([]map[string]interface{}, len(credentials))
	for i, cred := range credentials {
		devices[i] = map[string]interface{}{
			"id":            cred.ID,
			"friendly_name": cred.FriendlyName,
			"created_at":    cred.CreatedAt,
			"last_used_at":  cred.LastUsedAt,
			"is_synced":     cred.IsSynced,
			"is_backup":     cred.IsBackup,
		}
	}

	response := map[string]interface{}{
		"devices": devices,
		"total":   len(devices),
	}

	httphelpers.JSONResponse(w, http.StatusOK, response)
}

// Helper methods

// getCurrentUser extracts the current user from the request
func (h *Handler) getCurrentUser(r *http.Request) (*Session, *User, error) {
	// Get session token from cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil, nil, err
	}

	// Validate and get session
	return h.service.ValidateSession(r.Context(), cookie.Value)
}

// setSessionCookie sets the session cookie
func (h *Handler) setSessionCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		MaxAge:   int(DefaultSessionDuration.Seconds()),
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

// clearSessionCookie clears the session cookie
func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}

// Middleware for protecting routes

// RequireAuth is middleware that requires authentication
func (h *Handler) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, user, err := h.getCurrentUser(r)
		if err != nil {
			httphelpers.JSONError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		// Add user and session to request context
		ctx := r.Context()
		ctx = WithUser(ctx, user)
		ctx = WithSession(ctx, session)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// RequireAuthLevel is middleware that requires a specific auth level
func (h *Handler) RequireAuthLevel(level int) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return h.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
			session := GetSession(r.Context())
			if session == nil || session.AuthLevel < level {
				httphelpers.JSONError(w, http.StatusForbidden, "Insufficient authorization")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
