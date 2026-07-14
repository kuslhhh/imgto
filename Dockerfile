# =============================================================================
# Stage 1: Build the Go binary
# =============================================================================
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

# Cache dependencies first (layer is reused when go.mod/go.sum don't change)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build a statically-linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -extldflags=-static" \
    -o /build/ocr-mcp ./cmd/server

# =============================================================================
# Stage 2: Minimal runtime image
# =============================================================================
FROM gcr.io/distroless/static-debian12:nonroot

LABEL org.opencontainers.image.title="OCR MCP Server"
LABEL org.opencontainers.image.description="Go MCP server that converts images to structured text for text-only LLMs"
LABEL org.opencontainers.image.version="1.0.0"

COPY --from=builder /build/ocr-mcp /app/ocr-mcp
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER nonroot:nonroot

EXPOSE 7070

ENTRYPOINT ["/app/ocr-mcp"]
