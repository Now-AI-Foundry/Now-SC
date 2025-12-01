.PHONY: build build-all clean install test

# Binary name
BINARY_NAME=now-sc
VERSION?=1.0.0

# Build directory
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Main package path
MAIN_PATH=./cmd/now-sc

# Build the binary for current platform
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) -v $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for all platforms
build-all: clean
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)

	# Linux AMD64
	@echo "Building for Linux AMD64..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 -v $(MAIN_PATH)

	# Linux ARM64
	@echo "Building for Linux ARM64..."
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 -v $(MAIN_PATH)

	# macOS AMD64
	@echo "Building for macOS AMD64..."
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 -v $(MAIN_PATH)

	# macOS ARM64 (Apple Silicon)
	@echo "Building for macOS ARM64..."
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 -v $(MAIN_PATH)

	# Windows AMD64
	@echo "Building for Windows AMD64..."
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe -v $(MAIN_PATH)

	@echo "Build complete for all platforms!"
	@ls -lh $(BUILD_DIR)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete!"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "Dependencies installed!"

# Install binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)
	@echo "Installation complete!"

# Run tests
test:
	@echo "Running tests..."
	@$(GOTEST) -v ./...

# Run the application
run:
	@$(GOCMD) run $(MAIN_PATH)

# Display help
help:
	@echo "Available targets:"
	@echo "  build       - Build binary for current platform"
	@echo "  build-all   - Build binaries for all platforms"
	@echo "  clean       - Remove build artifacts"
	@echo "  deps        - Install dependencies"
	@echo "  install     - Install binary to GOPATH/bin"
	@echo "  test        - Run tests"
	@echo "  run         - Run the application"
	@echo "  help        - Display this help message"
