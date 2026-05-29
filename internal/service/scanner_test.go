package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"image-viewer/internal/config"
	"image-viewer/internal/repository"
)

func setupScannerTest(t *testing.T) (string, *ScannerService, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "scanner_test")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}

	// Create test photo directory
	photoDir := filepath.Join(tmpDir, "photos")
	os.MkdirAll(photoDir, 0755)

	// Create test files
	createFile := func(name string) {
		f, err := os.Create(filepath.Join(photoDir, name))
		if err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
		f.Close()
	}

	// Paired RAW+JPG
	createFile("DSC0001.ARW")
	createFile("DSC0001.JPG")
	// Orphan RAW
	createFile("DSC0002.ARW")
	// Orphan JPG
	createFile("DSC0003.JPG")
	// Paired with different case
	createFile("DSC0004.CR3")
	createFile("DSC0004.jpg")

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := repository.InitDB(dbPath)
	if err != nil {
		t.Fatalf("init db: %v", err)
	}

	cfg := &config.Config{
		CacheDir:         filepath.Join(tmpDir, "cache"),
		SupportedRawExts: []string{".arw", ".cr3", ".nef"},
		SupportedJpgExts: []string{".jpg", ".jpeg"},
		ConcurrencyLimit: 2,
	}
	os.MkdirAll(cfg.CacheDir, 0755)

	repo := repository.NewAssetRepository(db)
	svc := NewScannerService(cfg, repo)

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return photoDir, svc, cleanup
}

func TestScan_DualTrackMatching(t *testing.T) {
	photoDir, svc, cleanup := setupScannerTest(t)
	defer cleanup()

	ctx := context.Background()
	progressCh := make(chan ScanProgress, 100)

	err := svc.Scan(ctx, photoDir, progressCh)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	// Collect progress
	var lastProgress ScanProgress
	for p := range progressCh {
		lastProgress = p
	}

	if lastProgress.Phase != PhaseDone {
		t.Errorf("final phase = %q, want %q", lastProgress.Phase, PhaseDone)
	}
	if lastProgress.Found != 6 {
		t.Errorf("found = %d, want 6", lastProgress.Found)
	}
	if lastProgress.Matched != 2 {
		t.Errorf("matched = %d, want 2 (DSC0001 + DSC0004)", lastProgress.Matched)
	}
	if lastProgress.Orphans != 2 {
		t.Errorf("orphans = %d, want 2 (DSC0002 + DSC0003)", lastProgress.Orphans)
	}
}

func TestScan_EmptyDirectory(t *testing.T) {
	photoDir, svc, cleanup := setupScannerTest(t)
	defer cleanup()

	// Remove all files
	entries, _ := os.ReadDir(photoDir)
	for _, e := range entries {
		os.Remove(filepath.Join(photoDir, e.Name()))
	}

	ctx := context.Background()
	progressCh := make(chan ScanProgress, 100)

	err := svc.Scan(ctx, photoDir, progressCh)
	if err != nil {
		t.Fatalf("Scan empty dir: %v", err)
	}

	var lastProgress ScanProgress
	for p := range progressCh {
		lastProgress = p
	}

	if lastProgress.Found != 0 {
		t.Errorf("found = %d, want 0", lastProgress.Found)
	}
}

func TestScan_InvalidPath(t *testing.T) {
	_, svc, cleanup := setupScannerTest(t)
	defer cleanup()

	ctx := context.Background()
	progressCh := make(chan ScanProgress, 100)

	err := svc.Scan(ctx, "/nonexistent/path/12345", progressCh)
	if err == nil {
		t.Error("expected error for invalid path, got nil")
	}
}
