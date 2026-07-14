// Package main is the entrypoint for the OCR MCP server.
//
// It initializes configuration, logging, and starts the MCP server
// that provides image-to-text capabilities for text-only LLMs.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kush/ocr-mcp/configs"
	"github.com/kush/ocr-mcp/internal/mcp"
)

func main() {
	// Load configuration
	cfg := configs.LoadConfig()

	// Set up structured logging
	logLevel := parseLogLevel(cfg.LogLevel)
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})))

	adminAddr := fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort+1)

	slog.Info("starting OCR MCP server",
		slog.String("addr", cfg.ServerAddr()),
		slog.String("admin", adminAddr),
		slog.String("ocr_service", cfg.OCRServiceAddr()),
		slog.String("cache_type", cfg.CacheType),
		slog.Int("workers", cfg.WorkerCount),
	)

	// Create MCP server
	server, err := mcp.NewServer(cfg)
	if err != nil {
		slog.Error("failed to create MCP server", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		slog.Info("received signal, shutting down", slog.String("signal", sig.String()))
		cancel()
	}()

	// Start admin server (health + metrics)
	adminServer := mcp.NewAdminServer(server, adminAddr)
	adminServer.Start()

	// Start MCP server
	if err := server.Start(ctx, cfg.ServerAddr()); err != nil {
		slog.Error("server error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Shutdown admin server
	if shutdownErr := adminServer.Shutdown(context.Background()); shutdownErr != nil {
		slog.Warn("admin server shutdown", slog.String("error", shutdownErr.Error()))
	}

	slog.Info("server stopped gracefully")
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}


