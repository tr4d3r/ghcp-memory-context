.PHONY: build run test clean install deps lint fmt vet

# Variables
BINARY_NAME=ghcp-memory-server
BUILD_DIR=bin
MAIN_PATH=cmd/server/main.go

# Default target
all: deps fmt vet test build

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build the application
build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Run the application
run:
	go run $(MAIN_PATH)

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-cover:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Install the binary
install: build
	sudo mv $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

# Development server with auto-reload (requires air)
dev:
	air

# Docker build
docker-build:
	docker build -t ghcp-memory-server .

# Docker run
docker-run:
	docker run -p 8080:8080 ghcp-memory-server
