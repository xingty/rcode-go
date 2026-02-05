# Makefile for cross-platform compilation

# Platforms and architectures
PLATFORMS := windows linux darwin
ARCHS := amd64 386 arm64

# Output directory
DIST_DIR := dist

# Go build command
GOBUILD := go build

GOCACHE ?= $(CURDIR)/.gocache

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
VERSION_FLAGS := -ldflags "-X main.version=$(VERSION)"

EXE_SUFFIX = $(if $(filter windows,$(PLATFORM)),.exe,)
OUT_DIR = $(DIST_DIR)/$(PLATFORM)-$(ARCH)
BIN_DIR = $(OUT_DIR)/bin
GCODE_ALIASES := gcursor gwindsurf gzed gtrae

# Default target
.PHONY: all
all: build

# Build for all platforms and architectures
.PHONY: build
build:
	@set -e; for platform in $(PLATFORMS); do \
		if [ "$$platform" = "darwin" ]; then \
			archs="arm64 amd64"; \
		else \
			archs="$(ARCHS)"; \
		fi; \
		for arch in $$archs; do \
			echo "Building for $$platform/$$arch..."; \
			$(MAKE) build-one PLATFORM=$$platform ARCH=$$arch; \
		done; \
	done

# Build for a specific platform and architecture
.PHONY: build-one
build-one:
	@if [ -z "$(PLATFORM)" ] || [ -z "$(ARCH)" ]; then \
		echo "Error: PLATFORM and ARCH are required. Example: make build-one PLATFORM=linux ARCH=amd64"; \
		exit 2; \
	fi
	@if ! echo " $(PLATFORMS) " | grep -q " $(PLATFORM) "; then \
		echo "Error: unsupported PLATFORM '$(PLATFORM)'. Supported: $(PLATFORMS)"; \
		exit 2; \
	fi
	@if ! echo " $(ARCHS) " | grep -q " $(ARCH) "; then \
		echo "Error: unsupported ARCH '$(ARCH)'. Supported: $(ARCHS)"; \
		exit 2; \
	fi
	@mkdir -p $(BIN_DIR)
	@mkdir -p $(GOCACHE)

	@# Build gssh
	GOCACHE=$(GOCACHE) GOOS=$(PLATFORM) GOARCH=$(ARCH) $(GOBUILD) $(VERSION_FLAGS) -o $(BIN_DIR)/gssh$(EXE_SUFFIX) ./cmd/gssh
	
	@# Build gssh-ipc
	GOCACHE=$(GOCACHE) GOOS=$(PLATFORM) GOARCH=$(ARCH) $(GOBUILD) $(VERSION_FLAGS) -o $(BIN_DIR)/gssh-ipc$(EXE_SUFFIX) ./cmd/ipc
	
	@# Build gcode
	GOCACHE=$(GOCACHE) GOOS=$(PLATFORM) GOARCH=$(ARCH) $(GOBUILD) $(VERSION_FLAGS) -o $(BIN_DIR)/gcode$(EXE_SUFFIX) ./cmd/gcode
	
	@# Create gcode multicall aliases (argv0)
	@if [ "$(PLATFORM)" = "windows" ]; then \
		cp cmd/gcode/bat/ssh-wrapper.bat $(BIN_DIR)/ 2>/dev/null || true; \
		for name in $(GCODE_ALIASES); do \
			printf '@echo off\r\n\"%%~dp0gcode.exe\" --ide %s %%*\r\n' "$${name#g}" > $(BIN_DIR)/$$name.cmd; \
		done; \
	else \
		cp cmd/gcode/sh/ssh-wrapper $(BIN_DIR)/ 2>/dev/null || true; \
		chmod +x $(BIN_DIR)/ssh-wrapper 2>/dev/null || true; \
		for name in $(GCODE_ALIASES); do \
			ln -sf gcode $(BIN_DIR)/$$name; \
		done; \
	fi

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf -- $(DIST_DIR)

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
	@echo "Supported platforms: $(PLATFORMS)"
	@echo "Supported architectures: $(ARCHS)"
	
