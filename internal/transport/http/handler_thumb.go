package http

import (
	"strconv"

	"image-viewer/shared/types"

	"github.com/gin-gonic/gin"
)

// GetThumb returns the thumbnail file for an asset.
func (h *Handler) GetThumb(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, types.APIResponse{Success: false, Error: "invalid asset id"})
		return
	}

	size := c.DefaultQuery("size", "grid")
	cachedPath, err := h.thumbSvc.GetThumbPath(c.Request.Context(), id, size)
	if err != nil {
		c.JSON(404, types.APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.Header("Cache-Control", "public, max-age=86400")
	c.File(cachedPath)
}
