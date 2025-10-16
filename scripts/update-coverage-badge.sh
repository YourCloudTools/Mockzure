#!/bin/bash

# Update Coverage Badge Script
# Calculates comprehensive coverage and updates documentation badges

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
    echo "  --coverage-file FILE    Path to coverage.out file (default: coverage.out)"
    echo "  --docs-dir DIR          Path to docs directory (default: docs)"
    echo "  --dry-run               Show what would be updated without making changes"
    echo "  --help                  Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                      # Update badges with default settings"
    echo "  $0 --dry-run            # Preview changes without updating"
    echo "  $0 --coverage-file custom.out  # Use custom coverage file"
}

# Default values
COVERAGE_FILE="coverage.out"
DOCS_DIR="docs"
DRY_RUN=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --coverage-file)
            COVERAGE_FILE="$2"
            shift 2
            ;;
        --docs-dir)
            DOCS_DIR="$2"
            shift 2
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

# Check if coverage file exists
if [ ! -f "$COVERAGE_FILE" ]; then
    print_error "Coverage file not found: $COVERAGE_FILE"
    print_status "Run tests with coverage first: go test -coverprofile=coverage.out ./..."
    exit 1
fi

# Check if docs directory exists
if [ ! -d "$DOCS_DIR" ]; then
    print_error "Docs directory not found: $DOCS_DIR"
    exit 1
fi

print_status "Calculating comprehensive coverage from $COVERAGE_FILE..."

# Calculate total coverage percentage
TOTAL_COVERAGE=$(go tool cover -func="$COVERAGE_FILE" | tail -1 | awk '{print $3}' | sed 's/%//')

if [ -z "$TOTAL_COVERAGE" ]; then
    print_error "Failed to calculate coverage percentage"
    exit 1
fi

# Round to 1 decimal place
TOTAL_COVERAGE=$(printf "%.1f" "$TOTAL_COVERAGE")

print_status "Total coverage: ${TOTAL_COVERAGE}%"

# Determine badge color based on coverage percentage
if (( $(echo "$TOTAL_COVERAGE >= 80" | bc -l) )); then
    BADGE_COLOR="green"
    BADGE_LABEL="brightgreen"
elif (( $(echo "$TOTAL_COVERAGE >= 70" | bc -l) )); then
    BADGE_COLOR="yellow"
    BADGE_LABEL="yellow"
elif (( $(echo "$TOTAL_COVERAGE >= 40" | bc -l) )); then
    BADGE_COLOR="orange"
    BADGE_LABEL="orange"
else
    BADGE_COLOR="red"
    BADGE_LABEL="red"
fi

print_status "Badge color: $BADGE_COLOR"

# Generate badge URL
BADGE_URL="https://img.shields.io/badge/coverage-${TOTAL_COVERAGE}%25-${BADGE_LABEL}"

# Generate badge HTML
BADGE_HTML="<img src=\"${BADGE_URL}\" alt=\"Coverage\">"

print_status "Badge URL: $BADGE_URL"

# Function to update HTML files
update_html_file() {
    local file="$1"
    local temp_file="${file}.tmp"
    
    if [ ! -f "$file" ]; then
        print_warning "File not found: $file"
        return 1
    fi
    
    print_status "Updating $file..."
    
    if [ "$DRY_RUN" = true ]; then
        print_status "DRY RUN: Would update coverage badge in $file"
        grep -n "coverage.*%" "$file" || print_warning "No coverage badge found in $file"
        return 0
    fi
    
    # Create backup
    cp "$file" "${file}.backup"
    
    # Update the file
    if grep -q "<!-- COVERAGE_BADGE_START -->" "$file"; then
        # Use comment markers if they exist
        sed -i.tmp "/<!-- COVERAGE_BADGE_START -->/,/<!-- COVERAGE_BADGE_END -->/c\\
<!-- COVERAGE_BADGE_START -->\\
${BADGE_HTML}\\
<!-- COVERAGE_BADGE_END -->" "$file"
        rm -f "${file}.tmp"
    else
        # Fallback: replace any existing coverage badge
        sed -i.tmp "s|<img src=\"https://img.shields.io/badge/coverage-[0-9.]*%25-[^\"]*\" alt=\"Coverage\">|${BADGE_HTML}|g" "$file"
        rm -f "${file}.tmp"
    fi
    
    # Verify the update
    if grep -q "$TOTAL_COVERAGE%" "$file"; then
        print_success "Updated coverage badge in $file"
        rm -f "${file}.backup"
    else
        print_error "Failed to update $file, restoring backup"
        mv "${file}.backup" "$file"
        return 1
    fi
}

# Update HTML files in docs directory
HTML_FILES=$(find "$DOCS_DIR" -name "*.html" -type f)

if [ -z "$HTML_FILES" ]; then
    print_warning "No HTML files found in $DOCS_DIR"
else
    for file in $HTML_FILES; do
        update_html_file "$file"
    done
fi

# Update README.md if it exists
if [ -f "README.md" ]; then
    print_status "Updating README.md..."
    
    if [ "$DRY_RUN" = true ]; then
        print_status "DRY RUN: Would update coverage badge in README.md"
        grep -n "coverage.*%" "README.md" || print_warning "No coverage badge found in README.md"
    else
        # Create backup
        cp "README.md" "README.md.backup"
        
        # Update coverage badge in README
        sed -i.tmp "s|!\[Coverage\](https://img.shields.io/badge/coverage-[0-9.]*%25-[^)]*)|![Coverage](${BADGE_URL})|g" "README.md"
        rm -f "README.md.tmp"
        
        # Verify the update
        if grep -q "$TOTAL_COVERAGE%" "README.md"; then
            print_success "Updated coverage badge in README.md"
            rm -f "README.md.backup"
        else
            print_error "Failed to update README.md, restoring backup"
            mv "README.md.backup" "README.md"
        fi
    fi
fi

# Generate coverage summary
print_status "Generating coverage summary..."

# Get function-level coverage
FUNCTION_COVERAGE=$(go tool cover -func="$COVERAGE_FILE" | grep -v "total:" | wc -l)
COVERED_FUNCTIONS=$(go tool cover -func="$COVERAGE_FILE" | grep -v "total:" | awk '{print $3}' | sed 's/%//' | awk '$1 > 0' | wc -l)

print_status "Coverage Summary:"
print_status "  Total Coverage: ${TOTAL_COVERAGE}%"
print_status "  Functions: $FUNCTION_COVERAGE"
print_status "  Covered Functions: $COVERED_FUNCTIONS"
print_status "  Badge Color: $BADGE_COLOR"

# Create coverage summary file
if [ "$DRY_RUN" = false ]; then
    cat > "${DOCS_DIR}/coverage-summary.json" << EOF
{
  "generated_at": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "total_coverage": $TOTAL_COVERAGE,
  "total_functions": $FUNCTION_COVERAGE,
  "covered_functions": $COVERED_FUNCTIONS,
  "badge_color": "$BADGE_COLOR",
  "badge_url": "$BADGE_URL"
}
EOF
    print_success "Created coverage summary: ${DOCS_DIR}/coverage-summary.json"
fi

print_success "Coverage badge update completed!"
print_status "Coverage: ${TOTAL_COVERAGE}% (${BADGE_COLOR})"
