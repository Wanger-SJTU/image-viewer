package service

import (
	"fmt"
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
// ARW files are TIFF-based; the preview is typically in the first IFD's JPEGInterchangeFormat tag.
// This is a simplified implementation that searches for JPEG markers in the raw file.
func extractARWPreview(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read arw file: %w", err)
	}

	// Search for JPEG SOI marker (0xFF 0xD8) and EOI marker (0xFF 0xD9)
	// Start from offset 8 to skip potential TIFF header
	soi := -1
	for i := 8; i < len(data)-1; i++ {
		if data[i] == 0xFF && data[i+1] == 0xD8 {
			soi = i
			break
		}
	}

	if soi == -1 {
		return nil, fmt.Errorf("no JPEG preview found in ARW file")
	}

	// Find EOI marker after SOI
	eoi := -1
	for i := soi + 2; i < len(data)-1; i++ {
		if data[i] == 0xFF && data[i+1] == 0xD9 {
			eoi = i + 2
			break
		}
	}

	if eoi == -1 {
		return nil, fmt.Errorf("incomplete JPEG preview in ARW file")
	}

	return data[soi:eoi], nil
}
