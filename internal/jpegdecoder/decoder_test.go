package jpegdecoder

import (
	"testing"
)

func TestDecodeFile(t *testing.T) {
	img, err := DecodeFile("/home/wanger/Public/sort/jpg/2025/2025-09-21/DSC00016.JPG")
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	b := img.Bounds()
	if b.Dx() != 7008 || b.Dy() != 3944 {
		t.Errorf("expected 7008x3944, got %dx%d", b.Dx(), b.Dy())
	}
}
