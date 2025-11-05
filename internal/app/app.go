package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"circles.diy/internal/auth"
	"circles.diy/internal/config"
	"circles.diy/internal/storage"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// App represents the application with all its dependencies
type App struct {
	Config   *config.Config
	Logger   *zap.Logger
	Postgres *storage.Postgres
	Redis    *storage.Redis
	Server   *http.Server

	// Authentication components
	AuthService *auth.Service
	AuthHandler *auth.Handler
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

	// Initialize authentication system
	logger.Debug("initializing authentication system...")
	authRepo := auth.NewRepository(postgres.Pool)
	authConfig := auth.DefaultWebAuthnConfig()

	authService := auth.NewService(authRepo, authConfig, logger)
	authHandler := auth.NewHandler(authService, logger)

	app := &App{
		Config:      cfg,
		Logger:      logger,
		Postgres:    postgres,
		Redis:       redis,
		AuthService: authService,
		AuthHandler: authHandler,
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
		a.Logger.Info("starting http server",
			zap.String("port", a.Server.Addr),
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

		a.Logger.Info("server stopped gracefully")
	}

	return nil
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown(ctx context.Context) error {
	a.Logger.Info("initiating graceful shutdown...")

	// Shutdown HTTP server first
	if a.Server != nil {
		a.Logger.Info("shutting down http server...")
		if err := a.Server.Shutdown(ctx); err != nil {
			a.Logger.Error("error shutting down http server", zap.Error(err))
		}
	}

	// Close Redis connection
	if a.Redis != nil {
		a.Logger.Info("closing redis connection...")
		if err := a.Redis.Close(); err != nil {
			a.Logger.Error("error closing redis connection", zap.Error(err))
		}
	}

	// Close database connection pool
	if a.Postgres != nil {
		a.Logger.Info("closing database connection pool...")
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

	return nil
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
