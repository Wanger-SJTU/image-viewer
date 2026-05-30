# TODO — Image Viewer

## Completed Features

### Scanning & Matching
- [x] Concurrent WalkDir + bounded worker pool
- [x] Dual-track matching (dir+basename) — RAW ↔ JPG pairing
- [x] Time-based matching — cross-directory pairing via EXIF capture time
- [x] Scan progress reporting (phases: scanning → matching → exif → saving → done)

### EXIF Extraction
- [x] Full EXIF metadata: camera model, lens, focal length, aperture, shutter speed, ISO
- [x] Image dimensions, orientation, capture time
- [x] Bounded concurrent extraction (configurable concurrency limit)
- [x] EXIF stored in `media_files` table, `captured_at` propagated to `assets`

### Thumbnails
- [x] Dual-layer cache: grid (200px) + full (2048px)
- [x] CGO/libjpeg-turbo decoder with DCT-domain scaling (handles Sony JPEGs)
- [x] RAW embedded JPEG extraction (ARW) — iterates all SOI/EOI pairs, picks largest valid
- [x] On-demand generation + pre-generation after scan
- [x] Bilinear resize (pure stdlib)

### API
- [x] Full REST API: CRUD, rating, labeling, thumbnail serving, scan trigger
- [x] 12 filter parameters: rating, color_label, camera_model, file_type, focal_length range, aperture range, ISO range, date range, search
- [x] Pagination with metadata
- [x] Auto-port switching (up to 10 attempts)

### Frontend
- [x] Vue 3 Composition API + Pinia + TypeScript
- [x] Image grid with file type badges (JPG/RAW/RAW+JPG)
- [x] Image preview overlay with EXIF info display + rating + labeling
- [x] Comprehensive FilterBar: search, rating, date range, focal length, aperture, ISO, camera, file type
- [x] Keyboard shortcuts (1-5 rate, arrows navigate, Esc close, Delete)
- [x] Clear All with confirmation
- [x] Scan dialog

### Build & Deployment
- [x] go:embed single binary (frontend + backend)
- [x] build.sh script

## Pending

### RAW Format Support
- [ ] CR2 (Canon), CR3 (Canon), NEF (Nikon) embedded JPEG extraction
- [ ] ARW full-resolution JPEG extraction (currently picks largest valid, could prefer specific IFD)

### Thumbnails
- [ ] WebP encoding (currently JPEG quality 85)
- [ ] Cache eviction / size management

### Frontend
- [ ] Virtual scrolling (@tanstack/vue-virtual) for 10k+ libraries
- [ ] Color label filter UI
- [ ] Masonry / waterfall layout

### AI Extension (future)
- [ ] AI status field on assets (reserved in DB)
- [ ] Plugin interface for image quality assessment, closed-eye detection, etc.

### Operations
- [ ] SSE / WebSocket for real-time scan progress
- [ ] Configurable cache directory outside storage/
- [ ] Image deletion also removes source file (opt-in)
