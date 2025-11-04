package handlers

import (
	"log"
	"net/http"

	"circles.diy/internal/auth"
	"circles.diy/internal/middleware"
	"circles.diy/internal/templates"
)

func ChatHandler(w http.ResponseWriter, r *http.Request) {
	// Get mock chat data
	data := templates.GetMockChatData()

	// Add user from context to template data
	data.User = auth.GetUser(r.Context())
	data.CSRFToken = middleware.GetCSRFToken(r)

	// Render the chat template
	err := templates.GetTemplates().Chat.ExecuteTemplate(w, "chat", data)
	if err != nil {
		log.Printf("Error rendering chat template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}