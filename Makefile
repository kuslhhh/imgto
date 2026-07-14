.PHONY: build run test lint clean docker-build docker-run dev help
.PHONY: ocr-run ocr-docker-build ocr-docker-run ocr-stop up

APP_NAME       := ocr-mcp
BUILD_DIR      := ./build
CMD_DIR        := ./cmd/server
GO_FLAGS       := -ldflags="-s -w"
GO_TEST_FLAGS  := -v -race -count=1

OCR_SERVICE_DIR := services/ocr
OCR_IMAGE_NAME  := ocr-service
OCR_PORT        := 9090

# Default target
help:
	@echo "OCR MCP Server"
	@echo ""
	@echo "Go Server:"
	@echo "  make build        Build the server binary"
	@echo "  make run          Build and run the server"
	@echo "  make dev          Run with live reload (air)"
	@echo ""
	@echo "OCR Service (Python):"
	@echo "  make ocr-run        Run OCR service locally (requires Python deps)"
	@echo "  make ocr-docker-build Build OCR service Docker image"
	@echo "  make ocr-docker-run   Run OCR service in Docker container"
	@echo ""
	@echo "All Services:"
	@echo "  make up             Start everything with Docker Compose"
	@echo "  make ocr-stop       Stop OCR service (docker)"
	@echo ""
	@echo "Quality:"
	@echo "  make test         Run all Go tests"
	@echo "  make test-coverage Run tests with coverage report"
	@echo "  make lint         Run golangci-lint"
	@echo "  make fmt          Format code with gofumpt"
	@echo "  make clean        Clean build artifacts"
	@echo ""

# ---- Go Server ----

build:
	@mkdir -p $(BUILD_DIR)
	go build $(GO_FLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)
	@echo "Built $(BUILD_DIR)/$(APP_NAME)"

run: build
	@echo "Starting $(APP_NAME)..."
	$(BUILD_DIR)/$(APP_NAME)

dev:
	@which air > /dev/null 2>&1 || { echo "Installing air..."; go install github.com/air-verse/air@latest; }
	air

# ---- OCR Service (Python) ----

ocr-run:
	@echo "Starting OCR service..."
	@cd $(OCR_SERVICE_DIR) && python app.py

ocr-docker-build:
	docker build -f $(OCR_SERVICE_DIR)/Dockerfile.ocr -t $(OCR_IMAGE_NAME) .

ocr-docker-run:
	@echo "Starting OCR service on port $(OCR_PORT)..."
	docker run --rm -p $(OCR_PORT):$(OCR_PORT) --name $(OCR_IMAGE_NAME) $(OCR_IMAGE_NAME)

ocr-stop:
	@docker stop $(OCR_IMAGE_NAME) 2>/dev/null || true
	@echo "OCR service stopped"

# ---- All Services ----

up: ocr-docker-build
	docker-compose up --build

# ---- Quality ----

test:
	go test $(GO_TEST_FLAGS) ./...

test-coverage:
	go test $(GO_TEST_FLAGS) -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-race:
	go test -race -count=1 ./...

bench:
	go test -bench=. -benchmem ./...

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

fmt:
	gofumpt -l -w .

tidy:
	go mod tidy
	go mod verify

clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html tmp/

# ---- Docker ----
docker-build:
	docker build -t $(APP_NAME) .

docker-run:
	docker-compose up --build

docker-stop:
	docker-compose down

# All checks
check: fmt lint test
	@echo "All checks passed!"
