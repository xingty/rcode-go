# Makefile for cross-platform compilation

# Platforms and architectures
PLATFORMS := windows linux darwin
ARCHS := amd64 386 arm64

# Command names
CMDS := gssh gssh-ipc gcode

# Output directory
DIST_DIR := dist

# Go build command
GOBUILD := go build

# Default target
.PHONY: all
all: build

# Build for all platforms and architectures
.PHONY: build
build:
	@for platform in $(PLATFORMS); do \
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
	@mkdir -p $(DIST_DIR)/$(PLATFORM)_$(ARCH)/bin
	
	@# Set executable extension based on platform
	$(eval EXE_SUFFIX := $(if $(filter windows,$(PLATFORM)),.exe,))
	
	@# Build gssh
	GOOS=$(PLATFORM) GOARCH=$(ARCH) $(GOBUILD) -o $(DIST_DIR)/$(PLATFORM)_$(ARCH)/bin/gssh$(EXE_SUFFIX) ./cmd/gssh
	
	@# Build gssh-ipc
	GOOS=$(PLATFORM) GOARCH=$(ARCH) $(GOBUILD) -o $(DIST_DIR)/$(PLATFORM)_$(ARCH)/bin/gssh-ipc$(EXE_SUFFIX) ./cmd/ipc
	
	@# Build gcode
	GOOS=$(PLATFORM) GOARCH=$(ARCH) $(GOBUILD) -o $(DIST_DIR)/$(PLATFORM)_$(ARCH)/gcode$(EXE_SUFFIX) ./cmd/gcode
	
	@# Copy platform-specific scripts
	@if [ "$(PLATFORM)" = "windows" ]; then \
		cp cmd/gcode/bat/*.bat $(DIST_DIR)/$(PLATFORM)_$(ARCH)/bin/ 2>/dev/null || true; \
	else \
		cp cmd/gcode/sh/* $(DIST_DIR)/$(PLATFORM)_$(ARCH)/bin/ ; \
		chmod +x $(DIST_DIR)/$(PLATFORM)_$(ARCH)/bin/* 2>/dev/null || true; \
	fi

# Clean build artifacts
.PHONY: clean
clean:
	rm -r $(DIST_DIR)

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
	