package handlers

import (
	"log"
	"net/http"
	"strings"

	"circles.diy/internal/auth"
	"circles.diy/internal/domain"
	"circles.diy/internal/middleware"
	"circles.diy/internal/models"
	"circles.diy/internal/preferences"
	"circles.diy/internal/profile"
	"circles.diy/internal/templates"
	"go.uber.org/zap"
)

// ProfileHandler handles profile page requests with real data from services
type ProfileHandler struct {
	profileService *profile.Service
	prefsService   *preferences.Service
	logger         *zap.Logger
}

// NewProfileHandler creates a new profile handler with dependencies
func NewProfileHandler(
	profileService *profile.Service,
	prefsService *preferences.Service,
	logger *zap.Logger,
) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
		prefsService:   prefsService,
		logger:         logger,
	}
}

// Handle processes profile page requests
func (h *ProfileHandler) Handle(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Get authenticated user from context
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's active profile ID from session
	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		http.Error(w, "No active profile", http.StatusBadRequest)
		return
	}

	// Load theme for the current user
	theme, err := h.prefsService.GetEffectiveTheme(r.Context(), user.ID, session.ActiveProfileID)
	if err != nil {
		h.logger.Warn("failed to load theme, using defaults", zap.Error(err))
		theme = domain.DefaultEffectiveTheme()
	}

	themeSettings := models.ThemeSettings{
		Mode:   theme.Mode,
		Radius: theme.Radius,
	}

	if path == "/profile" {
		// Internal profile view (owner's profile dashboard)
		h.handleInternalProfile(w, r, user, *session.ActiveProfileID, themeSettings)
	} else if strings.HasPrefix(path, "/profile/") {
		// External profile view (/profile/:handle)
		handle := strings.TrimPrefix(path, "/profile/")
		if handle == "" {
			http.NotFound(w, r)
			return
		}
		h.handleExternalProfile(w, r, user, handle, themeSettings)
	} else {
		http.NotFound(w, r)
	}
}

// handleInternalProfile renders the owner's profile dashboard
func (h *ProfileHandler) handleInternalProfile(
	w http.ResponseWriter,
	r *http.Request,
	user *domain.User,
	activeProfileID string,
	themeSettings models.ThemeSettings,
) {
	// Fetch the user's active profile
	profile, err := h.profileService.GetByID(r.Context(), activeProfileID)
	if err != nil {
		h.logger.Error("failed to fetch profile", zap.Error(err))
		http.Error(w, "Failed to load profile", http.StatusInternalServerError)
		return
	}

	if profile == nil {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	// Fetch profile settings
	settings, err := h.profileService.GetSettings(r.Context(), activeProfileID)
	if err != nil {
		h.logger.Warn("failed to fetch profile settings", zap.Error(err))
		// Continue with nil settings - template should handle this
	}

	// Map domain profile to template model
	templateProfile := models.Profile{
		ID:          profile.ID,
		UserID:      profile.UserID,
		Handle:      profile.Handle,
		Name:        profile.Name,
		DisplayName: profile.DisplayName,
		Bio:         profile.Bio,
		AvatarURL:   profile.AvatarURL,
		BannerURL:   profile.BannerURL,
		IsActive:    profile.IsActive,
		IsPublic:    true, // Default
		Stats:       models.ProfileStats{}, // TODO: Fetch real stats from content service when implemented
		IsOwner:     true,
		IsConnected: false,
		IsVerified:  false, // TODO: Add verification system
		Circles: []models.ProfileCircle{
			{ID: "1", Name: "Music crew", Icon: "🎵", Color: "var(--warm-accent)"},
			{ID: "2", Name: "Housemates", Icon: "🏠", Color: "var(--cool-accent)"},
			{ID: "3", Name: "Climbing", Icon: "🧗", Color: "var(--earth-accent)"},
			{ID: "4", Name: "Coffee nerds", Icon: "☕", Color: "var(--sage-accent)"},
		}, // TODO: Fetch real circles from circle service
		Settings: models.ProfileSettings2{
			SerendipityMode:    true,
			AwayMode:           false,
			BatchNotifications: true,
			CoordinationAlerts: true,
		}, // TODO: Fetch real settings from preferences service
	}

	// Add settings if available
	if settings != nil {
		templateProfile.IsPublic = settings.IsPublic
		templateProfile.Location = settings.Location
		templateProfile.Website = settings.Website
		templateProfile.Interests = settings.Interests
		templateProfile.SocialLinks = settings.SocialLinks
	}

	// Build template data from real sources only
	data := &models.ProfileData{
		BaseData: models.BaseData{
			Title:     profile.DisplayName + " - Profile",
			ActiveNav: "profile",
			Theme:     themeSettings,
			User:      user,
			CSRFToken: middleware.GetCSRFToken(r),
		},
		Profile:      templateProfile,
		Posts:        []models.Post{},        // TODO: Fetch from content service
		PostOffset:   0,
		HasMorePosts: false,
		IsOwner:      true,
		Extensions:   []models.Extension{},   // TODO: Fetch profile extensions
		Analytics:    models.Analytics{},     // TODO: Fetch profile analytics
		Drafts:       []models.DraftPost{},   // TODO: Fetch drafts
		DraftCount:   0,
	}

	err = templates.GetTemplates().ProfileInternal.ExecuteTemplate(w, "profile-internal", data)
	if err != nil {
		log.Printf("Error rendering profile-internal template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// handleExternalProfile renders a public profile view by handle
func (h *ProfileHandler) handleExternalProfile(
	w http.ResponseWriter,
	r *http.Request,
	user *domain.User,
	handle string,
	themeSettings models.ThemeSettings,
) {
	// Fetch the profile by handle
	profile, err := h.profileService.GetByHandle(r.Context(), handle)
	if err != nil {
		h.logger.Error("failed to fetch profile by handle",
			zap.String("handle", handle),
			zap.Error(err))
		http.Error(w, "Failed to load profile", http.StatusInternalServerError)
		return
	}

	if profile == nil {
		http.NotFound(w, r)
		return
	}

	// Fetch profile settings
	settings, err := h.profileService.GetSettings(r.Context(), profile.ID)
	if err != nil {
		h.logger.Warn("failed to fetch profile settings",
			zap.String("profile_id", profile.ID),
			zap.Error(err))
	}

	// Check if viewer is the owner
	session := auth.GetSession(r.Context())
	isOwner := session != nil &&
		session.ActiveProfileID != nil &&
		*session.ActiveProfileID == profile.ID

	// Map domain profile to template model
	templateProfile := models.Profile{
		ID:          profile.ID,
		Handle:      profile.Handle,
		Name:        profile.Name,
		DisplayName: profile.DisplayName,
		Bio:         profile.Bio,
		AvatarURL:   profile.AvatarURL,
		BannerURL:   profile.BannerURL,
		IsPublic:    true, // Default
		Stats:       models.ProfileStats{}, // TODO: Fetch real stats from content service when implemented
		IsOwner:     isOwner,
		IsConnected: false, // TODO: Check connection status between profiles
		IsVerified:  false, // TODO: Add verification system
	}

	// Add settings if available (and if public or owner)
	if settings != nil && (settings.IsPublic || isOwner) {
		templateProfile.IsPublic = settings.IsPublic
		templateProfile.Location = settings.Location
		templateProfile.Website = settings.Website
		templateProfile.Interests = settings.Interests
		templateProfile.SocialLinks = settings.SocialLinks
	}

	// Build template data from real sources only
	data := &models.ProfileData{
		BaseData: models.BaseData{
			Title:     profile.DisplayName + "'s Profile",
			ActiveNav: "", // No active nav for external profile view
			Theme:     themeSettings,
			User:      user,
			CSRFToken: middleware.GetCSRFToken(r),
		},
		Profile:      templateProfile,
		Posts:        []models.Post{},        // TODO: Fetch from content service
		PostOffset:   0,
		HasMorePosts: false,
		IsOwner:      isOwner,
		Extensions:   []models.Extension{},   // Extensions not shown on external view
		Analytics:    models.Analytics{},     // Analytics not shown on external view
		Drafts:       []models.DraftPost{},   // Drafts not shown on external view
		DraftCount:   0,
	}

	err = templates.GetTemplates().ProfilePublic.ExecuteTemplate(w, "profile-public", data)
	if err != nil {
		log.Printf("Error rendering profile-public template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
