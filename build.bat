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

if not exist build mkdir build

echo.
echo [2/2] Building viewer.exe (standalone desktop app)...
go build -tags desktop -ldflags="-s -w -H windowsgui" -o build\viewer.exe .

echo.
echo === Build complete ===
echo build\viewer.exe
echo Double-click to run the standalone desktop app.
pause
