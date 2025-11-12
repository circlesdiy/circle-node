package handlers

import (
	"log"
	"net/http"

	"circles.diy/internal/auth"
	"circles.diy/internal/dashboard"
	"circles.diy/internal/middleware"
	"circles.diy/internal/models"
	"circles.diy/internal/profile"
	"circles.diy/internal/templates"
)

type DashboardHandler struct {
	service        *dashboard.Service
	profileService *profile.Service
}

func NewDashboardHandler(service *dashboard.Service, profileService *profile.Service) *DashboardHandler {
	return &DashboardHandler{
		service:        service,
		profileService: profileService,
	}
}

func (h *DashboardHandler) Handle(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get active profile to display profile name
	session := auth.GetSession(r.Context())
	var profileName string
	if session != nil && session.ActiveProfileID != nil {
		profile, err := h.profileService.GetByID(r.Context(), *session.ActiveProfileID)
		if err != nil {
			log.Printf("failed to get profile: %v", err)
		} else if profile != nil && profile.Name != "" {
			profileName = profile.DisplayName
		}
	}

	// If no profile name found, fall back to username
	if profileName == "" {
		profileName = user.Username
	}

	data, err := h.service.GetDashboardData(r.Context(), user.ID)
	if err != nil {
		log.Printf("dashboard error: %v", err)
		data = h.getEmptyDashboard()
	}

	data.BaseData = models.BaseData{
		Title:       "Dashboard",
		ActiveNav:   "dashboard",
		CSRFToken:   middleware.GetCSRFToken(r),
		User:        user,
		ProfileName: profileName,
	}

	if err := templates.GetTemplates().Dashboard.ExecuteTemplate(w, "dashboard", data); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *DashboardHandler) getEmptyDashboard() *models.DashboardData {
	return &models.DashboardData{
		HappeningNow:   []models.HappeningNowItem{},
		UpcomingEvents: []models.UpcomingEvent{},
		CirclesSummary: []models.CircleSummary{},
	}
}
