package handlers

import (
	"log"
	"net/http"

	"circles.diy/internal/auth"
	"circles.diy/internal/middleware"
	"circles.diy/internal/templates"
)

func MarketplaceHandler(w http.ResponseWriter, r *http.Request) {
	// Get mock marketplace data
	data := templates.GetMockMarketplaceData()

	// Add user from context to template data
	data.User = auth.GetUser(r.Context())
	data.CSRFToken = middleware.GetCSRFToken(r)

	// Render the marketplace template
	err := templates.GetTemplates().Marketplace.ExecuteTemplate(w, "marketplace", data)
	if err != nil {
		log.Printf("Error rendering marketplace template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
