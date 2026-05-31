package types

import "time"

// AssetFilter holds optional filtering criteria for listing assets.
// Zero values mean "no filter applied".
type AssetFilter struct {
	Rating         int       `json:"rating"`
	ColorLabel     string    `json:"color_label"`
	CameraModel    string    `json:"camera_model"`
	MatchStatus    string    `json:"match_status"`
	FileType       string    `json:"file_type"` // "jpg", "raw", "both"
	FocalLengthMin float64   `json:"focal_length_min"`
	FocalLengthMax float64   `json:"focal_length_max"`
	ApertureMin    float64   `json:"aperture_min"`
	ApertureMax    float64   `json:"aperture_max"`
	ISOMin         int       `json:"iso_min"`
	ISOMax         int       `json:"iso_max"`
	CapturedAfter  time.Time `json:"captured_after"`
	CapturedBefore time.Time `json:"captured_before"`
	Search         string    `json:"search"`
	Trashed        *bool     `json:"trashed,omitempty"` // nil=exclude trashed (default), true=only trashed
}
