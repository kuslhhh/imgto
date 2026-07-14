package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/kush/ocr-mcp/internal/ocr"
)

// RedisCache is a placeholder for the Redis-backed cache implementation.
// It will be fully implemented in a future phase when Redis support is needed.
// The go-redis/v9 dependency will be added at that time.
type RedisCache struct {
	// TODO: Add redis.Client when implementing
	// client *redis.Client
	addr     string
	password string
	db       int
}

// NewRedisCache creates a new Redis cache placeholder.
// Currently returns an error since Redis support is not yet implemented.
func NewRedisCache(cfg Config) (*RedisCache, error) {
	return nil, fmt.Errorf("Redis cache not yet implemented: " +
		"requires github.com/redis/go-redis/v9 dependency and running Redis server. " +
		"Use CACHE_TYPE=memory (default) for now.")
}

// Get retrieves a cached OCR result by key.
func (r *RedisCache) Get(_ context.Context, _ string) (*ocr.OCRResult, error) {
	return nil, ErrNotSupported
}

// Set stores an OCR result with the given key and TTL.
func (r *RedisCache) Set(_ context.Context, _ string, _ *ocr.OCRResult, _ time.Duration) error {
	return ErrNotSupported
}

// Delete removes a cached entry by key.
func (r *RedisCache) Delete(_ context.Context, _ string) error {
	return ErrNotSupported
}

// Close cleans up resources.
func (r *RedisCache) Close() error {
	return nil
}
