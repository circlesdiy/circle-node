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

	// Apply middleware chain
	handler := middleware.Chain(
		mux,
		middleware.SecurityMiddleware,
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

	// Manifesto & user research routes
	mux.HandleFunc("/", handlers.HomeHandler)
	mux.HandleFunc("/feedback", handlers.FeedbackHandler)

	// App routes
	mux.HandleFunc("/dashboard", handlers.DashboardHandler)
	mux.HandleFunc("/dashboard/", handlers.DashboardHandler)
	mux.HandleFunc("/profile", handlers.ProfileHandler)
	mux.HandleFunc("/profile/", handlers.ProfileHandler)
	mux.HandleFunc("/circles", handlers.CirclesHandler)
	mux.HandleFunc("/circles/", handlers.CirclesHandler)
	mux.HandleFunc("/chat", handlers.ChatHandler)
	mux.HandleFunc("/chat/", handlers.ChatHandler)
	mux.HandleFunc("/gather", handlers.GatherHandler)
	mux.HandleFunc("/gather/", handlers.GatherHandler)
	mux.HandleFunc("/marketplace", handlers.MarketplaceHandler)
	mux.HandleFunc("/marketplace/", handlers.MarketplaceHandler)

	// Static asset routes
	mux.HandleFunc("/static/css/style.css", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServeStaticFile(w, r, "static/css/style.css", "text/css; charset=utf-8")
	})
	mux.HandleFunc("/static/js/htmx.min.js", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServeStaticFile(w, r, "static/js/htmx.min.js", "application/javascript; charset=utf-8")
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
