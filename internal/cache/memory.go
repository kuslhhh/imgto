package cache

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/kush/ocr-mcp/internal/ocr"
)

// entry holds a cached value with its expiration time.
type entry struct {
	data   []byte
	expiry time.Time
}

// MemoryCache is a TTL-based in-memory cache safe for concurrent use.
// It periodically evicts expired entries (every 5 minutes).
type MemoryCache struct {
	mu       sync.RWMutex
	items    map[string]*entry
	maxSize  int
	defaultTTL time.Duration
	stopCh   chan struct{}
}

// NewMemoryCache creates a new in-memory cache with the given configuration.
func NewMemoryCache(cfg Config) *MemoryCache {
	c := &MemoryCache{
		items:      make(map[string]*entry),
		maxSize:    cfg.MaxSize,
		defaultTTL: cfg.TTL,
		stopCh:     make(chan struct{}),
	}

	// Start periodic cleanup goroutine
	go c.cleanup()

	return c
}

// Get retrieves a cached OCR result by key.
func (c *MemoryCache) Get(_ context.Context, key string) (*ocr.OCRResult, error) {
	c.mu.RLock()
	e, ok := c.items[key]
	c.mu.RUnlock()

	if !ok {
		return nil, ErrNotFound
	}

	// Check expiry
	if !e.expiry.IsZero() && time.Now().After(e.expiry) {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return nil, ErrNotFound
	}

	// Deserialize
	var result ocr.OCRResult
	if err := json.Unmarshal(e.data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Set stores an OCR result with the given key and TTL.
func (c *MemoryCache) Set(_ context.Context, key string, result *ocr.OCRResult, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict oldest entry if at capacity and key doesn't exist
	if len(c.items) >= c.maxSize {
		if _, exists := c.items[key]; !exists {
			c.evictOne()
		}
	}

	c.items[key] = &entry{
		data:   data,
		expiry: time.Now().Add(ttl),
	}

	return nil
}

// Delete removes a cached entry by key.
func (c *MemoryCache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
	return nil
}

// Close stops the cleanup goroutine and clears the cache.
func (c *MemoryCache) Close() error {
	close(c.stopCh)
	c.mu.Lock()
	c.items = make(map[string]*entry)
	c.mu.Unlock()
	return nil
}

// Len returns the number of items in the cache (for monitoring).
func (c *MemoryCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// cleanup periodically removes expired entries (runs every 5 minutes).
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.deleteExpired()
		case <-c.stopCh:
			return
		}
	}
}

// deleteExpired removes all expired entries from the cache.
func (c *MemoryCache) deleteExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, e := range c.items {
		if !e.expiry.IsZero() && now.After(e.expiry) {
			delete(c.items, key)
		}
	}
}

// evictOne removes a single entry when the cache is full.
// Uses a simple approach: remove the first expired entry, or the oldest entry.
func (c *MemoryCache) evictOne() {
	if len(c.items) == 0 {
		return
	}

	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, e := range c.items {
		// Prefer removing expired entries
		if !e.expiry.IsZero() && time.Now().After(e.expiry) {
			delete(c.items, key)
			return
		}
		// Track oldest for fallback
		if first || e.expiry.Before(oldestTime) {
			oldestKey = key
			oldestTime = e.expiry
			first = false
		}
	}

	// If no expired entries found, remove the oldest
	if oldestKey != "" {
		delete(c.items, oldestKey)
		slog.Debug("evicted oldest cache entry", slog.String("key", oldestKey[:16]))
	}
}
