package handlers

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"circles.diy/internal/auth"
	"circles.diy/internal/domain"
	"circles.diy/internal/events"
	"circles.diy/internal/gather"
	"circles.diy/internal/middleware"
	"circles.diy/internal/models"
	"circles.diy/internal/profile"
	"circles.diy/internal/templates"
)

type GatherHandler struct {
	service        *gather.Service
	eventService   *events.Service
	profileService *profile.Service
}

func NewGatherHandler(service *gather.Service, eventService *events.Service, profileService *profile.Service) *GatherHandler {
	return &GatherHandler{
		service:        service,
		eventService:   eventService,
		profileService: profileService,
	}
}

func (h *GatherHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Route to specific handlers based on path
	path := strings.TrimPrefix(r.URL.Path, "/gather")

	switch {
	case path == "" || path == "/":
		h.handleListPage(w, r)
	case path == "/create":
		h.handleCreatePage(w, r)
	case strings.HasPrefix(path, "/"):
		// Extract event ID from path
		parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
		if len(parts) == 0 {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		eventID := parts[0]

		if len(parts) == 1 {
			h.handleEventDetail(w, r, eventID)
		} else if len(parts) >= 2 && parts[1] == "manage" {
			h.handleEventManage(w, r, eventID)
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func (h *GatherHandler) handleListPage(w http.ResponseWriter, r *http.Request) {
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

func (h *GatherHandler) handleCreatePage(w http.ResponseWriter, r *http.Request) {
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

	// Get profile name
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

	// Get user circles for dropdown
	circles, err := h.service.GetUserCircles(r.Context(), profileID)
	if err != nil {
		log.Printf("failed to get user circles: %v", err)
		http.Error(w, "Failed to load circles", http.StatusInternalServerError)
		return
	}

	// Check for preselected circle from query param
	preselectedCircleID := r.URL.Query().Get("circle")

	data := &models.EventCreatePageData{
		BaseData: models.BaseData{
			Title:       "Create Gathering",
			ActiveNav:   "gather",
			CSRFToken:   middleware.GetCSRFToken(r),
			User:        user,
			ProfileName: profileName,
		},
		UserCircles:         circles,
		PreselectedCircleID: preselectedCircleID,
	}

	if err := templates.GetTemplates().EventCreate.ExecuteTemplate(w, "event-create", data); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *GatherHandler) handleEventDetail(w http.ResponseWriter, r *http.Request, eventID string) {
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

	// Get profile name
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

	// Get event details
	details, err := h.eventService.GetEventDetails(r.Context(), eventID, profileID)
	if err != nil {
		log.Printf("failed to get event details: %v", err)
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	// Convert to page data
	data := h.buildEventDetailPageData(details, profileID, user, profileName)
	data.CSRFToken = middleware.GetCSRFToken(r)

	if err := templates.GetTemplates().EventDetail.ExecuteTemplate(w, "event-detail", data); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *GatherHandler) handleEventManage(w http.ResponseWriter, r *http.Request, eventID string) {
	// TODO: Implement event management page
	http.Error(w, "Event management coming soon", http.StatusNotImplemented)
}

func (h *GatherHandler) buildEventDetailPageData(details *events.EventDetails, profileID string, user *domain.User, profileName string) *models.EventDetailPageData {
	// Convert attendees
	var attendees []models.EventAttendee
	for _, a := range details.Attendees {
		attendees = append(attendees, models.EventAttendee{
			User: models.User{
				ID:     a.ProfileID,
				Name:   a.Name,
				Avatar: a.AvatarURL,
			},
			RSVPStatus: a.RSVPStatus,
		})
	}

	// Convert coordination needs
	var coordinationNeeds []models.EventCoordinationNeed
	for _, n := range details.CoordinationNeeds {
		coordinationNeeds = append(coordinationNeeds, models.EventCoordinationNeed{
			ID:         n.ID,
			Type:       n.Type,
			Message:    n.Message,
			AuthorName: n.AuthorName,
			CreatedAt:  n.CreatedAt,
			ResolvedAt: n.ResolvedAt,
		})
	}

	return &models.EventDetailPageData{
		BaseData: models.BaseData{
			Title:       details.Event.Title,
			ActiveNav:   "gather",
			User:        user,
			ProfileName: profileName,
		},
		Event:              details.Event,
		CircleName:         details.CircleName,
		OrganizerName:      details.OrganizerName,
		FormattedStartTime: formatEventTime(details.Event.StartTime),
		FormattedEndTime:   formatEventTime(details.Event.EndTime),
		ViewerRSVPStatus:   details.ViewerRSVPStatus,
		IsOrganizer:        details.Event.OrganizerProfileID == profileID,
		AttendeeCount:      len(attendees),
		Attendees:          attendees,
		CoordinationNeeds:  coordinationNeeds,
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

func formatEventTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("Monday, January 2, 2006 at 3:04 PM")
}

// GetUserCircles is a convenience method to get user circles
func (h *GatherHandler) GetUserCircles(ctx context.Context, profileID string) ([]models.CircleOption, error) {
	circles, err := h.service.GetUserCircles(ctx, profileID)
	if err != nil {
		return nil, err
	}
	return circles, nil
}
