VERSION := v0.1.0
BINARY_NAME := engine
BUILD_DIR := bin
LD_FLAGS := -s -w

# Force PowerShell as the shell
SHELL := powershell.exe
.SHELLFLAGS := -Command

# Default Target
all: build-all

# Local Build
build:
	@$$env:GOOS=""; $$env:GOARCH=""; go build -ldflags="$(LD_FLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME).exe ./cmd/worker

# Build for all platforms
build-all: build-windows-amd64 build-windows-arm64 build-linux-amd64 build-linux-arm64

build-windows-amd64:
	@$$env:GOOS="windows"; $$env:GOARCH="amd64"; go build -ldflags="$(LD_FLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/worker

build-windows-arm64:
	@$$env:GOOS="windows"; $$env:GOARCH="arm64"; go build -ldflags="$(LD_FLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe ./cmd/worker

build-linux-amd64:
	@$$env:GOOS="linux"; $$env:GOARCH="amd64"; go build -ldflags="$(LD_FLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/worker

build-linux-arm64:
	@$$env:GOOS="linux"; $$env:GOARCH="arm64"; go build -ldflags="$(LD_FLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/worker

# Run Go tests
test:
	@go test -v ./...

# Clean build artifacts
clean:
	@if (Test-Path $(BUILD_DIR)) { Remove-Item -Recurse -Force $(BUILD_DIR) }
	@echo "[*] Cleaned $(BUILD_DIR)"

# Help Target
help:
	@echo "Rundown-Workers Build System (v0.1.0)"
	@echo ""
	@echo "Usage:"
	@echo "  make              Build for all platforms"
	@echo "  make build        Build for current platform"
	@echo "  make test         Run Go tests"
	@echo "  make clean        Remove build artifacts"