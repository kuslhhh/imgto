package mcp

import (
	"sync"
	"time"
)

// RateLimiter implements a simple token bucket rate limiter.
// It allows up to `rate` requests per minute per key (e.g., client IP).
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     int
	interval time.Duration
}

type bucket struct {
	tokens    int
	lastCheck time.Time
}

// NewRateLimiter creates a rate limiter allowing `rate` requests per minute.
func NewRateLimiter(rate int) *RateLimiter {
	if rate <= 0 {
		rate = 60
	}
	return &RateLimiter{
		buckets:  make(map[string]*bucket),
		rate:     rate,
		interval: time.Minute,
	}
}

// Allow checks if a request for the given key should be allowed.
// Returns true if the request is within the rate limit.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, exists := rl.buckets[key]
	if !exists {
		rl.buckets[key] = &bucket{tokens: rl.rate - 1, lastCheck: time.Now()}
		return true
	}

	// Refill tokens based on elapsed time
	elapsed := time.Since(b.lastCheck)
	// Refill proportionally: rate tokens per interval
	refill := int(elapsed * time.Duration(rl.rate) / rl.interval)
	if refill > 0 {
		b.tokens += refill
		if b.tokens > rl.rate {
			b.tokens = rl.rate
		}
		b.lastCheck = b.lastCheck.Add(time.Duration(refill) * rl.interval / time.Duration(rl.rate))
	}

	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}
