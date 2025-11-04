package handlers

import (
	"log"
	"net/http"

	"circles.diy/internal/auth"
	"circles.diy/internal/middleware"
	"circles.diy/internal/templates"
)

func GatherHandler(w http.ResponseWriter, r *http.Request) {
	// Get mock gather data
	data := templates.GetMockGatherData()

	// Add user from context to template data
	data.User = auth.GetUser(r.Context())
	data.CSRFToken = middleware.GetCSRFToken(r)

	// Render the gather template
	err := templates.GetTemplates().Gather.ExecuteTemplate(w, "gather", data)
	if err != nil {
		log.Printf("Error rendering gather template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
