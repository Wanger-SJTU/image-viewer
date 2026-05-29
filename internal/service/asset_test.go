package service

import (
	"context"
	"os"
	"testing"

	"image-viewer/internal/config"
	"image-viewer/internal/repository"
	"image-viewer/shared/types"
)

func setupAssetTest(t *testing.T) (*AssetService, repository.AssetRepository, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "asset_test")
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
	svc := NewAssetService(cfg, repo)

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return svc, repo, cleanup
}

func seedAsset(t *testing.T, repo repository.AssetRepository) int64 {
	t.Helper()

	assets := []*types.Asset{
		{Name: "DSC0001", DirPath: "/photos", MatchStatus: types.MatchStatusPaired},
	}
	files := []*types.MediaFile{
		{AssetID: 1, FilePath: "/photos/DSC0001.JPG", FileName: "DSC0001.JPG", MediaType: types.MediaTypeJPG},
	}
	if err := repo.BulkUpsert(context.Background(), assets, files); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return 1
}

func TestAssetService_GetByID(t *testing.T) {
	svc, repo, cleanup := setupAssetTest(t)
	defer cleanup()

	id := seedAsset(t, repo)

	asset, err := svc.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if asset.Name != "DSC0001" {
		t.Errorf("name = %q, want DSC0001", asset.Name)
	}
}

func TestAssetService_GetByID_NotFound(t *testing.T) {
	svc, _, cleanup := setupAssetTest(t)
	defer cleanup()

	asset, err := svc.GetByID(context.Background(), 999)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if asset != nil {
		t.Error("expected nil for non-existent asset")
	}
}

func TestAssetService_Rate(t *testing.T) {
	svc, repo, cleanup := setupAssetTest(t)
	defer cleanup()

	id := seedAsset(t, repo)

	// Valid rating
	if err := svc.Rate(context.Background(), id, 5); err != nil {
		t.Fatalf("Rate(5): %v", err)
	}

	asset, _ := svc.GetByID(context.Background(), id)
	if asset.Rating != 5 {
		t.Errorf("rating = %d, want 5", asset.Rating)
	}

	// Invalid rating
	if err := svc.Rate(context.Background(), id, 6); err == nil {
		t.Error("expected error for rating > 5")
	}
	if err := svc.Rate(context.Background(), id, -1); err == nil {
		t.Error("expected error for rating < 0")
	}
}

func TestAssetService_Label(t *testing.T) {
	svc, repo, cleanup := setupAssetTest(t)
	defer cleanup()

	id := seedAsset(t, repo)

	// Valid label
	if err := svc.Label(context.Background(), id, "blue"); err != nil {
		t.Fatalf("Label(blue): %v", err)
	}

	// Invalid label
	if err := svc.Label(context.Background(), id, "invalid"); err == nil {
		t.Error("expected error for invalid label")
	}
}

func TestAssetService_Delete(t *testing.T) {
	svc, repo, cleanup := setupAssetTest(t)
	defer cleanup()

	id := seedAsset(t, repo)

	if err := svc.Delete(context.Background(), id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	asset, _ := svc.GetByID(context.Background(), id)
	if asset != nil {
		t.Error("expected nil after delete")
	}
}
