package ocr

import (
	"context"
	"fmt"
)

// GoogleProvider is a stub that will be implemented in a future phase.
// It will use the Google Cloud Vision API for OCR.
type GoogleProvider struct {
	// TODO(Phase 10+): Add Google Vision client configuration
	// apiKey string
	// client *vision.Client
}

// NewGoogleProvider creates a new Google Cloud Vision OCR provider.
func NewGoogleProvider() (*GoogleProvider, error) {
	return nil, fmt.Errorf("Google Cloud Vision provider not yet implemented: " +
		"requires cloud.google.com/go/vision dependency and service account credentials")
}

// Name returns the provider name.
func (p *GoogleProvider) Name() string {
	return "Google Cloud Vision"
}

// ExtractText performs OCR on the provided image.
func (p *GoogleProvider) ExtractText(_ context.Context, _ []byte) (*OCRResult, error) {
	return nil, fmt.Errorf("Google Cloud Vision provider not yet implemented")
}
