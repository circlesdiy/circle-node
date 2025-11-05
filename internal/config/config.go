package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	Redis     RedisConfig     `yaml:"redis"`
	Auth      AuthConfig      `yaml:"auth"`
	Security  SecurityConfig  `yaml:"security"`
	Logging   LoggingConfig   `yaml:"logging"`
	Templates TemplatesConfig `yaml:"templates"`
	Static    StaticConfig    `yaml:"static"`
}

type ServerConfig struct {
	Port            string        `yaml:"port"`
	Environment     string        `yaml:"environment"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type DatabaseConfig struct {
	Host              string        `yaml:"host"`
	Port              int           `yaml:"port"`
	User              string        `yaml:"user"`
	Password          string        `yaml:"password"`
	Database          string        `yaml:"database"`
	SSLMode           string        `yaml:"ssl_mode"`
	MaxConnections    int           `yaml:"max_connections"`
	MinConnections    int           `yaml:"min_connections"`
	MaxConnLifetime   time.Duration `yaml:"max_conn_lifetime"`
	MaxConnIdleTime   time.Duration `yaml:"max_conn_idle_time"`
	HealthCheckPeriod time.Duration `yaml:"health_check_period"`
}

type RedisConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Password     string `yaml:"password"`
	Database     int    `yaml:"database"`
	MaxRetries   int    `yaml:"max_retries"`
	PoolSize     int    `yaml:"pool_size"`
	MinIdleConns int    `yaml:"min_idle_conns"`
}

type AuthConfig struct {
	SessionDuration      time.Duration `yaml:"session_duration"`
	RefreshTokenDuration time.Duration `yaml:"refresh_token_duration"`
	WebAuthnTimeout      time.Duration `yaml:"webauthn_timeout"`
	WebAuthnRPName       string        `yaml:"webauthn_rp_name"`
	WebAuthnRPID         string        `yaml:"webauthn_rp_id"`
	WebAuthnRPOrigin     string        `yaml:"webauthn_rp_origin"`
}

type SecurityConfig struct {
	CSRFSecret        string        `yaml:"csrf_secret"`
	RateLimitRequests int           `yaml:"rate_limit_requests"`
	RateLimitWindow   time.Duration `yaml:"rate_limit_window"`
	BCryptCost        int           `yaml:"bcrypt_cost"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

type TemplatesConfig struct {
	HotReload bool `yaml:"hot_reload"`
}

type StaticConfig struct {
	HotReload bool `yaml:"hot_reload"`
}

// Load reads configuration from YAML file and applies environment variable overrides
func Load(path string) (*Config, error) {
	// Set default config path if not provided
	if path == "" {
		path = "config.yaml"
	}

	// Create config with defaults
	cfg := &Config{
		Server: ServerConfig{
			Port:            "8080",
			Environment:     "development",
			ReadTimeout:     15 * time.Second,
			WriteTimeout:    15 * time.Second,
			IdleTimeout:     60 * time.Second,
			ShutdownTimeout: 30 * time.Second,
		},
		Database: DatabaseConfig{
			Host:              "localhost",
			Port:              5432,
			User:              "circles",
			Password:          "circles_dev_password",
			Database:          "circles_diy",
			SSLMode:           "disable",
			MaxConnections:    25,
			MinConnections:    5,
			MaxConnLifetime:   1 * time.Hour,
			MaxConnIdleTime:   10 * time.Minute,
			HealthCheckPeriod: 1 * time.Minute,
		},
		Redis: RedisConfig{
			Host:         "localhost",
			Port:         6379,
			Database:     0,
			MaxRetries:   3,
			PoolSize:     10,
			MinIdleConns: 5,
		},
		Auth: AuthConfig{
			SessionDuration:      720 * time.Hour,
			RefreshTokenDuration: 2160 * time.Hour,
			WebAuthnTimeout:      60 * time.Second,
			WebAuthnRPName:       "Circles.DIY",
			WebAuthnRPID:         "localhost",
			WebAuthnRPOrigin:     "http://localhost:8080",
		},
		Security: SecurityConfig{
			CSRFSecret:        "change-this-to-a-random-32-byte-string",
			RateLimitRequests: 100,
			RateLimitWindow:   1 * time.Minute,
			BCryptCost:        12,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Templates: TemplatesConfig{
			HotReload: true,
		},
		Static: StaticConfig{
			HotReload: true,
		},
	}

	// Try to read config file
	if data, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Apply environment variable overrides
	cfg.applyEnvOverrides()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// applyEnvOverrides applies environment variable overrides to config
func (c *Config) applyEnvOverrides() {
	// Server
	if v := os.Getenv("PORT"); v != "" {
		c.Server.Port = v
	}
	if v := os.Getenv("ENVIRONMENT"); v != "" {
		c.Server.Environment = v
	}

	// Database
	if v := os.Getenv("DB_HOST"); v != "" {
		c.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.Database.Port = port
		}
	}
	if v := os.Getenv("DB_USER"); v != "" {
		c.Database.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		c.Database.Password = v
	}
	if v := os.Getenv("DB_DATABASE"); v != "" {
		c.Database.Database = v
	}
	if v := os.Getenv("DB_SSL_MODE"); v != "" {
		c.Database.SSLMode = v
	}

	// Redis
	if v := os.Getenv("REDIS_HOST"); v != "" {
		c.Redis.Host = v
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.Redis.Port = port
		}
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		c.Redis.Password = v
	}
	if v := os.Getenv("REDIS_DATABASE"); v != "" {
		if db, err := strconv.Atoi(v); err == nil {
			c.Redis.Database = db
		}
	}

	// Auth
	if v := os.Getenv("WEBAUTHN_RP_NAME"); v != "" {
		c.Auth.WebAuthnRPName = v
	}
	if v := os.Getenv("WEBAUTHN_RP_ID"); v != "" {
		c.Auth.WebAuthnRPID = v
	}
	if v := os.Getenv("WEBAUTHN_RP_ORIGIN"); v != "" {
		c.Auth.WebAuthnRPOrigin = v
	}

	// Security
	if v := os.Getenv("CSRF_SECRET"); v != "" {
		c.Security.CSRFSecret = v
	}

	// Logging
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		c.Logging.Level = v
	}
	if v := os.Getenv("LOG_FORMAT"); v != "" {
		c.Logging.Format = v
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate environment
	if c.Server.Environment != "development" &&
		c.Server.Environment != "staging" &&
		c.Server.Environment != "production" {
		return fmt.Errorf("invalid environment: %s", c.Server.Environment)
	}

	// Validate database config
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Port == 0 {
		return fmt.Errorf("database port is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}

	// Validate Redis config
	if c.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}
	if c.Redis.Port == 0 {
		return fmt.Errorf("redis port is required")
	}

	// Validate auth config
	if c.Auth.WebAuthnRPID == "" {
		return fmt.Errorf("webauthn RP ID is required")
	}
	if c.Auth.WebAuthnRPOrigin == "" {
		return fmt.Errorf("webauthn RP origin is required")
	}

	// Warn about default CSRF secret in production
	if c.Server.Environment == "production" &&
		c.Security.CSRFSecret == "change-this-to-a-random-32-byte-string" {
		return fmt.Errorf("must set a secure CSRF secret in production")
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

// DatabaseDSN returns the PostgreSQL connection string
func (c *Config) DatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Database,
		c.Database.SSLMode,
	)
}

// RedisAddr returns the Redis connection address
func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}
