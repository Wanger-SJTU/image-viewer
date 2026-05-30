package http

import (
	"context"
	"log"
	"net/http"

	"image-viewer/internal/service"
	"image-viewer/shared/types"

	"github.com/gin-gonic/gin"
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
		if err := h.scannerSvc.Scan(context.Background(), req.Path, progressCh); err != nil {
			log.Printf("scan error: %v", err)
		}
		// Pre-generate grid thumbnails after scan completes
		h.thumbSvc.PreGenerateAll(context.Background())
	}()

	// Return accepted immediately; future enhancement: SSE on /scan/progress
	c.JSON(http.StatusAccepted, types.APIResponse{
		Success: true,
		Data:    gin.H{"path": req.Path, "status": "started"},
	})
}
