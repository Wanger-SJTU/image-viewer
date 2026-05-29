# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Image Viewer (Photo Sifter) is a web-based high-performance image viewer and photo culling/management app. The backend is Go, compiled as a **single binary with frontend static assets embedded via `go:embed`** — double-click to run locally, or serve over LAN for cross-device use.

## Tech Stack

- **Backend**: Go + Gin framework, SQLite for metadata
- **Frontend**: Vue 3 Composition API (Script Setup), Pinia state management
- **Shared types**: Go structs in `shared/types/` — these are the source of truth; frontend mirrors them as TypeScript types

## Directory Structure (planned)

```
shared/types/       # Go structs: Asset, MediaFile, ExifMeta — single source of truth
cmd/viewer/         # main.go: init SQLite, services, HTTP server
internal/
  config/           # Config: scan paths, cache paths, supported extensions, concurrency
  repository/       # SQLite CRUD (db.go, asset_repo.go)
  service/          # Business logic: scanner.go (concurrent scan + dual-track matching),
                    #   asset.go (rating, labels, filtering), thumb.go (thumbnail gen + .cache/ management)
  transport/http/   # router.go, handler.go — HTTP layer
web/
  dist/             # Built frontend (go:embed target)
  src/              # Vue 3 source
```

## Key Architectural Decisions

1. **Single-file distribution**: Go binary embeds frontend via `go:embed`. No separate deployment, no Node runtime needed at runtime.

2. **Dual-track matching engine**: RAW files (.CR3, .ARW, .NEF) and JPG files with the same base name are aggregated into a single "Asset" logical unit. Orphan files (no match) are flagged. Composite key for matching: `DirPath + "_" + Lowercase(AssetName)`.

3. **Dual-layer cache pipeline** (critical for performance):
   - Layer 1: 200px WebP thumbnails for grid/waterfall view
   - Layer 2: 2048px WebP for full-screen preview
   - RAW files: **always extract embedded preview JPEG** — never decode the full RAW file for thumbnails
   - Cache stored in `storage/cache/`, served from filesystem when available

4. **Scan workflow**: Frontend submits directory path → Backend returns `202 Accepted` immediately → async goroutine walk + concurrent worker pool via bounded channel → batch SQLite insert via transaction → progress reported via WebSocket or polling.

5. **Unidirectional data flow**: Frontend → HTTP → Handlers → Services → Repositories → SQLite. No circular dependencies.

6. **AI extension interface**: Data layer reserves `ai_status` fields on assets for future AI integration (quality assessment, closed-eye detection, semantic search).

## Core Features

- 0-5 star rating + color labels with keyboard shortcuts
- EXIF-based filtering (capture time, camera model, focal length, aperture, ISO)
- Virtual scrolling for 10k+ image libraries
- RAW format support: ARW (Sony), with CR2, CR3, NEF planned
