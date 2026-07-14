# OCR MCP Server — Build Plan

## Overview

Build a high-performance Go-based MCP server that converts images to structured text for text-only LLMs (DeepSeek, Kimi, GLM, Minimax, etc.). The server communicates with a long-running Python OCR service (PaddleOCR) over HTTP.

---

## Phase 0: Project Scaffolding & Tooling

### Tasks

- [ ] Initialize Go module (`go mod init github.com/<user>/ocr-mcp`)
- [ ] Install Go MCP SDK (`github.com/mark3labs/mcp-go` or similar)
- [ ] Create project directory structure:
  ```
  ocr-mcp/
  ├── cmd/
  │   └── server/          # Entrypoint
  ├── internal/
  │   ├── mcp/             # MCP tool definitions & handlers
  │   ├── ocr/             # OCR provider interface & implementations
  │   ├── preprocess/      # Image preprocessing
  │   ├── formatter/       # Markdown/JSON output formatting
  │   ├── cache/           # In-memory & Redis cache
  │   ├── vision/          # Optional vision service client
  │   ├── workers/         # Worker pool for concurrent processing
  │   └── utils/           # Shared utilities (hashing, etc.)
  ├── configs/
  │   └── config.go        # Configuration loading
  ├── Dockerfile
  ├── Makefile
  └── go.mod
  ```
- [ ] Create `Makefile` with targets: `build`, `run`, `test`, `lint`, `docker-build`, `docker-run`
- [ ] Set up linting (`golangci-lint`) and formatting conventions
- [ ] Create `.gitignore`

**Estimated effort**: 1 session

---

## Phase 1: Core Go MCP Server

### 1.1 Configuration

- [ ] Define config struct in `configs/config.go`
  - Server port, OCR service URL, cache type (memory/redis), worker count, timeouts, log level
- [ ] Load config from environment variables with sensible defaults
- [ ] Support optional config file (YAML/JSON)

### 1.2 MCP Server Skeleton

- [ ] Create `cmd/server/main.go`:
  - Initialize config
  - Set up logging
  - Initialize MCP server
  - Register tools
  - Start HTTP/SSE transport
  - Handle graceful shutdown (SIGINT/SIGTERM)
- [ ] Create `internal/mcp/server.go`:
  - Wraps MCP SDK server
  - Registers tool handlers
  - Handles request validation

### 1.3 MCP Tool: `read_image`

- [ ] Define tool input schema (image data as base64 or file path)
- [ ] Implement handler in `internal/mcp/tools.go`:
  - Validate input
  - Compute image hash
  - Check cache
  - Preprocess image
  - Send to OCR service
  - Format result as Markdown
  - Cache result
  - Return response

### 1.4 Error Handling

- [ ] Define consistent error types (validation errors, OCR failures, timeouts, cache errors)
- [ ] Return structured error responses via MCP

**Estimated effort**: 2 sessions

---

## Phase 2: OCR Provider Interface & Implementations

### 2.1 Interface Definition

- [ ] Define `internal/ocr/provider.go`:
  ```go
  type OCRResult struct {
      Text       string             `json:"text"`
      Confidence float64            `json:"confidence"`
      Blocks     []TextBlock        `json:"blocks,omitempty"`
      Tables     []Table            `json:"tables,omitempty"`
      Error      string             `json:"error,omitempty"`
  }

  type TextBlock struct {
      Text       string    `json:"text"`
      Confidence float64   `json:"confidence"`
      BoundingBox [4][2]int `json:"bounding_box"`
  }

  type Table struct {
      Rows [][]string `json:"rows"`
  }

  type OCRProvider interface {
      ExtractText(ctx context.Context, image []byte) (*OCRResult, error)
      Name() string
  }
  ```

### 2.2 HTTP Provider (for PaddleOCR service)

- [ ] Create `internal/ocr/http_provider.go`:
  - HTTP client with configurable timeout & retry
  - POST multipart/form-data with image bytes
  - Parse JSON response
  - Convert to `OCRResult`
- [ ] Implement retry logic (exponential backoff, max retries)

### 2.3 Fallback / Local Providers (future)

- [ ] Create `internal/ocr/tesseract_provider.go` (Tesseract via gosseract)
- [ ] Create `internal/ocr/google_provider.go` (Google Cloud Vision API)

### 2.4 Provider Registry

- [ ] Create provider factory that selects implementation by config
- [ ] Allow multiple providers with fallback chain

**Estimated effort**: 2 sessions

---

## Phase 3: Python OCR Service (PaddleOCR)

### 3.1 Service Scaffolding

- [ ] Create `services/ocr/` directory with:
  - `requirements.txt` (paddlepaddle, paddleocr, fastapi, uvicorn, pillow)
  - `app.py` — FastAPI application
  - `ocr_engine.py` — PaddleOCR wrapper
- [ ] Create `Dockerfile.ocr` for the OCR service

### 3.2 FastAPI Endpoints

- [ ] `POST /ocr` — Accept image bytes, run OCR, return JSON
- [ ] `GET /health` — Health check endpoint
- [ ] Support optional preprocessing params (grayscale, threshold, denoise, etc.)

### 3.3 PaddleOCR Wrapper

- [ ] Load PaddleOCR model at startup (not per-request)
- [ ] Accept image bytes as input
- [ ] Extract text blocks with confidence scores and bounding boxes
- [ ] Return structured JSON response
- [ ] Handle errors gracefully (no text found, low confidence, etc.)

### 3.4 Testing

- [ ] Write test with sample images
- [ ] Benchmark latency

**Estimated effort**: 2 sessions

---

## Phase 4: Image Preprocessing

### 4.1 Preprocessing Pipeline

- [ ] Create `internal/preprocess/preprocess.go`:
  - Load image (JPEG, PNG, WebP, BMP, TIFF)
  - Convert to grayscale (optional)
  - Apply adaptive thresholding (optional)
  - Denoise (optional)
  - Deskew (optional)
  - Resize if too large (maintain aspect ratio)
  - Return processed image bytes

### 4.2 Image Utilities

- [ ] Create `internal/preprocess/image.go`:
  - Image decoding/encoding
  - Format validation
  - Size limits (configurable max dimensions / file size)
- [ ] Use Go's `image` stdlib or `github.com/disintegration/imaging`

### 4.3 Testing

- [ ] Unit tests for each preprocessing step
- [ ] Test with various image formats

**Estimated effort**: 1 session

---

## Phase 5: Output Formatting

### 5.1 Markdown Formatter

- [ ] Create `internal/formatter/markdown.go`:
  ```
  # OCR Result

  ## Extracted Text

  [text content]

  ## Tables

  | Col1 | Col2 |
  |------|------|
  | ...  | ...  |

  ## UI Components

  - Button: "Submit"
  - Input field: "Email"

  ## Layout

  [layout description]

  ## Confidence

  98%
  ```

### 5.2 JSON Formatter

- [ ] Create `internal/formatter/json.go`:
  - Structured JSON output
  - Suitable for programmatic consumption

### 5.3 Formatter Selection

- [ ] Configurable output format (markdown/json/both)
- [ ] Allow LLM-specific format customization

**Estimated effort**: 1 session

---

## Phase 6: Caching

### 6.1 Cache Interface

- [ ] Create `internal/cache/cache.go`:
  ```go
  type Cache interface {
      Get(ctx context.Context, key string) (*OCRResult, error)
      Set(ctx context.Context, key string, result *OCRResult, ttl time.Duration) error
      Delete(ctx context.Context, key string) error
      Close() error
  }
  ```

### 6.2 In-Memory Cache

- [ ] Create `internal/cache/memory.go`:
  - `sync.RWMutex`-protected map or `sync.Map`
  - Configurable TTL with periodic cleanup
  - Max size limit with LRU eviction (optional)

### 6.3 Redis Cache

- [ ] Create `internal/cache/redis.go`:
  - `go-redis` client
  - JSON serialization of `OCRResult`
  - Configurable TTL

### 6.4 Cache Key Generation

- [ ] SHA-256 hashing of raw image bytes in `internal/utils/hash.go`
- [ ] Optional metadata-based key prefixes

**Estimated effort**: 1 session

---

## Phase 7: Worker Pool

### 7.1 Pool Implementation

- [ ] Create `internal/workers/pool.go`:
  ```go
  type Pool struct {
      workers    int
      jobQueue   chan Job
      ctx        context.Context
      cancel     context.CancelFunc
      wg         sync.WaitGroup
  }

  type Job struct {
      ID      string
      Image   []byte
      Options Options
      Result  chan<- *Result
  }
  ```

### 7.2 Job Scheduling

- [ ] Configurable worker count (default: `runtime.NumCPU()`)
- [ ] Buffered job queue
- [ ] Graceful shutdown (drain pending jobs)
- [ ] Timeout per job
- [ ] Error handling per job (isolate failures)

**Estimated effort**: 1 session

---

## Phase 8: Vision Service (Optional)

### 8.1 Vision Service Client

- [ ] Create `internal/vision/client.go`:
  - HTTP client for vision service (Qwen2.5-VL / Florence-2)
  - Request/response types
  - Timeout & retry configuration

### 8.2 UI Analysis

- [ ] UI layout detection
- [ ] Component identification (buttons, inputs, modals, navigation)
- [ ] Semantic description generation

### 8.3 Integration

- [ ] Merge OCR text output with vision semantic output
- [ ] Produce richer Markdown with UI context

### 8.4 Vision Python Service

- [ ] Create `services/vision/` with FastAPI app
- [ ] Load vision model at startup
- [ ] `POST /describe` endpoint
- [ ] `Dockerfile.vision`

**Estimated effort**: 2 sessions

---

## Phase 9: Docker & Deployment

### 9.1 Dockerfiles

- [ ] Create `Dockerfile` for Go MCP server (multi-stage build):
  - Stage 1: Build Go binary
  - Stage 2: Minimal `scratch` or `alpine` image
- [ ] Create `Dockerfile.ocr` for Python OCR service
- [ ] Create `Dockerfile.vision` for Python vision service (optional)

### 9.2 Docker Compose

- [ ] Create `docker-compose.yml`:
  ```yaml
  services:
    ocr-mcp:
      build: .
      ports: ["8080:8080"]
      depends_on: [ocr-service]
      environment: [...]

    ocr-service:
      build:
        dockerfile: Dockerfile.ocr
      ports: ["9090:9090"]
      deploy:
        resources:
          reservations:
            devices:
              - capabilities: [gpu]  # optional
  ```

### 9.3 Configuration & Environment

- [ ] Document all env vars in `.env.example`
- [ ] Production config recommendations (resource limits, log levels, etc.)

**Estimated effort**: 1 session

---

## Phase 10: Production Readiness

### 10.1 Logging

- [ ] Structured logging with `slog` (Go 1.21+)
- [ ] Correlation IDs per request
- [ ] Log levels: debug, info, warn, error

### 10.2 Metrics

- [ ] Prometheus metrics endpoint:
  - Request count, latency (histogram), error count
  - OCR latency, cache hit/miss rate, worker pool utilization
  - Active requests gauge

### 10.3 Health Checks

- [ ] MCP server health endpoint
- [ ] OCR service health check (periodic ping)
- [ ] Cache connectivity check
- [ ] Readiness / liveness probes for Docker/K8s

### 10.4 Rate Limiting

- [ ] Configurable per-client rate limiting
- [ ] Token bucket or sliding window

### 10.5 Authentication

- [ ] API key validation for MCP requests
- [ ] Env-var configured keys
- [ ] Future: JWT or OAuth2

**Estimated effort**: 2 sessions

---

## Phase 11: Testing & Quality

### 11.1 Unit Tests

- [ ] All `internal/` packages: `_test.go` files with table-driven tests
- [ ] Mock OCR provider for testing
- [ ] Mock cache for testing
- [ ] Target: >70% coverage

### 11.2 Integration Tests

- [ ] Start OCR service, run MCP server, send real image
- [ ] Test with sample images (receipt, screenshot, document, table)

### 11.3 Benchmarks

- [ ] `go test -bench=.` for preprocessing, formatting, hashing
- [ ] End-to-end latency benchmark
- [ ] Cache performance benchmark

### 11.4 Linting & Static Analysis

- [ ] `golangci-lint` config with common linters
- [ ] Add to CI pipeline and Makefile target

**Estimated effort**: 2 sessions

---

## Phase 12: Documentation

### 12.1 README

- [ ] Project overview
- [ ] Architecture diagram
- [ ] Quick start guide
- [ ] Configuration reference
- [ ] API documentation
- [ ] Development guide

### 12.2 Code Comments

- [ ] Go doc comments on all exported symbols
- [ ] Inline comments for complex logic

### 12.3 Examples

- [ ] Sample curl commands
- [ ] OpenCode/Claude Desktop integration guide
- [ ] Sample outputs (Markdown, JSON)

**Estimated effort**: 1 session

---

## Dependency Graph (Build Order)

```
Phase 0 (Scaffolding)
    │
    ▼
Phase 1 (Go MCP Server) ────────────────┐
    │                                     │
    ├── Phase 2 (OCR Interface) ──────────┤
    │         │                           │
    │         ▼                           │
    │    Phase 3 (Python OCR Service) ────┤
    │                                     │
    ├── Phase 4 (Preprocessing)           │
    ├── Phase 5 (Formatter)               │
    ├── Phase 6 (Cache)                   │
    ├── Phase 7 (Worker Pool)             │
    │                                     │
    ├── Phase 8 (Vision Service) ─────────┤  (Optional — can be skipped)
    │                                     │
    ▼                                     ▼
Phase 9 (Docker & Deployment) ───── Phase 10 (Production Readiness)
    │                                     │
    ▼                                     ▼
Phase 11 (Testing & Quality) ────── Phase 12 (Documentation)
```

**Bold** phases can be done in parallel once Phase 1 is complete:
- Phases 2, 4, 5, 6, 7 can be developed concurrently
- Phase 3 depends on Phase 2 (interface agreement)
- Phase 8 is fully optional
- Phases 9–12 are finalization

---

## Key Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| MCP SDK | `github.com/mark3labs/mcp-go` | Most mature Go MCP SDK |
| Image processing | `github.com/disintegration/imaging` | Simple API, fast, supports common formats |
| HTTP client | Go stdlib `net/http` | Sufficient, no external dep needed |
| Logging | `log/slog` (stdlib, Go 1.21+) | Structured, native, no external dep |
| Config | Environment variables + optional YAML | 12-factor app, overridable |
| Cache | In-memory (default) + Redis (optional) | Zero-dependency start, upgrade path |
| Worker pool | Custom with channels | Simple, idiomatic Go |
| OCR transport | HTTP/REST | Simple, language-agnostic |

---

## Effort Summary

| Phase | Sessions | Dependencies |
|-------|----------|-------------|
| Phase 0 — Scaffolding | 1 | None |
| Phase 1 — MCP Server | 2 | Phase 0 |
| Phase 2 — OCR Interface | 2 | Phase 0 |
| Phase 3 — OCR Service | 2 | Phase 2 |
| Phase 4 — Preprocessing | 1 | Phase 0 |
| Phase 5 — Formatter | 1 | Phase 0 |
| Phase 6 — Caching | 1 | Phase 0 |
| Phase 7 — Worker Pool | 1 | Phase 0 |
| Phase 8 — Vision Service | 2 | Phase 0 (optional) |
| Phase 9 — Docker | 1 | Phases 1-7 |
| Phase 10 — Production | 2 | Phases 1-7 |
| Phase 11 — Testing | 2 | Phases 1-8 |
| Phase 12 — Docs | 1 | All |
| **Total** | **19 sessions** | |
