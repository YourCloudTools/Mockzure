# Mockzure Makefile
# Build and test automation

.PHONY: help build test test-unit test-integration test-compatibility test-all coverage clean install deps lint format

# Default target
help:
	@echo "Mockzure Build and Test Automation"
	@echo "=================================="
	@echo ""
	@echo "Available targets:"
	@echo "  build           - Build the Mockzure binary"
	@echo "  test            - Run all tests"
	@echo "  test-unit       - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-compatibility - Run compatibility tests only"
	@echo "  test-all        - Run all tests with coverage"
	@echo "  coverage        - Generate comprehensive coverage report"
	@echo "  clean           - Clean build artifacts"
	@echo "  install         - Install dependencies"
	@echo "  deps            - Download and verify dependencies"
	@echo "  lint            - Run linter"
	@echo "  format          - Format code"
	@echo "  docs            - Generate documentation"
	@echo "  docker          - Build Docker image"
	@echo "  rpm             - Build RPM package"
	@echo ""
	@echo "Examples:"
	@echo "  make test-unit  # Run only unit tests"
	@echo "  make coverage   # Generate coverage report"
	@echo "  make build      # Build the binary"

# Build the Mockzure binary
build:
	@echo "🔨 Building Mockzure..."
	go build -v -o mockzure main.go
	@echo "✅ Build complete: mockzure"

# Run all tests
test:
	@echo "🧪 Running all tests..."
	./test.sh all

# Run unit tests only
test-unit:
	@echo "🧪 Running unit tests..."
	./test.sh unit

# Run integration tests only
test-integration:
	@echo "🧪 Running integration tests..."
	./test.sh integration

# Run compatibility tests only
test-compatibility:
	@echo "🧪 Running compatibility tests..."
	./test.sh compatibility

# Run all tests with coverage
test-all:
	@echo "🧪 Running all tests with coverage..."
	./test.sh all --coverage

# Generate comprehensive coverage report
coverage:
	@echo "📊 Generating coverage report..."
	./test.sh all --coverage
	@echo "📈 Coverage report generated: coverage.html"
	@echo "📈 Coverage summary:"
	@go tool cover -func=coverage.out | tail -1

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f mockzure
	rm -f coverage.out coverage.html coverage.json
	rm -f generate-report
	rm -rf dist/
	@echo "✅ Clean complete"

# Install dependencies
install: deps
	@echo "📦 Dependencies installed"

# Download and verify dependencies
deps:
	@echo "📥 Downloading dependencies..."
	go mod download
	go mod verify
	@echo "✅ Dependencies verified"

# Run linter
lint:
	@echo "🔍 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed, skipping linting"; \
		echo "   Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Format code
format:
	@echo "🎨 Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted"

# Generate documentation
docs:
	@echo "📚 Generating documentation..."
	@if [ -f "scripts/prepare-docs.sh" ]; then \
		chmod +x scripts/prepare-docs.sh; \
		./scripts/prepare-docs.sh; \
	else \
		echo "⚠️  Documentation preparation script not found"; \
	fi

# Build Docker image
docker:
	@echo "🐳 Building Docker image..."
	docker build -t mockzure:latest .
	@echo "✅ Docker image built: mockzure:latest"

# Build RPM package
rpm:
	@echo "📦 Building RPM package..."
	@if [ -f "deploy/rpm/build-rpm.sh" ]; then \
		chmod +x deploy/rpm/build-rpm.sh; \
		./deploy/rpm/build-rpm.sh; \
	else \
		echo "⚠️  RPM build script not found"; \
	fi

# Development setup
dev-setup: deps
	@echo "🛠️  Setting up development environment..."
	@if [ ! -f "config.json" ]; then \
		cp config.json.example config.json; \
		echo "📝 Created config.json from example"; \
	fi
	@echo "✅ Development environment ready"

# Run Mockzure locally
run: build
	@echo "🚀 Starting Mockzure..."
	./mockzure

# Run with Docker Compose
docker-run:
	@echo "🐳 Starting Mockzure with Docker Compose..."
	docker compose up -d
	@echo "✅ Mockzure started at http://localhost:8090"

# Stop Docker Compose
docker-stop:
	@echo "🛑 Stopping Mockzure..."
	docker compose down
	@echo "✅ Mockzure stopped"

# Update coverage badge
update-badge:
	@echo "🏷️  Updating coverage badge..."
	@if [ -f "scripts/update-coverage-badge.sh" ]; then \
		chmod +x scripts/update-coverage-badge.sh; \
		./scripts/update-coverage-badge.sh; \
	else \
		echo "⚠️  Coverage badge update script not found"; \
	fi

# Generate compatibility report
compatibility-report:
	@echo "📋 Generating compatibility report..."
	@if [ -f "cmd/generate_compatibility_report/main.go" ]; then \
		go run cmd/generate_compatibility_report/main.go; \
	else \
		echo "⚠️  Compatibility report generator not found"; \
	fi

# Full CI pipeline simulation
ci: deps lint test-all docs
	@echo "✅ CI pipeline completed successfully"

# Quick development cycle
dev: format test-unit build
	@echo "✅ Development cycle completed"

# Production build
prod: clean deps lint test-all build
	@echo "✅ Production build completed"
