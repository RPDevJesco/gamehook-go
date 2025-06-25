# GameHook Go Makefile

.PHONY: help build run test clean deps dev release

# Default target
help:
	@echo "GameHook Go Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build     - Build the gamehook binary"
	@echo "  run       - Build and run with default settings"
	@echo "  dev       - Run in development mode with auto-reload"
	@echo "  test      - Run all tests"
	@echo "  deps      - Download dependencies"
	@echo "  clean     - Clean build artifacts"
	@echo "  release   - Build release binaries for multiple platforms"
	@echo "  docker    - Build Docker image"
	@echo ""
	@echo "RetroArch setup:"
	@echo "  make retroarch-setup - Show RetroArch configuration instructions"

# Variables
BINARY_NAME=gamehook
MAIN_PATH=./cmd/gamehook
BUILD_DIR=./build
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

# Build the main binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

# Build and run with default settings
run: build
	@echo "Starting GameHook..."
	@echo "RetroArch should be running with network commands enabled on port 55355"
	@echo "Server will be available at http://localhost:8080"
	@echo ""
	$(BUILD_DIR)/$(BINARY_NAME)

# Development mode with auto-reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	@echo "Starting development mode with auto-reload..."
	@if command -v air >/dev/null 2>&1; then \
		air -c .air.toml; \
	else \
		echo "Air not found. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Falling back to regular run..."; \
		make run; \
	fi

# Run all tests
test:
	@echo "Running tests..."
	go test -v ./...

# Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	go clean

# Build release binaries for multiple platforms
release: clean
	@echo "Building release binaries..."
	@mkdir -p $(BUILD_DIR)/release

	# Linux
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)

	# Windows
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

	# macOS
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

	@echo "Release binaries built in $(BUILD_DIR)/release/"

# Docker build
docker:
	@echo "Building Docker image..."
	docker build -t gamehook:latest .
	@echo "Run with: docker run -p 8080:8080 gamehook:latest"

# Show RetroArch setup instructions
retroarch-setup:
	@echo "RetroArch Setup Instructions:"
	@echo "1. Open RetroArch"
	@echo "2. Go to Settings → Network"
	@echo "3. Enable 'Network Commands': ON"
	@echo "4. Set 'Network Command Port': 55355"
	@echo "5. Load a game (NES games work with the example mapper)"
	@echo "6. Run: make run"
	@echo ""
	@echo "Test connection:"
	@echo "  curl -X POST http://localhost:8080/api/mappers/super_mario_bros/load"

# Create example directories and files
setup-examples:
	@echo "Setting up example directories..."
	@mkdir -p mappers uis/mario-overlay uis/simple-stats
	@echo "Example directories created"
	@echo "Add your .cue files to mappers/"
	@echo "Add your UI folders to uis/"

# Validate CUE files
validate-mappers:
	@echo "Validating CUE mapper files..."
	@for file in mappers/*.cue; do \
		if [ -f "$$file" ]; then \
			echo "Validating $$file..."; \
			cue vet "$$file" || exit 1; \
		fi \
	done
	@echo "All mappers validated successfully"

# Development helpers
install-tools:
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install cuelang.org/go/cmd/cue@latest
	@echo "Development tools installed"

# Quick test with sample data (no RetroArch required)
test-run:
	@echo "Starting test run with sample data..."
	$(BUILD_DIR)/$(BINARY_NAME) --test-mode

# Show current configuration
info:
	@echo "GameHook Go Information:"
	@echo "  Version: $(VERSION)"
	@echo "  Go version: $(shell go version)"
	@echo "  Build directory: $(BUILD_DIR)"
	@echo "  Binary name: $(BINARY_NAME)"
	@echo ""
	@echo "Project structure:"
	@find . -type f -name "*.go" | head -10
	@echo "  (showing first 10 Go files)"

# Check for common issues
doctor:
	@echo "GameHook Doctor - Checking for common issues..."
	@echo ""

	# Check Go version
	@echo "✓ Checking Go installation..."
	@go version || (echo "✗ Go not found. Please install Go 1.21+"; exit 1)

	# Check CUE installation
	@echo "✓ Checking CUE installation..."
	@cue version >/dev/null 2>&1 || echo "⚠ CUE not found. Install with: go install cuelang.org/go/cmd/cue@latest"

	# Check directories
	@echo "✓ Checking directories..."
	@[ -d "mappers" ] || echo "⚠ mappers/ directory not found. Run: mkdir mappers"
	@[ -d "uis" ] || echo "⚠ uis/ directory not found. Run: mkdir uis"

	# Check for RetroArch (if running)
	@echo "✓ Checking RetroArch connection..."
	@nc -z -w1 127.0.0.1 55355 >/dev/null 2>&1 && echo "✓ RetroArch appears to be running" || echo "⚠ RetroArch not detected on port 55355"

	@echo ""
	@echo "Doctor check complete!"