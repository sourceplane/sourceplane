.PHONY: build build-sp build-thinci install install-all clean test fmt vet help release

# Binary names
SP_BINARY=sp
THINCI_BINARY=thinci
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Build all binaries
build: build-sp build-thinci

# Build the Sourceplane CLI
build-sp:
	@echo "Building $(SP_BINARY)..."
	@go build $(LDFLAGS) -o $(SP_BINARY) ./cmd/sp
	@echo "✅ Build complete: ./$(SP_BINARY)"

# Build the Thin-CI standalone binary
build-thinci:
	@echo "Building $(THINCI_BINARY)..."
	@go build $(LDFLAGS) -o $(THINCI_BINARY) ./cmd/thinci
	@echo "✅ Build complete: ./$(THINCI_BINARY)"

# Install Sourceplane to /usr/local/bin
install: build-sp
	@echo "Installing $(SP_BINARY) to /usr/local/bin..."
	@sudo cp $(SP_BINARY) /usr/local/bin/
	@echo "✅ Installation complete"

# Install both binaries to /usr/local/bin
install-all: build
	@echo "Installing $(SP_BINARY) and $(THINCI_BINARY) to /usr/local/bin..."
	@sudo cp $(SP_BINARY) /usr/local/bin/
	@sudo cp $(THINCI_BINARY) /usr/local/bin/
	@echo "✅ Installation complete"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(SP_BINARY) $(THINCI_BINARY)
	@rm -f blueprint.yaml
	@rm -rf example-api/
	@rm -rf dist/
	@echo "✅ Clean complete"

# Run tests
test:
	@echo "Running tests..."
	@go test ./... -v

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✅ Format complete"

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "✅ Vet complete"

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy
	@echo "✅ Tidy complete"

# Run all checks
check: fmt vet
	@echo "✅ All checks passed"

# Build release binaries for multiple platforms
release:
	@echo "Building release binaries..."
	@mkdir -p dist
	@echo "Building for Linux AMD64..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(SP_BINARY)-linux-amd64 ./cmd/sp
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(THINCI_BINARY)-linux-amd64 ./cmd/thinci
	@echo "Building for Linux ARM64..."
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(SP_BINARY)-linux-arm64 ./cmd/sp
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(THINCI_BINARY)-linux-arm64 ./cmd/thinci
	@echo "Building for macOS AMD64..."
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(SP_BINARY)-darwin-amd64 ./cmd/sp
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(THINCI_BINARY)-darwin-amd64 ./cmd/thinci
	@echo "Building for macOS ARM64..."
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(SP_BINARY)-darwin-arm64 ./cmd/sp
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(THINCI_BINARY)-darwin-arm64 ./cmd/thinci
	@echo "Building for Windows AMD64..."
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(SP_BINARY)-windows-amd64.exe ./cmd/sp
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(THINCI_BINARY)-windows-amd64.exe ./cmd/thinci
	@echo "✅ Release binaries built in dist/"

# Show help
help:
	@echo "Available targets:"
	@echo "  build        - Build both sp and thinci binaries"
	@echo "  build-sp     - Build only the Sourceplane CLI binary"
	@echo "  build-thinci - Build only the Thin-CI binary"
	@echo "  install      - Install sp to /usr/local/bin (requires sudo)"
	@echo "  install-all  - Install both binaries to /usr/local/bin (requires sudo)"
	@echo "  clean        - Remove build artifacts"
	@echo "  test         - Run tests"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  tidy         - Tidy dependencies"
	@echo "  check        - Run fmt and vet"
	@echo "  release      - Build release binaries for multiple platforms"
	@echo "  help         - Show this help message"
