#!/bin/bash
set -e

cd "$(dirname "$0")"

echo "=== Image Viewer Build ==="

# Build frontend (once, shared by all platforms)
echo ""
echo "[1/3] Building frontend..."
if [ -d web ]; then
    cd web
    npm install --silent 2>/dev/null || true
    npm run build
    cd ..
else
    echo "  (web/ directory not found, skipping)"
fi

mkdir -p build

# ── Linux ──────────────────────────────────────────────
echo ""
echo "[2/3] Building Linux..."

if pkg-config --exists gtk+-3.0 webkit2gtk-4.0 2>/dev/null; then
    echo "  desktop mode..."
    CGO_ENABLED=1 go build -tags desktop -ldflags="-s -w" -o build/viewer .
else
    echo "  server mode (gtk/webkit dev not found) ..."
    CGO_ENABLED=1 go build -ldflags="-s -w" -o build/viewer .
fi
echo "  -> build/viewer"

# ── Windows ────────────────────────────────────────────
echo ""
echo "[3/3] Building Windows..."

if command -v x86_64-w64-mingw32-gcc &>/dev/null; then
    echo "  desktop mode..."
    CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ \
        GOOS=windows GOARCH=amd64 \
        go build -tags desktop -ldflags="-s -w -H windowsgui" -o build/viewer.exe .
else
    echo "  skipping (mingw-w64 not found)"
    echo "  HINT: sudo apt install gcc-mingw-w64-x86-64 g++-mingw-w64-x86-64"
fi
echo "  -> build/viewer.exe"

echo ""
echo "=== Build complete ==="
ls -lh build/ 2>/dev/null || true
