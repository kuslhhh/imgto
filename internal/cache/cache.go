// Package cache provides caching interfaces and implementations for OCR results.
// It supports in-memory caching (default) and Redis caching (optional).
package cache

import (
	"context"
	"errors"
	"time"

	"github.com/kush/ocr-mcp/internal/ocr"
)

// Common cache errors.
var (
	ErrNotFound    = errors.New("cache: key not found")
	ErrCacheMiss   = errors.New("cache: miss")
	ErrCacheFull   = errors.New("cache: full")
	ErrNotSupported = errors.New("cache: operation not supported")
)

// Cache is the interface for OCR result caching.
// Implementations must be safe for concurrent use.
type Cache interface {
	// Get retrieves a cached OCR result by key.
	// Returns ErrNotFound if the key doesn't exist.
	Get(ctx context.Context, key string) (*ocr.OCRResult, error)

	// Set stores an OCR result with the given key and TTL.
	Set(ctx context.Context, key string, result *ocr.OCRResult, ttl time.Duration) error

	// Delete removes a cached entry by key.
	Delete(ctx context.Context, key string) error

	// Close cleans up any resources held by the cache.
	Close() error
}

// Config holds caching configuration.
type Config struct {
	// Type is the cache type: "memory" (default) or "redis".
	Type string

	// TTL is the default time-to-live for cached entries.
	TTL time.Duration

	// MaxSize is the maximum number of entries in the cache (in-memory only).
	MaxSize int

	// Redis connection settings (redis only).
	RedisURL      string
	RedisPassword string
	RedisDB       int
}

// DefaultConfig returns a default cache configuration.
func DefaultConfig() Config {
	return Config{
		Type:    "memory",
		TTL:     1 * time.Hour,
		MaxSize: 1000,
	}
}

// HexKey converts a byte slice (e.g., SHA-256 hash) to a hex string cache key.
func HexKey(hash []byte) string {
	return formatHex(hash)
}

// formatHex formats bytes as a lowercase hex string — avoids importing encoding/hex.
func formatHex(data []byte) string {
	const hexChars = "0123456789abcdef"
	buf := make([]byte, len(data)*2)
	for i, b := range data {
		buf[i*2] = hexChars[b>>4]
		buf[i*2+1] = hexChars[b&0x0F]
	}
	return string(buf)
}
