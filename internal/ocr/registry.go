package ocr

import (
	"fmt"
	"log/slog"
)

// ErrNoProviderConfigured is returned when no OCR provider has been configured.
var ErrNoProviderConfigured = fmt.Errorf("no OCR provider configured")

// NewProvider creates an OCR provider based on the provided configuration.
// Supported provider types:
//   - "paddleocr" (default): HTTP provider for the PaddleOCR service
//   - "" (empty): returns nil as a no-op, useful when OCR is optional
func NewProvider(cfg ProviderConfig) (OCRProvider, error) {
	providerType := cfg.Type
	if providerType == "" {
		providerType = "paddleocr"
	}

	switch providerType {
	case "paddleocr":
		slog.Info("creating OCR provider",
			slog.String("type", providerType),
			slog.String("url", fmt.Sprintf("%s:%d", cfg.ServiceURL, cfg.ServicePort)),
		)
		return NewHTTPProvider(cfg), nil

	case "tesseract":
		// Tesseract provider is a stub — will be implemented in a future phase.
		slog.Warn("tesseract OCR provider not yet implemented, falling back to no-op")
		return nil, fmt.Errorf("tesseract provider not yet implemented")

	case "google":
		// Google Cloud Vision provider is a stub — will be implemented in a future phase.
		slog.Warn("google Cloud Vision provider not yet implemented, falling back to no-op")
		return nil, fmt.Errorf("google Cloud Vision provider not yet implemented")

	default:
		return nil, fmt.Errorf("unknown OCR provider type: %q", providerType)
	}
}
