package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"circles.diy/internal/auth"
	"circles.diy/internal/circle"
	"circles.diy/internal/config"
	"circles.diy/internal/content"
	"circles.diy/internal/dashboard"
	"circles.diy/internal/domain"
	"circles.diy/internal/events"
	"circles.diy/internal/gather"
	"circles.diy/internal/handlers"
	"circles.diy/internal/preferences"
	"circles.diy/internal/profile"
	"circles.diy/internal/storage"
	"circles.diy/internal/user"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// App represents the application with all its dependencies
type App struct {
	Config      *config.Config
	Logger      *zap.Logger
	Postgres    *storage.Postgres
	Redis       *storage.Redis
	FileStorage domain.FileStorage
	Server      *http.Server

	// Domain services
	UserService        *user.Service
	ProfileService     *profile.Service
	PreferencesService *preferences.Service
	CircleService      *circle.Service
	DashboardService   *dashboard.Service
	GatherService      *gather.Service
	EventService       *events.Service
	ContentService     *content.Service

	// Authentication components
	AuthService *auth.Service
	AuthHandler *auth.Handler

	// Feature handlers
	ProfileHandler interface {
		Handle(w http.ResponseWriter, r *http.Request)
		HandleEdit(w http.ResponseWriter, r *http.Request)
		UpdateProfile(w http.ResponseWriter, r *http.Request)
	}
	ProfileUploadHandler *handlers.ProfileUploadHandler
	CircleHandler        *handlers.CircleHandler
	CircleUploadHandler  *handlers.CircleUploadHandler
	DashboardHandler     *handlers.DashboardHandler
	GatherHandler        *handlers.GatherHandler
	GatherAPIHandler     *handlers.GatherAPIHandler
	PostHandler          *handlers.PostHandler
}

// New creates a new application instance with all dependencies
func New(ctx context.Context) (*App, error) {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	logger, err := initLogger(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	logger.Debug("starting application",
		zap.String("environment", cfg.Server.Environment),
		zap.String("port", cfg.Server.Port),
	)

	// Initialize PostgreSQL
	logger.Debug("connecting to database...")
	postgres, err := storage.NewPostgres(ctx, cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run migrations
	logger.Debug("running database migrations...")
	if err := postgres.RunMigrations(ctx); err != nil {
		postgres.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Initialize Redis
	logger.Debug("connecting to redis...")
	redis, err := storage.NewRedis(ctx, cfg, logger)
	if err != nil {
		postgres.Close()
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	// Initialize file storage
	logger.Debug("initializing file storage...")
	var fileStorage domain.FileStorage
	switch cfg.Upload.Provider {
	case "local":
		fileStorage, err = storage.NewLocalStorage(cfg.Upload, logger)
		if err != nil {
			postgres.Close()
			redis.Close()
			return nil, fmt.Errorf("failed to initialize local storage: %w", err)
		}
	case "s3":
		postgres.Close()
		redis.Close()
		return nil, fmt.Errorf("s3 provider not yet implemented")
	case "minio":
		postgres.Close()
		redis.Close()
		return nil, fmt.Errorf("minio provider not yet implemented")
	default:
		postgres.Close()
		redis.Close()
		return nil, fmt.Errorf("unknown storage provider: %s", cfg.Upload.Provider)
	}

	// Initialize domain services
	logger.Debug("initializing domain services...")

	// User service
	userRepo := user.NewRepository(postgres.Pool)
	userService := user.NewService(userRepo, logger)

	// Profile service
	profileRepo := profile.NewRepository(postgres.Pool)
	profileService := profile.NewService(profileRepo, fileStorage, logger)

	// Preferences service
	prefsRepo := preferences.NewRepository(postgres.Pool)
	prefsService := preferences.NewService(prefsRepo, logger)

	// Circle service
	circleRepo := circle.NewRepository(postgres.Pool)
	circleService := circle.NewService(circleRepo, fileStorage, logger)

	// Dashboard service
	dashboardRepo := dashboard.NewPostgresRepository(postgres.Pool)
	dashboardService := dashboard.NewService(dashboardRepo, logger)

	// Gather service
	gatherRepo := gather.NewPostgresRepository(postgres.Pool)
	gatherService := gather.NewService(gatherRepo, logger)

	// Events service
	eventRepo := events.NewPostgresRepository(postgres.Pool)
	eventService := events.NewService(eventRepo, circleService, logger)

	// Content service
	contentRepo := content.NewRepository(postgres.Pool)
	reactionRepo := content.NewReactionRepository(postgres.Pool)
	contentService := content.NewService(contentRepo, reactionRepo)

	// Initialize authentication system
	logger.Debug("initializing authentication system...")
	authRepo := auth.NewRepository(postgres.Pool)
	authConfig := auth.DefaultWebAuthnConfig()

	authService := auth.NewService(authRepo, userService, profileService, prefsService, authConfig, logger)
	authHandler := auth.NewHandler(authService, logger)

	// Initialize feature handlers
	logger.Debug("initializing feature handlers...")
	profileHandler := handlers.NewProfileHandler(profileService, prefsService, circleService, logger)
	profileUploadHandler := handlers.NewProfileUploadHandler(profileService, logger)
	circleHandler := handlers.NewCircleHandler(circleService, contentService, prefsService, profileService, cfg.AssetVersion, logger)
	circleUploadHandler := handlers.NewCircleUploadHandler(circleService, logger)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService, profileService)
	gatherHandler := handlers.NewGatherHandler(gatherService, eventService, profileService)
	gatherAPIHandler := handlers.NewGatherAPIHandler(gatherService, eventService, logger)
	postHandler := handlers.NewPostHandler(contentService, profileService)

	app := &App{
		Config:               cfg,
		Logger:               logger,
		Postgres:             postgres,
		Redis:                redis,
		FileStorage:          fileStorage,
		UserService:          userService,
		ProfileService:       profileService,
		PreferencesService:   prefsService,
		CircleService:        circleService,
		DashboardService:     dashboardService,
		GatherService:        gatherService,
		EventService:         eventService,
		ContentService:       contentService,
		AuthService:          authService,
		AuthHandler:          authHandler,
		ProfileHandler:       profileHandler,
		ProfileUploadHandler: profileUploadHandler,
		CircleHandler:        circleHandler,
		CircleUploadHandler:  circleUploadHandler,
		DashboardHandler:     dashboardHandler,
		GatherHandler:        gatherHandler,
		GatherAPIHandler:     gatherAPIHandler,
		PostHandler:          postHandler,
	}

	return app, nil
}

// SetupHTTPServer configures and returns the HTTP server with the given handler
func (a *App) SetupHTTPServer(handler http.Handler) {
	a.Server = &http.Server{
		Addr:         ":" + a.Config.Server.Port,
		Handler:      handler,
		ReadTimeout:  a.Config.Server.ReadTimeout,
		WriteTimeout: a.Config.Server.WriteTimeout,
		IdleTimeout:  a.Config.Server.IdleTimeout,
	}
}

// Run starts the application and blocks until shutdown signal is received
func (a *App) Run() error {
	if a.Server == nil {
		return fmt.Errorf("HTTP server not configured")
	}

	// Create error channel for server errors
	serverErrors := make(chan error, 1)

	// Start HTTP server in a goroutine
	go func() {
		a.Logger.Info("Server",
			zap.String("url", formatServerURL(a.Server.Addr, a.Config.IsDevelopment())),
		)
		serverErrors <- a.Server.ListenAndServe()
	}()

	// Create channel to listen for interrupt signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or server error
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}

	case sig := <-shutdown:
		a.Logger.Info("shutdown signal received",
			zap.String("signal", sig.String()),
		)

		// Create context with timeout for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), a.Config.Server.ShutdownTimeout)
		defer cancel()

		// Attempt graceful shutdown
		if err := a.Shutdown(ctx); err != nil {
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}

		a.Logger.Info("server stopped successfully")
	}

	return nil
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown(ctx context.Context) error {
	a.Logger.Info("initiating graceful shutdown")

	// Shutdown HTTP server first
	if a.Server != nil {
		a.Logger.Info("shutting down http server")
		if err := a.Server.Shutdown(ctx); err != nil {
			a.Logger.Error("error shutting down http server", zap.Error(err))
		}
	}

	// Close Redis connection
	if a.Redis != nil {
		if err := a.Redis.Close(); err != nil {
			a.Logger.Error("error closing redis connection", zap.Error(err))
		}
	}

	// Close database connection pool
	if a.Postgres != nil {
		a.Postgres.Close()
	}

	// Sync logger
	a.Logger.Sync()

	return nil
}

// HealthCheck performs health checks on all dependencies
func (a *App) HealthCheck(ctx context.Context) error {
	// Check database
	if err := a.Postgres.HealthCheck(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Check Redis
	if err := a.Redis.HealthCheck(ctx); err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}

	// Check file storage
	if err := a.FileStorage.HealthCheck(ctx); err != nil {
		return fmt.Errorf("file storage health check failed: %w", err)
	}

	return nil
}

// formatServerURL formats the server listening URL
func formatServerURL(addr string, isDev bool) string {
	protocol := "http"
	if !isDev {
		protocol = "https"
	}
	// addr is in format ":8080", so we add localhost
	if addr[0] == ':' {
		return fmt.Sprintf("%s://localhost%s", protocol, addr)
	}
	return fmt.Sprintf("%s://%s", protocol, addr)
}

// initLogger creates and configures the logger based on config
func initLogger(cfg *config.Config) (*zap.Logger, error) {
	var zapConfig zap.Config

	if cfg.IsDevelopment() {
		zapConfig = zap.NewDevelopmentConfig()
		// Use console encoding for better readability in development
		zapConfig.Encoding = "console"
		zapConfig.EncoderConfig.TimeKey = "T"
		zapConfig.EncoderConfig.LevelKey = "L"
		zapConfig.EncoderConfig.NameKey = "N"
		zapConfig.EncoderConfig.CallerKey = "" // Disable caller for cleaner output
		zapConfig.EncoderConfig.MessageKey = "M"
		zapConfig.EncoderConfig.StacktraceKey = "S"
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
		zapConfig.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
		zapConfig.EncoderConfig.ConsoleSeparator = " "
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	// Set log level
	switch cfg.Logging.Level {
	case "debug":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	// Set output paths
	if cfg.Logging.Output != "" && cfg.Logging.Output != "stdout" {
		zapConfig.OutputPaths = []string{cfg.Logging.Output}
	}

	return zapConfig.Build()
}
