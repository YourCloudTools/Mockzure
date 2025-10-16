#!/bin/bash

# Prepare Documentation Script
# Prepares documentation for deployment with fresh coverage reports and validation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --docs-dir DIR          Path to docs directory (default: docs)"
    echo "  --coverage-file FILE    Path to coverage.out file (default: coverage.out)"
    echo "  --skip-coverage         Skip coverage report generation"
    echo "  --skip-validation       Skip link validation"
    echo "  --dry-run               Show what would be done without making changes"
    echo "  --help                  Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                      # Prepare docs with all steps"
    echo "  $0 --skip-coverage      # Prepare docs without coverage reports"
    echo "  $0 --dry-run            # Preview what would be done"
}

# Default values
DOCS_DIR="docs"
COVERAGE_FILE="coverage.out"
SKIP_COVERAGE=false
SKIP_VALIDATION=false
DRY_RUN=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --docs-dir)
            DOCS_DIR="$2"
            shift 2
            ;;
        --coverage-file)
            COVERAGE_FILE="$2"
            shift 2
            ;;
        --skip-coverage)
            SKIP_COVERAGE=true
            shift
            ;;
        --skip-validation)
            SKIP_VALIDATION=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

print_status "Preparing documentation for deployment..."

# Check if we're in the right directory
if [ ! -f "main.go" ]; then
    print_error "Please run this script from the Mockzure project root directory"
    exit 1
fi

# Check if docs directory exists
if [ ! -d "$DOCS_DIR" ]; then
    print_error "Docs directory not found: $DOCS_DIR"
    exit 1
fi

# Function to run command with dry-run support
run_command() {
    local cmd="$1"
    local description="$2"
    
    if [ "$DRY_RUN" = true ]; then
        print_status "DRY RUN: $description"
        print_status "  Command: $cmd"
    else
        print_status "$description"
        # Execute command directly instead of using eval for security
        case "$cmd" in
            "go test"*)
                go test -coverprofile="$COVERAGE_FILE" -covermode=atomic ./...
                ;;
            "go tool cover"*)
                go tool cover -html="$COVERAGE_FILE" -o "${DOCS_DIR}/coverage.html"
                ;;
            "scripts/update-coverage-badge.sh"*)
                ./scripts/update-coverage-badge.sh --coverage-file "$COVERAGE_FILE" --docs-dir "$DOCS_DIR"
                ;;
            "go run"*)
                go run cmd/generate_compatibility_report/main.go
                ;;
            *)
                print_error "Unknown command: $cmd"
                return 1
                ;;
        esac
    fi
}

# Step 1: Generate fresh coverage report
if [ "$SKIP_COVERAGE" = false ]; then
    if [ -f "$COVERAGE_FILE" ]; then
        print_status "Using existing coverage file: $COVERAGE_FILE"
    else
        print_warning "Coverage file not found: $COVERAGE_FILE"
        print_status "Generating coverage report..."
        run_command "go test -coverprofile=$COVERAGE_FILE -covermode=atomic ./..." "Running tests with coverage"
    fi
    
    # Generate HTML coverage report
    run_command "go tool cover -html=$COVERAGE_FILE -o ${DOCS_DIR}/coverage.html" "Generating HTML coverage report"
    
    # Update coverage badge
    if [ -f "scripts/update-coverage-badge.sh" ]; then
        run_command "scripts/update-coverage-badge.sh --coverage-file $COVERAGE_FILE --docs-dir $DOCS_DIR" "Updating coverage badges"
    else
        print_warning "Coverage badge update script not found"
    fi
else
    print_status "Skipping coverage report generation"
fi

# Step 2: Generate compatibility report
if [ -f "cmd/generate_compatibility_report/main.go" ]; then
    run_command "go run cmd/generate_compatibility_report/main.go" "Generating Azure API compatibility report"
    
    # Copy compatibility reports to docs
    if [ -f "docs/AZURE_API_COMPATIBILITY.md" ]; then
        print_success "Compatibility report generated: docs/AZURE_API_COMPATIBILITY.md"
    fi
    if [ -f "docs/AZURE_API_COMPATIBILITY.json" ]; then
        print_success "Compatibility report generated: docs/AZURE_API_COMPATIBILITY.json"
    fi
else
    print_warning "Compatibility report generator not found"
fi

# Step 3: Copy additional documentation assets
print_status "Copying documentation assets..."

# Ensure .nojekyll file exists for GitHub Pages
if [ "$DRY_RUN" = false ]; then
    touch "${DOCS_DIR}/.nojekyll"
    print_success "Created .nojekyll file for GitHub Pages"
else
    print_status "DRY RUN: Would create .nojekyll file"
fi

# Copy any additional assets if they exist
if [ -d "assets" ] && [ ! -d "${DOCS_DIR}/assets" ]; then
    run_command "cp -r assets ${DOCS_DIR}/" "Copying assets directory"
fi

# Step 4: Validate documentation files
if [ "$SKIP_VALIDATION" = false ]; then
    print_status "Validating documentation files..."
    
    # Check for required files
    REQUIRED_FILES=(
        "${DOCS_DIR}/index.html"
        "${DOCS_DIR}/getting-started.html"
        "${DOCS_DIR}/api-reference.html"
        "${DOCS_DIR}/architecture.html"
    )
    
    for file in "${REQUIRED_FILES[@]}"; do
        if [ -f "$file" ]; then
            print_success "Found: $file"
        else
            print_warning "Missing: $file"
        fi
    done
    
    # Validate HTML files
    HTML_FILES=$(find "$DOCS_DIR" -name "*.html" -type f)
    if [ -n "$HTML_FILES" ]; then
        print_status "Validating HTML files..."
        for file in $HTML_FILES; do
            if [ "$DRY_RUN" = false ]; then
                # Basic HTML validation (check for basic structure)
                if grep -q "<!DOCTYPE html>" "$file" && grep -q "</html>" "$file"; then
                    print_success "Valid HTML: $file"
                else
                    print_warning "Invalid HTML structure: $file"
                fi
            else
                print_status "DRY RUN: Would validate HTML structure in $file"
            fi
        done
    fi
    
    # Check for broken internal links (basic check)
    print_status "Checking for broken internal links..."
    for file in $HTML_FILES; do
        if [ "$DRY_RUN" = false ]; then
            # Extract href attributes and check if files exist
            hrefs=$(grep -o 'href="[^"]*"' "$file" | sed 's/href="//g' | sed 's/"//g' | grep -v '^http' | grep -v '^#' | grep -v '^mailto:')
            for href in $hrefs; do
                if [ ! -f "${DOCS_DIR}/${href}" ] && [ ! -f "${DOCS_DIR}/${href}.html" ]; then
                    print_warning "Broken link in $file: $href"
                fi
            done
        else
            print_status "DRY RUN: Would check internal links in $file"
        fi
    done
else
    print_status "Skipping validation"
fi

# Step 5: Generate documentation index
if [ "$DRY_RUN" = false ]; then
    print_status "Generating documentation index..."
    
    # Create a simple index of all documentation files
    cat > "${DOCS_DIR}/file-index.txt" << EOF
Mockzure Documentation Index
Generated: $(date -u +"%Y-%m-%d %H:%M:%S UTC")

HTML Files:
$(find "$DOCS_DIR" -name "*.html" -type f | sort)

Other Files:
$(find "$DOCS_DIR" -type f ! -name "*.html" | sort)

Coverage Report: $(if [ -f "${DOCS_DIR}/coverage.html" ]; then echo "Available"; else echo "Not available"; fi)
Compatibility Report: $(if [ -f "${DOCS_DIR}/AZURE_API_COMPATIBILITY.md" ]; then echo "Available"; else echo "Not available"; fi)
EOF
    
    print_success "Created documentation index: ${DOCS_DIR}/file-index.txt"
else
    print_status "DRY RUN: Would generate documentation index"
fi

# Step 6: Final validation
print_status "Performing final validation..."

# Check total size of docs directory
if [ "$DRY_RUN" = false ]; then
    DOCS_SIZE=$(du -sh "$DOCS_DIR" | cut -f1)
    print_status "Documentation size: $DOCS_SIZE"
    
    # Count files
    FILE_COUNT=$(find "$DOCS_DIR" -type f | wc -l)
    print_status "Total files: $FILE_COUNT"
    
    # Check for large files that might cause issues
    LARGE_FILES=$(find "$DOCS_DIR" -type f -size +1M)
    if [ -n "$LARGE_FILES" ]; then
        print_warning "Large files found:"
        echo "$LARGE_FILES"
    fi
else
    print_status "DRY RUN: Would perform final validation"
fi

print_success "Documentation preparation completed!"
print_status "Documentation directory: $DOCS_DIR"
print_status "Ready for deployment to GitHub Pages"

# Show summary
if [ "$DRY_RUN" = false ]; then
    echo ""
    print_status "Documentation Summary:"
    print_status "  - HTML files: $(find "$DOCS_DIR" -name "*.html" -type f | wc -l)"
    print_status "  - Coverage report: $(if [ -f "${DOCS_DIR}/coverage.html" ]; then echo "✓"; else echo "✗"; fi)"
    print_status "  - Compatibility report: $(if [ -f "${DOCS_DIR}/AZURE_API_COMPATIBILITY.md" ]; then echo "✓"; else echo "✗"; fi)"
    print_status "  - Total size: $(du -sh "$DOCS_DIR" | cut -f1)"
fi
