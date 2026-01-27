# Makefile for DevUp CLI

.PHONY: all build install clean test lint help bump-patch bump-minor bump-major version

# ==============================================================================
# Variables
# ==============================================================================

# BINARY_NAME: The name of the binary to build.
BINARY_NAME=devup

# BUILD_DIR: The directory where the built binary will be placed.
BUILD_DIR=build

# INSTALL_PATH: The directory where the binary will be installed.
INSTALL_PATH=/usr/local/bin

# VERSION: The version of the application. It's read from the VERSION file.
VERSION=$(shell cat VERSION 2>/dev/null || echo "1.0.0")

# ==============================================================================
# Go Parameters
# ==============================================================================

# GOCMD: The Go command.
GOCMD=go

# GOBUILD: The Go build command.
GOBUILD=$(GOCMD) build

# GOCLEAN: The Go clean command.
GOCLEAN=$(GOCMD) clean

# GOTEST: The Go test command.
GOTEST=$(GOCMD) test

# GOGET: The Go get command.
GOGET=$(GOCMD) get

# GOMOD: The Go mod command.
GOMOD=$(GOCMD) mod

# ==============================================================================
# Build Flags
# ==============================================================================

# LDFLAGS: Linker flags.
# -s: Omit the symbol table and debug information.
# -w: Omit the DWARF symbol table.
# -X: Set the value of a string variable in the application.
LDFLAGS=-ldflags "-s -w -X 'github.com/evertonmj/devup_v2/cmd.Version=$(VERSION)'"

all: test build

## version: Show current version
version:
	@echo "Current version: $(VERSION)"

## bump-patch: Bump patch version (1.0.0 -> 1.0.1)
bump-patch:
	@./scripts/bump-version.sh patch

## bump-minor: Bump minor version (1.0.0 -> 1.1.0)
bump-minor:
	@./scripts/bump-version.sh minor

## bump-major: Bump major version (1.0.0 -> 2.0.0)
bump-major:
	@./scripts/bump-version.sh major

## build: Build the binary without version bump
build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -ldflags "-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME) v$(VERSION)"

## build-release: Build and bump patch version (use for releases)
build-release:
	@echo "Bumping version..."
	@./scripts/bump-version.sh patch
	@echo ""
	@NEW_VERSION=$$(cat VERSION); \
	echo "Building $(BINARY_NAME) v$$NEW_VERSION..."; \
	mkdir -p $(BUILD_DIR); \
	$(GOBUILD) -ldflags "-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) .; \
	echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME) v$$NEW_VERSION"

## install: Build and install to system
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "✅ Installed successfully!"
	@echo "Run: $(BINARY_NAME) --version"

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -cover ./cmd/... ./internal/...

## test-coverage: Run tests and generate coverage report (cmd + internal)
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out ./cmd/... ./internal/...
	@echo "✅ Coverage report generated: coverage.out"
	@go tool cover -func=coverage.out | grep '^total:'
	@echo "To view the report, run: go tool cover -html=coverage.out"

## test-coverage-internal: Run tests for internal packages only; fail if coverage < 90%%
test-coverage-internal:
	@echo "Running internal package tests with coverage (target ≥90%)..."
	$(GOTEST) -coverprofile=coverage_internal.out ./internal/...
	@go tool cover -func=coverage_internal.out | grep '^total:'
	@go tool cover -func=coverage_internal.out | awk '/^total:/ { gsub(/%/,""); p=$$NF+0; if (p < 90) { printf "❌ Internal coverage %.1f%% is below 90%%\n", p; exit 1 }; printf "✅ Internal coverage %.1f%%\n", p }'

## lint: Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "⚠️  golangci-lint not installed. Install with:"; \
		echo "    brew install golangci-lint"; \
	fi

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@echo "✅ Clean complete"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "✅ Dependencies updated"

## run: Build and run with example config
run: build
	@echo "Running devup with example config..."
	./$(BUILD_DIR)/$(BINARY_NAME) list

## help: Show this help message
help:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "DevUp CLI - Build Commands"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "Current Version: $(VERSION)"
	@echo ""
	@echo "Usage: make <command>"
	@echo ""
	@grep -E '^## ' Makefile | sed 's/## /  /'
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "Note: 'make build' builds without version change"
	@echo "      Use 'make build-release' to build and bump version (1.0.0 → 1.0.1)"
	@echo ""
