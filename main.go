package main

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"syscall"

	"image-viewer/internal/config"
	"image-viewer/internal/repository"
	"image-viewer/internal/service"
	httptransport "image-viewer/internal/transport/http"
)

//go:embed web/dist/*
var webDist embed.FS

func main() {
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

	// Try ports starting from cfg.Port, auto-increment if bind fails
	port := cfg.Port
	const maxAttempts = 10
	var lnErr error
	for attempts := 0; attempts < maxAttempts; attempts++ {
		addr := fmt.Sprintf(":%d", port)
		log.Printf("Listening on http://localhost%s", addr)
		if lnErr = http.ListenAndServe(addr, router); lnErr != nil {
			if isAddrInUse(lnErr) && attempts < maxAttempts-1 {
				log.Printf("Port %d is in use, trying port %d...", port, port+1)
				port++
				continue
			}
			log.Fatalf("Server error: %v", lnErr)
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
