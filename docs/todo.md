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
- [x] EXIF orientation auto-rotation on thumbnails (8 orientations: 1-8)
- [x] Orientation read from ARW files directly via goexif (ARW embedded JPEG has no EXIF)

### Thumbnails
- [x] Dual-layer cache: grid (600px) + full (2048px)
- [x] CGO/libjpeg-turbo decoder with DCT-domain scaling (handles Sony JPEGs)
- [x] RAW embedded JPEG extraction (ARW) — iterates all SOI/EOI pairs, picks largest valid
- [x] On-demand generation + pre-generation after scan
- [x] Bilinear resize (pure stdlib, center crop not stretch)
- [x] `max-width: 100%` prevents upscaling blur in waterfall layout

### API
- [x] Full REST API: CRUD, rating, labeling, thumbnail serving, scan trigger
- [x] 12 filter parameters: rating, color_label, camera_model, file_type, focal_length range, aperture range, ISO range, date range, search
- [x] Pagination with metadata (max limit 10000 for fetch-all)
- [x] Auto-port switching (up to 10 attempts)

### Frontend — Layout
- [x] Vue 3 Composition API + Pinia + TypeScript
- [x] Waterfall/masonry layout via CSS columns (responsive: 4→3→2→1 columns)
- [x] Left sidebar filter panel (220px fixed width)
- [x] Image preview overlay with EXIF info + rating + color label
- [x] No pagination — all assets loaded at once, waterfall scroll

### Frontend — Filters & Search
- [x] Search input, rating stars, file type dropdown, camera dropdown
- [x] Custom dark-themed calendar date range picker (locale-aware weekday headers)
- [x] Focal length range, aperture range, ISO range (min/max inputs)
- [x] Color label filter, clear all button

### Frontend — i18n
- [x] Bilingual support (zh/en) via lightweight composable, no external deps
- [x] Language toggle button in toolbar (中文/English)
- [x] locale persisted in localStorage
- [x] All UI text internationalized: FilterBar, GalleryView, ImageCard, ImagePreview, DateRangePicker

### Frontend — Review Mode (审片模式)
- [x] Preview/Review mode toggle buttons in toolbar
- [x] Single large image display with left/right navigation arrows
- [x] EXIF metadata tags: camera model, focal length, aperture, ISO, shutter speed, capture date
- [x] Rating stars + color labels in info bar
- [x] Zoom: mouse wheel, +/- buttons, +/- keys (0.1x — 8x)
- [x] Rotate 90° CW: button + R key
- [x] Fit screen / 1:1 toggle: button + F key
- [x] Reset view button
- [x] Click-drag panning when zoomed in (cursor: grab/grabbing)
- [x] Pan resets on image switch, fit toggle, and reset
- [x] EXIF orientation auto-applied as CSS transform (fallback for stale thumbnail cache)
- [x] Keyboard shortcut hints bar
- [x] Rating/label changes sync between review mode and store

### Frontend — Keyboard Shortcuts
- [x] 1-5: set rating (works in preview overlay + review mode)
- [x] 0: clear rating
- [x] X: toggle red reject label (review mode)
- [x] Arrow keys: prev/next image (review mode) or overlay navigation
- [x] R: rotate (review mode)
- [x] F: fit/1:1 toggle (review mode)
- [x] +/-: zoom in/out (review mode)
- [x] Escape: close preview overlay
- [x] Delete: delete current asset

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
- [ ] Cache versioning to avoid stale thumbnails after code changes to orientation/resize

### Frontend
- [ ] Virtual scrolling for 10k+ libraries (currently works with CSS columns, not virtual)
- [ ] Double-click waterfall image to enter review mode at that position
- [ ] EXIF filtering: match_status, orientation
- [ ] i18n for keyboard hint labels ("Rate"/"Zero"/"Reject"/"Rotate"/"Zoom"/"Fit")

### AI Extension (future)
- [ ] AI status field on assets (reserved in DB)
- [ ] Plugin interface for image quality assessment, closed-eye detection, etc.

### Operations
- [ ] SSE / WebSocket for real-time scan progress
- [ ] Configurable cache directory outside storage/
- [ ] Image deletion also removes source file (opt-in)
