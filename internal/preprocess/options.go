// Package preprocess provides image preprocessing capabilities for OCR.
//
// It supports multiple preprocessing modes and configurable options
// for improving OCR accuracy on various types of images.
package preprocess

import (
	"errors"
	"strings"
)

// PreprocessMode represents the type of preprocessing to apply.
type PreprocessMode string

const (
	// ModeAuto applies all preprocessing steps adaptively based on image analysis.
	ModeAuto PreprocessMode = "auto"
	// ModeNone skips preprocessing entirely.
	ModeNone PreprocessMode = "none"
	// ModeGrayscale converts the image to grayscale only.
	ModeGrayscale PreprocessMode = "grayscale"
	// ModeThreshold applies grayscale conversion followed by adaptive thresholding.
	ModeThreshold PreprocessMode = "threshold"
	// ModeDenoise applies denoising only.
	ModeDenoise PreprocessMode = "denoise"
)

// String returns the string representation of the mode.
func (m PreprocessMode) String() string {
	return string(m)
}

// ParseMode parses a preprocessing mode from a string.
// Returns ModeAuto for unknown values as a safe default.
func ParseMode(s string) PreprocessMode {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "none":
		return ModeNone
	case "grayscale":
		return ModeGrayscale
	case "threshold":
		return ModeThreshold
	case "denoise":
		return ModeDenoise
	case "auto", "":
		return ModeAuto
	default:
		return ModeAuto
	}
}

// Options controls the behavior of the preprocessing pipeline.
type Options struct {
	// Mode selects the preprocessing mode.
	Mode PreprocessMode

	// MaxWidth is the maximum image width in pixels.
	// Images wider than this will be resized proportionally.
	MaxWidth int

	// MaxHeight is the maximum image height in pixels.
	// Images taller than this will be resized proportionally.
	MaxHeight int

	// AutoPreprocess enables automatic preprocessing in auto mode.
	AutoPreprocess bool
}

// Validate checks that the options are valid.
func (o Options) Validate() error {
	if o.MaxWidth < 0 {
		return errors.New("max_width must be non-negative")
	}
	if o.MaxHeight < 0 {
		return errors.New("max_height must be non-negative")
	}
	return nil
}

// DefaultOptions returns sensible defaults for preprocessing.
func DefaultOptions() Options {
	return Options{
		Mode:           ModeAuto,
		MaxWidth:       4096,
		MaxHeight:      4096,
		AutoPreprocess: true,
	}
}
