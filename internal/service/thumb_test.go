package service

import (
	"context"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"image-viewer/internal/config"
	"image-viewer/internal/repository"
	"image-viewer/shared/types"
)

func setupThumbTest(t *testing.T) (*ThumbService, *AssetService, repository.AssetRepository, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "thumb_test")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}

	dbPath := tmpDir + "/test.db"
	db, err := repository.InitDB(dbPath)
	if err != nil {
		t.Fatalf("init db: %v", err)
	}

	cfg := &config.Config{CacheDir: tmpDir + "/cache"}
	os.MkdirAll(cfg.CacheDir, 0755)

	repo := repository.NewAssetRepository(db)
	assetSvc := NewAssetService(cfg, repo)
	thumbSvc := NewThumbService(cfg, repo)

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return thumbSvc, assetSvc, repo, tmpDir, cleanup
}

// createTestImage creates a simple 100x100 JPEG file for testing.
func createTestImage(t *testing.T, path string) {
	t.Helper()
	os.MkdirAll(filepath.Dir(path), 0755)
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create image file: %v", err)
	}
	defer f.Close()
	if err := jpeg.Encode(f, img, nil); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}
}

func TestGenerateThumb_Grid(t *testing.T) {
	thumbSvc, _, repo, tmpDir, cleanup := setupThumbTest(t)
	defer cleanup()

	// Create a test image
	imgPath := filepath.Join(tmpDir, "photos/DSC0001.JPG")
	createTestImage(t, imgPath)

	// Seed asset
	assets := []*types.Asset{{
		Name: "DSC0001", DirPath: filepath.Join(tmpDir, "photos"),
		MatchStatus: types.MatchStatusOrphan,
	}}
	files := []*types.MediaFile{{
		AssetID: 1, FilePath: imgPath, FileName: "DSC0001.JPG",
		MediaType: types.MediaTypeJPG,
	}}
	repo.BulkUpsert(context.Background(), assets, files)

	asset, _ := repo.FindByID(context.Background(), 1)

	thumbPath, err := thumbSvc.GenerateThumb(context.Background(), asset, ThumbGrid)
	if err != nil {
		t.Fatalf("GenerateThumb(grid): %v", err)
	}
	if thumbPath == "" {
		t.Error("expected non-empty thumb path")
	}

	// Verify file exists
	if _, err := os.Stat(thumbPath); os.IsNotExist(err) {
		t.Errorf("thumb file does not exist: %s", thumbPath)
	}
}

func TestGenerateThumb_Full(t *testing.T) {
	thumbSvc, _, repo, tmpDir, cleanup := setupThumbTest(t)
	defer cleanup()

	imgPath := filepath.Join(tmpDir, "photos/DSC0001.JPG")
	createTestImage(t, imgPath)

	assets := []*types.Asset{{
		Name: "DSC0001", DirPath: filepath.Join(tmpDir, "photos"),
		MatchStatus: types.MatchStatusOrphan,
	}}
	files := []*types.MediaFile{{
		AssetID: 1, FilePath: imgPath, FileName: "DSC0001.JPG",
		MediaType: types.MediaTypeJPG,
	}}
	repo.BulkUpsert(context.Background(), assets, files)

	asset, _ := repo.FindByID(context.Background(), 1)

	thumbPath, err := thumbSvc.GenerateThumb(context.Background(), asset, ThumbFull)
	if err != nil {
		t.Fatalf("GenerateThumb(full): %v", err)
	}
	if _, err := os.Stat(thumbPath); os.IsNotExist(err) {
		t.Errorf("thumb file does not exist: %s", thumbPath)
	}
}

func TestGetThumbPath_CacheHit(t *testing.T) {
	thumbSvc, _, repo, tmpDir, cleanup := setupThumbTest(t)
	defer cleanup()

	imgPath := filepath.Join(tmpDir, "photos/DSC0001.JPG")
	createTestImage(t, imgPath)

	assets := []*types.Asset{{
		Name: "DSC0001", DirPath: filepath.Join(tmpDir, "photos"),
		MatchStatus: types.MatchStatusOrphan,
	}}
	files := []*types.MediaFile{{
		AssetID: 1, FilePath: imgPath, FileName: "DSC0001.JPG",
		MediaType: types.MediaTypeJPG,
	}}
	repo.BulkUpsert(context.Background(), assets, files)

	asset, _ := repo.FindByID(context.Background(), 1)

	// First call generates
	thumbPath, err := thumbSvc.GenerateThumb(context.Background(), asset, ThumbGrid)
	if err != nil {
		t.Fatalf("first GenerateThumb: %v", err)
	}

	// Second call should return cached path
	cachedPath, err := thumbSvc.GetThumbPath(context.Background(), asset.ID, ThumbGrid)
	if err != nil {
		t.Fatalf("GetThumbPath: %v", err)
	}
	if cachedPath != thumbPath {
		t.Errorf("cached path = %q, want %q", cachedPath, thumbPath)
	}
}

func TestGetThumbPath_InvalidSize(t *testing.T) {
	thumbSvc, _, _, _, cleanup := setupThumbTest(t)
	defer cleanup()

	_, err := thumbSvc.GetThumbPath(context.Background(), 1, "invalid")
	if err == nil {
		t.Error("expected error for invalid size")
	}
}
