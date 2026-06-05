@echo off
echo === Image Viewer Windows Build ===

echo.
echo [1/2] Building frontend...
if exist web (
    cd web
    call npm install --silent 2>nul
    call npm run build
    cd ..
) else (
    echo   (web/ directory not found, skipping)
)

echo.
echo [2/2] Building viewer.exe (standalone desktop app)...
go build -tags desktop -ldflags="-s -w -H windowsgui" -o viewer.exe .

echo.
echo === Build complete: viewer.exe ===
echo Double-click viewer.exe to run the standalone desktop app.
pause
