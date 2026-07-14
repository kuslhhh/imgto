package mcp

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/kush/ocr-mcp/internal/cache"
	"github.com/kush/ocr-mcp/internal/formatter"
	"github.com/kush/ocr-mcp/internal/ocr"
	"github.com/kush/ocr-mcp/internal/workers"

	"github.com/mark3labs/mcp-go/mcp"
)

// readImageTool defines the read_image tool for the MCP server.
func readImageTool() mcp.Tool {
	return mcp.NewTool("read_image",
		mcp.WithDescription("Extract text from an image using OCR. "+
			"Accepts image data as base64-encoded bytes or a file:// URL path to the image. "+
			"Returns structured Markdown with extracted text, tables, and confidence scores."),
		mcp.WithString("image_data",
			mcp.Required(),
			mcp.Description("Base64-encoded image data, or a file:// URL path to the image"),
		),
		mcp.WithString("image_path",
			mcp.Description("File path to the image (alternative to image_data for local files)"),
		),
		mcp.WithString("format",
			mcp.Description("Output format: 'markdown' (default) or 'json'"),
		),
		mcp.WithString("preprocess",
			mcp.Description("Preprocessing options: 'auto' (default), 'none', 'grayscale', 'threshold'"),
		),
	)
}

// handleReadImage handles the read_image MCP tool call.
// It performs auth, rate limiting, validation, caching, preprocessing, OCR, and formatting.
func (s *Server) handleReadImage(
	ctx context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid arguments type"), nil
	}

	// Track metrics
	s.metrics.RequestsTotal.Add(1)
	s.metrics.RequestsActive.Add(1)
	defer s.metrics.RequestsActive.Add(-1)

	// --- Step 0: Authentication ---
	if apiKey, _ := args["api_key"].(string); !s.Authenticate(apiKey) {
		s.metrics.AuthFailures.Add(1)
		return toolErrorToResult(ErrAuthFailed.WithDetails("invalid or missing API key"))
	}

	// --- Step 0b: Rate limiting ---
	clientIP, _ := args["client_id"].(string)
	if clientIP == "" {
		clientIP = "default"
	}
	if !s.ratelimit.Allow(clientIP) {
		s.metrics.RateLimited.Add(1)
		return toolErrorToResult(ErrRateLimited.WithDetails(
			"too many requests. Try again later."))
	}

	// --- Step 1: Extract and validate input ---
	imageData, format, preprocess, err := parseInputArgs(args)
	if err != nil {
		return toolErrorToResult(err)
	}

	slog.Debug("read_image called",
		slog.String("format", format),
		slog.String("preprocess", preprocess),
	)

	// --- Step 2: Decode base64 image data ---
	imageBytes, err := decodeImageInput(imageData)
	if err != nil {
		return toolErrorToResult(err)
	}

	// Validate image size against config limits
	maxSize := s.cfg.MaxImageSizeMB * 1024 * 1024
	if len(imageBytes) > maxSize {
		return toolErrorToResult(
			ErrImageTooLarge.WithDetails(
				fmt.Sprintf("image is %d bytes, max is %d MB", len(imageBytes), s.cfg.MaxImageSizeMB),
			),
		)
	}

	// --- Step 3: Compute image hash ---
	imageHash := computeImageHash(imageBytes)
	slog.Debug("image hash computed",
		slog.String("hash", fmt.Sprintf("%x", imageHash)),
	)

	// --- Step 4: Check cache ---
	cacheKey := cache.HexKey(imageHash)
	if s.cache != nil {
		cachedResult, err := s.cache.Get(ctx, cacheKey)
		if err == nil && cachedResult != nil {
			s.metrics.CacheHits.Add(1)
			slog.Debug("cache hit", slog.String("hash", cacheKey[:16]))
			formatted, fmtErr := formatter.FormatString(cachedResult, format)
			if fmtErr == nil {
				return mcp.NewToolResultText(formatted), nil
			}
			slog.Debug("cache hit but formatting failed, re-processing", slog.String("error", fmtErr.Error()))
		}
	}

	s.metrics.CacheMisses.Add(1)

	// --- Step 5: Preprocess image ---
	processedBytes := imageBytes
	if s.preproc != nil {
		result, err := s.preproc.Process(ctx, imageBytes, preprocess)
		if err != nil {
			return toolErrorToResult(NewToolError(ErrCodeInternal, "image preprocessing failed").WithDetails(err.Error()))
		}
		processedBytes = result
		slog.Debug("image preprocessed",
			slog.String("mode", preprocess),
			slog.Int("original_size", len(imageBytes)),
			slog.Int("processed_size", len(processedBytes)),
		)
	}

	// --- Step 6: Send to OCR service ---
	// Use processed bytes for OCR
	// Route through worker pool if available; fall back to direct call.
	if s.ocr != nil {
		var ocrResult *ocr.OCRResult
		var err error

		if s.pool != nil {
			// Route through worker pool
			resultCh := make(chan *workers.JobResult, 1)
			poolJob := workers.Job{
				ID:      fmt.Sprintf("%x", imageHash)[:12],
				Image:   processedBytes,
				Timeout: s.cfg.JobTimeout,
				Result:  resultCh,
				Ctx:     ctx,
			}

			// Submit job to pool (blocks if queue is full — provides backpressure)
			if submitErr := s.pool.Submit(poolJob); submitErr != nil {
				return toolErrorToResult(NewToolError(ErrCodeInternal, "failed to submit OCR job").WithDetails(submitErr.Error()))
			}

			select {
			case jobResult := <-resultCh:
				if jobResult != nil {
					if jobResult.Err != nil {
						s.metrics.OCRErrors.Add(1)
						return toolErrorToResult(ErrOCRFailed.WithDetails(jobResult.Err.Error()))
					}
					ocrResult = jobResult.Result
				} else {
					return toolErrorToResult(ErrOCRFailed.WithDetails("OCR job returned nil result"))
				}
			case <-ctx.Done():
				return toolErrorToResult(ErrTimeout.WithDetails("request cancelled while waiting for OCR"))
			}
		} else {
			// Direct call (no pool)
			ocrResult, err = s.ocr.ExtractText(ctx, processedBytes)
			if err != nil {
				s.metrics.OCRErrors.Add(1)
				return toolErrorToResult(ErrOCRFailed.WithDetails(err.Error()))
			}
		}

		slog.Debug("OCR completed",
			slog.String("hash", fmt.Sprintf("%x", imageHash)),
			slog.Float64("confidence", ocrResult.Confidence),
			slog.Int("text_length", len(ocrResult.Text)),
		)

		// --- Step 7: Format result ---
		formatted, err := formatter.FormatString(ocrResult, format)
		if err != nil {
			return toolErrorToResult(NewToolError(ErrCodeInternal, "failed to format OCR result").WithDetails(err.Error()))
		}

		// --- Step 8: Cache result ---
		if s.cache != nil {
			if err := s.cache.Set(ctx, cacheKey, ocrResult, s.cfg.CacheTTL); err != nil {
				slog.Warn("failed to cache result", slog.String("error", err.Error()))
			} else {
				slog.Debug("result cached", slog.String("key", cacheKey[:16]))
			}
		}

		return mcp.NewToolResultText(formatted), nil
	}

	slog.Debug("no OCR provider configured")
	result := fmt.Sprintf("# OCR Result\n\n"+
		"## Extracted Text\n\n"+
		"_No OCR provider is configured. The server started without an OCR backend._\n\n"+
		"## Confidence\n\nN/A\n\n---\n\n"+
		"*Set `OCR_SERVICE_URL` and `OCR_SERVICE_PORT` environment variables to configure the OCR service.*\n")

	return mcp.NewToolResultText(result), nil
}

// parseInputArgs extracts and validates arguments from the MCP tool request.
func parseInputArgs(args map[string]interface{}) (imageData string, format string, preprocess string, err error) {
	// Get image_data
	imageData, ok := args["image_data"].(string)
	if !ok || imageData == "" {
		// Try image_path as fallback
		imagePath, ok := args["image_path"].(string)
		if !ok || imagePath == "" {
			return "", "", "", ErrImageRequired
		}
		// file:// URL handling
		if strings.HasPrefix(imagePath, "file://") {
			imageData = imagePath
		} else {
			// Assume it's a local path — we'll need the MCP client to resolve it
			imageData = "file://" + imagePath
		}
	}

	// Get format (optional, defaults to "markdown")
	if f, ok := args["format"].(string); ok && f != "" {
		format = f
	} else {
		format = "markdown"
	}

	// Get preprocessing (optional, defaults to "auto")
	if p, ok := args["preprocess"].(string); ok && p != "" {
		preprocess = p
	} else {
		preprocess = "auto"
	}

	return imageData, format, preprocess, nil
}

// decodeImageInput decodes a base64-encoded image string.
// Supports both raw base64 and data URIs (e.g., data:image/png;base64,...).
func decodeImageInput(input string) ([]byte, error) {
	// Handle data URIs
	if strings.HasPrefix(input, "data:") {
		// Format: data:[mime];base64,[data]
		parts := strings.SplitN(input, ",", 2)
		if len(parts) != 2 {
			return nil, NewToolError(ErrCodeValidation, "invalid data URI format")
		}
		input = parts[1]
	}

	// Handle file:// URLs — these won't have data, so return error
	if strings.HasPrefix(input, "file://") {
		return nil, NewToolError(ErrCodeValidation,
			"file:// URLs require the MCP client to resolve the file. "+
				"Use the image_path parameter or provide base64-encoded image_data instead.")
	}

	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		// Try raw URL encoding (no padding)
		decoded, err = base64.RawURLEncoding.DecodeString(input)
		if err != nil {
			return nil, NewToolError(ErrCodeValidation,
				"invalid base64 encoding: image_data must be base64-encoded image bytes")
		}
	}

	return decoded, nil
}

// computeImageHash computes a SHA-256 hash of the image bytes.
func computeImageHash(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

// toolErrorToResult converts a ToolError to an MCP CallToolResult.
// This provides structured, machine-readable error responses to the LLM.
func toolErrorToResult(err error) (*mcp.CallToolResult, error) {
	var toolErr *ToolError
	if errors.As(err, &toolErr) {
		slog.Warn("tool error",
			slog.String("code", string(toolErr.Code)),
			slog.String("message", toolErr.Message),
		)
		return mcp.NewToolResultError(toolErr.Error()), nil
	}
	return mcp.NewToolResultError(err.Error()), nil
}


