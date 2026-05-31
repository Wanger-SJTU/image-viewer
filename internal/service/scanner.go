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

	// Step 2: Match by base filename (case-insensitive, cross-directory)
	groups := make(map[string]*types.Asset)

	for _, e := range entries {
		ext := strings.ToLower(filepath.Ext(e.name))
		base := strings.ToLower(strings.TrimSuffix(e.name, filepath.Ext(e.name)))
		dirPath := filepath.Dir(e.path)

		asset, exists := groups[base]
		if !exists {
			asset = &types.Asset{
				Name:        strings.TrimSuffix(e.name, filepath.Ext(e.name)),
				DirPath:     dirPath,
				MatchStatus: types.MatchStatusOrphan,
			}
			groups[base] = asset
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
			if asset.RawFile == nil {
				asset.RawFile = file
			}
		} else {
			if asset.JpgFile == nil {
				asset.JpgFile = file
			}
		}
	}

	// Step 3: Determine match by filename + capture date
	matched := 0
	orphans := 0
	var assets []*types.Asset
	var raws, jpgs []*types.Asset

	for _, a := range groups {
		if a.RawFile != nil && a.JpgFile != nil {
			// Verify capture dates match (within 1 second) before pairing
			rawTime, rawErr := extractCaptureTime(a.RawFile.FilePath)
			jpgTime, jpgErr := extractCaptureTime(a.JpgFile.FilePath)
			if rawErr == nil && jpgErr == nil {
				diff := rawTime.Sub(jpgTime)
				if diff < 0 {
					diff = -diff
				}
				if diff <= time.Second {
					a.MatchStatus = types.MatchStatusPaired
					matched++
					assets = append(assets, a)
					continue
				}
			}
			// Dates don't match or EXIF missing — split into separate orphans
			if a.RawFile != nil {
				orphan := &types.Asset{
					Name:        a.Name,
					DirPath:     filepath.Dir(a.RawFile.FilePath),
					RawFile:     a.RawFile,
					MatchStatus: types.MatchStatusOrphan,
				}
				raws = append(raws, orphan)
				orphans++
			}
			if a.JpgFile != nil {
				orphan := &types.Asset{
					Name:        a.Name,
					DirPath:     filepath.Dir(a.JpgFile.FilePath),
					JpgFile:     a.JpgFile,
					MatchStatus: types.MatchStatusOrphan,
				}
				jpgs = append(jpgs, orphan)
				orphans++
			}
		} else if a.RawFile != nil {
			raws = append(raws, a)
			orphans++
			assets = append(assets, a)
		} else if a.JpgFile != nil {
			jpgs = append(jpgs, a)
			orphans++
			assets = append(assets, a)
		}
	}

	// Step 4: Match remaining orphans by capture time
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

	// Step 5: Extract EXIF for remaining files that don't have it yet
	var wg sync.WaitGroup
	sem := make(chan struct{}, s.cfg.ConcurrencyLimit)
	exifCount := 0
	var exifMu sync.Mutex
	for _, a := range assets {
		for _, f := range []*types.MediaFile{a.RawFile, a.JpgFile} {
			if f == nil || f.Exif != nil {
				continue // already extracted during matching
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

	// Step 6: Batch insert — files are embedded in assets
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
