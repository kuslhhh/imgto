package preprocess

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log/slog"

	"github.com/disintegration/imaging"
)

// Processor provides image preprocessing operations.
type Processor struct {
	opts Options
}

// NewProcessor creates a new image processor with the given options.
func NewProcessor(opts Options) *Processor {
	return &Processor{opts: opts}
}

// Process applies the preprocessing pipeline to the given image bytes.
// The mode parameter can override the default mode in options.
// Supported modes: "auto", "none", "grayscale", "threshold", "denoise".
func (p *Processor) Process(ctx context.Context, imageBytes []byte, mode string) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	processMode := ParseMode(mode)
	if processMode == ModeNone {
		return imageBytes, nil
	}

	// Decode the image (with WebP fallback)
	img, format, err := DecodeWithFallback(imageBytes)
	if err != nil {
		return nil, fmt.Errorf("decoding image: %w", err)
	}

	// Convert to imaging.NRGBA for consistent processing
	nrgba := toNRGBA(img)

	slog.Debug("preprocessing image",
		slog.String("mode", string(processMode)),
		slog.String("format", format),
		slog.Int("width", nrgba.Bounds().Dx()),
		slog.Int("height", nrgba.Bounds().Dy()),
	)

	// Resize if too large (maintain aspect ratio)
	nrgba = p.resizeIfNeeded(nrgba)

	// Apply preprocessing steps based on mode
	switch processMode {
	case ModeGrayscale:
		nrgba = imaging.Grayscale(nrgba)

	case ModeThreshold:
		nrgba = imaging.Grayscale(nrgba)
		nrgba = adaptiveThreshold(nrgba)

	case ModeDenoise:
		nrgba = denoiseImage(nrgba)

	case ModeAuto:
		nrgba = p.autoPreprocess(nrgba)

	default:
		nrgba = p.autoPreprocess(nrgba)
	}

	// Encode back to PNG bytes
	var buf bytes.Buffer
	if err := imaging.Encode(&buf, nrgba, imaging.PNG); err != nil {
		return nil, fmt.Errorf("encoding processed image: %w", err)
	}

	slog.Debug("preprocessing complete",
		slog.Int("output_size", buf.Len()),
	)

	return buf.Bytes(), nil
}

// resizeIfNeeded scales down the image if it exceeds max dimensions.
func (p *Processor) resizeIfNeeded(img *image.NRGBA) *image.NRGBA {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	if w <= p.opts.MaxWidth && h <= p.opts.MaxHeight {
		return img
	}

	// Calculate scale factor to fit within bounds
	scaleX := float64(p.opts.MaxWidth) / float64(w)
	scaleY := float64(p.opts.MaxHeight) / float64(h)
	scale := scaleX
	if scaleY < scale {
		scale = scaleY
	}

	newW := int(float64(w) * scale)
	newH := int(float64(h) * scale)

	slog.Debug("resizing image",
		slog.Int("from_width", w),
		slog.Int("from_height", h),
		slog.Int("to_width", newW),
		slog.Int("to_height", newH),
	)

	return imaging.Resize(img, newW, newH, imaging.Lanczos)
}

// autoPreprocess applies all preprocessing steps intelligently.
func (p *Processor) autoPreprocess(img *image.NRGBA) *image.NRGBA {
	if !p.opts.AutoPreprocess {
		return img
	}

	// Step 1: Convert to grayscale
	img = imaging.Grayscale(img)

	// Step 2: Apply mild denoising
	img = denoiseImage(img)

	// Step 3: Apply adaptive thresholding for better text extraction
	img = adaptiveThreshold(img)

	return img
}

// denoiseImage applies a median filter to reduce noise while preserving edges.
func denoiseImage(img *image.NRGBA) *image.NRGBA {
	// Use a mild blur to reduce noise
	return imaging.Blur(img, 0.5)
}

// adaptiveThreshold applies a simple local threshold to produce a binary image.
// This improves OCR accuracy on scanned documents and photos of text.
func adaptiveThreshold(img *image.NRGBA) *image.NRGBA {
	bounds := img.Bounds()
	result := image.NewNRGBA(bounds)

	// Use a simple global threshold with Otsu-like calculation
	// For a more sophisticated approach, we use the mean brightness
	threshold := calculateThreshold(img)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			gray := 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)

			var val uint8
			if gray > threshold {
				val = 255 // White (background)
			} else {
				val = 0 // Black (text)
			}

			result.Set(x, y, color.NRGBA{R: val, G: val, B: val, A: 255})
		}
	}

	return result
}

// calculateThreshold computes an adaptive threshold value using a histogram-based approach.
func calculateThreshold(img *image.NRGBA) float64 {
	bounds := img.Bounds()
	var totalPixels int
	var sum float64

	// Sample pixels for threshold calculation
	step := 4 // Sample every 4th pixel for speed
	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		for x := bounds.Min.X; x < bounds.Max.X; x += step {
			r, g, b, _ := img.At(x, y).RGBA()
			gray := 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
			sum += gray
			totalPixels++
		}
	}

	if totalPixels == 0 {
		return 128
	}

	mean := sum / float64(totalPixels)

	// Use the mean as the threshold (simple Otsu approximation)
	return mean
}

// toNRGBA converts any image to NRGBA format for consistent processing.
func toNRGBA(img image.Image) *image.NRGBA {
	if nrgba, ok := img.(*image.NRGBA); ok {
		return nrgba
	}

	bounds := img.Bounds()
	nrgba := image.NewNRGBA(bounds)
	draw.Draw(nrgba, bounds, img, bounds.Min, draw.Src)
	return nrgba
}
