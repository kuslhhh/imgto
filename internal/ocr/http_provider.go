package ocr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"time"
)

// HTTPProvider communicates with a remote OCR service over HTTP.
// The remote service is expected to be a PaddleOCR FastAPI server
// (Phase 3) that accepts multipart image uploads and returns JSON.
type HTTPProvider struct {
	client     *http.Client
	baseURL    string
	timeout    time.Duration
	maxRetries int
}

// NewHTTPProvider creates a new HTTP OCR provider.
func NewHTTPProvider(cfg ProviderConfig) *HTTPProvider {
	transport := &http.Transport{
		MaxIdleConns:        20,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
	}

	return &HTTPProvider{
		client: &http.Client{
			Transport: transport,
			Timeout:   cfg.Timeout,
		},
		baseURL:    fmt.Sprintf("%s:%d", cfg.ServiceURL, cfg.ServicePort),
		timeout:    cfg.Timeout,
		maxRetries: cfg.MaxRetries,
	}
}

// Name returns the provider name.
func (p *HTTPProvider) Name() string {
	return "PaddleOCR"
}

// ExtractText sends an image to the OCR service and returns structured results.
// It implements exponential backoff retry for transient failures.
func (p *HTTPProvider) ExtractText(ctx context.Context, image []byte) (*OCRResult, error) {
	var lastErr error

	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 100ms, 200ms, 400ms
			backoff := time.Duration(100*(1<<uint(attempt-1))) * time.Millisecond
			slog.Debug("retrying OCR request",
				slog.Int("attempt", attempt),
				slog.Duration("backoff", backoff),
			)
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("ocr cancelled during retry: %w", ctx.Err())
			case <-time.After(backoff):
			}
		}

		result, err := p.doRequest(ctx, image)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Don't retry on context cancellation or validation errors
		if ctx.Err() != nil {
			return nil, fmt.Errorf("ocr request cancelled: %w", ctx.Err())
		}

		slog.Warn("OCR request failed",
			slog.Int("attempt", attempt+1),
			slog.Int("max_retries", p.maxRetries),
			slog.String("error", err.Error()),
		)
	}

	return nil, fmt.Errorf("ocr request failed after %d retries: %w", p.maxRetries, lastErr)
}

// ocrResponse mirrors the JSON response from the PaddleOCR service.
type ocrResponse struct {
	Text       string           `json:"text"`
	Confidence float64          `json:"confidence"`
	Blocks     []ocrTextBlock   `json:"blocks,omitempty"`
	Tables     []ocrTable       `json:"tables,omitempty"`
	Error      string           `json:"error,omitempty"`
}

type ocrTextBlock struct {
	Text       string    `json:"text"`
	Confidence float64   `json:"confidence"`
	BoundingBox [4][2]int `json:"bounding_box"`
}

type ocrTable struct {
	Rows [][]string `json:"rows"`
}

// doRequest performs a single HTTP request to the OCR service.
func (p *HTTPProvider) doRequest(ctx context.Context, image []byte) (*OCRResult, error) {
	url := fmt.Sprintf("%s/ocr", p.baseURL)

	// Build multipart form request
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("image", "image.png")
	if err != nil {
		return nil, fmt.Errorf("creating form file: %w", err)
	}

	if _, err := part.Write(image); err != nil {
		return nil, fmt.Errorf("writing image to form: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("closing multipart writer: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute request
	start := time.Now()
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ocr service returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var ocrResp ocrResponse
	if err := json.Unmarshal(body, &ocrResp); err != nil {
		return nil, fmt.Errorf("parsing OCR response: %w", err)
	}

	// Check for service-level error
	if ocrResp.Error != "" {
		return nil, fmt.Errorf("ocr service error: %s", ocrResp.Error)
	}

	// Convert to OCRResult
	result := &OCRResult{
		Text:           ocrResp.Text,
		Confidence:     ocrResp.Confidence,
		Blocks:         make([]TextBlock, len(ocrResp.Blocks)),
		Tables:         make([]Table, len(ocrResp.Tables)),
		OCRProvider:    p.Name(),
		ProcessingTime: time.Since(start),
	}

	for i, b := range ocrResp.Blocks {
		result.Blocks[i] = TextBlock{
			Text:        b.Text,
			Confidence:  b.Confidence,
			BoundingBox: b.BoundingBox,
		}
	}

	for i, t := range ocrResp.Tables {
		result.Tables[i] = Table{Rows: t.Rows}
	}

	return result, nil
}
