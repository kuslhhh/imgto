.PHONY: build run test lint clean docker-build docker-run dev help

APP_NAME    := ocr-mcp
BUILD_DIR   := ./build
CMD_DIR     := ./cmd/server
GO_FLAGS    := -ldflags="-s -w"
GO_TEST_FLAGS := -v -race -count=1

# Default target
help:
	@echo "OCR MCP Server"
	@echo ""
	@echo "Usage:"
	@echo "  make build        Build the server binary"
	@echo "  make run          Build and run the server"
	@echo "  make dev          Run with live reload (air)"
	@echo "  make test         Run all tests"
	@echo "  make test-coverage Run tests with coverage report"
	@echo "  make lint         Run golangci-lint"
	@echo "  make fmt          Format code with gofumpt"
	@echo "  make clean        Clean build artifacts"
	@echo "  make docker-build Build Docker image"
	@echo "  make docker-run   Run with Docker Compose"
	@echo ""

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

# Docker
docker-build:
	docker build -t $(APP_NAME) .

docker-run:
	docker-compose up --build

docker-stop:
	docker-compose down

# All checks
check: fmt lint test
	@echo "All checks passed!"
