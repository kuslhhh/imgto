package mcp

import (
	"context"
	"log/slog"

	"github.com/kush/ocr-mcp/internal/formatter"
	"github.com/mark3labs/mcp-go/mcp"
)

// describeImageTool defines the describe_image tool for the MCP server.
func describeImageTool() mcp.Tool {
	return mcp.NewTool("describe_image",
		mcp.WithDescription("Analyze an image and provide a semantic description, UI component detection, "+
			"and layout analysis. Uses a vision-language model (Florence-2 / Qwen2.5-VL). "+
			"Returns structured Markdown with image description, UI components, and layout information."),
		mcp.WithString("image_data",
			mcp.Required(),
			mcp.Description("Base64-encoded image data"),
		),
		mcp.WithString("detail_level",
			mcp.Description("Analysis detail: 'basic', 'detailed' (default), or 'ui' for UI-specific analysis"),
		),
		mcp.WithString("format",
			mcp.Description("Output format: 'markdown' (default) or 'json'"),
		),
	)
}

// handleDescribeImage handles the describe_image MCP tool call.
func (s *Server) handleDescribeImage(
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

	// Extract arguments
	imageData, _ := args["image_data"].(string)
	if imageData == "" {
		return toolErrorToResult(ErrImageRequired)
	}

	detailLevel, _ := args["detail_level"].(string)
	if detailLevel == "" {
		detailLevel = "detailed"
	}

	// Get output format (default to markdown using ParseFormat for consistency)
	formatRaw, _ := args["format"].(string)
	format := formatter.ParseFormat(formatRaw)

	// Check that vision service is configured
	if s.vision == nil {
		return toolErrorToResult(ErrVisionNotConfigured)
	}

	// Decode image input (base64, data URI, or file path)
	imageBytes, err := decodeImageInput(imageData)
	if err != nil {
		return toolErrorToResult(err)
	}

	slog.Debug("describe_image called",
		slog.String("detail_level", detailLevel),
		slog.String("format", string(format)),
		slog.Int("image_size", len(imageBytes)),
	)

	// Call vision service
	visionResult, err := s.vision.DescribeImage(ctx, imageBytes, detailLevel)
	if err != nil {
		slog.Warn("vision service failed",
			slog.String("error", err.Error()))
		return toolErrorToResult(ErrVisionFailed.WithDetails(err.Error()))
	}

	slog.Debug("vision completed",
		slog.String("description_prefix", safePrefix(visionResult.Description, 60)),
		slog.Int("components", len(visionResult.UIComponents)),
	)

	// Format result
	var formatted string
	switch format {
	case formatter.FormatJSON:
		formatted = formatter.FormatVisionJSON(visionResult)
	default:
		formatted = formatter.FormatVision(visionResult, detailLevel)
	}
	return mcp.NewToolResultText(formatted), nil
}

// safePrefix returns the first n characters of a string.
func safePrefix(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
