package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"circles.diy/internal/app"
	"circles.diy/internal/handlers"
	"circles.diy/internal/middleware"
	"circles.diy/internal/templates"
	"circles.diy/internal/utils"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	// Initialize application
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Build CSS
	if application.Config.IsDevelopment() && application.Config.Static.HotReload {
		application.Logger.Info("starting CSS file watcher...")
		go utils.WatchCSSFiles()
	} else {
		application.Logger.Info("building CSS...")
		if err := utils.BuildCSS(); err != nil {
			application.Logger.Fatal("failed to build CSS", zap.Error(err))
		}
	}

	// Initialize templates
	application.Logger.Info("initializing templates...")
	if err := templates.InitTemplates(); err != nil {
		application.Logger.Fatal("failed to initialize templates", zap.Error(err))
	}

	// Setup routes
	mux := setupRoutes(application)

	// Apply middleware chain with CSRF protection
	// Note: WebAuthn endpoints are exempted from CSRF as they have built-in challenge/origin validation
	csrfConfig := middleware.CSRFConfig{
		Secret:       application.Config.Security.CSRFSecret,
		SecureCookie: !application.Config.IsDevelopment(), // HTTPS only in production
		SkipPaths: []string{
			// Static assets don't need CSRF
			"/static/",
			// WebAuthn endpoints have built-in challenge/origin validation
			"/auth/register/begin",
			"/auth/register/finish",
			"/auth/login/begin",
			"/auth/login/finish",
			// Validation endpoints (public, safe for CSRF exemption)
			"/auth/check-username",
			"/auth/check-email",
			// API endpoints
			"/api/",
			// Health check
			"/health",
		},
	}

	handler := middleware.Chain(
		mux,
		middleware.SecurityMiddleware,
		middleware.CSRFMiddleware(csrfConfig),
		middleware.RateLimitMiddleware,
	)

	// Setup HTTP server
	application.SetupHTTPServer(handler)

	// Log available routes
	logRoutes(application.Logger)

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
	mux.HandleFunc("/dashboard", app.AuthHandler.RequireAuth(handlers.DashboardHandler))
	mux.HandleFunc("/dashboard/", app.AuthHandler.RequireAuth(handlers.DashboardHandler))
	mux.HandleFunc("/profile", app.AuthHandler.RequireAuth(handlers.ProfileHandler))
	mux.HandleFunc("/profile/", app.AuthHandler.RequireAuth(handlers.ProfileHandler))
	mux.HandleFunc("/circles", app.AuthHandler.RequireAuth(handlers.CirclesHandler))
	mux.HandleFunc("/circles/", app.AuthHandler.RequireAuth(handlers.CirclesHandler))
	mux.HandleFunc("/chat", app.AuthHandler.RequireAuth(handlers.ChatHandler))
	mux.HandleFunc("/chat/", app.AuthHandler.RequireAuth(handlers.ChatHandler))
	mux.HandleFunc("/gather", app.AuthHandler.RequireAuth(handlers.GatherHandler))
	mux.HandleFunc("/gather/", app.AuthHandler.RequireAuth(handlers.GatherHandler))
	mux.HandleFunc("/marketplace", app.AuthHandler.RequireAuth(handlers.MarketplaceHandler))
	mux.HandleFunc("/marketplace/", app.AuthHandler.RequireAuth(handlers.MarketplaceHandler))

	// Static asset routes
	mux.HandleFunc("/static/css/style.css", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServeStaticFile(w, r, "static/css/style.css", "text/css; charset=utf-8")
	})
	mux.HandleFunc("/static/js/htmx.min.js", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServeStaticFile(w, r, "static/js/htmx.min.js", "application/javascript; charset=utf-8")
	})
	mux.HandleFunc("/static/js/auth-login.js", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServeStaticFile(w, r, "static/js/auth-login.js", "application/javascript; charset=utf-8")
	})
	mux.HandleFunc("/static/js/auth-register.js", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServeStaticFile(w, r, "static/js/auth-register.js", "application/javascript; charset=utf-8")
	})
	mux.HandleFunc("/static/img/", handlers.ServeStaticImage)

	// PWA routes
	mux.HandleFunc("/manifest.json", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServeStaticFile(w, r, "static/manifest.json", "application/manifest+json")
	})
	mux.HandleFunc("/sw.js", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServeStaticFile(w, r, "static/js/sw.js", "application/javascript; charset=utf-8")
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

// logRoutes logs all available routes
func logRoutes(logger *zap.Logger) {
	logger.Info("routes configured",
		zap.Strings("routes", []string{
			"/ - Manifesto landing page",
			"/dashboard - Dashboard with templates + HTMX",
			"/profile - Profile page with templates + HTMX",
			"/circles - Circles page with templates + HTMX",
			"/chat - Chat page with templates + HTMX",
			"/gather - Events page with templates + HTMX",
			"/marketplace - Marketplace page with templates + HTMX",
			"/health - Health check endpoint",
		}),
	)
}
