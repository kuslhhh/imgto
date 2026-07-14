package mcp

import (
	"context"
	"encoding/json"
	"expvar"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// AdminServer provides health check and metrics endpoints for monitoring.
type AdminServer struct {
	server *http.Server
	srv    *Server
}

// NewAdminServer creates an admin HTTP server with health and metrics endpoints.
func NewAdminServer(srv *Server, addr string) *AdminServer {
	mux := http.NewServeMux()
	a := &AdminServer{
		server: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
		srv: srv,
	}

	mux.HandleFunc("/health", a.handleHealth)
	mux.HandleFunc("/metrics", a.handleMetrics)

	return a
}

// Start starts the admin server in a goroutine.
func (a *AdminServer) Start() {
	slog.Info("admin server listening", slog.String("addr", a.server.Addr))
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Warn("admin server error", slog.String("error", err.Error()))
		}
	}()
}

// Shutdown gracefully stops the admin server.
func (a *AdminServer) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}

// handleHealth returns the server health status.
func (a *AdminServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	status := "ok"
	if a.srv.ocr == nil {
		status = "degraded"
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   status,
		"service":  "ocr-mcp-server",
		"version":  "1.0.0",
		"ocr_ready": a.srv.ocr != nil,
		"cache":    a.srv.cfg.CacheType,
	})
}

// handleMetrics returns expvar metrics as JSON.
func (a *AdminServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	expvar.Do(func(kv expvar.KeyValue) {
		fmt.Fprintf(w, "%s: %s\n", kv.Key, kv.Value)
	})
}
