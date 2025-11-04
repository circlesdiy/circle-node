package handlers

import (
	"log"
	"net/http"

	"circles.diy/internal/auth"
	"circles.diy/internal/middleware"
	"circles.diy/internal/templates"
)

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	data := templates.GetMockDashboardData()

	// Add user from context to template data
	data.User = auth.GetUser(r.Context())

	// Add CSRF token to template data
	data.CSRFToken = middleware.GetCSRFToken(r)

	err := templates.GetTemplates().Dashboard.ExecuteTemplate(w, "dashboard", data)
	if err != nil {
		log.Printf("Error rendering dashboard template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}