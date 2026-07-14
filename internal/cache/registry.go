package cache

import (
	"fmt"
	"log/slog"
)

// NewCache creates a Cache implementation based on the provided configuration.
// Supported types:
//   - "memory" (default): TTL-based in-memory cache with periodic cleanup
//   - "redis": Redis-backed cache (not yet implemented)
func NewCache(cfg Config) (Cache, error) {
	cacheType := cfg.Type
	if cacheType == "" {
		cacheType = "memory"
	}

	switch cacheType {
	case "memory":
		slog.Info("creating in-memory cache",
			slog.Int("max_size", cfg.MaxSize),
			slog.Duration("ttl", cfg.TTL),
		)
		return NewMemoryCache(cfg), nil

	case "redis":
		slog.Info("creating Redis cache",
			slog.String("url", cfg.RedisURL),
		)
		rc, err := NewRedisCache(cfg)
		if err != nil {
			return nil, fmt.Errorf("redis cache: %w", err)
		}
		return rc, nil

	default:
		return nil, fmt.Errorf("unknown cache type: %q (supported: memory, redis)", cacheType)
	}
}
