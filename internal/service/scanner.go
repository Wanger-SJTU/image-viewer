package service

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"image-viewer/internal/config"
	"image-viewer/shared/types"
)

// ScanProgress represents a progress update during a scan operation.
type ScanProgress struct {
	Phase     string `json:"phase"`
	Found     int    `json:"found"`
	Processed int    `json:"processed"`
	Matched   int    `json:"matched"`
	Orphans   int    `json:"orphans"`
	Error     string `json:"error,omitempty"`
}

const (
	PhaseScanning = "scanning"
	PhaseMatching = "matching"
	PhaseSaving   = "saving"
	PhaseDone     = "done"
	PhaseError    = "error"
)

// scannerRepo defines the subset of repository methods used by ScannerService.
type scannerRepo interface {
	BulkUpsert(ctx context.Context, assets []*types.Asset, files []*types.MediaFile) error
}

// ScannerService handles concurrent file scanning and dual-track matching.
type ScannerService struct {
	cfg  *config.Config
	repo scannerRepo
	mu   sync.Mutex
}

// NewScannerService creates a new ScannerService.
func NewScannerService(cfg *config.Config, repo scannerRepo) *ScannerService {
	return &ScannerService{cfg: cfg, repo: repo}
}

// Scan walks the given directory, finds image files, and performs dual-track matching.
// Progress updates are sent to progressCh. The channel is closed when scanning is complete.
func (s *ScannerService) Scan(ctx context.Context, rootPath string, progressCh chan<- ScanProgress) error {
	defer close(progressCh)

	// Validate path
	info, err := os.Stat(rootPath)
	if err != nil {
		return fmt.Errorf("invalid path %s: %w", rootPath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", rootPath)
	}

	// Step 1: Walk and collect files
	s.sendProgress(progressCh, PhaseScanning, 0, 0, 0, 0, "")

	rawExts := makeSet(s.cfg.SupportedRawExts)
	jpgExts := makeSet(s.cfg.SupportedJpgExts)

	type fileEntry struct {
		path string
		name string
	}

	var entries []fileEntry
	err = filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible files
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && path != rootPath {
				return fs.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if rawExts[ext] || jpgExts[ext] {
			entries = append(entries, fileEntry{path: path, name: d.Name()})
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk directory: %w", err)
	}

	s.sendProgress(progressCh, PhaseMatching, len(entries), 0, 0, 0, "")

	// Step 2: Match by composite key
	type matchKey struct {
		dir  string
		base string
	}
	groups := make(map[matchKey]*types.Asset)

	for _, e := range entries {
		ext := strings.ToLower(filepath.Ext(e.name))
		base := strings.ToLower(strings.TrimSuffix(e.name, filepath.Ext(e.name)))
		dir := strings.ToLower(filepath.Dir(e.path))

		key := matchKey{dir: dir, base: base}
		asset, exists := groups[key]
		if !exists {
			asset = &types.Asset{
				Name:        strings.TrimSuffix(e.name, filepath.Ext(e.name)),
				DirPath:     filepath.Dir(e.path),
				MatchStatus: types.MatchStatusOrphan,
			}
			groups[key] = asset
		}

		mediaType := types.MediaTypeJPG
		if rawExts[ext] {
			mediaType = types.MediaTypeRAW
		}

		file := &types.MediaFile{
			FilePath:  e.path,
			FileName:  e.name,
			MediaType: mediaType,
		}

		// Get file size
		if fi, err := os.Stat(e.path); err == nil {
			file.FileSize = fi.Size()
		}

		if mediaType == types.MediaTypeRAW {
			asset.RawFile = file
		} else {
			asset.JpgFile = file
		}
	}

	// Determine match status
	matched := 0
	orphans := 0
	var assets []*types.Asset

	for _, a := range groups {
		if a.RawFile != nil && a.JpgFile != nil {
			a.MatchStatus = types.MatchStatusPaired
			matched++
		} else {
			orphans++
		}
		assets = append(assets, a)
	}

	s.sendProgress(progressCh, PhaseSaving, len(entries), len(entries), matched, orphans, "")

	// Step 3: Batch insert — files are embedded in assets
	const batchSize = 500
	for i := 0; i < len(assets); i += batchSize {
		end := i + batchSize
		if end > len(assets) {
			end = len(assets)
		}
		if err := s.repo.BulkUpsert(ctx, assets[i:end], nil); err != nil {
			s.sendProgress(progressCh, PhaseError, len(entries), len(entries), matched, orphans, err.Error())
			return fmt.Errorf("bulk upsert: %w", err)
		}
	}

	s.sendProgress(progressCh, PhaseDone, len(entries), len(entries), matched, orphans, "")
	return nil
}

func (s *ScannerService) sendProgress(ch chan<- ScanProgress, phase string, found, processed, matched, orphans int, errMsg string) {
	select {
	case ch <- ScanProgress{
		Phase:     phase,
		Found:     found,
		Processed: processed,
		Matched:   matched,
		Orphans:   orphans,
		Error:     errMsg,
	}:
	default:
	}
}

func makeSet(items []string) map[string]bool {
	s := make(map[string]bool, len(items))
	for _, item := range items {
		s[item] = true
	}
	return s
}
