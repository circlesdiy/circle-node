package http

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// Response helpers for consistent HTTP responses

// JSONResponse writes a JSON response with the given status code
func JSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// JSONError writes a JSON error response
func JSONError(w http.ResponseWriter, statusCode int, message string) {
	JSONResponse(w, statusCode, map[string]string{"error": message})
}

// HTMLResponse writes an HTML response
func HTMLResponse(w http.ResponseWriter, statusCode int, html []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write(html)
}

// InternalServerError writes a 500 error response
func InternalServerError(w http.ResponseWriter, logger *zap.Logger, err error, msg string) {
	logger.Error(msg, zap.Error(err))
	JSONError(w, http.StatusInternalServerError, "Internal server error")
}

// BadRequest writes a 400 error response
func BadRequest(w http.ResponseWriter, message string) {
	JSONError(w, http.StatusBadRequest, message)
}

// Unauthorized writes a 401 error response
func Unauthorized(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	JSONError(w, http.StatusUnauthorized, message)
}

// Forbidden writes a 403 error response
func Forbidden(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Forbidden"
	}
	JSONError(w, http.StatusForbidden, message)
}

// NotFound writes a 404 error response
func NotFound(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Not found"
	}
	JSONError(w, http.StatusNotFound, message)
}

// IsHTMXRequest checks if the request is an HTMX request
func IsHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// HTMXRedirect sends an HTMX redirect response
func HTMXRedirect(w http.ResponseWriter, url string) {
	w.Header().Set("HX-Redirect", url)
	w.WriteHeader(http.StatusNoContent)
}

// HTMXRefresh triggers a page refresh in HTMX
func HTMXRefresh(w http.ResponseWriter) {
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusNoContent)
}
