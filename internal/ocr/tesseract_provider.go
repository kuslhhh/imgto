package ocr

import (
	"context"
	"fmt"
)

// TesseractProvider is a stub that will be implemented in a future phase.
// It will use gosseract (Go bindings for Tesseract) to perform OCR locally.
type TesseractProvider struct {
	// TODO(Phase 10+): Add Tesseract engine configuration
	// lang string
	// engineMode int
}

// NewTesseractProvider creates a new Tesseract OCR provider.
func NewTesseractProvider() (*TesseractProvider, error) {
	return nil, fmt.Errorf("Tesseract provider not yet implemented: " +
		"requires gosseract dependency and Tesseract binaries installed on the system")
}

// Name returns the provider name.
func (p *TesseractProvider) Name() string {
	return "Tesseract"
}

// ExtractText performs OCR on the provided image.
func (p *TesseractProvider) ExtractText(_ context.Context, _ []byte) (*OCRResult, error) {
	return nil, fmt.Errorf("Tesseract provider not yet implemented")
}
