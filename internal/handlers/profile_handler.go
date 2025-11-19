package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"

	"circles.diy/internal/auth"
	"circles.diy/internal/circle"
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
	circleService  *circle.Service
	logger         *zap.Logger
}

// NewProfileHandler creates a new profile handler with dependencies
func NewProfileHandler(
	profileService *profile.Service,
	prefsService *preferences.Service,
	circleService *circle.Service,
	logger *zap.Logger,
) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
		prefsService:   prefsService,
		circleService:  circleService,
		logger:         logger,
	}
}

// mapCirclesToProfileCircles converts domain circles to template model circles
func mapCirclesToProfileCircles(circles []domain.Circle) []models.ProfileCircle {
	profileCircles := make([]models.ProfileCircle, 0, len(circles))
	circleColorPalette := []string{
		"var(--warm-accent)",
		"var(--cool-accent)",
		"var(--earth-accent)",
		"var(--sage-accent)",
	}

	for i, circle := range circles {
		color := template.CSS(circleColorPalette[i%len(circleColorPalette)])
		if circle.IconBgColor != "" {
			color = circle.IconBgColor
		}

		icon := "⭕" // Default icon
		if circle.Icon != "" {
			icon = circle.Icon
		}

		profileCircles = append(profileCircles, models.ProfileCircle{
			ID:    circle.ID,
			Name:  circle.Name,
			Icon:  icon,
			Color: color,
		})
	}

	return profileCircles
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

	// Fetch user's circles
	userCircles, err := h.circleService.GetUserCircles(r.Context(), activeProfileID)
	if err != nil {
		h.logger.Warn("failed to fetch user circles", zap.Error(err))
		userCircles = []domain.Circle{} // Continue with empty circles
	}

	// Map circles to profile circles
	profileCircles := mapCirclesToProfileCircles(userCircles)

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
		Circles:     profileCircles,
		Settings: models.ProfileSettings2{
			SerendipityMode:    true,  // TODO: Implement serendipity mode in preferences
			AwayMode:           false, // TODO: Implement away mode in preferences
			BatchNotifications: true,  // TODO: Implement notification preferences
			CoordinationAlerts: true,  // TODO: Implement coordination alert preferences
		},
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

	// Fetch user's circles
	userCircles, err := h.circleService.GetUserCircles(r.Context(), profile.ID)
	if err != nil {
		h.logger.Warn("failed to fetch user circles",
			zap.String("profile_id", profile.ID),
			zap.Error(err))
		userCircles = []domain.Circle{} // Continue with empty circles
	}

	// Filter circles based on visibility (show only public circles unless viewer is owner)
	visibleCircles := make([]domain.Circle, 0)
	for _, circle := range userCircles {
		if isOwner || circle.IsPublic() {
			visibleCircles = append(visibleCircles, circle)
		}
	}

	// Map circles to profile circles
	profileCircles := mapCirclesToProfileCircles(visibleCircles)

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
		Circles:     profileCircles,
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

// UpdateProfile handles PUT /api/profile - Updates profile text fields
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get authenticated session
	session := auth.GetSession(ctx)
	if session == nil || session.ActiveProfileID == nil {
		h.logger.Warn("update profile attempted without session")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profileID := *session.ActiveProfileID

	// Parse form data
	if err := r.ParseForm(); err != nil {
		h.logger.Error("failed to parse form data",
			zap.String("profile_id", profileID),
			zap.Error(err))
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Extract form values (only update if provided and non-empty)
	var req struct {
		Handle      *string
		Name        *string
		DisplayName *string
		Bio         *string
	}

	if val := r.FormValue("handle"); val != "" {
		req.Handle = &val
	}
	if val := r.FormValue("name"); val != "" {
		req.Name = &val
	}
	if val := r.FormValue("display_name"); val != "" {
		req.DisplayName = &val
	}
	if val := r.FormValue("bio"); val != "" {
		req.Bio = &val
	}

	// Get current profile
	profile, err := h.profileService.GetByID(ctx, profileID)
	if err != nil {
		h.logger.Error("failed to get profile",
			zap.String("profile_id", profileID),
			zap.Error(err))
		http.Error(w, "Failed to load profile", http.StatusInternalServerError)
		return
	}

	if profile == nil {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	// Update fields if provided
	if req.Handle != nil {
		// Check if handle is already taken (unless it's the same)
		if *req.Handle != profile.Handle {
			exists, err := h.profileService.CheckHandleExists(ctx, *req.Handle)
			if err != nil {
				h.logger.Error("failed to check handle existence",
					zap.String("handle", *req.Handle),
					zap.Error(err))
				http.Error(w, "Failed to validate handle", http.StatusInternalServerError)
				return
			}
			if exists {
				http.Error(w, "Handle is already taken", http.StatusConflict)
				return
			}
		}
		profile.Handle = *req.Handle
	}

	if req.Name != nil {
		profile.Name = *req.Name
	}

	if req.DisplayName != nil {
		profile.DisplayName = *req.DisplayName
	}

	if req.Bio != nil {
		profile.Bio = *req.Bio
	}

	// Update profile
	if err := h.profileService.Update(ctx, profile); err != nil {
		h.logger.Error("failed to update profile",
			zap.String("profile_id", profileID),
			zap.Error(err))
		http.Error(w, "Failed to update profile: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated profile
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"profile": map[string]string{
			"id":           profile.ID,
			"handle":       profile.Handle,
			"name":         profile.Name,
			"display_name": profile.DisplayName,
			"bio":          profile.Bio,
			"avatar_url":   profile.AvatarURL,
			"banner_url":   profile.BannerURL,
		},
		"message": "Profile updated successfully",
	})
}

// HandleEdit handles GET /profile/edit - Renders the profile edit page
func (h *ProfileHandler) HandleEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get authenticated user
	user := auth.GetUser(ctx)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's active profile ID from session
	session := auth.GetSession(ctx)
	if session == nil || session.ActiveProfileID == nil {
		http.Error(w, "No active profile", http.StatusBadRequest)
		return
	}

	activeProfileID := *session.ActiveProfileID

	// Load theme for the current user
	theme, err := h.prefsService.GetEffectiveTheme(ctx, user.ID, session.ActiveProfileID)
	if err != nil {
		h.logger.Warn("failed to load theme, using defaults", zap.Error(err))
		theme = domain.DefaultEffectiveTheme()
	}

	themeSettings := models.ThemeSettings{
		Mode:   theme.Mode,
		Radius: theme.Radius,
	}

	// Fetch the user's active profile
	profile, err := h.profileService.GetByID(ctx, activeProfileID)
	if err != nil {
		h.logger.Error("failed to fetch profile for editing", zap.Error(err))
		http.Error(w, "Failed to load profile", http.StatusInternalServerError)
		return
	}

	if profile == nil {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	// Fetch profile settings
	settings, err := h.profileService.GetSettings(ctx, activeProfileID)
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
	}

	// Add settings if available
	if settings != nil {
		templateProfile.IsPublic = settings.IsPublic
		templateProfile.Location = settings.Location
		templateProfile.Website = settings.Website
		templateProfile.Interests = settings.Interests
		templateProfile.SocialLinks = settings.SocialLinks
	}

	// Build template data
	data := &models.ProfileData{
		BaseData: models.BaseData{
			Title:     "Edit Profile - " + profile.DisplayName,
			ActiveNav: "profile",
			Theme:     themeSettings,
			User:      user,
			CSRFToken: middleware.GetCSRFToken(r),
		},
		Profile: templateProfile,
		IsOwner: true,
	}

	// Render the profile edit template
	err = templates.GetTemplates().ProfileEdit.ExecuteTemplate(w, "profile-edit", data)
	if err != nil {
		h.logger.Error("error rendering profile-edit template", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
