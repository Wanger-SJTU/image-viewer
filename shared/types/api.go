package types

// PaginatedRequest holds common pagination parameters.
type PaginatedRequest struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// DefaultPage returns the default page number if unset.
func (p *PaginatedRequest) DefaultPage() int {
	if p.Page <= 0 {
		return 1
	}
	return p.Page
}

// DefaultLimit returns the default limit if unset, clamped to max.
func (p *PaginatedRequest) DefaultLimit() int {
	switch {
	case p.Limit <= 0:
		return 50
	case p.Limit > 10000:
		return 10000
	default:
		return p.Limit
	}
}

// Offset returns the SQL offset for the current page.
func (p *PaginatedRequest) Offset() int {
	return (p.DefaultPage() - 1) * p.DefaultLimit()
}

// PaginationMeta holds pagination metadata for list responses.
type PaginationMeta struct {
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}

// APIResponse is the standard JSON response envelope for all API endpoints.
type APIResponse struct {
	Success bool            `json:"success"`
	Data    interface{}     `json:"data"`
	Error   string          `json:"error,omitempty"`
	Meta    *PaginationMeta `json:"meta,omitempty"`
}

// ScanRequest is the request body for triggering a directory scan.
type ScanRequest struct {
	Path string `json:"path"`
}

// RateRequest is the request body for setting an asset rating.
type RateRequest struct {
	Rating int `json:"rating"`
}

// LabelRequest is the request body for setting an asset color label.
type LabelRequest struct {
	ColorLabel string `json:"color_label"`
}

// FilterOptions holds the available filter values from the database.
type FilterOptions struct {
	CameraModels  []string  `json:"camera_models"`
	FocalLengths  []float64 `json:"focal_lengths"`
	Apertures     []float64 `json:"apertures"`
	ISOs          []int     `json:"isos"`
	ColorLabels   []string  `json:"color_labels"`
	FileTypes     []string  `json:"file_types"`
	PhotoDates    []string  `json:"photo_dates"`
}
