package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"circles.diy/internal/auth"
	"circles.diy/internal/events"
	"circles.diy/internal/gather"
	"go.uber.org/zap"
)

type GatherAPIHandler struct {
	service      *gather.Service
	eventService *events.Service
	logger       *zap.Logger
}

func NewGatherAPIHandler(service *gather.Service, eventService *events.Service, logger *zap.Logger) *GatherAPIHandler {
	return &GatherAPIHandler{
		service:      service,
		eventService: eventService,
		logger:       logger,
	}
}

// HandleRSVP handles POST /api/events/{id}/rsvp
func (h *GatherAPIHandler) HandleRSVP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get active profile
	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		RenderError(w, h.logger, http.StatusBadRequest, "No active profile")
		return
	}

	profileID := *session.ActiveProfileID

	// Extract event ID from URL path
	// URL format: /api/events/{id}/rsvp
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		RenderError(w, h.logger, http.StatusBadRequest, "Invalid URL")
		return
	}
	eventID := pathParts[2]

	// Parse request body
	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RenderError(w, h.logger, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Update RSVP
	if err := h.service.UpdateRSVP(r.Context(), eventID, profileID, req.Status); err != nil {
		h.logger.Error("failed to update RSVP",
			zap.String("event_id", eventID),
			zap.String("profile_id", profileID),
			zap.String("status", req.Status),
			zap.Error(err),
		)
		RenderError(w, h.logger, http.StatusInternalServerError, "Failed to update RSVP")
		return
	}

	// Get updated attendee count
	count, err := h.service.GetEventAttendeeCount(r.Context(), eventID)
	if err != nil {
		h.logger.Error("failed to get attendee count",
			zap.String("event_id", eventID),
			zap.Error(err),
		)
		count = 0
	}

	// Return success with updated count
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"status":        req.Status,
		"attendeeCount": count,
	})
}

// HandleCreateEvent handles POST /api/events
func (h *GatherAPIHandler) HandleCreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get active profile
	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		RenderError(w, h.logger, http.StatusBadRequest, "No active profile")
		return
	}

	profileID := *session.ActiveProfileID

	// Parse form data
	if err := r.ParseForm(); err != nil {
		RenderValidationError(w, h.logger, "form", "Invalid form data")
		return
	}

	// Extract and validate required fields
	circleID := r.FormValue("circle_id")
	title := r.FormValue("title")
	startTimeStr := r.FormValue("start_time")

	if circleID == "" || title == "" || startTimeStr == "" {
		RenderValidationError(w, h.logger, "form", "Missing required fields: circle, title, and start time are required")
		return
	}

	// Parse start time
	startTime, err := time.Parse("2006-01-02T15:04", startTimeStr)
	if err != nil {
		RenderValidationError(w, h.logger, "start_time", "Invalid start time format")
		return
	}

	// Parse optional end time
	var endTime time.Time
	if endTimeStr := r.FormValue("end_time"); endTimeStr != "" {
		t, err := time.Parse("2006-01-02T15:04", endTimeStr)
		if err != nil {
			RenderValidationError(w, h.logger, "end_time", "Invalid end time format")
			return
		}
		endTime = t
	}

	// Parse optional capacity
	var capacity int
	if capacityStr := r.FormValue("capacity"); capacityStr != "" {
		err := json.Unmarshal([]byte(capacityStr), &capacity)
		if err != nil {
			capacity = 0
		}
	}

	// Create event via service
	event, err := h.eventService.CreateEvent(
		r.Context(),
		title,
		r.FormValue("location"),
		r.FormValue("timezone"),
		startTime,
		endTime,
		circleID,
		profileID,
		capacity,
	)
	if err != nil {
		h.logger.Error("failed to create event",
			zap.String("profile_id", profileID),
			zap.String("circle_id", circleID),
			zap.Error(err),
		)
		RenderError(w, h.logger, http.StatusInternalServerError, "Failed to create event")
		return
	}

	// Redirect to event detail page
	w.Header().Set("HX-Redirect", "/gather/"+event.ID)
	w.WriteHeader(http.StatusOK)
}

// HandleUpdateEvent handles PUT /api/events/{id}
func (h *GatherAPIHandler) HandleUpdateEvent(w http.ResponseWriter, r *http.Request, eventID string) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get active profile
	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		RenderError(w, h.logger, http.StatusBadRequest, "No active profile")
		return
	}

	profileID := *session.ActiveProfileID

	// Parse form data
	if err := r.ParseForm(); err != nil {
		RenderValidationError(w, h.logger, "form", "Invalid form data")
		return
	}

	// Extract fields
	title := r.FormValue("title")
	location := r.FormValue("location")
	timezone := r.FormValue("timezone")
	startTimeStr := r.FormValue("start_time")
	endTimeStr := r.FormValue("end_time")
	capacityStr := r.FormValue("capacity")

	// Validate required fields
	if title == "" || startTimeStr == "" {
		RenderValidationError(w, h.logger, "form", "Title and start time are required")
		return
	}

	// Parse start time
	startTime, err := time.Parse("2006-01-02T15:04", startTimeStr)
	if err != nil {
		RenderValidationError(w, h.logger, "start_time", "Invalid start time format")
		return
	}

	// Parse optional end time
	var endTime time.Time
	if endTimeStr != "" {
		t, err := time.Parse("2006-01-02T15:04", endTimeStr)
		if err != nil {
			RenderValidationError(w, h.logger, "end_time", "Invalid end time format")
			return
		}
		endTime = t
	}

	// Parse optional capacity
	var capacity int
	if capacityStr != "" {
		err := json.Unmarshal([]byte(capacityStr), &capacity)
		if err != nil {
			capacity = 0
		}
	}

	// Update event via service (service will verify organizer ownership)
	if err := h.eventService.UpdateEvent(r.Context(), eventID, title, location, timezone, startTime, endTime, capacity, profileID); err != nil {
		h.logger.Error("failed to update event",
			zap.String("event_id", eventID),
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		RenderError(w, h.logger, http.StatusInternalServerError, "Failed to update event")
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Event updated successfully",
	})
}

// HandleDeleteEvent handles DELETE /api/events/{id}
func (h *GatherAPIHandler) HandleDeleteEvent(w http.ResponseWriter, r *http.Request, eventID string) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get active profile
	session := auth.GetSession(r.Context())
	if session == nil || session.ActiveProfileID == nil {
		RenderError(w, h.logger, http.StatusBadRequest, "No active profile")
		return
	}

	profileID := *session.ActiveProfileID

	// Delete event via service (includes organizer verification)
	if err := h.eventService.DeleteEvent(r.Context(), eventID, profileID); err != nil {
		h.logger.Error("failed to delete event",
			zap.String("event_id", eventID),
			zap.String("profile_id", profileID),
			zap.Error(err),
		)
		RenderError(w, h.logger, http.StatusInternalServerError, "Failed to delete event")
		return
	}

	// Redirect to gather page
	w.Header().Set("HX-Redirect", "/gather")
	w.WriteHeader(http.StatusOK)
}
