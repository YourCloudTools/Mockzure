#!/bin/bash

# Mockzure Startup Script

echo "🚀 Starting Mockzure..."
echo ""

# Build the binary if it doesn't exist or main.go is newer
if [ ! -f mockzure ] || [ main.go -nt mockzure ]; then
    echo "📦 Building Mockzure..."
    go build -o mockzure main.go
    if [ $? -ne 0 ]; then
        echo "❌ Build failed"
        exit 1
    fi
    echo "✅ Build complete"
fi

echo ""
echo "🌐 Mockzure Portal will be available at:"
echo "   http://localhost:8090"
echo ""
echo "📋 Quick Access:"
echo "   - Resource Groups: http://localhost:8090 (Resource Groups tab)"
echo "   - Entra ID: http://localhost:8090 (Entra ID tab)"
echo "   - Settings: http://localhost:8090 (Settings tab)"
echo ""
echo "📚 API Endpoints:"
echo "   - GET  /mock/azure/resource-groups"
echo "   - GET  /mock/azure/vms"
echo "   - GET  /mock/azure/users"
echo "   - GET  /mock/azure/apps"
echo "   - GET  /mock/azure/stats"
echo ""
echo "Press Ctrl+C to stop Mockzure"
echo ""

./mockzure

