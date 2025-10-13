package storage

import (
	"context"
	"fmt"
	"time"

	"circles.diy/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Redis manages the Redis client connection
type Redis struct {
	Client *redis.Client
	logger *zap.Logger
}

// NewRedis creates a new Redis client
func NewRedis(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*Redis, error) {
	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr(),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.Database,
		MaxRetries:   cfg.Redis.MaxRetries,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	logger.Info("redis connection established",
		zap.String("addr", cfg.RedisAddr()),
		zap.Int("database", cfg.Redis.Database),
	)

	return &Redis{
		Client: client,
		logger: logger,
	}, nil
}

// Close closes the Redis client connection
func (r *Redis) Close() error {
	r.logger.Info("closing redis connection")
	return r.Client.Close()
}

// HealthCheck performs a health check on Redis
func (r *Redis) HealthCheck(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

// Set stores a key-value pair with optional expiration
func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value by key
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

// Delete removes a key
func (r *Redis) Delete(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists
func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.Client.Exists(ctx, key).Result()
	return count > 0, err
}

// Expire sets an expiration time on a key
func (r *Redis) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.Client.Expire(ctx, key, expiration).Err()
}

// HSet sets a hash field
func (r *Redis) HSet(ctx context.Context, key string, field string, value interface{}) error {
	return r.Client.HSet(ctx, key, field, value).Err()
}

// HGet retrieves a hash field
func (r *Redis) HGet(ctx context.Context, key string, field string) (string, error) {
	return r.Client.HGet(ctx, key, field).Result()
}

// HGetAll retrieves all fields of a hash
func (r *Redis) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.Client.HGetAll(ctx, key).Result()
}

// HDel deletes hash fields
func (r *Redis) HDel(ctx context.Context, key string, fields ...string) error {
	return r.Client.HDel(ctx, key, fields...).Err()
}

// Incr increments a counter
func (r *Redis) Incr(ctx context.Context, key string) (int64, error) {
	return r.Client.Incr(ctx, key).Result()
}

// Decr decrements a counter
func (r *Redis) Decr(ctx context.Context, key string) (int64, error) {
	return r.Client.Decr(ctx, key).Result()
}
