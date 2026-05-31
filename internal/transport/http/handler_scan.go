package http

import (
	"context"
	"log"
	"net/http"
	"sync"

	"image-viewer/internal/service"
	"image-viewer/shared/types"

	"github.com/gin-gonic/gin"
)

var (
	latestProgress   service.ScanProgress
	progressMu       sync.RWMutex
)

// StartScan triggers an async directory scan.
func (h *Handler) StartScan(c *gin.Context) {
	var req types.ScanRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Path == "" {
		respondError(c, http.StatusBadRequest, "path is required")
		return
	}

	progressCh := make(chan service.ScanProgress, 100)
	go func() {
		// Consume progress updates
		for p := range progressCh {
			progressMu.Lock()
			latestProgress = p
			progressMu.Unlock()
		}
	}()

	go func() {
		if err := h.scannerSvc.Scan(context.Background(), req.Path, progressCh); err != nil {
			log.Printf("scan error: %v", err)
		}
		// Pre-generate grid thumbnails after scan completes
		h.thumbSvc.PreGenerateAll(context.Background())
	}()

	// Return accepted immediately
	c.JSON(http.StatusAccepted, types.APIResponse{
		Success: true,
		Data:    gin.H{"path": req.Path, "status": "started"},
	})
}

// ScanStatus returns the latest scan progress.
func (h *Handler) ScanStatus(c *gin.Context) {
	progressMu.RLock()
	p := latestProgress
	progressMu.RUnlock()
	c.JSON(http.StatusOK, types.APIResponse{
		Success: true,
		Data:    p,
	})
}
