# Cortex Makefile
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

BINARY_NAME := cortex
BUILD_DIR := build
INSTALL_PATH := /usr/local/bin

.PHONY: all build clean install uninstall test release

all: build

# Build for current platform
build:
	@echo "Building $(BINARY_NAME)..."
	@go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/agentflow/
	@echo "Built: ./$(BINARY_NAME)"

# Install to /usr/local/bin
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@sudo cp $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installed: $(INSTALL_PATH)/$(BINARY_NAME)"
	@echo "Run 'cortex --help' to get started"

# Uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Uninstalled $(BINARY_NAME)"

# Clean build artifacts
clean:
	@rm -rf $(BINARY_NAME) $(BUILD_DIR)
	@echo "Cleaned"

# Run tests
test:
	@go test -v ./...

# Build for all platforms
release: clean
	@mkdir -p $(BUILD_DIR)
	@echo "Building for all platforms..."

	@echo "  → darwin/amd64"
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/agentflow/

	@echo "  → darwin/arm64"
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/agentflow/

	@echo "  → linux/amd64"
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/agentflow/

	@echo "  → linux/arm64"
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/agentflow/

	@echo "  → windows/amd64"
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/agentflow/

	@echo "Release binaries built in $(BUILD_DIR)/"
	@ls -la $(BUILD_DIR)/

# Create release archives
dist: release
	@echo "Creating distribution archives..."
	@cd $(BUILD_DIR) && \
		tar -czf $(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64 && \
		tar -czf $(BINARY_NAME)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64 && \
		tar -czf $(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 && \
		tar -czf $(BINARY_NAME)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64 && \
		zip $(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	@echo "Distribution archives created"
	@ls -la $(BUILD_DIR)/*.tar.gz $(BUILD_DIR)/*.zip

# Development helpers
dev: build
	@./$(BINARY_NAME) validate

run: build
	@./$(BINARY_NAME) run

.DEFAULT_GOAL := build
