package service

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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
	PhaseExif     = "exif"
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

	// Step 2: Match by composite key (dir + basename)
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

	// Determine match status — first pass by dir+basename
	matched := 0
	orphans := 0
	var assets []*types.Asset
	var raws, jpgs []*types.Asset

	for _, a := range groups {
		if a.RawFile != nil && a.JpgFile != nil {
			a.MatchStatus = types.MatchStatusPaired
			matched++
		} else if a.RawFile != nil {
			raws = append(raws, a)
			orphans++
		} else if a.JpgFile != nil {
			jpgs = append(jpgs, a)
			orphans++
		}
		assets = append(assets, a)
	}

	// Second pass: match orphans by capture time (also extracts EXIF for these files)
	if len(raws) > 0 && len(jpgs) > 0 {
		rawByTime := make(map[string]*types.Asset)
		for _, raw := range raws {
			exifData, err := extractExif(raw.RawFile.FilePath)
			if err != nil {
				continue
			}
			raw.RawFile.Exif = exifData
			key := exifData.CapturedAt.UTC().Round(time.Second).Format(time.RFC3339)
			rawByTime[key] = raw
		}

		timeMatched := make(map[*types.Asset]bool)
		for _, jpg := range jpgs {
			exifData, err := extractExif(jpg.JpgFile.FilePath)
			if err != nil {
				continue
			}
			jpg.JpgFile.Exif = exifData
			key := exifData.CapturedAt.UTC().Round(time.Second).Format(time.RFC3339)
			if raw, ok := rawByTime[key]; ok {
				raw.JpgFile = jpg.JpgFile
				raw.MatchStatus = types.MatchStatusPaired
				timeMatched[jpg] = true
				matched++
				orphans -= 2 // one raw + one jpg became a pair
			}
		}

		// Remove matched JPG orphans from assets
		if len(timeMatched) > 0 {
			filtered := make([]*types.Asset, 0, len(assets))
			for _, a := range assets {
				if timeMatched[a] {
					continue
				}
				filtered = append(filtered, a)
			}
			assets = filtered
		}
	}

	s.sendProgress(progressCh, PhaseExif, len(entries), 0, matched, orphans, "")

	// Step 3: Extract EXIF for remaining files that don't have it yet
	var wg sync.WaitGroup
	sem := make(chan struct{}, s.cfg.ConcurrencyLimit)
	exifCount := 0
	var exifMu sync.Mutex
	for _, a := range assets {
		for _, f := range []*types.MediaFile{a.RawFile, a.JpgFile} {
			if f == nil || f.Exif != nil {
				continue // already extracted during time-matching
			}
			wg.Add(1)
			sem <- struct{}{}
			go func(mf *types.MediaFile) {
				defer func() {
					<-sem
					wg.Done()
				}()
				if exifData, err := extractExif(mf.FilePath); err == nil {
					mf.Exif = exifData
					exifMu.Lock()
					exifCount++
					exifMu.Unlock()
				}
			}(f)
		}
	}
	wg.Wait()

	// Propagate capture time from media files to asset
	for _, a := range assets {
		for _, f := range []*types.MediaFile{a.RawFile, a.JpgFile} {
			if f != nil && f.Exif != nil && !f.Exif.CapturedAt.IsZero() {
				if a.CapturedAt == nil || f.Exif.CapturedAt.Before(*a.CapturedAt) {
					t := f.Exif.CapturedAt
					a.CapturedAt = &t
				}
			}
		}
	}

	s.sendProgress(progressCh, PhaseSaving, len(entries), exifCount, matched, orphans, "")

	// Step 4: Batch insert — files are embedded in assets
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
