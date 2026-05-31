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

	"github.com/rwcarlsen/goexif/exif"

	"image-viewer/internal/config"
	"image-viewer/internal/jpegdecoder"
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
		FindAllIDs(ctx context.Context) ([]int64, error)
	}
}

// NewThumbService creates a new ThumbService.
func NewThumbService(cfg *config.Config, repo interface {
	FindByID(ctx context.Context, id int64) (*types.Asset, error)
	UpdateThumbnails(ctx context.Context, id int64, gridThumb, fullThumb string) error
	FindAllIDs(ctx context.Context) ([]int64, error)
}) *ThumbService {
	return &ThumbService{cfg: cfg, repo: repo}
}

// GetThumbPath returns the cached thumbnail path for an asset, or generates it if missing.
// fileType can be "jpg" or "raw" to force a specific source; empty string auto-selects (JPG preferred).
func (s *ThumbService) GetThumbPath(ctx context.Context, assetID int64, size string, fileType string) (string, error) {
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

	// If a specific file type is requested, generate a type-specific cache file
	if fileType != "" {
		thumbName := fmt.Sprintf("%d_%s_%s.jpg", assetID, size, fileType)
		thumbPath := filepath.Join(s.cfg.CacheDir, thumbName)
		if _, err := os.Stat(thumbPath); err == nil {
			return thumbPath, nil
		}
		return s.generateForType(ctx, asset, size, fileType)
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

// generateForType generates a thumbnail from a specific file type (jpg or raw).
// Does NOT update the database columns — saves to a type-specific cache file.
func (s *ThumbService) generateForType(ctx context.Context, asset *types.Asset, size string, fileType string) (string, error) {
	var srcPath string
	switch fileType {
	case "jpg":
		if asset.JpgFile != nil {
			srcPath = asset.JpgFile.FilePath
		} else {
			return "", fmt.Errorf("asset %d has no JPG file", asset.ID)
		}
	case "raw":
		if asset.RawFile != nil {
			srcPath = asset.RawFile.FilePath
		} else {
			return "", fmt.Errorf("asset %d has no RAW file", asset.ID)
		}
	default:
		return "", fmt.Errorf("invalid file type: %q", fileType)
	}

	targetSize := 600
	if size == ThumbFull {
		targetSize = 2048
	}

	var img image.Image
	var err error

	isRaw := fileType == "raw"
	if isRaw {
		jpegData, extractErr := ExtractEmbeddedJPEG(srcPath)
		if extractErr != nil {
			return "", fmt.Errorf("raw preview extraction: %w", extractErr)
		}
		img, _, err = image.Decode(bytes.NewReader(jpegData))
		if err != nil {
			return "", fmt.Errorf("decode extracted preview: %w", err)
		}
		if o := readOrientation(srcPath); o > 1 {
			img = applyOrientation(img, o)
		}
	} else {
		scale := pickScale(srcPath, targetSize)
		if scale > 1 {
			img, err = jpegdecoder.DecodeFileScaled(srcPath, scale)
		} else {
			img, err = jpegdecoder.DecodeFile(srcPath)
		}
		if err != nil {
			return "", fmt.Errorf("decode jpeg: %w", err)
		}
		if o := readOrientation(srcPath); o > 1 {
			img = applyOrientation(img, o)
		}
	}

	resized := resizeImage(img, targetSize)

	thumbName := fmt.Sprintf("%d_%s_%s.jpg", asset.ID, size, fileType)
	thumbPath := filepath.Join(s.cfg.CacheDir, thumbName)

	outFile, err := os.Create(thumbPath)
	if err != nil {
		return "", fmt.Errorf("create thumb file: %w", err)
	}
	defer outFile.Close()

	if err := jpeg.Encode(outFile, resized, &jpeg.Options{Quality: 85}); err != nil {
		return "", fmt.Errorf("encode thumb: %w", err)
	}

	return thumbPath, nil
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

	// Determine target size
	targetSize := 600
	if size == ThumbFull {
		targetSize = 2048
	}

	// Decode using libjpeg-turbo with DCT-domain scaling for speed
	var img image.Image
	var err error
	if asset.JpgFile != nil {
		// Pick scale factor: decode just above target resolution, then bilinear down
		scale := pickScale(srcPath, targetSize)
		if scale > 1 {
			img, err = jpegdecoder.DecodeFileScaled(srcPath, scale)
		} else {
			img, err = jpegdecoder.DecodeFile(srcPath)
		}
		if err != nil {
			return "", fmt.Errorf("decode jpeg: %w", err)
		}
	} else if asset.RawFile != nil {
		// RAW file: try extracting embedded JPEG preview
		jpegData, extractErr := ExtractEmbeddedJPEG(srcPath)
		if extractErr != nil {
			return "", fmt.Errorf("raw preview extraction: %w", extractErr)
		}
		img, _, err = image.Decode(bytes.NewReader(jpegData))
		if err != nil {
			return "", fmt.Errorf("decode extracted preview: %w", err)
		}
		// Read orientation from embedded JPEG - goexif cannot parse RAW files
		if o := readOrientation(srcPath); o > 1 {
			img = applyOrientation(img, o)
		}
	} else {
		return "", fmt.Errorf("no source file for asset %d", asset.ID)
	}

	// Apply EXIF orientation before resize (JPG only — RAW handled above)
	if asset.JpgFile != nil {
		if o := readOrientation(srcPath); o > 1 {
			img = applyOrientation(img, o)
		}
	}

	// Resize
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

// applyOrientation transforms an image according to the EXIF orientation tag.
// Returns the original image unchanged if orientation is 0, 1, or unrecognized.
func applyOrientation(img image.Image, orientation int) image.Image {
	if orientation <= 1 {
		return img
	}

	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	var dst *image.RGBA

	switch orientation {
	case 2: // flip X
		dst = image.NewRGBA(image.Rect(0, 0, w, h))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dst.Set(x, y, img.At(w-1-x+bounds.Min.X, y+bounds.Min.Y))
			}
		}
	case 3: // rotate 180°
		dst = image.NewRGBA(image.Rect(0, 0, w, h))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dst.Set(x, y, img.At(w-1-x+bounds.Min.X, h-1-y+bounds.Min.Y))
			}
		}
	case 4: // flip Y
		dst = image.NewRGBA(image.Rect(0, 0, w, h))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dst.Set(x, y, img.At(x+bounds.Min.X, h-1-y+bounds.Min.Y))
			}
		}
	case 5: // transpose
		dst = image.NewRGBA(image.Rect(0, 0, h, w))
		for y := 0; y < w; y++ {
			for x := 0; x < h; x++ {
				dst.Set(x, y, img.At(y+bounds.Min.X, x+bounds.Min.Y))
			}
		}
	case 6: // rotate 90° CW
		dst = image.NewRGBA(image.Rect(0, 0, h, w))
		for y := 0; y < w; y++ {
			for x := 0; x < h; x++ {
				dst.Set(x, y, img.At(y+bounds.Min.X, h-1-x+bounds.Min.Y))
			}
		}
	case 7: // transverse
		dst = image.NewRGBA(image.Rect(0, 0, h, w))
		for y := 0; y < w; y++ {
			for x := 0; x < h; x++ {
				dst.Set(x, y, img.At(w-1-y+bounds.Min.X, h-1-x+bounds.Min.Y))
			}
		}
	case 8: // rotate 90° CCW
		dst = image.NewRGBA(image.Rect(0, 0, h, w))
		for y := 0; y < w; y++ {
			for x := 0; x < h; x++ {
				dst.Set(x, y, img.At(w-1-y+bounds.Min.X, x+bounds.Min.Y))
			}
		}
	default:
		return img
	}
	return dst
}

// readOrientationFromBytes reads the EXIF orientation tag from JPEG bytes.
// Returns 1 (normal) if the orientation cannot be read.
func readOrientationFromBytes(data []byte) int {
	x, err := exif.Decode(bytes.NewReader(data))
	if err != nil {
		return 1
	}

	tag, err := x.Get(exif.Orientation)
	if err != nil {
		return 1
	}

	v, err := tag.Int(0)
	if err != nil {
		return 1
	}

	return v
}

// readOrientation reads the EXIF orientation tag from a file path.
// Returns 1 (normal) if the orientation cannot be read.
func readOrientation(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 1
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		return 1
	}

	tag, err := x.Get(exif.Orientation)
	if err != nil {
		return 1
	}

	v, err := tag.Int(0)
	if err != nil {
		return 1
	}

	return v
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

// ClearCache removes all cached thumbnail files.
func (s *ThumbService) ClearCache() error {
	entries, err := os.ReadDir(s.cfg.CacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		_ = os.Remove(filepath.Join(s.cfg.CacheDir, entry.Name()))
	}
	return nil
}

// PreGenerateAll generates grid thumbnails for all assets that don't have them yet.
// Runs in the background; intended to be called after a scan completes.
func (s *ThumbService) PreGenerateAll(ctx context.Context) {
	ids, err := s.repo.FindAllIDs(ctx)
	if err != nil {
		return
	}

	// Process with bounded concurrency
	sem := make(chan struct{}, s.cfg.ConcurrencyLimit)
	for _, id := range ids {
		select {
		case <-ctx.Done():
			return
		default:
		}
		sem <- struct{}{}
		go func(assetID int64) {
			defer func() { <-sem }()
			// Only generate grid thumbs; full thumbs on demand
			_, _ = s.GetThumbPath(ctx, assetID, ThumbGrid, "")
		}(id)
	}
}

// pickScale returns the best DCT scale denominator for decoding.
// Chooses the largest scale where the output dimension still exceeds targetSize,
// avoiding full-resolution decode when we only need a small thumbnail.
func pickScale(path string, targetSize int) int {
	w, h, err := jpegdecoder.ReadDimensions(path)
	if err != nil {
		return 1 // fallback to full decode
	}
	longSide := w
	if h > w {
		longSide = h
	}
	// Libjpeg supports 1/1, 1/2, 1/4, 1/8
	for _, s := range []int{8, 4, 2} {
		if longSide/s >= targetSize {
			return s
		}
	}
	return 1
}
