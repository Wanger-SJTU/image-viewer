package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"syscall"

	"image-viewer/internal/config"
	"image-viewer/internal/desktop"
	"image-viewer/internal/repository"
	"image-viewer/internal/service"
	httptransport "image-viewer/internal/transport/http"
)

//go:embed web/dist/*
var webDist embed.FS

func main() {
	desktopMode := flag.Bool("desktop", false, "Run as standalone desktop app (requires -tags desktop build)")
	flag.Parse()

	cfg := config.Load()

	log.Printf("Image Viewer starting...")
	log.Printf("  Port: %d", cfg.Port)
	log.Printf("  DB Path: %s", cfg.DBPath)
	log.Printf("  Cache Dir: %s", cfg.CacheDir)
	log.Printf("  Concurrency: %d", cfg.ConcurrencyLimit)

	// Initialize database
	db, err := repository.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize repository
	repo := repository.NewAssetRepository(db)

	// Initialize services
	assetSvc := service.NewAssetService(cfg, repo)
	scannerSvc := service.NewScannerService(cfg, repo)
	thumbSvc := service.NewThumbService(cfg, repo)

	// Extract embedded frontend
	var distFS fs.FS
	distFS, err = fs.Sub(webDist, "web/dist")
	if err != nil {
		log.Printf("Warning: frontend not embedded, API-only mode")
		distFS = nil
	}

	// Setup router
	router := httptransport.NewRouter(assetSvc, scannerSvc, thumbSvc, distFS)

	if *desktopMode {
		desktop.Run(router, cfg.Port)
		return
	}

	// Server mode — bind to all interfaces
	startServer(router, cfg.Port)
}

func startServer(handler http.Handler, startPort int) {
	const maxAttempts = 10
	port := startPort
	for attempts := 0; attempts < maxAttempts; attempts++ {
		addr := fmt.Sprintf("0.0.0.0:%d", port)
		log.Printf("Listening on http://%s", addr)
		if err := http.ListenAndServe(addr, handler); err != nil {
			if isAddrInUse(err) && attempts < maxAttempts-1 {
				log.Printf("Port %d is in use, trying port %d...", port, port+1)
				port++
				continue
			}
			log.Fatalf("Server error: %v", err)
		}
		break
	}
}

// isAddrInUse returns true if the error indicates the address is already in use.
func isAddrInUse(err error) bool {
	var sysErr syscall.Errno
	if errors.As(err, &sysErr) {
		return sysErr == syscall.EADDRINUSE
	}
	return false
}
