package middleware

import (
	"net/http"
	"strings"
)

// SecurityMiddleware applies security headers to all routes
func SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Apply security headers based on route type
		isStaticAsset := strings.HasPrefix(r.URL.Path, "/static/")

		// Core security headers (apply to all routes)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy - tailored for current site content
		csp := "default-src 'self'; " +
			"style-src 'self' 'unsafe-inline'; " + // Allow inline styles for theme system and index.html
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " + // Allow self-hosted JS (htmx.min.js), inline scripts (theme manager), and eval for HTMX
			"img-src 'self' https://images.unsplash.com https://unsplash.com https://media.tenor.com data:; " + // Allow Unsplash images, Tenor GIFs, and data URIs
			"media-src 'self' https://www.pexels.com https://videos.pexels.com data:; " + // Allow Pexels videos and data URIs
			"font-src 'self'; " + // Only self-hosted fonts
			"connect-src 'self'; " + // Allow HTMX AJAX requests to same origin
			"object-src 'none'; " + // Block plugins
			"base-uri 'self'; " + // Restrict base URI
			"form-action 'self'; " + // Only allow form submissions to same origin
			"frame-ancestors 'none'" // Prevent framing (redundant with X-Frame-Options but good defense)

		w.Header().Set("Content-Security-Policy", csp)

		// Additional headers for non-static content
		if !isStaticAsset {
			w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		}

		next.ServeHTTP(w, r)
	})
}
