package service

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"os"
	"path/filepath"
	"strings"
)

// ExtractEmbeddedJPEG extracts the embedded JPEG preview from a RAW file.
// Currently supports ARW (Sony) format; other formats return ErrUnsupportedRawFormat.
func ExtractEmbeddedJPEG(rawPath string) ([]byte, error) {
	ext := strings.ToLower(filepath.Ext(rawPath))

	switch ext {
	case ".arw":
		return extractARWPreview(rawPath)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedRawFormat, ext)
	}
}

// ErrUnsupportedRawFormat is returned when a RAW format is not yet supported.
var ErrUnsupportedRawFormat = fmt.Errorf("unsupported raw format")

// extractARWPreview extracts the embedded JPEG from a Sony ARW file.
// ARW files are TIFF-based and contain multiple JPEG previews (thumbnail + full-size).
// We search for JPEG SOI/EOI pairs and return the largest valid JPEG that decodes successfully.
func extractARWPreview(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read arw file: %w", err)
	}

	var bestJPEG []byte
	bestW, bestH := 0, 0

	for i := 8; i < len(data)-1; i++ {
		if data[i] != 0xFF || data[i+1] != 0xD8 {
			continue
		}
		soi := i

		// Find matching EOI after this SOI
		eoi := -1
		for j := soi + 2; j < len(data)-1; j++ {
			if data[j] == 0xFF && data[j+1] == 0xD9 {
				eoi = j + 2
				break
			}
		}
		if eoi < 0 {
			continue
		}

		jpegData := data[soi:eoi]
		cfg, format, err := image.DecodeConfig(bytes.NewReader(jpegData))
		if err != nil {
			continue // not a valid JPEG, try next SOI
		}
		_ = format

		// Prefer larger images (measure by total pixels)
		pixels := cfg.Width * cfg.Height
		if pixels > bestW*bestH {
			bestW, bestH = cfg.Width, cfg.Height
			bestJPEG = jpegData
		}
	}

	if len(bestJPEG) == 0 {
		return nil, fmt.Errorf("no JPEG preview found in ARW file")
	}

	return bestJPEG, nil
}
