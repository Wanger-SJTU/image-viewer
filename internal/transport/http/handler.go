package http

import (
	"net/http"
	"strconv"

	"image-viewer/internal/service"
	"image-viewer/shared/types"

	"github.com/gin-gonic/gin"
)

// Handler holds all HTTP handler dependencies.
type Handler struct {
	assetSvc   *service.AssetService
	scannerSvc *service.ScannerService
	thumbSvc   *service.ThumbService
}

// NewHandler creates a new Handler.
func NewHandler(assetSvc *service.AssetService, scannerSvc *service.ScannerService, thumbSvc *service.ThumbService) *Handler {
	return &Handler{
		assetSvc:   assetSvc,
		scannerSvc: scannerSvc,
		thumbSvc:   thumbSvc,
	}
}

// --- Response helpers ---

func respondOK(c *gin.Context, data interface{}, meta *types.PaginationMeta) {
	c.JSON(http.StatusOK, types.APIResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

func respondError(c *gin.Context, code int, message string) {
	c.JSON(code, types.APIResponse{
		Success: false,
		Error:   message,
	})
}

// --- Health ---

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, types.APIResponse{Success: true, Data: "ok"})
}

// --- Asset CRUD ---

func (h *Handler) ListAssets(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	var filter types.AssetFilter
	if r := c.Query("rating"); r != "" {
		filter.Rating, _ = strconv.Atoi(r)
	}
	filter.ColorLabel = c.Query("color_label")
	filter.CameraModel = c.Query("camera_model")
	filter.MatchStatus = c.Query("match_status")
	filter.Search = c.Query("search")

	assets, total, err := h.assetSvc.List(c.Request.Context(), &filter, page, limit)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	meta := &types.PaginationMeta{Total: total, Page: page, Limit: limit}
	respondOK(c, assets, meta)
}

func (h *Handler) GetAsset(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid asset id")
		return
	}

	asset, err := h.assetSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if asset == nil {
		respondError(c, http.StatusNotFound, "asset not found")
		return
	}

	respondOK(c, asset, nil)
}

func (h *Handler) RateAsset(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid asset id")
		return
	}

	var req types.RateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.assetSvc.Rate(c.Request.Context(), id, req.Rating); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	respondOK(c, gin.H{"id": id, "rating": req.Rating}, nil)
}

func (h *Handler) LabelAsset(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid asset id")
		return
	}

	var req types.LabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.assetSvc.Label(c.Request.Context(), id, req.ColorLabel); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	respondOK(c, gin.H{"id": id, "color_label": req.ColorLabel}, nil)
}

func (h *Handler) DeleteAsset(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid asset id")
		return
	}

	if err := h.assetSvc.Delete(c.Request.Context(), id); err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, types.APIResponse{Success: true, Data: gin.H{"id": id}})
}
