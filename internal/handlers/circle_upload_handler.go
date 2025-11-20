package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"circles.diy/internal/auth"
	"circles.diy/internal/circle"
	"go.uber.org/zap"
)

// CircleUploadHandler handles circle image upload operations
type CircleUploadHandler struct {
	circleService *circle.Service
	logger        *zap.Logger
}

// NewCircleUploadHandler creates a new circle upload handler
func NewCircleUploadHandler(circleService *circle.Service, logger *zap.Logger) *CircleUploadHandler {
	return &CircleUploadHandler{
		circleService: circleService,
		logger:        logger,
	}
}

// UploadAvatar handles POST /api/circles/{id}/avatar
func (h *CircleUploadHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get authenticated session
	session := auth.GetSession(ctx)
	if session == nil || session.ActiveProfileID == nil {
		h.logger.Warn("upload circle avatar attempted without session")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	requesterProfileID := *session.ActiveProfileID

	// Extract circle ID from URL path
	circleID := extractCircleIDFromPath(r.URL.Path)
	if circleID == "" {
		http.Error(w, "Invalid circle ID", http.StatusBadRequest)
		return
	}

	// Parse multipart form (max 32MB in memory)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		h.logger.Error("failed to parse multipart form",
			zap.String("circle_id", circleID),
			zap.Error(err))
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("avatar")
	if err != nil {
		h.logger.Error("failed to get avatar file from form",
			zap.String("circle_id", circleID),
			zap.Error(err))
		http.Error(w, "No avatar file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Upload avatar via service (service will verify ownership)
	if err := h.circleService.UpdateAvatar(ctx, circleID, file, header, requesterProfileID); err != nil {
		h.logger.Error("failed to upload circle avatar",
			zap.String("circle_id", circleID),
			zap.String("requester_profile_id", requesterProfileID),
			zap.Error(err))

		// Check if it's a permission error
		if strings.Contains(err.Error(), "only circle owner") {
			http.Error(w, "Only circle owner can update avatar", http.StatusForbidden)
			return
		}

		http.Error(w, "Failed to upload avatar: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get updated circle to return new avatar URL
	updatedCircle, err := h.circleService.GetCircleByID(ctx, circleID)
	if err != nil {
		h.logger.Error("failed to get updated circle",
			zap.String("circle_id", circleID),
			zap.Error(err))
		http.Error(w, "Failed to retrieve updated circle", http.StatusInternalServerError)
		return
	}

	// Return success with new avatar URL
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"avatar_url": updatedCircle.AvatarURL,
		"message":    "Avatar uploaded successfully",
	})
}

// UploadBanner handles POST /api/circles/{id}/banner
func (h *CircleUploadHandler) UploadBanner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get authenticated session
	session := auth.GetSession(ctx)
	if session == nil || session.ActiveProfileID == nil {
		h.logger.Warn("upload circle banner attempted without session")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	requesterProfileID := *session.ActiveProfileID

	// Extract circle ID from URL path
	circleID := extractCircleIDFromPath(r.URL.Path)
	if circleID == "" {
		http.Error(w, "Invalid circle ID", http.StatusBadRequest)
		return
	}

	// Parse multipart form (max 32MB in memory)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		h.logger.Error("failed to parse multipart form",
			zap.String("circle_id", circleID),
			zap.Error(err))
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("banner")
	if err != nil {
		h.logger.Error("failed to get banner file from form",
			zap.String("circle_id", circleID),
			zap.Error(err))
		http.Error(w, "No banner file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Upload banner via service (service will verify ownership)
	if err := h.circleService.UpdateBanner(ctx, circleID, file, header, requesterProfileID); err != nil {
		h.logger.Error("failed to upload circle banner",
			zap.String("circle_id", circleID),
			zap.String("requester_profile_id", requesterProfileID),
			zap.Error(err))

		// Check if it's a permission error
		if strings.Contains(err.Error(), "only circle owner") {
			http.Error(w, "Only circle owner can update banner", http.StatusForbidden)
			return
		}

		http.Error(w, "Failed to upload banner: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get updated circle to return new banner URL
	updatedCircle, err := h.circleService.GetCircleByID(ctx, circleID)
	if err != nil {
		h.logger.Error("failed to get updated circle",
			zap.String("circle_id", circleID),
			zap.Error(err))
		http.Error(w, "Failed to retrieve updated circle", http.StatusInternalServerError)
		return
	}

	// Return success with new banner URL
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"banner_url": updatedCircle.BannerURL,
		"message":    "Banner uploaded successfully",
	})
}

// DeleteAvatar handles DELETE /api/circles/{id}/avatar
func (h *CircleUploadHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get authenticated session
	session := auth.GetSession(ctx)
	if session == nil || session.ActiveProfileID == nil {
		h.logger.Warn("delete circle avatar attempted without session")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	requesterProfileID := *session.ActiveProfileID

	// Extract circle ID from URL path
	circleID := extractCircleIDFromPath(r.URL.Path)
	if circleID == "" {
		http.Error(w, "Invalid circle ID", http.StatusBadRequest)
		return
	}

	// Remove avatar via service (service will verify ownership)
	if err := h.circleService.RemoveAvatar(ctx, circleID, requesterProfileID); err != nil {
		h.logger.Error("failed to delete circle avatar",
			zap.String("circle_id", circleID),
			zap.String("requester_profile_id", requesterProfileID),
			zap.Error(err))

		// Check if it's a permission error
		if strings.Contains(err.Error(), "only circle owner") {
			http.Error(w, "Only circle owner can remove avatar", http.StatusForbidden)
			return
		}

		http.Error(w, "Failed to delete avatar: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Avatar deleted successfully",
	})
}

// DeleteBanner handles DELETE /api/circles/{id}/banner
func (h *CircleUploadHandler) DeleteBanner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get authenticated session
	session := auth.GetSession(ctx)
	if session == nil || session.ActiveProfileID == nil {
		h.logger.Warn("delete circle banner attempted without session")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	requesterProfileID := *session.ActiveProfileID

	// Extract circle ID from URL path
	circleID := extractCircleIDFromPath(r.URL.Path)
	if circleID == "" {
		http.Error(w, "Invalid circle ID", http.StatusBadRequest)
		return
	}

	// Remove banner via service (service will verify ownership)
	if err := h.circleService.RemoveBanner(ctx, circleID, requesterProfileID); err != nil {
		h.logger.Error("failed to delete circle banner",
			zap.String("circle_id", circleID),
			zap.String("requester_profile_id", requesterProfileID),
			zap.Error(err))

		// Check if it's a permission error
		if strings.Contains(err.Error(), "only circle owner") {
			http.Error(w, "Only circle owner can remove banner", http.StatusForbidden)
			return
		}

		http.Error(w, "Failed to delete banner: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Banner deleted successfully",
	})
}

// extractCircleIDFromPath extracts the circle ID from URL paths like /api/circles/{id}/avatar
func extractCircleIDFromPath(path string) string {
	// Split path: /api/circles/{id}/avatar -> ["", "api", "circles", "{id}", "avatar"]
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 && parts[0] == "api" && parts[1] == "circles" {
		return parts[2]
	}
	return ""
}
