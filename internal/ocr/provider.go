// Package ocr provides the OCR provider interface and implementations
// for extracting text from images using various backends.
package ocr

import (
	"context"
	"time"
)

// OCRResult represents the result of an OCR extraction.
type OCRResult struct {
	// Text is the full extracted text content.
	Text string `json:"text"`
	// Confidence is an overall confidence score (0.0 to 1.0).
	Confidence float64 `json:"confidence"`
	// Blocks are individual text blocks with position and confidence.
	Blocks []TextBlock `json:"blocks,omitempty"`
	// Tables are detected table structures.
	Tables []Table `json:"tables,omitempty"`
	// OCRProvider indicates which provider produced this result.
	OCRProvider string `json:"ocr_provider,omitempty"`
	// ProcessingTime is how long the OCR took.
	ProcessingTime time.Duration `json:"processing_time_ms,omitempty"`
	// Error contains any error message from the provider.
	Error string `json:"error,omitempty"`
}

// TextBlock represents a single block of detected text with its position.
type TextBlock struct {
	// Text is the detected text content.
	Text string `json:"text"`
	// Confidence is the confidence score for this block (0.0 to 1.0).
	Confidence float64 `json:"confidence"`
	// BoundingBox is the position of the text block as four corner points.
	// Each point is [x, y] in pixel coordinates.
	// Order: top-left, top-right, bottom-right, bottom-left.
	BoundingBox [4][2]int `json:"bounding_box"`
}

// Table represents a detected table structure within the image.
type Table struct {
	// Rows contains the table data as rows of cell strings.
	Rows [][]string `json:"rows"`
}

// OCRProvider is the interface that all OCR backends must implement.
type OCRProvider interface {
	// ExtractText performs OCR on the provided image bytes and returns structured results.
	ExtractText(ctx context.Context, image []byte) (*OCRResult, error)

	// Name returns a human-readable name for this provider (e.g., "PaddleOCR", "Tesseract").
	Name() string
}

// ProviderConfig holds configuration for initializing an OCR provider.
type ProviderConfig struct {
	// Type is the provider type (e.g., "paddleocr", "tesseract", "google").
	Type string
	// ServiceURL is the base URL of the OCR HTTP service.
	ServiceURL string
	// ServicePort is the port of the OCR HTTP service.
	ServicePort int
	// Timeout is the maximum time to wait for an OCR request.
	Timeout time.Duration
	// MaxRetries is the maximum number of retry attempts for failed requests.
	MaxRetries int
	// APIKey is an optional API key for cloud providers.
	APIKey string
}
