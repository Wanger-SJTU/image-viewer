//go:build nolibjpeg

package jpegdecoder

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
)

// ReadDimensions reads the image dimensions from a JPEG header without decompressing.
func ReadDimensions(path string) (int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	cfg, err := jpeg.DecodeConfig(f)
	if err != nil {
		return 0, 0, fmt.Errorf("jpeg read header: %w", err)
	}
	return cfg.Width, cfg.Height, nil
}

// DecodeFile decodes a JPEG file at full resolution.
func DecodeFile(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, err := jpeg.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("jpeg decode: %w", err)
	}
	return img, nil
}

// DecodeFileScaled decodes a JPEG at reduced resolution.
// The pure Go path always decodes at full res; scale is only used to log intent.
func DecodeFileScaled(path string, scale int) (image.Image, error) {
	return DecodeFile(path)
}
