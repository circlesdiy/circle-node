package middleware

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"circles.diy/internal/utils"
)

const (
	csrfTokenKey contextKey = "csrf_token"
	csrfCookie   string     = "csrf_token"
	csrfHeader   string     = "X-CSRF-Token"
)

// CSRFConfig holds CSRF middleware configuration
type CSRFConfig struct {
	Secret        string
	CookieName    string
	HeaderName    string
	SkipPaths     []string
	SecureCookie  bool
	SameSite      http.SameSite
}

// CSRFMiddleware provides CSRF protection
func CSRFMiddleware(config CSRFConfig) func(http.Handler) http.Handler {
	if config.CookieName == "" {
		config.CookieName = csrfCookie
	}
	if config.HeaderName == "" {
		config.HeaderName = csrfHeader
	}
	if config.SameSite == 0 {
		config.SameSite = http.SameSiteLaxMode
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip CSRF check for safe methods
			if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
				var token string

				// Check if valid token already exists in cookie
				if existingCookie, err := r.Cookie(config.CookieName); err == nil && existingCookie.Value != "" {
					// Reuse existing token
					token = existingCookie.Value
				} else {
					// Generate new token only if none exists
					var err error
					token, err = generateCSRFToken(config.Secret)
					if err != nil {
						http.Error(w, "Failed to generate CSRF token", http.StatusInternalServerError)
						return
					}

					// Set cookie with new token
					http.SetCookie(w, &http.Cookie{
						Name:     config.CookieName,
						Value:    token,
						Path:     "/",
						HttpOnly: true,
						Secure:   config.SecureCookie,
						SameSite: config.SameSite,
						MaxAge:   86400, // 24 hours
					})
				}

				// Add token to context
				ctx := context.WithValue(r.Context(), csrfTokenKey, token)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Check if path should skip CSRF (supports both exact match and prefix match)
			for _, path := range config.SkipPaths {
				// If path ends with /, treat as prefix match
				if strings.HasSuffix(path, "/") {
					if strings.HasPrefix(r.URL.Path, path) {
						next.ServeHTTP(w, r)
						return
					}
				} else if r.URL.Path == path {
					// Exact match
					next.ServeHTTP(w, r)
					return
				}
			}

			// For unsafe methods, verify CSRF token
			cookieToken, err := r.Cookie(config.CookieName)
			if err != nil {
				http.Error(w, "CSRF cookie not found", http.StatusForbidden)
				return
			}

			// Get token from header or form
			headerToken := r.Header.Get(config.HeaderName)
			if headerToken == "" {
				// Try to get from form
				if err := r.ParseForm(); err == nil {
					headerToken = r.FormValue("csrf_token")
				}
			}

			if headerToken == "" {
				http.Error(w, "CSRF token not found", http.StatusForbidden)
				return
			}

			// Verify tokens match
			if !verifyCSRFToken(cookieToken.Value, headerToken) {
				http.Error(w, "CSRF token mismatch", http.StatusForbidden)
				return
			}

			// Add token to context
			ctx := context.WithValue(r.Context(), csrfTokenKey, cookieToken.Value)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetCSRFToken retrieves the CSRF token from the request context
func GetCSRFToken(r *http.Request) string {
	if val, ok := r.Context().Value(csrfTokenKey).(string); ok {
		return val
	}
	return ""
}

// generateCSRFToken generates a new CSRF token
func generateCSRFToken(secret string) (string, error) {
	// Generate random bytes
	tokenBytes, err := utils.GenerateRandomBytes(32)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode to base64
	token := base64.URLEncoding.EncodeToString(tokenBytes)
	return token, nil
}

// verifyCSRFToken verifies that two CSRF tokens match
func verifyCSRFToken(token1, token2 string) bool {
	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(token1), []byte(token2)) == 1
}
