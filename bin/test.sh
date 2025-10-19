#!/bin/bash

# Mockzure Test Script
# Runs all tests including compatibility tests and generates reports

set -e

echo "🧪 Running Mockzure Test Suite"
echo "================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the right directory
if [ ! -f "main.go" ]; then
    print_error "Please run this script from the Mockzure project root directory"
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go to run tests."
    exit 1
fi

print_status "Go version: $(go version)"

# Parse command line arguments
TEST_TYPE="all"
COVERAGE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        unit)
            TEST_TYPE="unit"
            shift
            ;;
        integration)
            TEST_TYPE="integration"
            shift
            ;;
        compatibility)
            TEST_TYPE="compatibility"
            shift
            ;;
        all)
            TEST_TYPE="all"
            shift
            ;;
        --coverage)
            COVERAGE=true
            shift
            ;;
        --help)
            echo "Usage: $0 [TEST_TYPE] [OPTIONS]"
            echo ""
            echo "Test Types:"
            echo "  unit          Run unit tests only"
            echo "  integration   Run integration tests only"
            echo "  compatibility Run compatibility tests only"
            echo "  all           Run all tests (default)"
            echo ""
            echo "Options:"
            echo "  --coverage    Generate coverage report"
            echo "  --help        Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                    # Run all tests"
            echo "  $0 unit               # Run unit tests only"
            echo "  $0 integration --coverage  # Run integration tests with coverage"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

print_status "Test type: $TEST_TYPE"

# Run tests based on type
case $TEST_TYPE in
    unit)
        print_status "Running unit tests..."
        if go test -v -run "TestHelper|TestJWT|TestAuth|TestPermission|TestRender|TestConfig" ./...; then
            print_success "Unit tests passed"
        else
            print_error "Unit tests failed"
            exit 1
        fi
        ;;
    integration)
        print_status "Running integration tests..."
        if go test -v -run "TestServiceAccount|TestAdmin|TestUser|TestMain" ./...; then
            print_success "Integration tests passed"
        else
            print_error "Integration tests failed"
            exit 1
        fi
        ;;
    compatibility)
        print_status "Running Azure API compatibility tests..."
        if go test -v -run "TestMicrosoft|TestAzure|TestRBAC" ./...; then
            print_success "Azure API compatibility tests passed"
        else
            print_warning "Some Azure API compatibility tests failed (this may be expected)"
        fi
        ;;
    all)
        print_status "Running all tests..."
        
        # Run unit tests
        print_status "Running unit tests..."
        if go test -v -run "TestHelper|TestJWT|TestAuth|TestPermission|TestRender|TestConfig" ./...; then
            print_success "Unit tests passed"
        else
            print_error "Unit tests failed"
            exit 1
        fi
        
        # Run integration tests
        print_status "Running integration tests..."
        if go test -v -run "TestServiceAccount|TestAdmin|TestUser|TestMain" ./...; then
            print_success "Integration tests passed"
        else
            print_error "Integration tests failed"
            exit 1
        fi
        
        # Run compatibility tests
        print_status "Running Azure API compatibility tests..."
        if go test -v -run "TestMicrosoft|TestAzure|TestRBAC" ./...; then
            print_success "Azure API compatibility tests passed"
        else
            print_warning "Some Azure API compatibility tests failed (this may be expected)"
        fi
        ;;
esac

# Generate compatibility report
print_status "Generating Azure API compatibility report..."
if go run cmd/generate_compatibility_report/main.go; then
    print_success "Compatibility report generated successfully"
    if [ -f "docs/AZURE_API_COMPATIBILITY.md" ]; then
        print_status "Report location: docs/AZURE_API_COMPATIBILITY.md"
    fi
else
    print_warning "Failed to generate compatibility report (continuing anyway)"
fi

# Run any additional test scripts if they exist
if [ -f "test-integration.sh" ]; then
    print_status "Running integration tests..."
    if bash test-integration.sh; then
        print_success "Integration tests passed"
    else
        print_error "Integration tests failed"
        exit 1
    fi
fi

# Summary
echo ""
echo "================================"
echo "🎉 Test Suite Complete!"
echo "================================"
print_success "All tests completed successfully"

# Show test coverage if requested
if [ "$COVERAGE" = true ]; then
    print_status "Generating test coverage report..."
    # Generate coverage for all packages
    go test -coverprofile=coverage.out -covermode=atomic ./...
    go tool cover -html=coverage.out -o coverage.html
    if [ -f "coverage.html" ]; then
        print_success "Coverage report generated: coverage.html"
        print_status "Coverage summary:"
        go tool cover -func=coverage.out | tail -1
        print_status "Open coverage.html in your browser to view detailed coverage"
        
        # Update coverage badge if script exists
        if [ -f "bin/update-coverage-badge.sh" ]; then
            print_status "Updating coverage badge..."
            chmod +x bin/update-coverage-badge.sh
            ./bin/update-coverage-badge.sh --coverage-file coverage.out --docs-dir docs
        fi
    else
        print_error "Failed to generate coverage report"
    fi
fi

# Show compatibility report summary if it exists
if [ -f "docs/AZURE_API_COMPATIBILITY.md" ]; then
    echo ""
    print_status "Azure API Compatibility Summary:"
    echo "----------------------------------------"
    grep -A 10 "## Compatibility Matrix" docs/AZURE_API_COMPATIBILITY.md | head -15
fi

echo ""
print_success "Test script completed successfully!"
