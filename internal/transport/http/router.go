package http

import (
	"io/fs"
	"net/http"

	"image-viewer/internal/service"

	"github.com/gin-gonic/gin"
)

// NewRouter creates and configures the Gin router with all API routes.
// If webDist is not nil, it serves the embedded frontend from that filesystem.
func NewRouter(
	assetSvc *service.AssetService,
	scannerSvc *service.ScannerService,
	thumbSvc *service.ThumbService,
	webDist fs.FS,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), CORS())

	h := NewHandler(assetSvc, scannerSvc, thumbSvc)

	api := r.Group("/api/v1")
	{
		api.GET("/health", h.Health)

		api.GET("/assets", h.ListAssets)
		api.GET("/assets/:id", h.GetAsset)
		api.POST("/assets/:id/rate", h.RateAsset)
		api.POST("/assets/:id/label", h.LabelAsset)
		api.DELETE("/assets", h.ClearAssets)
		api.DELETE("/assets/:id", h.DeleteAsset)
		api.POST("/assets/:id/trash", h.TrashAsset)
		api.POST("/assets/:id/restore", h.RestoreAsset)
		api.POST("/assets/:id/purge", h.PurgeAsset)

		api.GET("/thumbs/:id", h.GetThumb)

		api.GET("/filters", h.GetFilterOptions)

		api.POST("/scan", h.StartScan)
		api.GET("/scan/status", h.ScanStatus)
	}

	// Serve embedded frontend if available (already Sub'd by caller)
	if webDist != nil {
		r.NoRoute(gin.WrapH(http.FileServer(http.FS(webDist))))
	}

	return r
}
