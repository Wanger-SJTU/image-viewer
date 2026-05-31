package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"image-viewer/shared/types"
)

// AssetRepository defines the data access interface for assets and media files.
type AssetRepository interface {
	BulkUpsert(ctx context.Context, assets []*types.Asset, files []*types.MediaFile) error
	FindByID(ctx context.Context, id int64) (*types.Asset, error)
	List(ctx context.Context, filter *types.AssetFilter, page, limit int) ([]*types.Asset, int64, error)
	UpdateRating(ctx context.Context, id int64, rating int) error
	UpdateColorLabel(ctx context.Context, id int64, label string) error
	UpdateThumbnails(ctx context.Context, id int64, gridThumb, fullThumb string) error
	Delete(ctx context.Context, id int64) ([]string, error) // returns file paths to clean up
	DeleteAll(ctx context.Context) (int64, error)           // returns count of deleted assets
	SoftDelete(ctx context.Context, id int64) error
	Restore(ctx context.Context, id int64) error
	Purge(ctx context.Context, id int64, fileType string) ([]string, error) // returns file paths to clean up
	ExistsByDirName(ctx context.Context, dirPath, name string) (bool, int64, error)
	FindAllIDs(ctx context.Context) ([]int64, error)
	GetFilterOptions(ctx context.Context) (*types.FilterOptions, error)
}

type assetRepo struct {
	db *sql.DB
}

// NewAssetRepository creates a new AssetRepository backed by SQLite.
func NewAssetRepository(db *sql.DB) AssetRepository {
	return &assetRepo{db: db}
}

// BulkUpsert inserts or updates assets and their embedded media files in a single transaction.
func (r *assetRepo) BulkUpsert(ctx context.Context, assets []*types.Asset, files []*types.MediaFile) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	assetStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO assets (name, dir_path, match_status, captured_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(dir_path, name) DO UPDATE SET
			match_status = excluded.match_status,
			captured_at = COALESCE(excluded.captured_at, assets.captured_at),
			updated_at = excluded.updated_at,
			deleted_at = NULL
	`)
	if err != nil {
		return fmt.Errorf("prepare asset stmt: %w", err)
	}
	defer assetStmt.Close()

	fileStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO media_files (asset_id, file_path, file_name, file_size, media_type,
			camera_model, lens_model, focal_length, aperture, shutter_speed, iso,
			captured_at, width, height, orientation)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(file_path) DO UPDATE SET
			asset_id = excluded.asset_id,
			file_size = excluded.file_size,
			media_type = excluded.media_type,
			camera_model = excluded.camera_model,
			lens_model = excluded.lens_model,
			focal_length = excluded.focal_length,
			aperture = excluded.aperture,
			shutter_speed = excluded.shutter_speed,
			iso = excluded.iso,
			captured_at = excluded.captured_at,
			width = excluded.width,
			height = excluded.height,
			orientation = excluded.orientation
	`)
	if err != nil {
		return fmt.Errorf("prepare file stmt: %w", err)
	}
	defer fileStmt.Close()

	now := time.Now()
	for _, a := range assets {
		var capturedAt interface{}
		if a.CapturedAt != nil {
			capturedAt = a.CapturedAt.Format(time.RFC3339)
		}
		result, err := assetStmt.ExecContext(ctx, a.Name, a.DirPath, a.MatchStatus, capturedAt, now, now)
		if err != nil {
			return fmt.Errorf("insert asset: %w", err)
		}
		if a.ID == 0 {
			a.ID, _ = result.LastInsertId()
		}

		// Insert the asset's associated media files with correct AssetID
		for _, f := range []*types.MediaFile{a.RawFile, a.JpgFile} {
			if f == nil {
				continue
			}
			f.AssetID = a.ID
			if err := r.insertMediaFile(ctx, fileStmt, f); err != nil {
				return err
			}
		}
	}

	// Also insert any standalone files (from BulkUpsert calls that pass files separately)
	for _, f := range files {
		if f.AssetID == 0 {
			continue // skip files already handled above
		}
		if err := r.insertMediaFile(ctx, fileStmt, f); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *assetRepo) insertMediaFile(ctx context.Context, stmt *sql.Stmt, f *types.MediaFile) error {
	var capturedAt interface{}
	if f.Exif != nil && !f.Exif.CapturedAt.IsZero() {
		capturedAt = f.Exif.CapturedAt.Format(time.RFC3339)
	}
	cam, lens, fl, ap, ss, iso, w, h, ori := "", "", 0.0, 0.0, "", 0, 0, 0, 1
	if f.Exif != nil {
		cam = f.Exif.CameraModel
		lens = f.Exif.LensModel
		fl = f.Exif.FocalLength
		ap = f.Exif.Aperture
		ss = f.Exif.ShutterSpeed
		iso = f.Exif.ISO
		w = f.Exif.Width
		h = f.Exif.Height
		ori = f.Exif.Orientation
		if ori == 0 {
			ori = 1
		}
	}
	_, err := stmt.ExecContext(ctx, f.AssetID, f.FilePath, f.FileName, f.FileSize, f.MediaType,
		cam, lens, fl, ap, ss, iso, capturedAt, w, h, ori)
	if err != nil {
		return fmt.Errorf("insert media_file %s: %w", f.FilePath, err)
	}
	return nil
}

// FindByID returns a single asset with its media files populated.
func (r *assetRepo) FindByID(ctx context.Context, id int64) (*types.Asset, error) {
	asset := &types.Asset{}
	var capturedAt sql.NullString
	var deletedAt sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, dir_path, match_status, rating, color_label, ai_status,
			grid_thumb, full_thumb, captured_at, deleted_at, created_at, updated_at
		FROM assets WHERE id = ?`, id).Scan(
		&asset.ID, &asset.Name, &asset.DirPath, &asset.MatchStatus,
		&asset.Rating, &asset.ColorLabel, &asset.AiStatus,
		&asset.GridThumb, &asset.FullThumb, &capturedAt, &deletedAt,
		&asset.CreatedAt, &asset.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query asset: %w", err)
	}
	if capturedAt.Valid {
		t, _ := time.Parse(time.RFC3339, capturedAt.String)
		asset.CapturedAt = &t
	}
	if deletedAt.Valid {
		t, _ := time.Parse(time.RFC3339, deletedAt.String)
		asset.DeletedAt = &t
	}

	files, err := r.findMediaFiles(ctx, id)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if f.MediaType == types.MediaTypeRAW {
			asset.RawFile = f
		} else {
			asset.JpgFile = f
		}
	}
	return asset, nil
}

func (r *assetRepo) findMediaFiles(ctx context.Context, assetID int64) ([]*types.MediaFile, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, asset_id, file_path, file_name, file_size, media_type,
			camera_model, lens_model, focal_length, aperture, shutter_speed, iso,
			captured_at, width, height, orientation, created_at
		FROM media_files WHERE asset_id = ?`, assetID)
	if err != nil {
		return nil, fmt.Errorf("query media_files: %w", err)
	}
	defer rows.Close()

	files := make([]*types.MediaFile, 0)
	for rows.Next() {
		f := &types.MediaFile{Exif: &types.ExifMeta{}}
		var capturedAt sql.NullString
		err := rows.Scan(&f.ID, &f.AssetID, &f.FilePath, &f.FileName, &f.FileSize, &f.MediaType,
			&f.Exif.CameraModel, &f.Exif.LensModel, &f.Exif.FocalLength, &f.Exif.Aperture,
			&f.Exif.ShutterSpeed, &f.Exif.ISO, &capturedAt,
			&f.Exif.Width, &f.Exif.Height, &f.Exif.Orientation, &f.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan media_file: %w", err)
		}
		if capturedAt.Valid {
			f.Exif.CapturedAt, _ = time.Parse(time.RFC3339, capturedAt.String)
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

// List returns paginated assets with optional filtering.
func (r *assetRepo) List(ctx context.Context, filter *types.AssetFilter, page, limit int) ([]*types.Asset, int64, error) {
	req := &types.PaginatedRequest{Page: page, Limit: limit}
	page = req.DefaultPage()
	limit = req.DefaultLimit()

	var conditions []string
	var args []interface{}

	if filter != nil {
		if filter.Rating > 0 {
			conditions = append(conditions, "a.rating >= ?")
			args = append(args, filter.Rating)
		}
		if filter.ColorLabel != "" {
			conditions = append(conditions, "a.color_label = ?")
			args = append(args, filter.ColorLabel)
		}
		if filter.MatchStatus != "" {
			conditions = append(conditions, "a.match_status = ?")
			args = append(args, filter.MatchStatus)
		}
		if filter.FileType != "" {
			switch filter.FileType {
			case "jpg":
				conditions = append(conditions, "a.id IN (SELECT asset_id FROM media_files WHERE media_type = 'jpg') AND a.id NOT IN (SELECT asset_id FROM media_files WHERE media_type = 'raw')")
			case "raw":
				conditions = append(conditions, "a.id IN (SELECT asset_id FROM media_files WHERE media_type = 'raw') AND a.id NOT IN (SELECT asset_id FROM media_files WHERE media_type = 'jpg')")
			case "both":
				conditions = append(conditions, "a.match_status = 'paired'")
			}
		}
		if filter.CameraModel != "" {
			conditions = append(conditions, "a.id IN (SELECT asset_id FROM media_files WHERE camera_model LIKE ?)")
			args = append(args, "%"+filter.CameraModel+"%")
		}
		if filter.FocalLengthMin > 0 {
			conditions = append(conditions, "a.id IN (SELECT asset_id FROM media_files WHERE focal_length >= ?)")
			args = append(args, filter.FocalLengthMin)
		}
		if filter.FocalLengthMax > 0 {
			conditions = append(conditions, "a.id IN (SELECT asset_id FROM media_files WHERE focal_length <= ?)")
			args = append(args, filter.FocalLengthMax)
		}
		if filter.ApertureMin > 0 {
			conditions = append(conditions, "a.id IN (SELECT asset_id FROM media_files WHERE aperture >= ?)")
			args = append(args, filter.ApertureMin)
		}
		if filter.ApertureMax > 0 {
			conditions = append(conditions, "a.id IN (SELECT asset_id FROM media_files WHERE aperture <= ?)")
			args = append(args, filter.ApertureMax)
		}
		if filter.ISOMin > 0 {
			conditions = append(conditions, "a.id IN (SELECT asset_id FROM media_files WHERE iso >= ?)")
			args = append(args, filter.ISOMin)
		}
		if filter.ISOMax > 0 {
			conditions = append(conditions, "a.id IN (SELECT asset_id FROM media_files WHERE iso <= ?)")
			args = append(args, filter.ISOMax)
		}
		if !filter.CapturedAfter.IsZero() {
			conditions = append(conditions, "a.captured_at >= ?")
			args = append(args, filter.CapturedAfter.Format(time.RFC3339))
		}
		if !filter.CapturedBefore.IsZero() {
			conditions = append(conditions, "a.captured_at <= ?")
			args = append(args, filter.CapturedBefore.Format(time.RFC3339))
		}
		if filter.Search != "" {
			conditions = append(conditions, "a.name LIKE ?")
			args = append(args, "%"+filter.Search+"%")
		}
		if filter.Trashed != nil {
			if *filter.Trashed {
				conditions = append(conditions, "a.deleted_at IS NOT NULL")
			} else {
				conditions = append(conditions, "a.deleted_at IS NULL")
			}
		}
	}

	// By default, exclude trashed assets
	if filter == nil || filter.Trashed == nil {
		conditions = append(conditions, "a.deleted_at IS NULL")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM assets a %s", whereClause)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count assets: %w", err)
	}

	// Query page
	offset := (page - 1) * limit
	query := fmt.Sprintf(`
		SELECT a.id, a.name, a.dir_path, a.match_status, a.rating, a.color_label,
			a.ai_status, a.grid_thumb, a.full_thumb, a.captured_at, a.deleted_at, a.created_at, a.updated_at,
			(SELECT COUNT(*) FROM media_files m WHERE m.asset_id = a.id AND m.media_type = 'raw') AS has_raw,
			(SELECT COUNT(*) FROM media_files m WHERE m.asset_id = a.id AND m.media_type = 'jpg') AS has_jpg
		FROM assets a %s ORDER BY a.captured_at ASC, a.name ASC LIMIT ? OFFSET ?`, whereClause)
	queryArgs := append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list assets: %w", err)
	}
	defer rows.Close()

	assets := make([]*types.Asset, 0)
	for rows.Next() {
		a := &types.Asset{}
		var capturedAt sql.NullString
		var deletedAt sql.NullString
		var hasRaw, hasJpg int
		err := rows.Scan(&a.ID, &a.Name, &a.DirPath, &a.MatchStatus, &a.Rating, &a.ColorLabel,
			&a.AiStatus, &a.GridThumb, &a.FullThumb, &capturedAt, &deletedAt, &a.CreatedAt, &a.UpdatedAt,
			&hasRaw, &hasJpg)
		if err != nil {
			return nil, 0, fmt.Errorf("scan asset: %w", err)
		}
		if capturedAt.Valid {
			t, _ := time.Parse(time.RFC3339, capturedAt.String)
			a.CapturedAt = &t
		}
		if deletedAt.Valid {
			t, _ := time.Parse(time.RFC3339, deletedAt.String)
			a.DeletedAt = &t
		}
		if hasRaw > 0 {
			a.RawFile = &types.MediaFile{}
		}
		if hasJpg > 0 {
			a.JpgFile = &types.MediaFile{}
		}
		assets = append(assets, a)
	}

	return assets, total, rows.Err()
}

// UpdateRating sets the rating for an asset.
func (r *assetRepo) UpdateRating(ctx context.Context, id int64, rating int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE assets SET rating = ?, updated_at = ? WHERE id = ?`,
		rating, time.Now(), id)
	return err
}

// UpdateColorLabel sets the color label for an asset.
func (r *assetRepo) UpdateColorLabel(ctx context.Context, id int64, label string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE assets SET color_label = ?, updated_at = ? WHERE id = ?`,
		label, time.Now(), id)
	return err
}

// UpdateThumbnails sets the thumbnail paths for an asset.
func (r *assetRepo) UpdateThumbnails(ctx context.Context, id int64, gridThumb, fullThumb string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE assets SET grid_thumb = ?, full_thumb = ?, updated_at = ? WHERE id = ?`,
		gridThumb, fullThumb, time.Now(), id)
	return err
}

// Delete removes an asset and its media files, returning the file paths for cleanup.
func (r *assetRepo) Delete(ctx context.Context, id int64) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT file_path FROM media_files WHERE asset_id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("query file paths: %w", err)
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		paths = append(paths, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// CASCADE will delete media_files rows
	_, err = r.db.ExecContext(ctx, `DELETE FROM assets WHERE id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("delete asset: %w", err)
	}

	return paths, nil
}

func (r *assetRepo) DeleteAll(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM assets`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count assets: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `DELETE FROM assets`)
	if err != nil {
		return 0, fmt.Errorf("delete all: %w", err)
	}
	return count, nil
}

// SoftDelete marks an asset as trashed by setting deleted_at.
func (r *assetRepo) SoftDelete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `UPDATE assets SET deleted_at = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL`,
		time.Now(), time.Now(), id)
	if err != nil {
		return fmt.Errorf("soft delete asset %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("asset %d not found or already trashed", id)
	}
	return nil
}

// Restore clears the deleted_at timestamp on an asset.
func (r *assetRepo) Restore(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `UPDATE assets SET deleted_at = NULL, updated_at = ? WHERE id = ? AND deleted_at IS NOT NULL`,
		time.Now(), id)
	if err != nil {
		return fmt.Errorf("restore asset %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("asset %d not found or not trashed", id)
	}
	return nil
}

// Purge permanently deletes specific media files or the entire asset.
// fileType: "both" deletes the whole asset, "jpg" or "raw" deletes only that file.
// Returns the file paths to clean up from disk.
func (r *assetRepo) Purge(ctx context.Context, id int64, fileType string) ([]string, error) {
	switch fileType {
	case "both", "":
		// Full permanent delete — CASCADE removes media_files
		return r.Delete(ctx, id)
	case "jpg", "raw":
		// Delete only the specified media file type
		rows, err := r.db.QueryContext(ctx, `SELECT file_path FROM media_files WHERE asset_id = ? AND media_type = ?`, id, fileType)
		if err != nil {
			return nil, fmt.Errorf("query file paths: %w", err)
		}
		defer rows.Close()

		var paths []string
		for rows.Next() {
			var p string
			if err := rows.Scan(&p); err != nil {
				return nil, err
			}
			paths = append(paths, p)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}

		_, err = r.db.ExecContext(ctx, `DELETE FROM media_files WHERE asset_id = ? AND media_type = ?`, id, fileType)
		if err != nil {
			return nil, fmt.Errorf("delete media_files: %w", err)
		}

		// Update match_status: if only one file type remains, asset becomes orphan
		var remainingCount int
		if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM media_files WHERE asset_id = ?`, id).Scan(&remainingCount); err != nil {
			return nil, fmt.Errorf("check remaining files: %w", err)
		}
		if remainingCount == 0 {
			// No files left — delete the asset row
			_, err = r.db.ExecContext(ctx, `DELETE FROM assets WHERE id = ?`, id)
			if err != nil {
				return nil, fmt.Errorf("delete empty asset: %w", err)
			}
		} else {
			_, err = r.db.ExecContext(ctx, `UPDATE assets SET match_status = 'orphan', updated_at = ? WHERE id = ?`, time.Now(), id)
			if err != nil {
				return nil, fmt.Errorf("update match_status: %w", err)
			}
		}
		return paths, nil
	default:
		return nil, fmt.Errorf("invalid file_type: %s (must be both, jpg, or raw)", fileType)
	}
}

// ExistsByDirName checks if an asset already exists for the given directory and base name.
func (r *assetRepo) ExistsByDirName(ctx context.Context, dirPath, name string) (bool, int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `SELECT id FROM assets WHERE dir_path = ? AND name = ?`, dirPath, name).Scan(&id)
	if err == sql.ErrNoRows {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, err
	}
	return true, id, nil
}

func (r *assetRepo) FindAllIDs(ctx context.Context) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id FROM assets WHERE deleted_at IS NULL ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// GetFilterOptions returns distinct values for filter dropdowns.
func (r *assetRepo) GetFilterOptions(ctx context.Context) (*types.FilterOptions, error) {
	opts := &types.FilterOptions{
		CameraModels: []string{},
		FocalLengths: []float64{},
		Apertures:    []float64{},
		ISOs:         []int{},
		ColorLabels:  []string{"red", "orange", "yellow", "green", "blue", "purple"},
		FileTypes:    []string{"jpg", "raw", "both"},
		PhotoDates:   []string{},
	}

	// Camera models
	rows, err := r.db.QueryContext(ctx, `SELECT DISTINCT camera_model FROM media_files WHERE camera_model != '' ORDER BY camera_model`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var s string
			if rows.Scan(&s) == nil {
				opts.CameraModels = append(opts.CameraModels, s)
			}
		}
	}

	// Focal lengths (distinct, sorted)
	rows2, err := r.db.QueryContext(ctx, `SELECT DISTINCT focal_length FROM media_files WHERE focal_length > 0 ORDER BY focal_length`)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var v float64
			if rows2.Scan(&v) == nil {
				opts.FocalLengths = append(opts.FocalLengths, v)
			}
		}
	}

	// Apertures (distinct, sorted)
	rows3, err := r.db.QueryContext(ctx, `SELECT DISTINCT aperture FROM media_files WHERE aperture > 0 ORDER BY aperture`)
	if err == nil {
		defer rows3.Close()
		for rows3.Next() {
			var v float64
			if rows3.Scan(&v) == nil {
				opts.Apertures = append(opts.Apertures, v)
			}
		}
	}

	// ISOs (distinct, sorted)
	rows4, err := r.db.QueryContext(ctx, `SELECT DISTINCT iso FROM media_files WHERE iso > 0 ORDER BY iso`)
	if err == nil {
		defer rows4.Close()
		for rows4.Next() {
			var v int
			if rows4.Scan(&v) == nil {
				opts.ISOs = append(opts.ISOs, v)
			}
		}
	}

	// Photo dates (distinct dates, sorted)
	rows5, err := r.db.QueryContext(ctx, `SELECT DISTINCT date(captured_at) FROM assets WHERE captured_at IS NOT NULL AND deleted_at IS NULL ORDER BY date(captured_at)`)
	if err == nil {
		defer rows5.Close()
		for rows5.Next() {
			var s string
			if rows5.Scan(&s) == nil {
				opts.PhotoDates = append(opts.PhotoDates, s)
			}
		}
	}
	return opts, nil
}

// ensure interface satisfaction
var _ AssetRepository = (*assetRepo)(nil)
