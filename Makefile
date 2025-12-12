.PHONY: help test test-coverage lint run build clean migrate docker-up docker-down

# Default target
.DEFAULT_GOAL := help

# Variables
APP_NAME := thums-up-backend
GO_FILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")
MAIN_PATH := ./main.go

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test: ## Run all tests
	@echo "Running tests..."
	go test -v -race ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	go test -v -short ./...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	go test -v -run Integration ./...

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run --timeout=5m

fmt: ## Format code
	@echo "Formatting code..."
	gofmt -s -w $(GO_FILES)
	goimports -w $(GO_FILES)

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

build: ## Build the application
	@echo "Building..."
	go build -o bin/$(APP_NAME) $(MAIN_PATH)

run: ## Run the application
	@echo "Running application..."
	go run $(MAIN_PATH) server

run-subscriber: ## Run the pub/sub subscriber
	@echo "Running subscriber..."
	go run ./cmd/subscriber/main.go

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

migrate: ## Run database migrations
	@echo "Running migrations..."
	go run $(MAIN_PATH) migrate

docker-up: ## Start docker services
	@echo "Starting docker services..."
	docker-compose up -d

docker-down: ## Stop docker services
	@echo "Stopping docker services..."
	docker-compose down

docker-logs: ## Show docker logs
	docker-compose logs -f

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

security: ## Run security checks
	@echo "Running security checks..."
	gosec ./...

benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

ci: lint test ## Run CI pipeline locally

all: clean deps lint test build ## Run all checks and build
