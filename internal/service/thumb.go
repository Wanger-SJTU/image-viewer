package service

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"os"
	"path/filepath"

	"image-viewer/internal/config"
	"image-viewer/shared/types"
)

const (
	ThumbGrid = "grid"
	ThumbFull = "full"
)

// ThumbService handles thumbnail generation and caching.
type ThumbService struct {
	cfg  *config.Config
	repo interface {
		FindByID(ctx context.Context, id int64) (*types.Asset, error)
		UpdateThumbnails(ctx context.Context, id int64, gridThumb, fullThumb string) error
	}
}

// NewThumbService creates a new ThumbService.
func NewThumbService(cfg *config.Config, repo interface {
	FindByID(ctx context.Context, id int64) (*types.Asset, error)
	UpdateThumbnails(ctx context.Context, id int64, gridThumb, fullThumb string) error
}) *ThumbService {
	return &ThumbService{cfg: cfg, repo: repo}
}

// GetThumbPath returns the cached thumbnail path for an asset, or generates it if missing.
func (s *ThumbService) GetThumbPath(ctx context.Context, assetID int64, size string) (string, error) {
	if size != ThumbGrid && size != ThumbFull {
		return "", fmt.Errorf("invalid size: %q, must be %q or %q", size, ThumbGrid, ThumbFull)
	}

	asset, err := s.repo.FindByID(ctx, assetID)
	if err != nil {
		return "", fmt.Errorf("find asset: %w", err)
	}
	if asset == nil {
		return "", fmt.Errorf("asset %d not found", assetID)
	}

	var thumbPath string
	if size == ThumbGrid {
		thumbPath = asset.GridThumb
	} else {
		thumbPath = asset.FullThumb
	}

	// Check cache
	if thumbPath != "" {
		if _, err := os.Stat(thumbPath); err == nil {
			return thumbPath, nil
		}
	}

	// Generate if missing
	return s.GenerateThumb(ctx, asset, size)
}

// GenerateThumb generates a thumbnail for the given asset and size.
func (s *ThumbService) GenerateThumb(ctx context.Context, asset *types.Asset, size string) (string, error) {
	if size != ThumbGrid && size != ThumbFull {
		return "", fmt.Errorf("invalid size: %q, must be %q or %q", size, ThumbGrid, ThumbFull)
	}

	// Determine source file: prefer JPG, fallback to RAW
	var srcPath string
	if asset.JpgFile != nil {
		srcPath = asset.JpgFile.FilePath
	} else if asset.RawFile != nil {
		srcPath = asset.RawFile.FilePath
	} else {
		return "", fmt.Errorf("no source file for asset %d", asset.ID)
	}

	// Open source
	f, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("open source: %w", err)
	}
	defer f.Close()

	// Decode — try standard decode, then RAW preview extraction as fallback
	img, _, err := image.Decode(f)
	if err != nil {
		if asset.RawFile != nil {
			jpegData, extractErr := ExtractEmbeddedJPEG(srcPath)
			if extractErr != nil {
				return "", fmt.Errorf("decode image: %w (raw extraction: %v)", err, extractErr)
			}
			img, _, err = image.Decode(bytes.NewReader(jpegData))
			if err != nil {
				return "", fmt.Errorf("decode extracted preview: %w", err)
			}
		} else {
			return "", fmt.Errorf("decode image: %w", err)
		}
	}

	// Resize
	targetSize := 200
	if size == ThumbFull {
		targetSize = 2048
	}
	resized := resizeImage(img, targetSize)

	// Save to cache
	thumbName := fmt.Sprintf("%d_%s.jpg", asset.ID, size)
	thumbPath := filepath.Join(s.cfg.CacheDir, thumbName)

	outFile, err := os.Create(thumbPath)
	if err != nil {
		return "", fmt.Errorf("create thumb file: %w", err)
	}
	defer outFile.Close()

	if err := jpeg.Encode(outFile, resized, &jpeg.Options{Quality: 85}); err != nil {
		return "", fmt.Errorf("encode thumb: %w", err)
	}

	// Update database
	if size == ThumbGrid {
		if err := s.repo.UpdateThumbnails(ctx, asset.ID, thumbPath, asset.FullThumb); err != nil {
			return "", fmt.Errorf("update thumbnails: %w", err)
		}
	} else {
		if err := s.repo.UpdateThumbnails(ctx, asset.ID, asset.GridThumb, thumbPath); err != nil {
			return "", fmt.Errorf("update thumbnails: %w", err)
		}
	}

	return thumbPath, nil
}

// resizeImage resizes an image to fit within targetSize while maintaining aspect ratio.
// Uses bilinear interpolation via the standard library.
func resizeImage(img image.Image, targetSize int) image.Image {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	if w <= targetSize && h <= targetSize {
		return img
	}

	var newW, newH int
	if w > h {
		newW = targetSize
		newH = int(math.Max(1, float64(h)*float64(targetSize)/float64(w)))
	} else {
		newH = targetSize
		newW = int(math.Max(1, float64(w)*float64(targetSize)/float64(h)))
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))

	// Bilinear interpolation
	xRatio := float64(w) / float64(newW)
	yRatio := float64(h) / float64(newH)

	for dy := 0; dy < newH; dy++ {
		for dx := 0; dx < newW; dx++ {
			px := float64(dx)*xRatio + 0.5
			py := float64(dy)*yRatio + 0.5

			x0 := int(px)
			y0 := int(py)
			x1 := x0 + 1
			y1 := y0 + 1

			// Clamp to bounds
			if x0 < 0 {
				x0 = 0
			}
			if x0 >= w {
				x0 = w - 1
			}
			if y0 < 0 {
				y0 = 0
			}
			if y0 >= h {
				y0 = h - 1
			}
			if x1 >= w {
				x1 = w - 1
			}
			if y1 >= h {
				y1 = h - 1
			}

			xFrac := px - float64(x0)
			yFrac := py - float64(y0)

			c00 := img.At(x0+bounds.Min.X, y0+bounds.Min.Y)
			c10 := img.At(x1+bounds.Min.X, y0+bounds.Min.Y)
			c01 := img.At(x0+bounds.Min.X, y1+bounds.Min.Y)
			c11 := img.At(x1+bounds.Min.X, y1+bounds.Min.Y)

			r00, g00, b00, a00 := c00.RGBA()
			r10, g10, b10, a10 := c10.RGBA()
			r01, g01, b01, a01 := c01.RGBA()
			r11, g11, b11, a11 := c11.RGBA()

			// Interpolate
			r := lerp(lerp(float64(r00), float64(r10), xFrac), lerp(float64(r01), float64(r11), xFrac), yFrac)
			g := lerp(lerp(float64(g00), float64(g10), xFrac), lerp(float64(g01), float64(g11), xFrac), yFrac)
			b := lerp(lerp(float64(b00), float64(b10), xFrac), lerp(float64(b01), float64(b11), xFrac), yFrac)
			a := lerp(lerp(float64(a00), float64(a10), xFrac), lerp(float64(a01), float64(a11), xFrac), yFrac)

			const max16 = 65535.0
			dst.Set(dx, dy, color.RGBA64{
				R: uint16(clamp(r, 0, max16)),
				G: uint16(clamp(g, 0, max16)),
				B: uint16(clamp(b, 0, max16)),
				A: uint16(clamp(a, 0, max16)),
			})
		}
	}

	return dst
}

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
