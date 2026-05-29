#!/bin/bash
set -e

cd "$(dirname "$0")"

echo "=== Image Viewer Build ==="

# Build frontend
echo ""
echo "[1/2] Building frontend..."
if [ -d web ]; then
    cd web
    npm install --silent 2>/dev/null || true
    npm run build
    cd ..
else
    echo "  (web/ directory not found, skipping)"
fi

# Build backend with embedded frontend
echo ""
echo "[2/2] Building viewer binary..."
CGO_ENABLED=1 go build -ldflags="-s -w" -o viewer ./main.go

echo ""
echo "=== Build complete: ./viewer ==="
echo "Run with: ./viewer"
