package types

import "time"

// MediaType represents the type of a media file.
type MediaType string

const (
	MediaTypeRAW MediaType = "raw"
	MediaTypeJPG MediaType = "jpg"
)

// MatchStatus represents the pairing status of an asset.
type MatchStatus string

const (
	MatchStatusPaired MatchStatus = "paired"
	MatchStatusOrphan MatchStatus = "orphan"
)

// ValidColorLabels is the set of allowed color label values.
var ValidColorLabels = map[string]bool{
	"":       true,
	"red":    true,
	"orange": true,
	"yellow": true,
	"green":  true,
	"blue":   true,
	"purple": true,
}

// ExifMeta holds EXIF metadata extracted from an image file.
type ExifMeta struct {
	CameraModel  string    `json:"camera_model"`
	LensModel    string    `json:"lens_model,omitempty"`
	FocalLength  float64   `json:"focal_length,omitempty"`
	Aperture     float64   `json:"aperture,omitempty"`
	ShutterSpeed string    `json:"shutter_speed,omitempty"`
	ISO          int       `json:"iso,omitempty"`
	CapturedAt   time.Time `json:"captured_at,omitempty"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	Orientation  int       `json:"orientation,omitempty"`
}

// MediaFile represents a single physical image file on disk.
type MediaFile struct {
	ID        int64     `json:"id"`
	AssetID   int64     `json:"asset_id"`
	FilePath  string    `json:"file_path"`
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`
	MediaType MediaType `json:"media_type"`
	Exif      *ExifMeta `json:"exif,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Asset represents a logical photo asset (RAW + JPG pair or orphan).
type Asset struct {
	ID          int64       `json:"id"`
	Name        string      `json:"name"`
	DirPath     string      `json:"dir_path"`
	RawFile     *MediaFile  `json:"raw_file,omitempty"`
	JpgFile     *MediaFile  `json:"jpg_file,omitempty"`
	MatchStatus MatchStatus `json:"match_status"`
	Rating      int         `json:"rating"`
	ColorLabel  string      `json:"color_label"`
	AiStatus    string      `json:"ai_status"`
	CapturedAt  *time.Time  `json:"captured_at,omitempty"`
	GridThumb   string      `json:"grid_thumb"`
	FullThumb   string      `json:"full_thumb"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
