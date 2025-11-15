package handlers

import (
	"net/http"

	"circles.diy/internal/templates"
	"go.uber.org/zap"
)

// IsHTMXRequest checks if the request is from HTMX
func IsHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// RenderFragment renders a template fragment for HTMX responses
func RenderFragment(w http.ResponseWriter, logger *zap.Logger, templateName string, data interface{}) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl := templates.GetTemplates()
	var err error

	switch templateName {
	case "toast":
		err = tmpl.Toast.ExecuteTemplate(w, "toast", data)
	case "modal":
		err = tmpl.Modal.ExecuteTemplate(w, "modal", data)
	case "circle-card":
		err = tmpl.CircleCard.ExecuteTemplate(w, "circle-card", data)
	case "circle-form":
		err = tmpl.CircleForm.ExecuteTemplate(w, "circle-form", data)
	case "empty-state":
		err = tmpl.EmptyState.ExecuteTemplate(w, "empty-state", data)
	default:
		logger.Error("unknown template fragment", zap.String("template", templateName))
		http.Error(w, "Unknown template", http.StatusInternalServerError)
		return err
	}

	if err != nil {
		logger.Error("failed to render fragment",
			zap.String("template", templateName),
			zap.Error(err),
		)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return err
	}

	return nil
}

// RenderError renders an error response
func RenderError(w http.ResponseWriter, logger *zap.Logger, statusCode int, message string) {
	w.WriteHeader(statusCode)

	data := map[string]interface{}{
		"Type":    "error",
		"Message": message,
	}

	if err := RenderFragment(w, logger, "toast", data); err != nil {
		// Fallback to plain text if template fails
		http.Error(w, message, statusCode)
	}
}

// RenderSuccess renders a success toast
func RenderSuccess(w http.ResponseWriter, logger *zap.Logger, message string) {
	data := map[string]interface{}{
		"Type":    "success",
		"Message": message,
	}

	RenderFragment(w, logger, "toast", data)
}

// SendHTMXRedirect sends an HX-Redirect header
// Note: This must be called before any response body is written
func SendHTMXRedirect(w http.ResponseWriter, url string) {
	w.Header().Set("HX-Redirect", url)
}

// SendHTMXRefresh tells HTMX to refresh the page
// Note: This must be called before any response body is written
func SendHTMXRefresh(w http.ResponseWriter) {
	w.Header().Set("HX-Refresh", "true")
}

// RenderValidationError renders a field-specific validation error
func RenderValidationError(w http.ResponseWriter, logger *zap.Logger, field, message string) {
	w.WriteHeader(http.StatusBadRequest)

	data := map[string]interface{}{
		"Type":    "error",
		"Message": message,
		"Field":   field,
	}

	RenderFragment(w, logger, "toast", data)
}
