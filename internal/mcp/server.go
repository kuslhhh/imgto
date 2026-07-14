// Package mcp provides the MCP (Model Context Protocol) server implementation
// for the OCR service. It defines tools that text-only LLMs can call to
// extract text from images.
package mcp

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kush/ocr-mcp/configs"
	"github.com/kush/ocr-mcp/internal/cache"
	"github.com/kush/ocr-mcp/internal/ocr"
	"github.com/kush/ocr-mcp/internal/preprocess"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Server wraps the MCP server and provides OCR-related tools.
type Server struct {
	mcpServer   *server.MCPServer
	cfg         *configs.Config
	ocr         ocr.OCRProvider
	preproc     *preprocess.Processor
	cache       cache.Cache
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

	// Initialize image preprocessor
	s.preproc = preprocess.NewProcessor(preprocess.Options{
		MaxWidth:       cfg.MaxImageWidth,
		MaxHeight:      cfg.MaxImageHeight,
		AutoPreprocess: cfg.AutoPreprocess,
	})
	slog.Info("image preprocessor initialized",
		slog.Int("max_width", cfg.MaxImageWidth),
		slog.Int("max_height", cfg.MaxImageHeight),
	)

	// Initialize cache
	s.cache = cache.NewMemoryCache(cache.Config{
		MaxSize: cfg.CacheMaxSize,
		TTL:     cfg.CacheTTL,
	})
	slog.Info("cache initialized",
		slog.String("type", "memory"),
		slog.Int("max_size", cfg.CacheMaxSize),
		slog.Duration("ttl", cfg.CacheTTL),
	)

	// Initialize OCR provider
	ocrProvider, err := ocr.NewProvider(ocr.ProviderConfig{
		Type:        "paddleocr",
		ServiceURL:  cfg.OCRServiceURL,
		ServicePort: cfg.OCRServicePort,
		Timeout:     cfg.OCRTimeout,
		MaxRetries:  cfg.OCRMaxRetries,
	})
	if err != nil {
		slog.Warn("failed to initialize OCR provider, server will start with limited functionality",
			slog.String("error", err.Error()),
		)
	} else {
		s.ocr = ocrProvider
		slog.Info("OCR provider initialized", slog.String("provider", ocrProvider.Name()))
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


