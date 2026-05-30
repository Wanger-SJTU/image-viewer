# Architecture — Image Viewer (v1.5)

## 1. System Architecture

```
┌─────────────────┐             ┌─────────────────┐
│                 │             │                 │
│   Frontend      │────────────▶│   Backend       │
│   (Vue 3)       │   HTTP      │   (Go + Gin)    │
│                 │             │                 │
└─────────────────┘             └────────┬────────┘
                                         │
                         ┌───────────────┴───────────────┐
                         ▼                               ▼
              ┌─────────────────┐             ┌─────────────────┐
              │   File System   │             │     SQLite      │
              │ (直接读取本地RAW/JPG)│             │  (viewer.db)    │
              └─────────────────┘             └─────────────────┘
                         │
                         ▼
              ┌─────────────────┐
              │  storage/cache/  │
              │ (200px grid +     │
              │  2048px full)    │
              └─────────────────┘
```

## 2. Project Directory Structure

```text
image-viewer/
├── main.go                      # Entry point + go:embed frontend
├── go.mod / go.sum
├── build.sh                     # Build script (frontend + backend)
├── viewer                       # Compiled single binary
├── cmd/viewer/main.go           # Placeholder
├── shared/types/                # Source of truth: Asset, MediaFile, ExifMeta, Filter
│   ├── asset.go
│   ├── filter.go
│   └── api.go
├── internal/
│   ├── config/config.go         # Port, DB path, cache dir, supported extensions
│   ├── repository/
│   │   ├── db.go                # SQLite init + migrations (assets + media_files)
│   │   └── asset_repo.go        # Full CRUD, dynamic filtering, pagination
│   ├── service/
│   │   ├── scanner.go           # WalkDir + dual-track matching + time-based matching
│   │   ├── asset.go             # Rating, labeling, listing
│   │   ├── thumb.go             # Two-layer thumbnail generation + cache
│   │   ├── raw_preview.go       # ARW embedded JPEG extraction
│   │   └── exif.go              # Full EXIF metadata extraction (goexif)
│   ├── transport/http/
│   │   ├── router.go            # Gin router + embedded SPA fallback
│   │   ├── handler.go           # CRUD handlers + filter parsing
│   │   ├── handler_scan.go      # Async scan trigger
│   │   └── handler_thumb.go     # Thumbnail serving
│   └── jpegdecoder/
│       └── decoder.go           # CGO wrapper: libjpeg-turbo with DCT-domain scaling
├── web/src/                     # Vue 3 Composition API
│   ├── api/                     # Axios client + API functions
│   ├── components/              # ImageCard, ImageGrid, ImagePreview, FilterBar, etc.
│   ├── composables/             # useKeyboardShortcut
│   ├── stores/assets.ts         # Pinia store
│   ├── types/                   # TypeScript mirror of shared/types/
│   └── views/GalleryView.vue    # Main gallery page
├── storage/
│   ├── viewer.db                # SQLite database
│   └── cache/                   # Grid/full thumbnails
└── docs/                        # Documentation
```

## 3. Core Data Flows

### 3.1 Asset Scanning & Dual-Track Matching

```
WalkDir → Collect entries (RAW + JPG ext filter)
  → First pass: match by DirPath + Lowercase(BaseName)
  → Second pass: match orphan RAW/JPG by EXIF capture time
  → Extract full EXIF for all media files (bounded concurrency)
  → Propagate captured_at to asset
  → Batch insert via SQLite transaction (500/batch)
```

Matching key for first pass: `DirPath + Lowercase(BaseName)`
Matching key for second pass: `EXIF DateTimeOriginal.UTC().Round(1s)`

### 3.2 Thumbnail Cache Pipeline

```
Request: GET /thumbs/:id?size=grid|full
  → Check asset.GridThumb / FullThumb in DB
  → Check file exists on disk (storage/cache/)
  → If missing: generate on-demand
    - JPG: libjpeg-turbo DCT-domain scaling → bilinear resize → JPEG encode
    - RAW: extract largest valid embedded JPEG → bilinear resize → JPEG encode
    - Save to storage/cache/{id}_{size}.jpg
    - Update DB with cache path
  → Serve file
```

Cache layers:
- Grid: 200px on long side
- Full: 2048px on long side

### 3.3 Filtering Pipeline

```
Frontend FilterBar → emits Partial<AssetFilter>
  → Pinia store updates filter
  → listAssets(filter, page, limit) → GET /assets?params...
  → Handler parses all query params → AssetFilter struct
  → Repository builds dynamic SQL WHERE from non-zero filter fields
  → Returns paginated results with metadata
```

Supported filters: rating, color_label, camera_model, file_type (jpg/raw/both),
focal_length_min/max, aperture_min/max, iso_min/max, captured_after/before, search

## 4. Key Technical Decisions

| Decision | Rationale |
|----------|-----------|
| CGO + libjpeg-turbo | Go stdlib fails on many Sony JPEGs; libjpeg-turbo handles them + DCT scaling |
| Time-based matching | RAW/JPG in different directories (full-width period in path) need EXIF-based pairing |
| Bounded concurrency | Worker pool via buffered channel prevents I/O overload during scan + EXIF extraction |
| Auto-port switching | Detects EADDRINUSE and increments port up to 10 times |
| go:embed single binary | Entire Vue 3 frontend embedded in Go binary; no Node.js at runtime |
