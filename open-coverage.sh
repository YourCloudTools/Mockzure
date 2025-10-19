#!/bin/bash

# Open Coverage Report Script
# Opens the generated coverage report in the default browser

if [ ! -f "coverage.html" ]; then
    echo "❌ Coverage report not found. Run './coverage.sh' first."
    exit 1
fi

echo "🌐 Opening coverage report in browser..."
echo "📊 Coverage report: $(pwd)/coverage.html"

# Try to open with different commands based on OS
if command -v open >/dev/null 2>&1; then
    # macOS
    open coverage.html
elif command -v xdg-open >/dev/null 2>&1; then
    # Linux
    xdg-open coverage.html
elif command -v start >/dev/null 2>&1; then
    # Windows
    start coverage.html
else
    echo "❌ Could not find a command to open the browser."
    echo "Please manually open: $(pwd)/coverage.html"
fi
