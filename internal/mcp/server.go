// Package mcp provides the MCP (Model Context Protocol) server implementation
// for the OCR service. It defines tools that text-only LLMs can call to
// extract text from images.
package mcp

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kush/ocr-mcp/configs"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Server wraps the MCP server and provides OCR-related tools.
type Server struct {
	mcpServer *server.MCPServer
	cfg       *configs.Config
}

// NewServer creates a new MCP server with OCR tools registered.
func NewServer(cfg *configs.Config) (*Server, error) {
	s := &Server{
		mcpServer: server.NewMCPServer(
			"OCR MCP Server",
			"1.0.0",
			server.WithResourceCapabilities(true, true),
			server.WithPromptCapabilities(true),
			server.WithToolCapabilities(true),
			server.WithLogging(),
		),
		cfg: cfg,
	}

	// Register tools
	if err := s.registerTools(); err != nil {
		return nil, fmt.Errorf("registering tools: %w", err)
	}

	return s, nil
}

// Start starts the MCP server and blocks until the context is cancelled.
func (s *Server) Start(ctx context.Context, addr string) error {
	// Use SSE transport for MCP communication
	sseServer := server.NewSSEServer(s.mcpServer,
		server.WithBaseURL(fmt.Sprintf("http://%s", addr)),
		server.WithBasePath("/mcp"),
	)

	slog.Info("MCP server listening",
		slog.String("addr", addr),
		slog.String("endpoint", "/mcp"),
	)

	// Start HTTP server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		if err := sseServer.Start(addr); err != nil {
			errCh <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// Wait for shutdown signal
	select {
	case <-ctx.Done():
		slog.Info("shutting down MCP server")
		return nil
	case err := <-errCh:
		return err
	}
}

// registerTools registers all available MCP tools.
func (s *Server) registerTools() error {
	tools := []mcp.Tool{
		readImageTool(),
	}

	for _, tool := range tools {
		s.mcpServer.AddTool(tool, s.handleReadImage)
		slog.Debug("registered tool", slog.String("name", tool.Name))
	}

	return nil
}

// readImageTool defines the read_image tool for the MCP server.
func readImageTool() mcp.Tool {
	return mcp.NewTool("read_image",
		mcp.WithDescription("Extract text from an image using OCR. "+
			"Accepts image data as base64-encoded bytes or a file path. "+
			"Returns structured Markdown with extracted text, tables, and confidence scores."),
		mcp.WithString("image_data",
			mcp.Required(),
			mcp.Description("Base64-encoded image data, or a file:// URL path to the image"),
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
// This is a placeholder that will be fully implemented in Phase 1.
func (s *Server) handleReadImage(
	ctx context.Context,
	req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	// Extract image data from request
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid arguments type"), nil
	}

	imageData, ok := args["image_data"].(string)
	if !ok {
		return mcp.NewToolResultError("image_data must be a string"), nil
	}

	// TODO: Phase 1 — full implementation
	// - Validate input
	// - Compute image hash
	// - Check cache
	// - Preprocess image
	// - Send to OCR service
	// - Format result as Markdown/JSON
	// - Cache result
	// - Return response

	slog.Debug("read_image called",
		slog.String("image_data_prefix", safePrefix(imageData, 50)),
	)

	result := fmt.Sprintf(`# OCR Result

## Extracted Text

OCR processing not yet implemented. Received image data: %s...

## Confidence

N/A

---

*This is a placeholder response. The OCR pipeline will be implemented in a future phase.*
`, safePrefix(imageData, 30))

	return mcp.NewToolResultText(result), nil
}

// safePrefix returns the first n characters of a string for logging.
func safePrefix(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
