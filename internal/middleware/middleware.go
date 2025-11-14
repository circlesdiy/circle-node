package middleware

import (
	"net/http"
	"log/slog"

	"circles.diy/internal/auth"
)

func Chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

func SessionValidationMiddleware(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionCookie, err := r.Cookie("session_id")
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			session, user, err := authService.ValidateSession(r.Context(), sessionCookie.Value)
			if err != nil {
				slog.Info("Session validation failed", "error", err, "path", r.URL.Path)
				http.SetCookie(w, &http.Cookie{
					Name:     "session_id",
					Value:    "",
					Path:     "/",
					MaxAge:   -1,
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteLaxMode,
				})
				next.ServeHTTP(w, r)
				return
			}

			ctx := auth.WithUser(r.Context(), user)
			ctx = auth.WithSession(ctx, session)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsAuthenticated(r.Context()) {
			if IsHTMXRequest(r) {
				w.Header().Set("HX-Redirect", "/")
			} else {
				http.Redirect(w, r, "/", http.StatusSeeOther)
			}
			return
		}
		next.ServeHTTP(w, r)
	})
}

// NoCacheMiddleware adds cache control headers to prevent browser caching of dynamic HTML
func NoCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set headers to prevent caching of dynamic HTML pages
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, private")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
}