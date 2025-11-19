package handlers

import (
	"encoding/json"
	"net/http"

	"circles.diy/internal/auth"
	"circles.diy/internal/profile"
	"go.uber.org/zap"
)

// ProfileUploadHandler handles profile image upload operations
type ProfileUploadHandler struct {
	profileService *profile.Service
	logger         *zap.Logger
}

// NewProfileUploadHandler creates a new profile upload handler
func NewProfileUploadHandler(profileService *profile.Service, logger *zap.Logger) *ProfileUploadHandler {
	return &ProfileUploadHandler{
		profileService: profileService,
		logger:         logger,
	}
}

// UploadAvatar handles POST /api/profile/avatar
func (h *ProfileUploadHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get authenticated session
	session := auth.GetSession(ctx)
	if session == nil || session.ActiveProfileID == nil {
		h.logger.Warn("upload avatar attempted without session")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profileID := *session.ActiveProfileID

	// Parse multipart form (max 32MB in memory)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		h.logger.Error("failed to parse multipart form",
			zap.String("profile_id", profileID),
			zap.Error(err))
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("avatar")
	if err != nil {
		h.logger.Error("failed to get avatar file from form",
			zap.String("profile_id", profileID),
			zap.Error(err))
		http.Error(w, "No avatar file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Upload avatar via service
	if err := h.profileService.UpdateAvatar(ctx, profileID, file, header); err != nil {
		h.logger.Error("failed to upload avatar",
			zap.String("profile_id", profileID),
			zap.Error(err))
		http.Error(w, "Failed to upload avatar: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get updated profile to return new avatar URL
	updatedProfile, err := h.profileService.GetByID(ctx, profileID)
	if err != nil {
		h.logger.Error("failed to get updated profile",
			zap.String("profile_id", profileID),
			zap.Error(err))
		http.Error(w, "Failed to retrieve updated profile", http.StatusInternalServerError)
		return
	}

	// Return success with new avatar URL
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"avatar_url": updatedProfile.AvatarURL,
		"message":    "Avatar uploaded successfully",
	})
}

// UploadBanner handles POST /api/profile/banner
func (h *ProfileUploadHandler) UploadBanner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get authenticated session
	session := auth.GetSession(ctx)
	if session == nil || session.ActiveProfileID == nil {
		h.logger.Warn("upload banner attempted without session")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profileID := *session.ActiveProfileID

	// Parse multipart form (max 32MB in memory)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		h.logger.Error("failed to parse multipart form",
			zap.String("profile_id", profileID),
			zap.Error(err))
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("banner")
	if err != nil {
		h.logger.Error("failed to get banner file from form",
			zap.String("profile_id", profileID),
			zap.Error(err))
		http.Error(w, "No banner file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Upload banner via service
	if err := h.profileService.UpdateBanner(ctx, profileID, file, header); err != nil {
		h.logger.Error("failed to upload banner",
			zap.String("profile_id", profileID),
			zap.Error(err))
		http.Error(w, "Failed to upload banner: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get updated profile to return new banner URL
	updatedProfile, err := h.profileService.GetByID(ctx, profileID)
	if err != nil {
		h.logger.Error("failed to get updated profile",
			zap.String("profile_id", profileID),
			zap.Error(err))
		http.Error(w, "Failed to retrieve updated profile", http.StatusInternalServerError)
		return
	}

	// Return success with new banner URL
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"banner_url": updatedProfile.BannerURL,
		"message":    "Banner uploaded successfully",
	})
}

// DeleteAvatar handles DELETE /api/profile/avatar
func (h *ProfileUploadHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get authenticated session
	session := auth.GetSession(ctx)
	if session == nil || session.ActiveProfileID == nil {
		h.logger.Warn("delete avatar attempted without session")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profileID := *session.ActiveProfileID

	// Remove avatar via service
	if err := h.profileService.RemoveAvatar(ctx, profileID); err != nil {
		h.logger.Error("failed to delete avatar",
			zap.String("profile_id", profileID),
			zap.Error(err))
		http.Error(w, "Failed to delete avatar: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Avatar deleted successfully",
	})
}

// DeleteBanner handles DELETE /api/profile/banner
func (h *ProfileUploadHandler) DeleteBanner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get authenticated session
	session := auth.GetSession(ctx)
	if session == nil || session.ActiveProfileID == nil {
		h.logger.Warn("delete banner attempted without session")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profileID := *session.ActiveProfileID

	// Remove banner via service
	if err := h.profileService.RemoveBanner(ctx, profileID); err != nil {
		h.logger.Error("failed to delete banner",
			zap.String("profile_id", profileID),
			zap.Error(err))
		http.Error(w, "Failed to delete banner: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Banner deleted successfully",
	})
}
