package main

import (
	"fmt"
	"log"
	"net/http"

	"image-viewer/internal/config"
	"image-viewer/internal/repository"
	"image-viewer/internal/service"
	httptransport "image-viewer/internal/transport/http"
)

func main() {
	cfg := config.Load()

	log.Printf("Image Viewer starting...")
	log.Printf("  Port: %d", cfg.Port)
	log.Printf("  DB Path: %s", cfg.DBPath)
	log.Printf("  Cache Dir: %s", cfg.CacheDir)
	log.Printf("  Concurrency: %d", cfg.ConcurrencyLimit)
	log.Printf("  RAW extensions: %v", cfg.SupportedRawExts)
	log.Printf("  JPG extensions: %v", cfg.SupportedJpgExts)

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

	// Setup router
	router := httptransport.NewRouter(assetSvc, scannerSvc, thumbSvc, nil)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
