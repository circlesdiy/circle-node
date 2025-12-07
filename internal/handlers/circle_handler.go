package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"circles.diy/internal/auth"
	"circles.diy/internal/circle"
	"circles.diy/internal/content"
	"circles.diy/internal/domain"
	"circles.diy/internal/middleware"
	"circles.diy/internal/models"
	"circles.diy/internal/preferences"
	"circles.diy/internal/profile"
	"circles.diy/internal/templates"
	"go.uber.org/zap"
)

// CircleHandler handles circle-related HTTP requests
type CircleHandler struct {
	circleService  *circle.Service
	contentService *content.Service
	prefsService   *preferences.Service
	profileService *profile.Service
	assetVersion   string
	logger         *zap.Logger
}

// NewCircleHandler creates a new circle handler
func NewCircleHandler(
	circleService *circle.Service,
	contentService *content.Service,
	prefsService *preferences.Service,
	profileService *profile.Service,
	assetVersion string,
	logger *zap.Logger,
) *CircleHandler {
	return &CircleHandler{
		circleService:  circleService,
		contentService: contentService,
		prefsService:   prefsService,
		profileService: profileService,
		assetVersion:   assetVersion,
		logger:         logger,
	}
}

// Handle is the main router for circle requests
func (h *CircleHandler) Handle(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/circles")

	switch {
	case path == "" || path == "/":
		h.handleListCircles(w, r)
	case strings.HasPrefix(path, "/new"):
		h.handleNewCircle(w, r)
	case strings.HasPrefix(path, "/"):
		// Extract circle ID from path
		parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
		if len(parts) == 0 {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		circleID := parts[0]

		// Route to specific circle actions
		if len(parts) == 1 {
			h.handleViewCircle(w, r, circleID)
		} else if len(parts) >= 2 {
			action := parts[1]
			switch action {
			case "edit":
				h.handleEditCircle(w, r, circleID)
			case "delete":
				h.handleDeleteCircle(w, r, circleID)
			case "invite":
				h.handleInviteMember(w, r, circleID)
			case "join":
				h.handleJoinCircle(w, r, circleID)
			case "leave":
				h.handleLeaveCircle(w, r, circleID)
			case "members":
				if len(parts) == 4 {
					memberID := parts[2]
					memberAction := parts[3]
					switch memberAction {
					case "ban":
						h.handleBanMember(w, r, circleID, memberID)
					case "remove":
						h.handleRemoveMember(w, r, circleID, memberID)
					default:
						http.Error(w, "Not found", http.StatusNotFound)
					}
				} else {
					http.Error(w, "Not found", http.StatusNotFound)
				}
			case "invites":
				if len(parts) == 4 {
					profileID := parts[2]
					inviteAction := parts[3]
					switch inviteAction {
					case "resend":
						h.handleResendInvite(w, r, circleID, profileID)
					case "revoke":
						h.handleRevokeInvite(w, r, circleID, profileID)
					default:
						http.Error(w, "Not found", http.StatusNotFound)
					}
				} else {
					http.Error(w, "Not found", http.StatusNotFound)
				}
			default:
				http.Error(w, "Not found", http.StatusNotFound)
			}
		}
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// RegisterRoutes registers all circle routes with the mux
func (h *CircleHandler) RegisterRoutes(mux *http.ServeMux, authHandler *auth.Handler, uploadHandler *CircleUploadHandler) {
	// Page routes
	mux.HandleFunc("/circles", authHandler.RequireAuth(h.Handle))
	mux.HandleFunc("/circles/", authHandler.RequireAuth(h.Handle))

	// API routes for HTMX
	mux.HandleFunc("/api/circles", authHandler.RequireAuth(h.handleAPICreateCircle))
	mux.HandleFunc("/api/circles/form", authHandler.RequireAuth(h.handleAPICircleForm))
	mux.HandleFunc("/api/circles/", authHandler.RequireAuth(h.handleAPICircleActions))

	// Tab lazy-loading endpoints
	mux.HandleFunc("GET /api/circles/{id}/tabs/chat", authHandler.RequireAuth(h.handleAPITabChat))
	mux.HandleFunc("GET /api/circles/{id}/tabs/members", authHandler.RequireAuth(h.handleAPITabMembers))
	mux.HandleFunc("GET /api/circles/{id}/tabs/gatherings", authHandler.RequireAuth(h.handleAPITabGatherings))
	mux.HandleFunc("GET /api/circles/{id}/tabs/files", authHandler.RequireAuth(h.handleAPITabFiles))
	mux.HandleFunc("GET /api/circles/{id}/tabs/settings", authHandler.RequireAuth(h.handleAPITabSettings))

	// Image upload routes (use specific paths to avoid conflicts with handleAPICircleActions)
	mux.HandleFunc("POST /api/circles/{id}/avatar", authHandler.RequireAuth(uploadHandler.UploadAvatar))
	mux.HandleFunc("POST /api/circles/{id}/banner", authHandler.RequireAuth(uploadHandler.UploadBanner))
	mux.HandleFunc("DELETE /api/circles/{id}/avatar", authHandler.RequireAuth(uploadHandler.DeleteAvatar))
	mux.HandleFunc("DELETE /api/circles/{id}/banner", authHandler.RequireAuth(uploadHandler.DeleteBanner))
}

// handleListCircles shows the main circles page with user's circles and public circles
func (h *CircleHandler) handleListCircles(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		http.Error(w, "No active profile", http.StatusBadRequest)
		return
	}

	profileID := *session.ActiveProfileID

	// Load theme preferences
	effectiveTheme, err := h.prefsService.GetEffectiveTheme(r.Context(), user.ID, session.ActiveProfileID)
	if err != nil {
		h.logger.Error("failed to load theme preferences",
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		// Continue with default theme
		effectiveTheme = &domain.EffectiveTheme{
			BaseThemeSettings: domain.BaseThemeSettings{
				Mode:   "system",
				Radius: "6",
			},
		}
	}

	// Get pending invitations with inviter information
	pendingInvitations := make([]models.CircleInvitation, 0)
	invitations, err := h.circleService.GetPendingInvitationsWithInviter(r.Context(), profileID)
	if err != nil {
		h.logger.Error("failed to get pending invitations",
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		// Don't fail the whole page, just log the error
	} else {
		for _, inv := range invitations {
			// Get circle details
			circle, err := h.circleService.GetCircleByID(r.Context(), inv.CircleID)
			if err != nil {
				h.logger.Error("failed to get circle for invitation",
					zap.String("circle_id", inv.CircleID),
					zap.Error(err),
				)
				continue
			}

			// Use inviter name from the query (joined with profiles table)
			inviterName := "Someone"
			if inv.InviterName != "" {
				inviterName = inv.InviterName
			}

			pendingInvitations = append(pendingInvitations, models.CircleInvitation{
				MembershipID:  inv.ID,
				CircleID:      circle.ID,
				CircleName:    circle.Name,
				CircleIcon:    circle.Icon,
				CircleBgColor: circle.IconBgColor,
				CircleAvatar:  circle.AvatarURL,
				InviterName:   inviterName,
				InviterHandle: inv.InviterHandle,
				InvitedAt:     inv.CreatedAt.Format("Jan 2, 2006"),
			})
		}
	}

	// Get user's circles (owned + member)
	userCircles, err := h.circleService.GetUserCircles(r.Context(), profileID)
	if err != nil {
		h.logger.Error("failed to get user circles",
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		http.Error(w, "Failed to load circles", http.StatusInternalServerError)
		return
	}

	// Get public circles
	publicCircles, err := h.circleService.GetPublicCircles(r.Context(), 20, 0)
	if err != nil {
		h.logger.Error("failed to get public circles",
			zap.Error(err),
		)
		http.Error(w, "Failed to load public circles", http.StatusInternalServerError)
		return
	}

	// Convert domain circles to template models
	templateCircles := make([]models.Circle, 0, len(userCircles))
	for _, c := range userCircles {
		// Get member count
		memberCount, _ := h.circleService.CountMembers(r.Context(), c.ID)

		// Determine user role
		userRole := "member"
		if c.OwnerProfileID == profileID {
			userRole = "owner"
		}

		templateCircles = append(templateCircles, models.Circle{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
			MemberCount: strconv.Itoa(memberCount),
			UserRole:    userRole,
			Active:      true,
			Icon:        c.Icon,
			IconBgColor: c.IconBgColor,
			AvatarURL:   c.AvatarURL,
			Banner:      c.BannerURL,
		})
	}

	// Add public circles that user is not already a member of
	existingIDs := make(map[string]bool)
	for _, c := range userCircles {
		existingIDs[c.ID] = true
	}

	featuredCircles := make([]models.Circle, 0)
	for _, c := range publicCircles {
		if !existingIDs[c.ID] {
			memberCount, _ := h.circleService.CountMembers(r.Context(), c.ID)
			featuredCircles = append(featuredCircles, models.Circle{
				ID:          c.ID,
				Name:        c.Name,
				Description: c.Description,
				MemberCount: strconv.Itoa(memberCount),
				UserRole:    "",
				Active:      false,
				Icon:        c.Icon,
				IconBgColor: c.IconBgColor,
				AvatarURL:   c.AvatarURL,
				Banner:      c.BannerURL,
			})
		}
	}

	// Build page data
	data := &models.CirclesPageData{
		BaseData: models.BaseData{
			Title:     "Circles",
			ActiveNav: "circles",
			Theme: models.ThemeSettings{
				Mode:   effectiveTheme.Mode,
				Radius: effectiveTheme.Radius,
			},
			User:         user,
			CSRFToken:    middleware.GetCSRFToken(r),
			AssetVersion: h.assetVersion,
		},
		Circles:            templateCircles,
		PendingInvitations: pendingInvitations,
		FeaturedCircles:    featuredCircles,
		RecentActivity:     []models.CircleActivity{}, // TODO: Implement activity feed
		Stats: models.CircleStats{
			TotalPosts:     0, // TODO: Implement stats
			ActiveMembers:  len(templateCircles),
			RecentActivity: "Today",
		},
	}

	err = templates.GetTemplates().Circles.ExecuteTemplate(w, "circles", data)
	if err != nil {
		h.logger.Error("failed to render circles template",
			zap.Error(err),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// handleViewCircle shows a specific circle's detail page
func (h *CircleHandler) handleViewCircle(w http.ResponseWriter, r *http.Request, circleID string) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		http.Error(w, "No active profile", http.StatusBadRequest)
		return
	}

	profileID := *session.ActiveProfileID

	// Get circle
	circle, err := h.circleService.GetCircleByID(r.Context(), circleID)
	if err != nil {
		h.logger.Error("failed to get circle",
			zap.String("circle_id", circleID),
			zap.Error(err),
		)
		http.Error(w, "Failed to load circle", http.StatusInternalServerError)
		return
	}

	if circle == nil {
		http.Error(w, "Circle not found", http.StatusNotFound)
		return
	}

	// Check if user can view this circle
	canView, err := h.circleService.CanView(r.Context(), circleID, profileID)
	if err != nil {
		h.logger.Error("failed to check view permission",
			zap.String("circle_id", circleID),
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
		return
	}

	if !canView {
		http.Error(w, "You don't have permission to view this circle", http.StatusForbidden)
		return
	}

	// Get user preferences for theme
	effectiveTheme, err := h.prefsService.GetEffectiveTheme(r.Context(), user.ID, &profileID)
	if err != nil {
		h.logger.Warn("failed to get user preferences",
			zap.String("user_id", user.ID),
			zap.Error(err),
		)
		effectiveTheme = domain.DefaultEffectiveTheme()
	}

	// Build page data
	data := h.buildCircleDetailPageData(r.Context(), circle, profileID, user, effectiveTheme)
	data.CSRFToken = middleware.GetCSRFToken(r)

	// Render template
	if err := templates.GetTemplates().CircleDetail.ExecuteTemplate(w, "circle-detail", data); err != nil {
		h.logger.Error("template error",
			zap.String("circle_id", circleID),
			zap.Error(err),
		)
		// Note: Cannot send error response here as headers are already sent
	}
}

// buildCircleDetailPageData builds the circle detail page data
// This now only loads minimal data + chat tab for initial render
// Other tabs are loaded lazily via HTMX
func (h *CircleHandler) buildCircleDetailPageData(ctx context.Context, circle *domain.Circle, profileID string, user *domain.User, effectiveTheme *domain.EffectiveTheme) *models.CircleDetailPageData {
	// Determine user role and permissions (minimal checks)
	isOwner, _ := h.circleService.IsOwner(ctx, circle.ID, profileID)
	isAdmin, _ := h.circleService.IsAdmin(ctx, circle.ID, profileID)
	isMember, _ := h.circleService.IsMember(ctx, circle.ID, profileID)

	// Get member count for header badge
	memberCount, _ := h.circleService.CountMembers(ctx, circle.ID)

	// Load ONLY chat tab data for initial render
	posts, _ := h.contentService.GetPostsByCircle(ctx, circle.ID, 20, 0)
	recentPosts := h.buildPostModels(ctx, posts, profileID)

	// Build lightweight stats for sidebar
	stats := models.CircleDetailStats{
		TotalPosts:   len(posts),
		CreatedAt:    circle.CreatedAt.Format("Jan 2, 2006"),
		LastActivity: "Recently",
	}

	return &models.CircleDetailPageData{
		BaseData: models.BaseData{
			Title:     circle.Name,
			ActiveNav: "circles",
			Theme: models.ThemeSettings{
				Mode:   effectiveTheme.Mode,
				Radius: effectiveTheme.Radius,
			},
			User:         user,
			AssetVersion: h.assetVersion,
		},
		Circle:          *circle,
		MemberCount:     memberCount,
		RecentPosts:     recentPosts, // Only chat data
		IsOwner:         isOwner,
		IsAdmin:         isAdmin,
		IsMember:        isMember,
		ActiveProfileID: profileID,
		CircleStats:     stats,
		ActiveTab:       "chat",

		// These are now loaded on-demand via HTMX:
		Members:            nil,
		PendingInvites:     nil,
		UpcomingGatherings: nil,
		SharedFiles:        nil,
		CanInvite:          false, // Loaded in members tab
		CanEditInfo:        false, // Loaded in settings tab
		CanEditVisibility:  false, // Loaded in settings tab
		CanEditPermissions: false, // Loaded in settings tab
		UserRole:           "",    // Loaded if needed
	}
}

// formatTimeAgo formats a time as a relative string
func formatTimeAgo(t time.Time) string {
	now := time.Now().UTC()
	diff := now.Sub(t.UTC())

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else if diff < 30*24*time.Hour {
		weeks := int(diff.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	} else if diff < 365*24*time.Hour {
		months := int(diff.Hours() / (24 * 30))
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	} else {
		years := int(diff.Hours() / (24 * 365))
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

// Helper methods for building tab data models

// buildPostModels converts domain posts to template models
func (h *CircleHandler) buildPostModels(ctx context.Context, posts []*domain.Post, viewerProfileID string) []models.CirclePost {
	if len(posts) == 0 {
		return []models.CirclePost{}
	}

	// Collect unique profile IDs for batch fetching
	profileIDSet := make(map[string]bool)
	for _, post := range posts {
		profileIDSet[post.AuthorProfileID] = true
	}

	// Batch fetch profiles
	profileMap := make(map[string]*domain.Profile)
	for authorProfileID := range profileIDSet {
		profile, err := h.profileService.GetByID(ctx, authorProfileID)
		if err == nil && profile != nil {
			profileMap[authorProfileID] = profile
		}
	}

	recentPosts := make([]models.CirclePost, 0, len(posts))
	for _, post := range posts {
		// Get reaction counts
		reactionCounts, _ := h.contentService.GetReactionCounts(ctx, "post", post.ID)
		likeCount := reactionCounts["like"]

		// Check if user has liked
		userHasLiked := false
		if viewerProfileID != "" {
			userReaction, _ := h.contentService.GetUserReaction(ctx, "post", post.ID, viewerProfileID)
			userHasLiked = userReaction != nil && userReaction.Key == "like"
		}

		// Check if user can edit
		canEdit := post.AuthorProfileID == viewerProfileID

		// Format time
		formattedTime := formatTimeAgo(post.CreatedAt)

		// Check if edited
		isEdited := post.EditedAt != nil
		var editedAtStr *string
		if isEdited {
			editedStr := formatTimeAgo(*post.EditedAt)
			editedAtStr = &editedStr
		}

		// Get author profile information
		authorName := "Circle Member"
		authorAvatar := ""
		if authorProfile, ok := profileMap[post.AuthorProfileID]; ok {
			if authorProfile.DisplayName != "" {
				authorName = authorProfile.DisplayName
			} else if authorProfile.Name != "" {
				authorName = authorProfile.Name
			} else if authorProfile.Handle != "" {
				authorName = authorProfile.Handle
			}
			authorAvatar = authorProfile.AvatarURL
		}

		recentPosts = append(recentPosts, models.CirclePost{
			ID:              post.ID,
			AuthorProfileID: post.AuthorProfileID,
			AuthorName:      authorName,
			AuthorAvatar:    authorAvatar,
			Body:            post.Body,
			BodyFormat:      post.BodyFormat,
			ContentWarning:  post.ContentWarning,
			Visibility:      post.Visibility,
			CreatedAt:       post.CreatedAt.Format(time.RFC3339),
			EditedAt:        editedAtStr,
			FormattedTime:   formattedTime,
			IsEdited:        isEdited,
			ReplyCount:      post.ReplyCount,
			LikeCount:       likeCount,
			UserHasLiked:    userHasLiked,
			CanEdit:         canEdit,
			ShowComments:    false,
		})
	}

	return recentPosts
}

// buildMemberModels fetches and builds member models for the members tab
func (h *CircleHandler) buildMemberModels(ctx context.Context, circleID string, ownerProfileID string) ([]models.CircleMember, []models.CircleMember) {
	// Get members
	memberships, _ := h.circleService.GetCircleMembers(ctx, circleID, 100, 0)

	// Batch fetch member profiles
	memberProfileIDSet := make(map[string]bool)
	for _, m := range memberships {
		memberProfileIDSet[m.ProfileID] = true
	}

	memberProfileMap := make(map[string]*domain.Profile)
	for memberProfileID := range memberProfileIDSet {
		profile, err := h.profileService.GetByID(ctx, memberProfileID)
		if err == nil && profile != nil {
			memberProfileMap[memberProfileID] = profile
		}
	}

	// Separate pending invites from active members
	members := make([]models.CircleMember, 0)
	pendingInvites := make([]models.CircleMember, 0)

	for _, m := range memberships {
		// Determine role for this member
		role := "member"
		if m.ProfileID == ownerProfileID {
			role = "owner"
		}
		// TODO: Check if member is admin once roles are implemented

		// Get profile information
		name := "Member"
		username := ""
		avatar := ""
		if memberProfile, ok := memberProfileMap[m.ProfileID]; ok {
			if memberProfile.DisplayName != "" {
				name = memberProfile.DisplayName
			} else if memberProfile.Name != "" {
				name = memberProfile.Name
			} else if memberProfile.Handle != "" {
				name = memberProfile.Handle
			}
			username = memberProfile.Handle
			avatar = memberProfile.AvatarURL
		}

		circleMember := models.CircleMember{
			ProfileID: m.ProfileID,
			Name:      name,
			Username:  username,
			Avatar:    avatar,
			Role:      role,
			State:     m.State,
			JoinedAt:  m.JoinedAt.Format("Jan 2, 2006"),
		}

		// Separate invited vs active members
		if m.State == "invited" {
			pendingInvites = append(pendingInvites, circleMember)
		} else if m.State == "active" {
			members = append(members, circleMember)
		}
	}

	return members, pendingInvites
}

// Tab endpoint handlers for lazy loading

// API wrappers for tab endpoints

// handleAPITabChat is the API wrapper for chat tab
func (h *CircleHandler) handleAPITabChat(w http.ResponseWriter, r *http.Request) {
	circleID := r.PathValue("id")
	h.handleTabChat(w, r, circleID)
}

// handleAPITabMembers is the API wrapper for members tab
func (h *CircleHandler) handleAPITabMembers(w http.ResponseWriter, r *http.Request) {
	circleID := r.PathValue("id")
	h.handleTabMembers(w, r, circleID)
}

// handleAPITabGatherings is the API wrapper for gatherings tab
func (h *CircleHandler) handleAPITabGatherings(w http.ResponseWriter, r *http.Request) {
	circleID := r.PathValue("id")
	h.handleTabGatherings(w, r, circleID)
}

// handleAPITabFiles is the API wrapper for files tab
func (h *CircleHandler) handleAPITabFiles(w http.ResponseWriter, r *http.Request) {
	circleID := r.PathValue("id")
	h.handleTabFiles(w, r, circleID)
}

// handleAPITabSettings is the API wrapper for settings tab
func (h *CircleHandler) handleAPITabSettings(w http.ResponseWriter, r *http.Request) {
	circleID := r.PathValue("id")
	h.handleTabSettings(w, r, circleID)
}

// handleTabChat returns the chat tab content
func (h *CircleHandler) handleTabChat(w http.ResponseWriter, r *http.Request, circleID string) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		http.Error(w, "No active profile", http.StatusBadRequest)
		return
	}

	profileID := *session.ActiveProfileID

	// Permission check
	canView, err := h.circleService.CanView(r.Context(), circleID, profileID)
	if err != nil {
		h.logger.Error("failed to check view permission",
			zap.String("circle_id", circleID),
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
		return
	}

	if !canView {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Load only chat-related data
	circle, err := h.circleService.GetCircleByID(r.Context(), circleID)
	if err != nil {
		h.logger.Error("failed to get circle",
			zap.String("circle_id", circleID),
			zap.Error(err),
		)
		http.Error(w, "Failed to load circle", http.StatusInternalServerError)
		return
	}

	isMember, _ := h.circleService.IsMember(r.Context(), circleID, profileID)
	posts, _ := h.contentService.GetPostsByCircle(r.Context(), circleID, 20, 0)

	// Build post models
	recentPosts := h.buildPostModels(r.Context(), posts, profileID)

	data := map[string]interface{}{
		"Circle":      circle,
		"IsMember":    isMember,
		"RecentPosts": recentPosts,
	}

	RenderFragment(w, h.logger, "circle-tab-chat", data)
}

// handleTabMembers returns the members tab content
func (h *CircleHandler) handleTabMembers(w http.ResponseWriter, r *http.Request, circleID string) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		http.Error(w, "No active profile", http.StatusBadRequest)
		return
	}

	profileID := *session.ActiveProfileID

	canView, err := h.circleService.CanView(r.Context(), circleID, profileID)
	if err != nil {
		http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
		return
	}

	if !canView {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	circle, err := h.circleService.GetCircleByID(r.Context(), circleID)
	if err != nil {
		http.Error(w, "Failed to load circle", http.StatusInternalServerError)
		return
	}

	isOwner := circle.OwnerProfileID == profileID
	isAdmin, _ := h.circleService.IsAdmin(r.Context(), circleID, profileID)
	canInvite, _ := h.circleService.CanInvite(r.Context(), circleID, profileID)
	memberCount, _ := h.circleService.CountMembers(r.Context(), circleID)

	// Load members with profiles
	members, pendingInvites := h.buildMemberModels(r.Context(), circleID, circle.OwnerProfileID)

	data := map[string]interface{}{
		"Circle":          circle,
		"Members":         members,
		"PendingInvites":  pendingInvites,
		"MemberCount":     memberCount,
		"IsOwner":         isOwner,
		"IsAdmin":         isAdmin,
		"CanInvite":       canInvite,
		"ActiveProfileID": profileID,
	}

	RenderFragment(w, h.logger, "circle-tab-members", data)
}

// handleTabGatherings returns the gatherings tab content
func (h *CircleHandler) handleTabGatherings(w http.ResponseWriter, r *http.Request, circleID string) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		http.Error(w, "No active profile", http.StatusBadRequest)
		return
	}

	profileID := *session.ActiveProfileID

	canView, err := h.circleService.CanView(r.Context(), circleID, profileID)
	if err != nil {
		http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
		return
	}

	if !canView {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	circle, err := h.circleService.GetCircleByID(r.Context(), circleID)
	if err != nil {
		http.Error(w, "Failed to load circle", http.StatusInternalServerError)
		return
	}

	isMember, _ := h.circleService.IsMember(r.Context(), circleID, profileID)

	// TODO: Load gatherings from service when implemented
	upcomingGatherings := []models.GatheringItem{}

	data := map[string]interface{}{
		"Circle":             circle,
		"IsMember":           isMember,
		"UpcomingGatherings": upcomingGatherings,
	}

	RenderFragment(w, h.logger, "circle-tab-gatherings", data)
}

// handleTabFiles returns the files tab content
func (h *CircleHandler) handleTabFiles(w http.ResponseWriter, r *http.Request, circleID string) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		http.Error(w, "No active profile", http.StatusBadRequest)
		return
	}

	profileID := *session.ActiveProfileID

	canView, err := h.circleService.CanView(r.Context(), circleID, profileID)
	if err != nil {
		http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
		return
	}

	if !canView {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	circle, err := h.circleService.GetCircleByID(r.Context(), circleID)
	if err != nil {
		http.Error(w, "Failed to load circle", http.StatusInternalServerError)
		return
	}

	isMember, _ := h.circleService.IsMember(r.Context(), circleID, profileID)

	// TODO: Load files from service when implemented
	sharedFiles := []models.CircleFile{}

	data := map[string]interface{}{
		"Circle":      circle,
		"IsMember":    isMember,
		"SharedFiles": sharedFiles,
	}

	RenderFragment(w, h.logger, "circle-tab-files", data)
}

// handleTabSettings returns the settings tab content
func (h *CircleHandler) handleTabSettings(w http.ResponseWriter, r *http.Request, circleID string) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		http.Error(w, "No active profile", http.StatusBadRequest)
		return
	}

	profileID := *session.ActiveProfileID

	canView, err := h.circleService.CanView(r.Context(), circleID, profileID)
	if err != nil {
		http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
		return
	}

	if !canView {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	circle, err := h.circleService.GetCircleByID(r.Context(), circleID)
	if err != nil {
		http.Error(w, "Failed to load circle", http.StatusInternalServerError)
		return
	}

	isOwner := circle.OwnerProfileID == profileID
	canEditSettings, _ := h.circleService.CanEditSettings(r.Context(), circleID, profileID)

	data := map[string]interface{}{
		"Circle":            circle,
		"IsOwner":           isOwner,
		"CanEditInfo":       canEditSettings,
		"CanEditVisibility": canEditSettings,
		"CSRFToken":         middleware.GetCSRFToken(r),
	}

	RenderFragment(w, h.logger, "circle-tab-settings", data)
}

// handleNewCircle shows the create circle form or handles creation
func (h *CircleHandler) handleNewCircle(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		http.Error(w, "No active profile", http.StatusBadRequest)
		return
	}

	profileID := *session.ActiveProfileID

	if r.Method == http.MethodGet {
		// TODO: Show create circle form
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<!DOCTYPE html>
			<html>
			<head><title>Create Circle</title></head>
			<body>
				<h1>Create a New Circle</h1>
				<form method="POST">
					<div>
						<label>Name: <input type="text" name="name" required></label>
					</div>
					<div>
						<label>Description: <textarea name="description"></textarea></label>
					</div>
					<div>
						<label>Visibility:
							<select name="visibility">
								<option value="public">Public</option>
								<option value="private">Private</option>
								<option value="unlisted">Unlisted</option>
							</select>
						</label>
					</div>
					<div>
						<label><input type="checkbox" name="auto_mod_enabled"> Enable Auto-Moderation</label>
					</div>
					<button type="submit">Create Circle</button>
				</form>
			</body>
			</html>
		`))
		return
	}

	if r.Method == http.MethodPost {
		// Parse form
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		name := r.FormValue("name")
		description := r.FormValue("description")
		visibility := r.FormValue("visibility")
		autoModEnabled := r.FormValue("auto_mod_enabled") == "on"

		// Create circle (legacy form - no icon/color support)
		circle, err := h.circleService.CreateCircle(
			r.Context(),
			name,
			description,
			visibility,
			autoModEnabled,
			"", // icon
			"", // iconBgColor
			profileID,
		)
		if err != nil {
			h.logger.Error("failed to create circle",
				zap.String("profile_id", profileID),
				zap.Error(err),
			)
			http.Error(w, "Failed to create circle: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Redirect to circle page
		http.Redirect(w, r, "/circles/"+circle.ID, http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// handleEditCircle handles circle editing (owner only)
func (h *CircleHandler) handleEditCircle(w http.ResponseWriter, r *http.Request, circleID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		http.Error(w, "No active profile", http.StatusBadRequest)
		return
	}

	profileID := *session.ActiveProfileID

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if name := r.FormValue("name"); name != "" {
		updates["name"] = name
	}
	if description := r.FormValue("description"); description != "" {
		updates["description"] = description
	}
	if visibility := r.FormValue("visibility"); visibility != "" {
		updates["visibility"] = visibility
	}
	updates["auto_mod_enabled"] = r.FormValue("auto_mod_enabled") == "on"

	// Update circle
	err := h.circleService.UpdateCircle(r.Context(), circleID, updates, profileID)
	if err != nil {
		h.logger.Error("failed to update circle",
			zap.String("circle_id", circleID),
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		http.Error(w, "Failed to update circle: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Redirect back to circle
	http.Redirect(w, r, "/circles/"+circleID, http.StatusSeeOther)
}

// handleDeleteCircle handles circle deletion (owner only)
func (h *CircleHandler) handleDeleteCircle(w http.ResponseWriter, r *http.Request, circleID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		http.Error(w, "No active profile", http.StatusBadRequest)
		return
	}

	profileID := *session.ActiveProfileID

	err := h.circleService.DeleteCircle(r.Context(), circleID, profileID)
	if err != nil {
		h.logger.Error("failed to delete circle",
			zap.String("circle_id", circleID),
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		http.Error(w, "Failed to delete circle: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Redirect to circles list
	http.Redirect(w, r, "/circles", http.StatusSeeOther)
}

// handleInviteMember handles inviting a member to a circle
func (h *CircleHandler) handleInviteMember(w http.ResponseWriter, r *http.Request, circleID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		http.Error(w, "No active profile", http.StatusBadRequest)
		return
	}

	profileID := *session.ActiveProfileID

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	inviteeProfileID := r.FormValue("profile_id")
	if inviteeProfileID == "" {
		http.Error(w, "Profile ID is required", http.StatusBadRequest)
		return
	}

	err := h.circleService.InviteMember(r.Context(), circleID, inviteeProfileID, profileID)
	if err != nil {
		h.logger.Error("failed to invite member",
			zap.String("circle_id", circleID),
			zap.String("invitee_profile_id", inviteeProfileID),
			zap.Error(err),
		)
		http.Error(w, "Failed to invite member: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Member invited successfully"))
}

// handleJoinCircle handles accepting a circle invitation
func (h *CircleHandler) handleJoinCircle(w http.ResponseWriter, r *http.Request, circleID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		http.Error(w, "No active profile", http.StatusBadRequest)
		return
	}

	profileID := *session.ActiveProfileID

	err := h.circleService.AcceptInvitation(r.Context(), circleID, profileID)
	if err != nil {
		h.logger.Error("failed to join circle",
			zap.String("circle_id", circleID),
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		http.Error(w, "Failed to join circle: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Redirect to circle
	http.Redirect(w, r, "/circles/"+circleID, http.StatusSeeOther)
}

// handleLeaveCircle handles leaving a circle
func (h *CircleHandler) handleLeaveCircle(w http.ResponseWriter, r *http.Request, circleID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		http.Error(w, "No active profile", http.StatusBadRequest)
		return
	}

	profileID := *session.ActiveProfileID

	err := h.circleService.LeaveCircle(r.Context(), circleID, profileID)
	if err != nil {
		h.logger.Error("failed to leave circle",
			zap.String("circle_id", circleID),
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		http.Error(w, "Failed to leave circle: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Redirect to circles list
	http.Redirect(w, r, "/circles", http.StatusSeeOther)
}

// handleBanMember handles banning a member from a circle
func (h *CircleHandler) handleBanMember(w http.ResponseWriter, r *http.Request, circleID, memberProfileID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		http.Error(w, "No active profile", http.StatusBadRequest)
		return
	}

	profileID := *session.ActiveProfileID

	err := h.circleService.BanMember(r.Context(), circleID, memberProfileID, profileID)
	if err != nil {
		h.logger.Error("failed to ban member",
			zap.String("circle_id", circleID),
			zap.String("member_profile_id", memberProfileID),
			zap.Error(err),
		)
		http.Error(w, "Failed to ban member: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Member banned successfully"))
}

// handleRemoveMember handles removing a member from a circle
func (h *CircleHandler) handleRemoveMember(w http.ResponseWriter, r *http.Request, circleID, memberProfileID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		http.Error(w, "No active profile", http.StatusBadRequest)
		return
	}

	profileID := *session.ActiveProfileID

	err := h.circleService.RemoveMember(r.Context(), circleID, memberProfileID, profileID)
	if err != nil {
		h.logger.Error("failed to remove member",
			zap.String("circle_id", circleID),
			zap.String("member_profile_id", memberProfileID),
			zap.Error(err),
		)
		http.Error(w, "Failed to remove member: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Member removed successfully"))
}

// handleResendInvite resends an invitation to a pending member
func (h *CircleHandler) handleResendInvite(w http.ResponseWriter, r *http.Request, circleID, invitedProfileID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		http.Error(w, "No active profile", http.StatusBadRequest)
		return
	}

	profileID := *session.ActiveProfileID

	// Check if user has permission to invite (owner or admin)
	circle, err := h.circleService.GetCircleByID(r.Context(), circleID)
	if err != nil {
		http.Error(w, "Circle not found", http.StatusNotFound)
		return
	}

	isOwner := circle.OwnerProfileID == profileID
	isAdmin, _ := h.circleService.IsAdmin(r.Context(), circleID, profileID)

	if !isOwner && !isAdmin {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	// Update the invitation timestamp (resend is just updating the created_at time)
	// For now, we'll just return success - in a real implementation, you'd update the timestamp
	// or send a notification
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<div class="toast">Invitation resent</div>`))
}

// handleRevokeInvite revokes a pending invitation
func (h *CircleHandler) handleRevokeInvite(w http.ResponseWriter, r *http.Request, circleID, invitedProfileID string) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		http.Error(w, "No active profile", http.StatusBadRequest)
		return
	}

	profileID := *session.ActiveProfileID

	// Check if user has permission to revoke invitations (owner or admin)
	circle, err := h.circleService.GetCircleByID(r.Context(), circleID)
	if err != nil {
		http.Error(w, "Circle not found", http.StatusNotFound)
		return
	}

	isOwner := circle.OwnerProfileID == profileID
	isAdmin, _ := h.circleService.IsAdmin(r.Context(), circleID, profileID)

	if !isOwner && !isAdmin {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	// Use DeclineInvitation to revoke (it deletes the membership record)
	err = h.circleService.DeclineInvitation(r.Context(), circleID, invitedProfileID)
	if err != nil {
		h.logger.Error("failed to revoke invitation",
			zap.String("circle_id", circleID),
			zap.String("invited_profile_id", invitedProfileID),
			zap.Error(err),
		)
		http.Error(w, "Failed to revoke invitation: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Return empty response to remove the card from DOM
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(""))
}

// API Handlers for HTMX

// handleAPICircleForm renders the circle form in a modal
func (h *CircleHandler) handleAPICircleForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if this is for editing an existing circle
	circleID := r.URL.Query().Get("id")

	var data map[string]interface{}

	if circleID != "" {
		// Get the circle for editing
		circle, err := h.circleService.GetCircleByID(r.Context(), circleID)
		if err != nil {
			h.logger.Error("failed to get circle",
				zap.String("circle_id", circleID),
				zap.Error(err),
			)
			RenderError(w, h.logger, http.StatusInternalServerError, "Failed to load circle")
			return
		}

		if circle == nil {
			RenderError(w, h.logger, http.StatusNotFound, "Circle not found")
			return
		}

		data = map[string]interface{}{
			"CircleID":       circle.ID,
			"Name":           circle.Name,
			"Description":    circle.Description,
			"Visibility":     circle.Visibility,
			"AutoModEnabled": circle.AutoModEnabled,
			"Icon":           circle.Icon,
			"IconBgColor":    circle.IconBgColor,
		}
	} else {
		// New circle form
		data = map[string]interface{}{
			"CircleID":       "",
			"Name":           "",
			"Description":    "",
			"Visibility":     "private",
			"AutoModEnabled": false,
			"Icon":           "",
			"IconBgColor":    "",
		}
	}

	RenderFragment(w, h.logger, "circle-form", data)
}

// handleAPIInviteMemberForm renders the invite member form in a modal
func (h *CircleHandler) handleAPIInviteMemberForm(w http.ResponseWriter, r *http.Request, circleID string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if user == nil {
		RenderError(w, h.logger, http.StatusUnauthorized, "Unauthorized")
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		RenderError(w, h.logger, http.StatusBadRequest, "No active profile")
		return
	}

	profileID := *session.ActiveProfileID

	// Verify user has permission to invite
	canInvite, err := h.circleService.CanInvite(r.Context(), circleID, profileID)
	if err != nil {
		h.logger.Error("failed to check invite permission",
			zap.String("circle_id", circleID),
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		RenderError(w, h.logger, http.StatusInternalServerError, "Failed to check permissions")
		return
	}

	if !canInvite {
		RenderError(w, h.logger, http.StatusForbidden, "You don't have permission to invite members")
		return
	}

	data := map[string]interface{}{
		"CircleID": circleID,
	}

	RenderFragment(w, h.logger, "invite-member-form", data)
}

// handleAPICreateCircle handles circle creation via HTMX
func (h *CircleHandler) handleAPICreateCircle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if user == nil {
		RenderError(w, h.logger, http.StatusUnauthorized, "Unauthorized")
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		RenderError(w, h.logger, http.StatusBadRequest, "No active profile")
		return
	}

	profileID := *session.ActiveProfileID

	// Parse multipart form (max 15MB for avatar 5MB + banner 10MB)
	if err := r.ParseMultipartForm(15 << 20); err != nil {
		RenderError(w, h.logger, http.StatusBadRequest, "Invalid form data")
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	description := strings.TrimSpace(r.FormValue("description"))
	visibility := r.FormValue("visibility")
	autoModEnabled := r.FormValue("auto_mod_enabled") == "on"
	icon := strings.TrimSpace(r.FormValue("icon"))
	iconBgColor := r.FormValue("icon_bg_color")

	h.logger.Debug("creating circle with form data",
		zap.String("name", name),
		zap.String("description", description),
		zap.String("visibility", visibility),
		zap.Bool("auto_mod_enabled", autoModEnabled),
		zap.String("icon", icon),
		zap.String("icon_bg_color", iconBgColor),
	)

	// Validate
	if name == "" {
		RenderValidationError(w, h.logger, "name", "Circle name is required")
		return
	}

	// Validate icon length (optional field)
	if icon != "" && len([]rune(icon)) > 2 {
		RenderValidationError(w, h.logger, "icon", "Icon must be 2 characters or less")
		return
	}

	// Validate iconBgColor (optional field)
	if iconBgColor != "" {
		validColors := []string{"warm-accent", "cool-accent", "earth-accent", "sage-accent"}
		valid := false
		for _, color := range validColors {
			if iconBgColor == color {
				valid = true
				iconBgColor = fmt.Sprintf("var(--%s)", iconBgColor)
				break
			}
		}
		if !valid {
			RenderValidationError(w, h.logger, "icon_bg_color", "Invalid icon background color")
			return
		}
	}

	// Create circle
	circle, err := h.circleService.CreateCircle(
		r.Context(),
		name,
		description,
		visibility,
		autoModEnabled,
		icon,
		iconBgColor,
		profileID,
	)
	if err != nil {
		h.logger.Error("failed to create circle",
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		RenderError(w, h.logger, http.StatusBadRequest, err.Error())
		return
	}

	h.logger.Info("circle created via API",
		zap.String("circle_id", circle.ID),
		zap.String("profile_id", profileID),
	)

	// Handle optional file uploads
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		// Upload avatar if provided
		if avatarFiles, ok := r.MultipartForm.File["avatar-file-input"]; ok && len(avatarFiles) > 0 {
			avatarFile, err := avatarFiles[0].Open()
			if err == nil {
				defer avatarFile.Close()
				if err := h.circleService.UpdateAvatar(r.Context(), circle.ID, avatarFile, avatarFiles[0], profileID); err != nil {
					h.logger.Warn("failed to upload avatar during circle creation",
						zap.String("circle_id", circle.ID),
						zap.Error(err),
					)
				}
			}
		}

		// Upload banner if provided
		if bannerFiles, ok := r.MultipartForm.File["banner-file-input"]; ok && len(bannerFiles) > 0 {
			bannerFile, err := bannerFiles[0].Open()
			if err == nil {
				defer bannerFile.Close()
				if err := h.circleService.UpdateBanner(r.Context(), circle.ID, bannerFile, bannerFiles[0], profileID); err != nil {
					h.logger.Warn("failed to upload banner during circle creation",
						zap.String("circle_id", circle.ID),
						zap.Error(err),
					)
				}
			}
		}
	}

	// Redirect to circle page via HX-Redirect header
	w.Header().Set("HX-Redirect", fmt.Sprintf("/circles/%s", circle.ID))
	w.WriteHeader(http.StatusOK)
}

// handleAPICircleActions handles various circle API actions
func (h *CircleHandler) handleAPICircleActions(w http.ResponseWriter, r *http.Request) {
	// Extract circle ID from path /api/circles/{id}/...
	path := strings.TrimPrefix(r.URL.Path, "/api/circles/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	circleID := parts[0]

	// Route to specific action
	if len(parts) == 1 {
		// /api/circles/{id} - Update or delete
		switch r.Method {
		case http.MethodPost:
			h.handleAPIUpdateCircle(w, r, circleID)
		case http.MethodDelete:
			h.handleAPIDeleteCircle(w, r, circleID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else if len(parts) >= 2 {
		action := parts[1]
		switch action {
		case "form":
			// GET /api/circles/{id}/form - Get edit form
			if r.Method == http.MethodGet {
				r.URL.RawQuery = "id=" + circleID
				h.handleAPICircleForm(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		case "leave":
			// POST /api/circles/{id}/leave
			if r.Method == http.MethodPost {
				h.handleAPILeaveCircle(w, r, circleID)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		case "join":
			// POST /api/circles/{id}/join
			if r.Method == http.MethodPost {
				h.handleAPIJoinCircle(w, r, circleID)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		case "invite":
			// Handle invitation sub-routes
			if len(parts) >= 3 {
				inviteAction := parts[2]
				switch inviteAction {
				case "accept":
					// POST /api/circles/{id}/invite/accept
					if r.Method == http.MethodPost {
						h.handleAPIAcceptInvitation(w, r, circleID)
					} else {
						http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
					}
				case "decline":
					// POST /api/circles/{id}/invite/decline
					if r.Method == http.MethodPost {
						h.handleAPIDeclineInvitation(w, r, circleID)
					} else {
						http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
					}
				case "form":
					// GET /api/circles/{id}/invite/form
					if r.Method == http.MethodGet {
						h.handleAPIInviteMemberForm(w, r, circleID)
					} else {
						http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
					}
				default:
					http.Error(w, "Not found", http.StatusNotFound)
				}
			} else if r.Method == http.MethodPost {
				// POST /api/circles/{id}/invite - Send invitation
				h.handleAPIInviteMemberSubmit(w, r, circleID)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		default:
			http.Error(w, "Not found", http.StatusNotFound)
		}
	}
}

// handleAPIUpdateCircle handles updating a circle via HTMX
func (h *CircleHandler) handleAPIUpdateCircle(w http.ResponseWriter, r *http.Request, circleID string) {
	user := auth.GetUser(r.Context())
	if user == nil {
		RenderError(w, h.logger, http.StatusUnauthorized, "Unauthorized")
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		RenderError(w, h.logger, http.StatusBadRequest, "No active profile")
		return
	}

	profileID := *session.ActiveProfileID

	// Parse form
	if err := r.ParseForm(); err != nil {
		RenderError(w, h.logger, http.StatusBadRequest, "Invalid form data")
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if name := strings.TrimSpace(r.FormValue("name")); name != "" {
		updates["name"] = name
	}
	if description := r.FormValue("description"); description != "" {
		updates["description"] = description
	}
	if visibility := r.FormValue("visibility"); visibility != "" {
		updates["visibility"] = visibility
	}
	updates["auto_mod_enabled"] = r.FormValue("auto_mod_enabled") == "on"

	// Update circle
	err := h.circleService.UpdateCircle(r.Context(), circleID, updates, profileID)
	if err != nil {
		h.logger.Error("failed to update circle",
			zap.String("circle_id", circleID),
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		RenderError(w, h.logger, http.StatusBadRequest, err.Error())
		return
	}

	h.logger.Info("circle updated via API",
		zap.String("circle_id", circleID),
		zap.String("profile_id", profileID),
	)

	// Set refresh header before rendering response
	SendHTMXRefresh(w)
	RenderSuccess(w, h.logger, "Circle updated successfully!")
}

// handleAPIDeleteCircle handles deleting a circle via HTMX
func (h *CircleHandler) handleAPIDeleteCircle(w http.ResponseWriter, r *http.Request, circleID string) {
	user := auth.GetUser(r.Context())
	if user == nil {
		RenderError(w, h.logger, http.StatusUnauthorized, "Unauthorized")
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		RenderError(w, h.logger, http.StatusBadRequest, "No active profile")
		return
	}

	profileID := *session.ActiveProfileID

	err := h.circleService.DeleteCircle(r.Context(), circleID, profileID)
	if err != nil {
		h.logger.Error("failed to delete circle",
			zap.String("circle_id", circleID),
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		RenderError(w, h.logger, http.StatusBadRequest, err.Error())
		return
	}

	h.logger.Info("circle deleted via API",
		zap.String("circle_id", circleID),
		zap.String("profile_id", profileID),
	)

	// Return success - the card will be removed by HTMX
	RenderSuccess(w, h.logger, "Circle deleted successfully")
	w.WriteHeader(http.StatusOK)
}

// handleAPILeaveCircle handles leaving a circle via HTMX
func (h *CircleHandler) handleAPILeaveCircle(w http.ResponseWriter, r *http.Request, circleID string) {
	user := auth.GetUser(r.Context())
	if user == nil {
		RenderError(w, h.logger, http.StatusUnauthorized, "Unauthorized")
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		RenderError(w, h.logger, http.StatusBadRequest, "No active profile")
		return
	}

	profileID := *session.ActiveProfileID

	err := h.circleService.LeaveCircle(r.Context(), circleID, profileID)
	if err != nil {
		h.logger.Error("failed to leave circle",
			zap.String("circle_id", circleID),
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		RenderError(w, h.logger, http.StatusBadRequest, err.Error())
		return
	}

	h.logger.Info("member left circle via API",
		zap.String("circle_id", circleID),
		zap.String("profile_id", profileID),
	)

	// Return success - the card will be removed by HTMX
	RenderSuccess(w, h.logger, "You left the circle")
	w.WriteHeader(http.StatusOK)
}

// handleAPIJoinCircle handles joining a public circle via HTMX
func (h *CircleHandler) handleAPIJoinCircle(w http.ResponseWriter, r *http.Request, circleID string) {
	user := auth.GetUser(r.Context())
	if user == nil {
		RenderError(w, h.logger, http.StatusUnauthorized, "Unauthorized")
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		RenderError(w, h.logger, http.StatusBadRequest, "No active profile")
		return
	}

	profileID := *session.ActiveProfileID

	err := h.circleService.JoinPublicCircle(r.Context(), circleID, profileID)
	if err != nil {
		h.logger.Error("failed to join circle",
			zap.String("circle_id", circleID),
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		RenderError(w, h.logger, http.StatusBadRequest, err.Error())
		return
	}

	h.logger.Info("member joined circle via API",
		zap.String("circle_id", circleID),
		zap.String("profile_id", profileID),
	)

	// Return the "joined" button state
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<button class="join-btn joined" disabled>Joined</button>`))
}

// handleAPIAcceptInvitation handles accepting a circle invitation via HTMX
func (h *CircleHandler) handleAPIAcceptInvitation(w http.ResponseWriter, r *http.Request, circleID string) {
	user := auth.GetUser(r.Context())
	if user == nil {
		RenderError(w, h.logger, http.StatusUnauthorized, "Unauthorized")
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		RenderError(w, h.logger, http.StatusBadRequest, "No active profile")
		return
	}

	profileID := *session.ActiveProfileID

	err := h.circleService.AcceptInvitation(r.Context(), circleID, profileID)
	if err != nil {
		h.logger.Error("failed to accept invitation",
			zap.String("circle_id", circleID),
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		RenderError(w, h.logger, http.StatusBadRequest, err.Error())
		return
	}

	h.logger.Info("invitation accepted via API",
		zap.String("circle_id", circleID),
		zap.String("profile_id", profileID),
	)

	// Return success message that will replace the invitation card
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<div class="card" style="padding: 1rem; background: var(--success-bg); border: 1px solid var(--success-border); color: var(--success-text);">
		<p style="margin: 0;">✓ Invitation accepted! Refresh to see the circle in your list.</p>
	</div>`))
}

// handleAPIDeclineInvitation handles declining a circle invitation via HTMX
func (h *CircleHandler) handleAPIDeclineInvitation(w http.ResponseWriter, r *http.Request, circleID string) {
	user := auth.GetUser(r.Context())
	if user == nil {
		RenderError(w, h.logger, http.StatusUnauthorized, "Unauthorized")
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		RenderError(w, h.logger, http.StatusBadRequest, "No active profile")
		return
	}

	profileID := *session.ActiveProfileID

	err := h.circleService.DeclineInvitation(r.Context(), circleID, profileID)
	if err != nil {
		h.logger.Error("failed to decline invitation",
			zap.String("circle_id", circleID),
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		RenderError(w, h.logger, http.StatusBadRequest, err.Error())
		return
	}

	h.logger.Info("invitation declined via API",
		zap.String("circle_id", circleID),
		zap.String("profile_id", profileID),
	)

	// Return empty div to remove the invitation card
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(``))
}

// handleAPIInviteMemberSubmit handles sending a circle invitation via HTMX
func (h *CircleHandler) handleAPIInviteMemberSubmit(w http.ResponseWriter, r *http.Request, circleID string) {
	user := auth.GetUser(r.Context())
	if user == nil {
		RenderError(w, h.logger, http.StatusUnauthorized, "Unauthorized")
		return
	}

	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		RenderError(w, h.logger, http.StatusBadRequest, "No active profile")
		return
	}

	profileID := *session.ActiveProfileID

	// Parse form
	if err := r.ParseForm(); err != nil {
		RenderError(w, h.logger, http.StatusBadRequest, "Invalid form data")
		return
	}

	username := r.FormValue("username")
	role := r.FormValue("role")

	if username == "" {
		RenderError(w, h.logger, http.StatusBadRequest, "Username is required")
		return
	}

	// Look up the profile by handle (username)
	inviteeProfile, err := h.profileService.GetByHandle(r.Context(), username)
	if err != nil {
		h.logger.Error("failed to find profile by handle",
			zap.String("handle", username),
			zap.Error(err),
		)
		RenderError(w, h.logger, http.StatusBadRequest, "User not found")
		return
	}

	if inviteeProfile == nil {
		RenderError(w, h.logger, http.StatusBadRequest, "User not found")
		return
	}

	// Send the invitation (role is captured for logging but not used yet in backend)
	err = h.circleService.InviteMember(r.Context(), circleID, inviteeProfile.ID, profileID)
	if err != nil {
		h.logger.Error("failed to invite member",
			zap.String("circle_id", circleID),
			zap.String("invitee_profile_id", inviteeProfile.ID),
			zap.String("role", role),
			zap.Error(err),
		)
		RenderError(w, h.logger, http.StatusBadRequest, err.Error())
		return
	}

	h.logger.Info("member invited via API",
		zap.String("circle_id", circleID),
		zap.String("invitee_profile_id", inviteeProfile.ID),
		zap.String("handle", username),
		zap.String("role", role),
	)

	// Return success message that replaces the modal
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<div style="padding: 2rem; text-align: center;">
		<div style="font-size: 3rem; margin-bottom: 1rem;">✓</div>
		<h3 style="margin: 0 0 0.5rem 0; font-size: 1.25rem; font-weight: 600;">Invitation sent!</h3>
		<p style="margin: 0; color: var(--text-muted);">@` + username + ` will receive the invitation.</p>
		<button onclick="htmx.find('#modal-body').innerHTML = ''; location.reload();" class="btn-primary" style="margin-top: 1.5rem; padding: 0.75rem 1.5rem;">
			Done
		</button>
	</div>`))
}
