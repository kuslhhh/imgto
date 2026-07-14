package preprocess

import (
	"bytes"
	"fmt"
	"image"
	"strings"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
)

// init registers extended image format decoders so that image.Decode
// can handle formats beyond the standard library defaults (PNG, JPEG, GIF).
func init() {
	image.RegisterFormat("bmp", "BM", bmp.Decode, bmp.DecodeConfig)
	image.RegisterFormat("tiff", "II\x2A\x00", tiff.Decode, tiff.DecodeConfig)
	image.RegisterFormat("tiff", "MM\x00\x2A", tiff.Decode, tiff.DecodeConfig)
	// WebP doesn't have a RegisterFormat-compatible magic header,
	// so we handle it separately via DetectFormat and direct decode.
}

// SupportedImageFormats returns a list of supported image format extensions.
func SupportedImageFormats() []string {
	return []string{"jpg", "jpeg", "png", "gif", "bmp", "tiff", "tif", "webp"}
}

// IsSupportedFormat checks if the given extension is a supported image format.
func IsSupportedFormat(ext string) bool {
	ext = strings.TrimPrefix(strings.ToLower(ext), ".")
	for _, supported := range SupportedImageFormats() {
		if ext == supported {
			return true
		}
	}
	return false
}

// DetectFormat attempts to detect the image format from the file header (magic bytes).
func DetectFormat(data []byte) (string, error) {
	if len(data) < 12 {
		return "", fmt.Errorf("image data too short to detect format")
	}

	switch {
	case len(data) > 8 && string(data[0:8]) == "\x89PNG\r\n\x1a\n":
		return "png", nil
	case len(data) > 2 && data[0] == 0xFF && data[1] == 0xD8:
		return "jpeg", nil
	case len(data) > 6 && string(data[0:6]) == "GIF87a" || string(data[0:6]) == "GIF89a":
		return "gif", nil
	case len(data) > 2 && data[0] == 0x42 && data[1] == 0x4D:
		return "bmp", nil
	case len(data) > 4 && string(data[0:4]) == "II\x2A\x00":
		return "tiff", nil
	case len(data) > 4 && string(data[0:4]) == "MM\x00\x2A":
		return "tiff", nil
	case len(data) > 4 && string(data[0:4]) == "RIFF":
		return "webp", nil
	default:
		_, format, err := image.Decode(bytes.NewReader(data))
		if err == nil && format != "" {
			return format, nil
		}
		return "", fmt.Errorf("unable to detect image format from header")
	}
}

// DecodeWithFallback decodes an image from bytes, trying standard formats first
// then falling back to WebP (which cannot be registered via image.RegisterFormat).
func DecodeWithFallback(data []byte) (image.Image, string, error) {
	// Try standard decoding first (PNG, JPEG, GIF, BMP, TIFF via init registration)
	img, format, err := image.Decode(bytes.NewReader(data))
	if err == nil {
		return img, format, nil
	}

	// Fallback: try WebP
	img, err = webp.Decode(bytes.NewReader(data))
	if err == nil {
		return img, "webp", nil
	}

	return nil, "", fmt.Errorf("decoding image: %w", err)
}
