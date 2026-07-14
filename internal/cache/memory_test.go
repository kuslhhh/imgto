package cache

import (
	"context"
	"testing"
	"time"

	"github.com/kush/ocr-mcp/internal/ocr"
)

func TestMemoryCacheSetGet(t *testing.T) {
	c := NewMemoryCache(Config{MaxSize: 100, TTL: 5 * time.Minute})
	defer c.Close()

	key := "test-key"
	result := &ocr.OCRResult{
		Text:       "Hello",
		Confidence: 0.95,
		Blocks:     []ocr.TextBlock{{Text: "Hello", Confidence: 0.95}},
	}

	err := c.Set(context.Background(), key, result, 5*time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	got, err := c.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.Text != "Hello" {
		t.Errorf("Text = %q, want %q", got.Text, "Hello")
	}
	if got.Confidence != 0.95 {
		t.Errorf("Confidence = %v, want %v", got.Confidence, 0.95)
	}
	if len(got.Blocks) != 1 {
		t.Errorf("len(Blocks) = %d, want %d", len(got.Blocks), 1)
	}
}

func TestMemoryCacheNotFound(t *testing.T) {
	c := NewMemoryCache(Config{MaxSize: 100, TTL: 5 * time.Minute})
	defer c.Close()

	_, err := c.Get(context.Background(), "nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryCacheDelete(t *testing.T) {
	c := NewMemoryCache(Config{MaxSize: 100, TTL: 5 * time.Minute})
	defer c.Close()

	c.Set(context.Background(), "key", &ocr.OCRResult{Text: "test"}, time.Minute)
	c.Delete(context.Background(), "key")

	_, err := c.Get(context.Background(), "key")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestMemoryCacheExpiry(t *testing.T) {
	c := NewMemoryCache(Config{MaxSize: 100, TTL: 5 * time.Minute})
	defer c.Close()

	// Use a very short TTL (1ms) so it expires before we read it
	c.Set(context.Background(), "key", &ocr.OCRResult{Text: "test"}, 1*time.Millisecond)
	time.Sleep(5 * time.Millisecond)

	_, err := c.Get(context.Background(), "key")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound for expired entry, got %v", err)
	}
}

func TestMemoryCacheEviction(t *testing.T) {
	c := NewMemoryCache(Config{MaxSize: 2, TTL: 5 * time.Minute})
	defer c.Close()

	c.Set(context.Background(), "key1", &ocr.OCRResult{Text: "one"}, time.Minute)
	c.Set(context.Background(), "key2", &ocr.OCRResult{Text: "two"}, time.Minute)
	c.Set(context.Background(), "key3", &ocr.OCRResult{Text: "three"}, time.Minute)

	if c.Len() > 2 {
		t.Errorf("expected at most 2 entries after eviction, got %d", c.Len())
	}
}

func TestHexKey(t *testing.T) {
	hash := []byte{0x00, 0xFF, 0xAB, 0x12}
	key := HexKey(hash)
	if key != "00ffab12" {
		t.Errorf("HexKey = %q, want %q", key, "00ffab12")
	}
}
