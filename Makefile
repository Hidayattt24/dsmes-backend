# ─────────────────────────────────────────────────────────────────────────────
# DSMES Backend — Makefile
#
# Usage:
#   make dev        Start development server with Air hot-reload
#   make build      Compile production binary
#   make run        Run the compiled binary
#   make test       Run all tests
#   make swag       Generate Swagger documentation
#   make lint       Run golangci-lint
#   make tidy       Tidy and verify go.mod / go.sum
#   make clean      Remove build artifacts
#   make docker-up  Start Docker Compose (PostgreSQL)
#   make docker-down Stop Docker Compose
# ─────────────────────────────────────────────────────────────────────────────

APP_NAME    := dsmes-backend
BINARY      := ./tmp/main
CMD_PATH    := ./cmd/api
DOCKER_COMP := docker-compose.yml

.PHONY: all dev build run test swag lint tidy clean docker-up docker-down help

# Default target
all: build

# ── Development ───────────────────────────────────────────────────────────────

## dev: Start hot-reload development server using Air
dev:
	@echo "Starting development server with Air..."
	air

## build: Compile the production binary
build:
	@echo "Building $(APP_NAME)..."
	go build -ldflags="-s -w" -o $(BINARY) $(CMD_PATH)
	@echo "Binary built at $(BINARY)"

## run: Run the compiled binary directly
run: build
	@echo "Running $(APP_NAME)..."
	$(BINARY)

## test: Run all tests with race detector
test:
	@echo "Running tests..."
	go test -v -race -count=1 ./...

## test-cover: Run tests with HTML coverage report
test-cover:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# ── Code Quality ──────────────────────────────────────────────────────────────

## swag: Generate Swagger documentation from annotations
swag:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/api/main.go -o ./docs --parseDependency --parseInternal
	@echo "Swagger docs generated in ./docs"

## lint: Run golangci-lint static analysis
lint:
	@echo "Running golangci-lint..."
	golangci-lint run ./...

## tidy: Tidy and verify go modules
tidy:
	@echo "Tidying modules..."
	go mod tidy
	go mod verify

## vet: Run go vet
vet:
	go vet ./...

# ── Docker ────────────────────────────────────────────────────────────────────

## docker-up: Start all Docker Compose services (PostgreSQL)
docker-up:
	@echo "Starting Docker services..."
	docker compose -f $(DOCKER_COMP) up -d
	@echo "Services started. PostgreSQL available on port 5432."

## docker-down: Stop all Docker Compose services
docker-down:
	@echo "Stopping Docker services..."
	docker compose -f $(DOCKER_COMP) down

## docker-logs: Tail Docker Compose service logs
docker-logs:
	docker compose -f $(DOCKER_COMP) logs -f

## docker-build: Build the application Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):latest .

# ── Utilities ─────────────────────────────────────────────────────────────────

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf ./tmp ./coverage.out ./coverage.html
	@echo "Clean complete."

## install-tools: Install required development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/air-verse/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed."

## help: Show this help message
help:
	@echo ""
	@echo "DSMES Backend — Available Commands:"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  make /'
	@echo ""
