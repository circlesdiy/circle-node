package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"circles.diy/internal/auth"
	"circles.diy/internal/circle"
	"circles.diy/internal/domain"
	"circles.diy/internal/middleware"
	"circles.diy/internal/models"
	"circles.diy/internal/preferences"
	"circles.diy/internal/templates"
	"go.uber.org/zap"
)

// CircleHandler handles circle-related HTTP requests
type CircleHandler struct {
	circleService *circle.Service
	prefsService  *preferences.Service
	logger        *zap.Logger
}

// NewCircleHandler creates a new circle handler
func NewCircleHandler(
	circleService *circle.Service,
	prefsService *preferences.Service,
	logger *zap.Logger,
) *CircleHandler {
	return &CircleHandler{
		circleService: circleService,
		prefsService:  prefsService,
		logger:        logger,
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
			default:
				http.Error(w, "Not found", http.StatusNotFound)
			}
		}
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// RegisterRoutes registers all circle routes with the mux
func (h *CircleHandler) RegisterRoutes(mux *http.ServeMux, authHandler *auth.Handler) {
	mux.HandleFunc("/circles", authHandler.RequireAuth(h.Handle))
	mux.HandleFunc("/circles/", authHandler.RequireAuth(h.Handle))
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
			User:      user,
			CSRFToken: middleware.GetCSRFToken(r),
		},
		Circles:         templateCircles,
		FeaturedCircles: featuredCircles,
		RecentActivity:  []models.CircleActivity{}, // TODO: Implement activity feed
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

	// For now, just return a simple response
	// TODO: Implement full circle detail page with posts, members, etc.
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
		<!DOCTYPE html>
		<html>
		<head><title>` + circle.Name + `</title></head>
		<body>
			<h1>` + circle.Name + `</h1>
			<p>` + circle.Description + `</p>
			<p>Visibility: ` + circle.Visibility + `</p>
			<p><a href="/circles">Back to Circles</a></p>
		</body>
		</html>
	`))
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

		// Create circle
		circle, err := h.circleService.CreateCircle(
			r.Context(),
			name,
			description,
			visibility,
			autoModEnabled,
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
