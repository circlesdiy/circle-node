package handlers

import (
	"log"
	"net/http"

	"circles.diy/internal/auth"
	"circles.diy/internal/gather"
	"circles.diy/internal/middleware"
	"circles.diy/internal/models"
	"circles.diy/internal/profile"
	"circles.diy/internal/templates"
)

type GatherHandler struct {
	service        *gather.Service
	profileService *profile.Service
}

func NewGatherHandler(service *gather.Service, profileService *profile.Service) *GatherHandler {
	return &GatherHandler{
		service:        service,
		profileService: profileService,
	}
}

func (h *GatherHandler) Handle(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get active profile
	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		http.Error(w, "No active profile", http.StatusBadRequest)
		return
	}

	profileID := *session.ActiveProfileID

	// Get profile to display profile name
	var profileName string
	profile, err := h.profileService.GetByID(r.Context(), profileID)
	if err != nil {
		log.Printf("failed to get profile: %v", err)
		profileName = user.Username
	} else if profile != nil && profile.Name != "" {
		profileName = profile.DisplayName
	} else {
		profileName = user.Username
	}

	// Get gather page data from service using profile ID
	data, err := h.service.GetGatherPageData(r.Context(), profileID)
	if err != nil {
		log.Printf("gather error: %v", err)
		data = h.getEmptyGatherData()
	}

	// Set base data
	data.BaseData = models.BaseData{
		Title:       "Gather",
		ActiveNav:   "gather",
		CSRFToken:   middleware.GetCSRFToken(r),
		User:        user,
		ProfileName: profileName,
	}

	// Render the gather template
	if err := templates.GetTemplates().Gather.ExecuteTemplate(w, "gather", data); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *GatherHandler) getEmptyGatherData() *models.GatherPageData {
	return &models.GatherPageData{
		CircleGatherings:   []models.CircleGathering{},
		DomainDiscovery:    []models.DomainSection{},
		UserCircles:        []models.CircleOption{},
		ActiveCoordination: []models.CoordinationItem{},
		UserHosting:        []models.HostingEvent{},
	}
}
