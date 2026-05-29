package service

import (
	"context"
	"fmt"

	"image-viewer/internal/config"
	"image-viewer/shared/types"
)

// AssetFilter is re-exported for service layer use.
type AssetFilter = types.AssetFilter

// AssetService handles business logic for asset CRUD, rating, and labeling.
type AssetService struct {
	cfg  *config.Config
	repo interface {
		FindByID(ctx context.Context, id int64) (*types.Asset, error)
		List(ctx context.Context, filter *types.AssetFilter, page, limit int) ([]*types.Asset, int64, error)
		UpdateRating(ctx context.Context, id int64, rating int) error
		UpdateColorLabel(ctx context.Context, id int64, label string) error
		UpdateThumbnails(ctx context.Context, id int64, gridThumb, fullThumb string) error
		Delete(ctx context.Context, id int64) ([]string, error)
	}
}

// NewAssetService creates a new AssetService.
func NewAssetService(cfg *config.Config, repo interface {
	FindByID(ctx context.Context, id int64) (*types.Asset, error)
	List(ctx context.Context, filter *types.AssetFilter, page, limit int) ([]*types.Asset, int64, error)
	UpdateRating(ctx context.Context, id int64, rating int) error
	UpdateColorLabel(ctx context.Context, id int64, label string) error
	UpdateThumbnails(ctx context.Context, id int64, gridThumb, fullThumb string) error
	Delete(ctx context.Context, id int64) ([]string, error)
}) *AssetService {
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

// Delete removes an asset and returns the file paths for cleanup.
func (s *AssetService) Delete(ctx context.Context, id int64) error {
	_, err := s.repo.Delete(ctx, id)
	return err
}
