#!/bin/bash
set -e

cd "$(dirname "$0")"

MODE="${1:-server}"

echo "=== Image Viewer Build ==="
echo "  Mode: $MODE"

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
mkdir -p build

case "$MODE" in
  desktop)
    CGO_ENABLED=1 go build -tags desktop -ldflags="-s -w" -o build/viewer .
    echo ""
    echo "=== Build complete: build/viewer ==="
    echo "Run with: ./build/viewer --desktop"
    ;;
  windows)
    if [ "$(go env GOOS)" != "windows" ]; then
      echo "  Cross-compiling for Windows..."
      CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -tags desktop -ldflags="-s -w -H windowsgui" -o build/viewer.exe .
    else
      CGO_ENABLED=1 go build -tags desktop -ldflags="-s -w -H windowsgui" -o build/viewer.exe .
    fi
    echo ""
    echo "=== Build complete: build/viewer.exe ==="
    echo "Distribute viewer.exe as a standalone Windows app"
    ;;
  server|*)
    CGO_ENABLED=1 go build -ldflags="-s -w" -o build/viewer .
    echo ""
    echo "=== Build complete: build/viewer ==="
    echo "Run with: ./build/viewer"
    ;;
esac
