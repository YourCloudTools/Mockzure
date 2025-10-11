#!/bin/bash

# Mockzure Startup Script
# Builds and starts Mockzure mock Azure server

set -e
cd "$(dirname "$0")/.."

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_success() { echo -e "${GREEN}✅ $1${NC}"; }
print_error() { echo -e "${RED}❌ $1${NC}"; }
print_status() { echo -e "${BLUE}ℹ️  $1${NC}"; }

# Parse arguments
FOREGROUND=false
QUIET=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --foreground|-f)
            FOREGROUND=true
            shift
            ;;
        --quiet|-q)
            QUIET=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --foreground, -f  Run in foreground (default: background)"
            echo "  --quiet, -q       Minimal output"
            echo "  --help, -h        Show this help"
            echo ""
            echo "Examples:"
            echo "  $0                # Start in background"
            echo "  $0 --foreground   # Run in foreground with logs"
            echo "  $0 --quiet        # Minimal output (for scripts)"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Run '$0 --help' for usage"
            exit 1
            ;;
    esac
done

[ "$QUIET" = false ] && echo -e "${BLUE}🔷 Mockzure Startup${NC}"
[ "$QUIET" = false ] && echo "==================="
[ "$QUIET" = false ] && echo ""

# Check if already running
if lsof -Pi :8090 -sTCP:LISTEN -t >/dev/null 2>&1; then
    print_error "Port 8090 is already in use"
    [ "$QUIET" = false ] && echo "To kill existing process: lsof -ti:8090 | xargs kill -9"
    exit 1
fi

# Build Mockzure if needed
[ "$QUIET" = false ] && print_status "Checking Mockzure binary..."

if [ ! -f "mockzure" ] || [ "main.go" -nt "mockzure" ]; then
    [ "$QUIET" = false ] && print_status "Building Mockzure..."
    go build -o mockzure main.go
    [ "$QUIET" = false ] && print_success "Mockzure built"
else
    [ "$QUIET" = false ] && print_success "Mockzure is up to date"
fi

# Create logs directory
mkdir -p ../logs

# Start Mockzure
if [ "$FOREGROUND" = true ]; then
    [ "$QUIET" = false ] && print_status "Starting Mockzure on :8090 (foreground)..."
    [ "$QUIET" = false ] && print_status "Press Ctrl+C to stop"
    [ "$QUIET" = false ] && echo ""
    ./mockzure
else
    [ "$QUIET" = false ] && print_status "Starting Mockzure on :8090 (background)..."
    ./mockzure > ../logs/mockzure.log 2>&1 &
    MOCKZURE_PID=$!
    
    # Wait and verify
    sleep 2
    
    if curl -s http://localhost:8090/mock/azure/stats > /dev/null 2>&1; then
        [ "$QUIET" = false ] && print_success "Mockzure running on http://localhost:8090"
        [ "$QUIET" = false ] && print_status "PID: $MOCKZURE_PID"
        [ "$QUIET" = false ] && print_status "Logs: tail -f ../logs/mockzure.log"
        [ "$QUIET" = false ] && echo ""
        [ "$QUIET" = false ] && print_status "To stop: kill $MOCKZURE_PID"
        
        # Export PID for parent script to use (logs dir already exists from above)
        echo "$MOCKZURE_PID" > ../logs/mockzure.pid
    else
        print_error "Failed to start Mockzure"
        cat ../logs/mockzure.log || true
        exit 1
    fi
fi

