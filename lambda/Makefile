# Makefile for location Lambda function

.PHONY: help build test lint clean deps tidy vet fmt check-fmt

# Default target
help:
	@echo "Available commands:"
	@echo "  build     - Build the Lambda function binary"
	@echo "  test      - Run all tests"
	@echo "  lint      - Run linting checks"
	@echo "  vet       - Run go vet"
	@echo "  fmt       - Format code with gofmt"
	@echo "  check-fmt - Check if code is properly formatted"
	@echo "  deps      - Download dependencies"
	@echo "  tidy      - Clean up go.mod and go.sum"
	@echo "  clean     - Clean build artifacts"
	@echo "  all       - Run tests, lint, and build"

# Variables
BINARY_NAME=location-handler
BUILD_DIR=build
LAMBDA_ZIP=$(BUILD_DIR)/$(BINARY_NAME).zip

# Go build settings
GOOS=linux
GOARCH=amd64
CGO_ENABLED=0

# Build the Lambda function
build: clean
	@echo "Building Lambda function..."
	@mkdir -p $(BUILD_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) go build \
		-ldflags="-s -w" \
		-o $(BUILD_DIR)/$(BINARY_NAME) \
		./cmd/handler

# Create deployment zip
zip: build
	@echo "Creating deployment zip..."
	@cd $(BUILD_DIR) && zip $(BINARY_NAME).zip $(BINARY_NAME)

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download

# Clean up go.mod and go.sum
tidy:
	@echo "Tidying go modules..."
	go mod tidy

# Run all tests
test:
	@echo "Running tests..."
	go test -v -race -cover ./...

# Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Check if code is properly formatted
check-fmt:
	@echo "Checking code formatting..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "The following files need formatting:"; \
		gofmt -l .; \
		exit 1; \
	fi

# Install golangci-lint if not present
install-lint:
	@which golangci-lint > /dev/null || \
		(echo "Installing golangci-lint..." && \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2)

# Run linting
lint: install-lint
	@echo "Running linting..."
	golangci-lint run ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Run all checks and build
all: check-fmt vet lint test build

# Development workflow
dev: fmt vet test

# Continuous integration workflow
ci: deps tidy check-fmt vet lint test build

# Install development dependencies
dev-deps:
	@echo "Installing development dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest