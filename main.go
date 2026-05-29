package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"

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

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
