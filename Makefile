# Application name
APP_NAME := yt-dlp-utils

# Output directory
BIN_DIR := ./bin

# CMD dir 
#CMD_DIR := ./cmd/$(APP_NAME)

# Go settings
GO := go
GOFLAGS :=

# Default target
.PHONY: all
all: build

# Build the application
.PHONY: build
build:
	mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(BIN_DIR)/$(APP_NAME) .

# Run the application
.PHONY: run
run:
	$(GO) run $(CMD_DIR)

# Run tests
.PHONY: test
test:
	$(GO) test ./...

# Format source code
.PHONY: fmt
fmt:
	$(GO) fmt ./...

# Vet source code
.PHONY: vet
vet:
	$(GO) vet ./...

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf $(BIN_DIR)

# Build for Linux AMD64
.PHONY: build-linux
build-linux:
	mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build -o $(BIN_DIR)/$(APP_NAME)-linux-amd64 .

# Build for Windows AMD64
.PHONY: build-windows
build-windows:
	mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build -o $(BIN_DIR)/$(APP_NAME)-windows-amd64.exe .

# Build for macOS ARM64 (Apple Silicon)
.PHONY: build-darwin
build-darwin:
	mkdir -p $(BIN_DIR)
	GOOS=darwin GOARCH=arm64 $(GO) build -o $(BIN_DIR)/$(APP_NAME)-darwin-arm64 .