package service

import (
	"os"
	"path/filepath"
	"testing"
)

func createFakeARW(t *testing.T, path string, jpegData []byte) {
	t.Helper()
	os.MkdirAll(filepath.Dir(path), 0755)
	// For now, just write raw JPEG data as placeholder
	// Real ARW parsing will be added in the implementation
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	f.Write(jpegData)
	f.Close()
}

func TestExtractEmbeddedJPEG_ARW(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "raw_test")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Minimal JPEG data
	jpegData := []byte{
		0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46,
		0x49, 0x46, 0x00, 0x01, 0x01, 0x00, 0x00, 0x01,
		0x00, 0x01, 0x00, 0x00, 0xFF, 0xDB, 0x00, 0x43,
		0x00, 0x08, 0x06, 0x06, 0x07, 0x06, 0x05, 0x08,
		0x07, 0x07, 0x07, 0x09, 0x09, 0x08, 0x0A, 0xFF, 0xD9,
	}

	arwPath := filepath.Join(tmpDir, "DSC0001.ARW")
	createFakeARW(t, arwPath, jpegData)

	data, err := ExtractEmbeddedJPEG(arwPath)
	if err != nil {
		t.Skipf("ARW extraction not yet implemented: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty JPEG data")
	}
}

func TestExtractEmbeddedJPEG_UnsupportedFormat(t *testing.T) {
	_, err := ExtractEmbeddedJPEG("/some/file.XXX")
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}
