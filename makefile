# Makefile for cross-platform compilation

DIST_DIR := dist
GOCACHE ?= $(CURDIR)/.gocache
export GOCACHE

# Optional override; when empty, tools/build will auto-detect.
VERSION ?=

# Default target
.PHONY: all
all: build

.PHONY: build
build:
	go run ./tools/build -dist=$(DIST_DIR) -version=$(VERSION) -all

.PHONY: build-one
build-one:
	go run ./tools/build -dist=$(DIST_DIR) -version=$(VERSION) -platform=$(PLATFORM) -arch=$(ARCH)

# Clean build artifacts
.PHONY: clean
clean:
	go run ./tools/build -dist=$(DIST_DIR) -clean

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all (default) - Build for all platforms and architectures"
	@echo "  build         - Same as 'all'"
	@echo "  build-one     - Build for a specific platform and architecture"
	@echo "                  Usage: make build-one PLATFORM=linux ARCH=amd64"
	@echo "  clean         - Remove all build artifacts"
	@echo "  help          - Show this help message"
	@echo ""
	@echo "Supported platforms: windows linux darwin"
	@echo "Supported architectures: amd64 386 arm64"
