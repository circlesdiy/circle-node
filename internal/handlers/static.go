package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func ServeStaticFile(w http.ResponseWriter, r *http.Request, filePath, contentType string, isDevelopment bool) {
	// Security: prevent path traversal
	if strings.Contains(filePath, "..") {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Get file info for cache headers
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set cache headers based on environment
	lastModified := fileInfo.ModTime().UTC().Format(http.TimeFormat)
	ifModifiedSince := r.Header.Get("If-Modified-Since")

	if ifModifiedSince != "" {
		if t, err := time.Parse(http.TimeFormat, ifModifiedSince); err == nil {
			if fileInfo.ModTime().UTC().Before(t.Add(1 * time.Second)) {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
	}

	// CRITICAL: Set Content-Type header BEFORE writing any content
	// This ensures it's not overridden and respects X-Content-Type-Options: nosniff
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Last-Modified", lastModified)

	// Smart caching: Short cache in development, long cache with immutable in production
	if isDevelopment {
		// Development: 10 second cache for quick iteration
		w.Header().Set("Cache-Control", "public, max-age=10")
	} else {
		// Production: 1 year cache with immutable (safe with versioned URLs)
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}

	// For CSS files, set additional headers
	if strings.HasSuffix(filePath, ".css") {
		w.Header().Set("Vary", "Accept-Encoding")
	}

	// Write the file content
	w.Write(content)
}

func ServeStaticImage(w http.ResponseWriter, r *http.Request, isDevelopment bool) {
	// Extract the image path from URL
	imagePath := strings.TrimPrefix(r.URL.Path, "/static/img/")

	// Security: prevent path traversal
	if strings.Contains(imagePath, "..") || imagePath == "" {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	// Construct full file path
	fullPath := filepath.Join("static", "img", imagePath)

	// Determine content type based on file extension
	ext := strings.ToLower(filepath.Ext(imagePath))
	var contentType string

	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	case ".svg":
		contentType = "image/svg+xml"
	case ".ico":
		contentType = "image/x-icon"
	default:
		http.Error(w, "Unsupported image format", http.StatusBadRequest)
		return
	}

	// Use the existing ServeStaticFile function
	ServeStaticFile(w, r, fullPath, contentType, isDevelopment)
}