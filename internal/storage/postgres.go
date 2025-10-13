package storage

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"
	"time"

	"circles.diy/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Postgres manages the PostgreSQL connection pool
type Postgres struct {
	Pool   *pgxpool.Pool
	logger *zap.Logger
}

// NewPostgres creates a new PostgreSQL connection pool
func NewPostgres(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*Postgres, error) {
	// Build connection pool config
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure pool settings
	poolConfig.MaxConns = int32(cfg.Database.MaxConnections)
	poolConfig.MinConns = int32(cfg.Database.MinConnections)
	poolConfig.MaxConnLifetime = cfg.Database.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.Database.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = cfg.Database.HealthCheckPeriod

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("database connection established",
		zap.String("host", cfg.Database.Host),
		zap.Int("port", cfg.Database.Port),
		zap.String("database", cfg.Database.Database),
	)

	return &Postgres{
		Pool:   pool,
		logger: logger,
	}, nil
}

// Close closes the database connection pool
func (p *Postgres) Close() {
	p.logger.Info("closing database connection pool")
	p.Pool.Close()
}

// RunMigrations runs all embedded SQL migrations
func (p *Postgres) RunMigrations(ctx context.Context) error {
	p.logger.Info("starting database migrations")

	// Create migrations table if it doesn't exist
	if err := p.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of migration files
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Sort migrations by filename (should be numbered: 001_initial.sql, 002_add_posts.sql, etc.)
	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Get already applied migrations
	appliedMigrations, err := p.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Run pending migrations
	for _, filename := range migrationFiles {
		if appliedMigrations[filename] {
			p.logger.Debug("skipping already applied migration", zap.String("file", filename))
			continue
		}

		p.logger.Info("applying migration", zap.String("file", filename))

		// Read migration file
		content, err := migrationsFS.ReadFile("migrations/" + filename)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", filename, err)
		}

		// Execute migration in a transaction
		tx, err := p.Pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", filename, err)
		}

		// Execute migration SQL
		if _, err := tx.Exec(ctx, string(content)); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		// Record migration as applied
		if _, err := tx.Exec(ctx,
			"INSERT INTO schema_migrations (filename, applied_at) VALUES ($1, $2)",
			filename, time.Now(),
		); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		// Commit transaction
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", filename, err)
		}

		p.logger.Info("successfully applied migration", zap.String("file", filename))
	}

	p.logger.Info("database migrations completed", zap.Int("total", len(migrationFiles)))
	return nil
}

// createMigrationsTable creates the schema_migrations table if it doesn't exist
func (p *Postgres) createMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`
	_, err := p.Pool.Exec(ctx, query)
	return err
}

// getAppliedMigrations returns a map of already applied migrations
func (p *Postgres) getAppliedMigrations(ctx context.Context) (map[string]bool, error) {
	rows, err := p.Pool.Query(ctx, "SELECT filename FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, err
		}
		applied[filename] = true
	}

	return applied, rows.Err()
}

// HealthCheck performs a health check on the database
func (p *Postgres) HealthCheck(ctx context.Context) error {
	return p.Pool.Ping(ctx)
}
