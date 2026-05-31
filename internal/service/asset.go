package service

import (
	"context"
	"fmt"
	"os"

	"image-viewer/internal/config"
	"image-viewer/shared/types"
)

// AssetFilter is re-exported for service layer use.
type AssetFilter = types.AssetFilter

type assetRepo interface {
	FindByID(ctx context.Context, id int64) (*types.Asset, error)
	List(ctx context.Context, filter *types.AssetFilter, page, limit int) ([]*types.Asset, int64, error)
	UpdateRating(ctx context.Context, id int64, rating int) error
	UpdateColorLabel(ctx context.Context, id int64, label string) error
	UpdateThumbnails(ctx context.Context, id int64, gridThumb, fullThumb string) error
	Delete(ctx context.Context, id int64) ([]string, error)
	DeleteAll(ctx context.Context) (int64, error)
	SoftDelete(ctx context.Context, id int64) error
	Restore(ctx context.Context, id int64) error
	Purge(ctx context.Context, id int64, fileType string) ([]string, error)
	GetFilterOptions(ctx context.Context) (*types.FilterOptions, error)
}

// AssetService handles business logic for asset CRUD, rating, and labeling.
type AssetService struct {
	cfg  *config.Config
	repo assetRepo
}

// NewAssetService creates a new AssetService.
func NewAssetService(cfg *config.Config, repo assetRepo) *AssetService {
	return &AssetService{cfg: cfg, repo: repo}
}

// GetByID returns an asset by its ID, or nil if not found.
func (s *AssetService) GetByID(ctx context.Context, id int64) (*types.Asset, error) {
	return s.repo.FindByID(ctx, id)
}

// List returns paginated assets with optional filtering.
func (s *AssetService) List(ctx context.Context, filter *AssetFilter, page, limit int) ([]*types.Asset, int64, error) {
	return s.repo.List(ctx, filter, page, limit)
}

// Rate sets the rating for an asset (valid: 0-5).
func (s *AssetService) Rate(ctx context.Context, id int64, rating int) error {
	if rating < 0 || rating > 5 {
		return fmt.Errorf("rating must be 0-5, got %d", rating)
	}
	return s.repo.UpdateRating(ctx, id, rating)
}

// Label sets the color label for an asset.
func (s *AssetService) Label(ctx context.Context, id int64, label string) error {
	if !types.ValidColorLabels[label] {
		return fmt.Errorf("invalid color label: %q", label)
	}
	return s.repo.UpdateColorLabel(ctx, id, label)
}

// Delete soft-deletes an asset (moves to trash). Use Purge for permanent deletion.
func (s *AssetService) Delete(ctx context.Context, id int64) error {
	return s.repo.SoftDelete(ctx, id)
}

// Trash is an alias for Delete (moves asset to trash).
func (s *AssetService) Trash(ctx context.Context, id int64) error {
	return s.repo.SoftDelete(ctx, id)
}

// Restore moves a trashed asset back to active.
func (s *AssetService) Restore(ctx context.Context, id int64) error {
	return s.repo.Restore(ctx, id)
}

// Purge permanently deletes an asset or specific file types, returning file paths for cleanup.
func (s *AssetService) Purge(ctx context.Context, id int64, fileType string) ([]string, error) {
	return s.repo.Purge(ctx, id, fileType)
}

// PurgeWithCleanup permanently deletes and removes physical files from disk.
func (s *AssetService) PurgeWithCleanup(ctx context.Context, id int64, fileType string) error {
	paths, err := s.repo.Purge(ctx, id, fileType)
	if err != nil {
		return err
	}
	for _, p := range paths {
		os.Remove(p)
	}
	return nil
}

// ClearAll removes all assets from the database.
func (s *AssetService) ClearAll(ctx context.Context) (int64, error) {
	return s.repo.DeleteAll(ctx)
}

// GetFilterOptions returns distinct filter values for dropdown selectors.
func (s *AssetService) GetFilterOptions(ctx context.Context) (*types.FilterOptions, error) {
	return s.repo.GetFilterOptions(ctx)
}
