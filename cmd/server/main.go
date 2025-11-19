package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"circles.diy/internal/app"
	"circles.diy/internal/handlers"
	"circles.diy/internal/middleware"
	"circles.diy/internal/templates"
	"circles.diy/internal/utils"
	"circles.diy/internal/version"
	"go.uber.org/zap"
)

func main() {
	startTime := time.Now()
	ctx := context.Background()

	// Initialize application
	application, err := app.New(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	// Build CSS
	if application.Config.IsDevelopment() && application.Config.Static.HotReload {
		application.Logger.Debug("starting CSS file watcher...")
		go utils.WatchCSSFiles(application.Logger)
	} else {
		application.Logger.Debug("building CSS...")
		if err := utils.BuildCSS(application.Logger); err != nil {
			application.Logger.Fatal("failed to build CSS", zap.Error(err))
		}
	}

	// Initialize templates
	application.Logger.Debug("initializing templates...")
	if err := templates.InitTemplates(); err != nil {
		application.Logger.Fatal("failed to initialize templates", zap.Error(err))
	}

	// Setup routes
	mux := setupRoutes(application)

	// Apply middleware chain with CSRF protection
	// Note: WebAuthn endpoints are exempted from CSRF as they have built-in challenge/origin validation
	// csrfConfig := middleware.CSRFConfig{
	// 	Secret:       application.Config.Security.CSRFSecret,
	// 	SecureCookie: !application.Config.IsDevelopment(), // HTTPS only in production
	// 	SkipPaths: []string{
	// 		// Static assets don't need CSRF
	// 		"/static/",
	// 		// WebAuthn endpoints have built-in challenge/origin validation
	// 		"/auth/register/begin",
	// 		"/auth/register/finish",
	// 		"/auth/login/begin",
	// 		"/auth/login/finish",
	// 		// Validation endpoints (public, safe for CSRF exemption)
	// 		"/auth/check-username",
	// 		"/auth/check-email",
	// 		// API endpoints
	// 		"/api/",
	// 		// Health check
	// 		"/health",
	// 	},
	// }

	handler := middleware.Chain(
		mux,
		// middleware.SecurityMiddleware,
		// middleware.CSRFMiddleware(csrfConfig),
		middleware.NoCacheMiddleware,
		middleware.RateLimitMiddleware,
	)

	// Setup HTTP server
	application.SetupHTTPServer(handler)

	// Log startup information
	logStartupInfo(application, startTime)

	// Start background cleanup goroutine for expired auth records
	go func() {
		application.Logger.Debug("🧹 cleanup job started (runs hourly)")
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		// Run cleanup immediately on startup
		cleanupCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		if err := application.AuthService.RunCleanup(cleanupCtx); err != nil {
			application.Logger.Warn("cleanup failed", zap.Error(err))
		}
		cancel()

		// Then run periodically
		for range ticker.C {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := application.AuthService.RunCleanup(cleanupCtx); err != nil {
				application.Logger.Warn("cleanup failed", zap.Error(err))
			}
			cancel()
		}
	}()

	// Run application (blocks until shutdown signal)
	if err := application.Run(); err != nil {
		application.Logger.Fatal("application error", zap.Error(err))
		os.Exit(1)
	}
}

// setupRoutes configures all HTTP routes
func setupRoutes(app *app.App) *http.ServeMux {
	mux := http.NewServeMux()

	// Register authentication routes
	app.AuthHandler.RegisterRoutes(mux)

	// Manifesto & user research routes (public)
	mux.HandleFunc("/", handlers.HomeHandler)
	mux.HandleFunc("/feedback", handlers.FeedbackHandler)

	// App routes (protected - require authentication)
	mux.HandleFunc("/dashboard", app.AuthHandler.RequireAuth(app.DashboardHandler.Handle))
	mux.HandleFunc("/dashboard/", app.AuthHandler.RequireAuth(app.DashboardHandler.Handle))
	mux.HandleFunc("/profile", app.AuthHandler.RequireAuth(app.ProfileHandler.Handle))
	mux.HandleFunc("/profile/", app.AuthHandler.RequireAuth(app.ProfileHandler.Handle))

	// Register circle routes
	app.CircleHandler.RegisterRoutes(mux, app.AuthHandler)

	mux.HandleFunc("/chat", app.AuthHandler.RequireAuth(handlers.ChatHandler))
	mux.HandleFunc("/chat/", app.AuthHandler.RequireAuth(handlers.ChatHandler))
	mux.HandleFunc("/gather", app.AuthHandler.RequireAuth(app.GatherHandler.Handle))
	mux.HandleFunc("/gather/", app.AuthHandler.RequireAuth(app.GatherHandler.Handle))
	mux.HandleFunc("/marketplace", app.AuthHandler.RequireAuth(handlers.MarketplaceHandler))
	mux.HandleFunc("/marketplace/", app.AuthHandler.RequireAuth(handlers.MarketplaceHandler))

	// Register profile API routes (requires authentication)
	mux.HandleFunc("PUT /api/profile", app.AuthHandler.RequireAuth(app.ProfileHandler.UpdateProfile))
	mux.HandleFunc("POST /api/profile/avatar", app.AuthHandler.RequireAuth(app.ProfileUploadHandler.UploadAvatar))
	mux.HandleFunc("POST /api/profile/banner", app.AuthHandler.RequireAuth(app.ProfileUploadHandler.UploadBanner))
	mux.HandleFunc("DELETE /api/profile/avatar", app.AuthHandler.RequireAuth(app.ProfileUploadHandler.DeleteAvatar))
	mux.HandleFunc("DELETE /api/profile/banner", app.AuthHandler.RequireAuth(app.ProfileUploadHandler.DeleteBanner))

	// Register post API routes (requires authentication)
	mux.HandleFunc("/api/posts", app.AuthHandler.RequireAuth(app.PostHandler.HandlePosts))
	mux.HandleFunc("/api/posts/htmx", app.AuthHandler.RequireAuth(app.PostHandler.HandleCreatePostHTMX))
	mux.HandleFunc("/api/posts/", app.AuthHandler.RequireAuth(app.PostHandler.HandlePost))
	// Comment routes - TODO: Enable when comment handlers are implemented
	// mux.HandleFunc("/api/comments/", app.AuthHandler.RequireAuth(app.PostHandler.HandleComment))

	// API routes
	mux.HandleFunc("/api/events", app.AuthHandler.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			app.GatherAPIHandler.HandleCreateEvent(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/events/", app.AuthHandler.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		// Parse URL to extract event ID
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 3 {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}
		eventID := parts[2]

		if len(parts) == 4 && parts[3] == "rsvp" {
			app.GatherAPIHandler.HandleRSVP(w, r)
		} else if len(parts) == 3 {
			// /api/events/{id}
			switch r.Method {
			case http.MethodPut, http.MethodPost:
				app.GatherAPIHandler.HandleUpdateEvent(w, r, eventID)
			case http.MethodDelete:
				app.GatherAPIHandler.HandleDeleteEvent(w, r, eventID)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	}))

	// Upload file serving (publicly accessible)
	uploadFS := http.FileServer(http.Dir(app.Config.Upload.StorageDir))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", uploadFS))

	// Static asset routes
	isDev := app.Config.IsDevelopment()
	mux.HandleFunc("/static/css/style.css", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServeStaticFile(w, r, "static/css/style.css", "text/css; charset=utf-8", isDev)
	})
	mux.HandleFunc("/static/js/htmx.min.js", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServeStaticFile(w, r, "static/js/htmx.min.js", "application/javascript; charset=utf-8", isDev)
	})
	mux.HandleFunc("/static/js/htmx-helpers.js", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServeStaticFile(w, r, "static/js/htmx-helpers.js", "application/javascript; charset=utf-8", isDev)
	})
	mux.HandleFunc("/static/js/post-cards.js", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServeStaticFile(w, r, "static/js/post-cards.js", "application/javascript; charset=utf-8", isDev)
	})
	mux.HandleFunc("/static/js/auth-login.js", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServeStaticFile(w, r, "static/js/auth-login.js", "application/javascript; charset=utf-8", isDev)
	})
	mux.HandleFunc("/static/js/auth-register.js", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServeStaticFile(w, r, "static/js/auth-register.js", "application/javascript; charset=utf-8", isDev)
	})
	mux.HandleFunc("/static/img/", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServeStaticImage(w, r, isDev)
	})

	// PWA routes
	mux.HandleFunc("/manifest.json", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServeStaticFile(w, r, "static/manifest.json", "application/manifest+json", isDev)
	})
	mux.HandleFunc("/sw.js", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServeStaticFile(w, r, "static/js/sw.js", "application/javascript; charset=utf-8", isDev)
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := app.HealthCheck(r.Context()); err != nil {
			app.Logger.Error("health check failed", zap.Error(err))
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("unhealthy"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("healthy"))
	})

	return mux
}

// logStartupInfo logs concise startup information
func logStartupInfo(app *app.App, startTime time.Time) {
	logger := app.Logger
	cfg := app.Config

	elapsed := time.Since(startTime)

	// Clean startup message for users
	logger.Info("Starting circles.diy",
		zap.String("version", version.Version),
	)
	logger.Info("Environment",
		zap.String("environment", cfg.Server.Environment),
	)

	// Detailed info for debugging
	logger.Debug("startup details",
		zap.String("url", formatServerURL(cfg.Server.Port, cfg.IsDevelopment())),
		zap.Int("pid", os.Getpid()),
		zap.String("started", formatDuration(elapsed)),
	)
}

// maskPassword formats database URL with masked password
func maskPassword(host string, port int, user, database string) string {
	return fmt.Sprintf("postgres://%s:***@%s:%d/%s", user, host, port, database)
}

// formatRedisURL formats Redis connection URL
func formatRedisURL(host string, port, database int) string {
	return fmt.Sprintf("redis://%s:%d/%d", host, port, database)
}

// formatServerURL formats the server listening URL
func formatServerURL(port string, isDev bool) string {
	protocol := "http"
	if !isDev {
		protocol = "https"
	}
	return fmt.Sprintf("%s://localhost:%s", protocol, port)
}

// formatDuration formats duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}
