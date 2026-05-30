package service

import (
	"fmt"
	"os"
	"time"

	"github.com/rwcarlsen/goexif/exif"

	"image-viewer/shared/types"
)

// extractExif reads all relevant EXIF metadata from an image file.
func extractExif(path string) (*types.ExifMeta, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode exif: %w", err)
	}

	meta := &types.ExifMeta{}

	// Camera model
	if tag, err := x.Get(exif.Model); err == nil {
		if s, err := tag.StringVal(); err == nil {
			meta.CameraModel = s
		}
	}

	// Lens model
	if tag, err := x.Get(exif.LensModel); err == nil {
		if s, err := tag.StringVal(); err == nil {
			meta.LensModel = s
		}
	}

	// Focal length
	if tag, err := x.Get(exif.FocalLength); err == nil {
		if num, den, err := tag.Rat2(0); err == nil && den > 0 {
			meta.FocalLength = float64(num) / float64(den)
		}
	}

	// Aperture / FNumber
	if tag, err := x.Get(exif.FNumber); err == nil {
		if num, den, err := tag.Rat2(0); err == nil && den > 0 {
			meta.Aperture = float64(num) / float64(den)
		}
	}

	// Shutter speed / ExposureTime
	if tag, err := x.Get(exif.ExposureTime); err == nil {
		if num, den, err := tag.Rat2(0); err == nil && den > 0 {
			if num < den {
				meta.ShutterSpeed = fmt.Sprintf("%d/%d", num, den)
			} else if den == 1 {
				meta.ShutterSpeed = fmt.Sprintf("%d", num)
			} else {
				meta.ShutterSpeed = fmt.Sprintf("%.1f", float64(num)/float64(den))
			}
		}
	}

	// ISO
	if tag, err := x.Get(exif.ISOSpeedRatings); err == nil {
		if v, err := tag.Int(0); err == nil {
			meta.ISO = v
		}
	}

	// Image dimensions
	if tag, err := x.Get(exif.PixelXDimension); err == nil {
		if v, err := tag.Int(0); err == nil {
			meta.Width = v
		}
	}
	if tag, err := x.Get(exif.PixelYDimension); err == nil {
		if v, err := tag.Int(0); err == nil {
			meta.Height = v
		}
	}

	// Orientation
	if tag, err := x.Get(exif.Orientation); err == nil {
		if v, err := tag.Int(0); err == nil {
			meta.Orientation = v
		}
	}

	// Capture date/time
	if dt, err := x.DateTime(); err == nil {
		meta.CapturedAt = dt
	}

	return meta, nil
}

// extractCaptureTime reads only the DateTimeOriginal from an image file.
func extractCaptureTime(path string) (time.Time, error) {
	f, err := os.Open(path)
	if err != nil {
		return time.Time{}, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		return time.Time{}, fmt.Errorf("decode exif: %w", err)
	}

	dt, err := x.DateTime()
	if err != nil {
		return time.Time{}, fmt.Errorf("read datetime: %w", err)
	}

	return dt, nil
}
