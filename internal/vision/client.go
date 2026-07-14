package vision

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

// Client communicates with a remote vision service over HTTP.
// The remote service is expected to be a FastAPI server running a
// vision model (Qwen2.5-VL or Florence-2).
type Client struct {
	client  *http.Client
	baseURL string
	timeout time.Duration
}

// NewClient creates a new vision service HTTP client.
func NewClient(serviceURL string, timeout time.Duration) *Client {
	return &Client{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:    10,
				IdleConnTimeout: 90 * time.Second,
			},
		},
		baseURL: serviceURL,
		timeout: timeout,
	}
}

// Name returns the provider name for this client.
func (c *Client) Name() string {
	return "VisionService"
}

// DescribeImage sends an image to the vision service for semantic description,
// UI component detection, and layout analysis.
func (c *Client) DescribeImage(ctx context.Context, image []byte, detailLevel string) (*VisionResult, error) {
	url := fmt.Sprintf("%s/describe", c.baseURL)

	// Build multipart form request
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add image file
	part, err := writer.CreateFormFile("image", "image.png")
	if err != nil {
		return nil, fmt.Errorf("creating form file: %w", err)
	}
	if _, err := part.Write(image); err != nil {
		return nil, fmt.Errorf("writing image to form: %w", err)
	}

	// Add optional detail level
	if detailLevel != "" {
		writer.WriteField("detail_level", detailLevel)
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

	// Execute
	start := time.Now()
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vision service returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var descResp DescribeResponse
	if err := json.Unmarshal(body, &descResp); err != nil {
		return nil, fmt.Errorf("parsing vision response: %w", err)
	}

	if descResp.Error != "" {
		return nil, fmt.Errorf("vision service error: %s", descResp.Error)
	}

	// Convert to VisionResult
	result := &VisionResult{
		Description:    descResp.Description,
		UIComponents:   make([]UIComponent, len(descResp.UIComponents)),
		Tags:           descResp.Tags,
		VisionProvider: c.Name(),
		ProcessingTime: time.Since(start),
	}

	for i, comp := range descResp.UIComponents {
		result.UIComponents[i] = UIComponent{
			Type:         comp.Type,
			Label:        comp.Label,
			Description:  comp.Description,
			BoundingBox:  comp.BoundingBox,
		}
	}

	result.Layout = LayoutInfo{
		Type:        descResp.Layout.Type,
		Description: descResp.Layout.Description,
		Regions:     make([]LayoutRegion, len(descResp.Layout.Regions)),
	}
	for i, reg := range descResp.Layout.Regions {
		result.Layout.Regions[i] = LayoutRegion{
			Name:        reg.Name,
			Description: reg.Description,
			BoundingBox: reg.BoundingBox,
		}
	}

	slog.Debug("vision completed",
		slog.String("provider", result.VisionProvider),
		slog.Int("components", len(result.UIComponents)),
		slog.Int("tags", len(result.Tags)),
	)

	return result, nil
}

// Health checks if the vision service is reachable.
func (c *Client) Health(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
