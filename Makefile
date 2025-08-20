.PHONY: build run test clean install deps lint fmt vet vscode-setup test-mcp help

# Variables
BINARY_NAME=ghcp-memory-context
BUILD_DIR=bin
MAIN_PATH=cmd/server/main.go
DATA_DIR=.memory-context
VERSION=1.0.0

# Colors for output
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

# Default target
all: deps fmt vet test build

help: ## Show available targets
	@echo "$(GREEN)GHCP Memory Context Server$(NC)"
	@echo "$(BLUE)Available targets:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Install dependencies
deps: ## Download and tidy dependencies
	go mod download
	go mod tidy

# Build the application
build: ## Build binary for VS Code integration
	@echo "$(GREEN)Building $(BINARY_NAME)...$(NC)"
	go build -ldflags="-s -w" -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)✓ Binary built: $(BINARY_NAME)$(NC)"

# Build for bin directory (legacy)
build-bin: ## Build binary in bin/ directory
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Run the HTTP server
run: ## Run HTTP server
	go run $(MAIN_PATH)

# Run MCP stdio server
run-mcp: build ## Run MCP stdio server for testing
	@echo "$(GREEN)Starting MCP stdio server...$(NC)"
	./$(BINARY_NAME) --mcp-stdio

# Run tests
test: ## Run all tests
	go test -v ./...

# Test MCP protocol functionality
test-mcp: build ## Test MCP protocol functionality
	@echo "$(GREEN)Testing MCP protocol...$(NC)"
	@echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}' | ./$(BINARY_NAME) --mcp-stdio 2>/dev/null | grep -q "ghcp-memory-context" && echo "$(GREEN)✓ MCP initialize working$(NC)" || echo "❌ MCP initialize failed"
	@echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./$(BINARY_NAME) --mcp-stdio 2>/dev/null | grep -q "remember_fact" && echo "$(GREEN)✓ MCP tools working$(NC)" || echo "❌ MCP tools failed"

# Run tests with coverage
test-cover: ## Run tests with coverage report
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Format code
fmt: ## Format Go code
	go fmt ./...

# Vet code
vet: ## Run go vet
	go vet ./...

# Lint code (requires golangci-lint)
lint: ## Run linter
	golangci-lint run

# Clean build artifacts
clean: ## Clean build artifacts and data
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html $(BINARY_NAME)
	rm -rf $(DATA_DIR)
	@echo "$(GREEN)✓ Cleaned$(NC)"

# Install the binary
install: build ## Install binary to /usr/local/bin
	@echo "$(GREEN)Installing $(BINARY_NAME) to /usr/local/bin...$(NC)"
	sudo cp $(BINARY_NAME) /usr/local/bin/
	@echo "$(GREEN)✓ Installed: /usr/local/bin/$(BINARY_NAME)$(NC)"

# VS Code setup
vscode-setup: build ## Setup VS Code integration
	@echo "$(GREEN)Setting up VS Code integration...$(NC)"
	@if [ ! -f .vscode/mcp.json ]; then \
		mkdir -p .vscode; \
		echo '{\n  "mcpServers": {\n    "memory-context": {\n      "command": "./$(BINARY_NAME)",\n      "args": ["--mcp-stdio"],\n      "description": "Persistent memory context for project development"\n    }\n  }\n}' > .vscode/mcp.json; \
		echo "$(GREEN)✓ Created .vscode/mcp.json$(NC)"; \
	else \
		echo "$(YELLOW)✓ .vscode/mcp.json already exists$(NC)"; \
	fi
	@echo "$(BLUE)VS Code setup complete!$(NC)"
	@echo "$(BLUE)1. Start VS Code: code .$(NC)"
	@echo "$(BLUE)2. Open Copilot Chat$(NC)"
	@echo "$(BLUE)3. Test: 'Remember that we use conventional commits'$(NC)"

# Development server with auto-reload (requires air)
dev: ## Run development server with auto-reload
	air

# Docker build
docker-build: ## Build Docker image
	docker build -t ghcp-memory-context .

# Docker run
docker-run: ## Run Docker container
	docker run -d --name ghcp-memory-context -p 8080:8080 -v $(PWD)/$(DATA_DIR):/app/$(DATA_DIR) ghcp-memory-context
