package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"image-viewer/shared/types"
)

func setupDB(t *testing.T) *assetRepo {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "viewer_test")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	db, err := InitDB(tmpDir + "/test.db")
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	return NewAssetRepository(db).(*assetRepo)
}

func TestBulkUpsertAndFindByID(t *testing.T) {
	repo := setupDB(t)
	ctx := context.Background()

	capturedAt := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)
	assets := []*types.Asset{
		{
			Name: "DSC0001", DirPath: "/photos/2024", MatchStatus: types.MatchStatusPaired,
			CapturedAt: &capturedAt,
		},
	}
	files := []*types.MediaFile{
		{
			AssetID: 1, FilePath: "/photos/2024/DSC0001.ARW", FileName: "DSC0001.ARW",
			FileSize: 50000000, MediaType: types.MediaTypeRAW,
			Exif: &types.ExifMeta{CameraModel: "Sony A7M4", Width: 7008, Height: 4672},
		},
	}

	err := repo.BulkUpsert(ctx, assets, files)
	if err != nil {
		t.Fatalf("BulkUpsert: %v", err)
	}

	asset, err := repo.FindByID(ctx, 1)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if asset == nil {
		t.Fatal("expected asset, got nil")
	}
	if asset.Name != "DSC0001" {
		t.Errorf("name = %q, want DSC0001", asset.Name)
	}
	if asset.MatchStatus != types.MatchStatusPaired {
		t.Errorf("match_status = %q, want paired", asset.MatchStatus)
	}
	if asset.RawFile == nil {
		t.Fatal("expected raw_file, got nil")
	}
	if asset.RawFile.Exif.CameraModel != "Sony A7M4" {
		t.Errorf("camera_model = %q, want Sony A7M4", asset.RawFile.Exif.CameraModel)
	}
}

func TestListAssets(t *testing.T) {
	repo := setupDB(t)
	ctx := context.Background()

	capturedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	assets := []*types.Asset{
		{Name: "DSC0001", DirPath: "/p", MatchStatus: types.MatchStatusPaired, CapturedAt: &capturedAt},
		{Name: "DSC0002", DirPath: "/p", MatchStatus: types.MatchStatusOrphan, CapturedAt: &capturedAt},
		{Name: "DSC0003", DirPath: "/p", MatchStatus: types.MatchStatusPaired, CapturedAt: &capturedAt},
	}
	files := []*types.MediaFile{
		{AssetID: 1, FilePath: "/p/DSC0001.ARW", FileName: "DSC0001.ARW", MediaType: types.MediaTypeRAW,
			Exif: &types.ExifMeta{CameraModel: "Sony A7M4", FocalLength: 50, ISO: 400}},
		{AssetID: 2, FilePath: "/p/DSC0002.JPG", FileName: "DSC0002.JPG", MediaType: types.MediaTypeJPG,
			Exif: &types.ExifMeta{CameraModel: "Canon R5", FocalLength: 85, ISO: 800}},
		{AssetID: 3, FilePath: "/p/DSC0003.ARW", FileName: "DSC0003.ARW", MediaType: types.MediaTypeRAW,
			Exif: &types.ExifMeta{CameraModel: "Sony A7M4", FocalLength: 35, ISO: 200}},
	}

	err := repo.BulkUpsert(ctx, assets, files)
	if err != nil {
		t.Fatalf("BulkUpsert: %v", err)
	}

	// List all
	list, total, err := repo.List(ctx, nil, 1, 50)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(list) != 3 {
		t.Errorf("len(list) = %d, want 3", len(list))
	}

	// Filter by match_status
	filter := &types.AssetFilter{MatchStatus: "paired"}
	list, total, err = repo.List(ctx, filter, 1, 50)
	if err != nil {
		t.Fatalf("List filtered: %v", err)
	}
	if total != 2 {
		t.Errorf("filtered total = %d, want 2", total)
	}

	// Pagination
	list, total, err = repo.List(ctx, nil, 1, 1)
	if err != nil {
		t.Fatalf("List page 1: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(list) != 1 {
		t.Errorf("len(page1) = %d, want 1", len(list))
	}
}

func TestUpdateRating(t *testing.T) {
	repo := setupDB(t)
	ctx := context.Background()

	assets := []*types.Asset{{Name: "DSC0001", DirPath: "/p", MatchStatus: types.MatchStatusOrphan}}
	files := []*types.MediaFile{{AssetID: 1, FilePath: "/p/DSC0001.JPG", FileName: "DSC0001.JPG", MediaType: types.MediaTypeJPG}}

	repo.BulkUpsert(ctx, assets, files)

	if err := repo.UpdateRating(ctx, 1, 5); err != nil {
		t.Fatalf("UpdateRating: %v", err)
	}

	asset, _ := repo.FindByID(ctx, 1)
	if asset.Rating != 5 {
		t.Errorf("rating = %d, want 5", asset.Rating)
	}
}

func TestUpdateColorLabel(t *testing.T) {
	repo := setupDB(t)
	ctx := context.Background()

	assets := []*types.Asset{{Name: "DSC0001", DirPath: "/p", MatchStatus: types.MatchStatusOrphan}}
	files := []*types.MediaFile{{AssetID: 1, FilePath: "/p/DSC0001.JPG", FileName: "DSC0001.JPG", MediaType: types.MediaTypeJPG}}

	repo.BulkUpsert(ctx, assets, files)

	if err := repo.UpdateColorLabel(ctx, 1, "red"); err != nil {
		t.Fatalf("UpdateColorLabel: %v", err)
	}

	asset, _ := repo.FindByID(ctx, 1)
	if asset.ColorLabel != "red" {
		t.Errorf("color_label = %q, want red", asset.ColorLabel)
	}
}

func TestDelete(t *testing.T) {
	repo := setupDB(t)
	ctx := context.Background()

	assets := []*types.Asset{{Name: "DSC0001", DirPath: "/p", MatchStatus: types.MatchStatusOrphan}}
	files := []*types.MediaFile{
		{AssetID: 1, FilePath: "/p/DSC0001.JPG", FileName: "DSC0001.JPG", MediaType: types.MediaTypeJPG},
		{AssetID: 1, FilePath: "/p/DSC0001.ARW", FileName: "DSC0001.ARW", MediaType: types.MediaTypeRAW},
	}

	repo.BulkUpsert(ctx, assets, files)

	paths, err := repo.Delete(ctx, 1)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if len(paths) != 2 {
		t.Errorf("len(paths) = %d, want 2", len(paths))
	}

	asset, _ := repo.FindByID(ctx, 1)
	if asset != nil {
		t.Error("expected nil after delete")
	}
}
