# API Design — Image Viewer

Base URL: `http://localhost:8080/api/v1`

All responses use a consistent envelope:

```json
{
  "success": true,
  "data": {},
  "error": null,
  "meta": { "total": 100, "page": 1, "limit": 50 }
}
```

## Endpoints

### Health
| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |

### Assets
| Method | Path | Description |
|--------|------|-------------|
| GET | `/assets` | List assets (paginated, filterable) |
| GET | `/assets/:id` | Single asset with full media file + EXIF |
| POST | `/assets/:id/rate` | Set rating `{ "rating": 0-5 }` |
| POST | `/assets/:id/label` | Set color label `{ "color_label": "red" }` |
| DELETE | `/assets/:id` | Delete single asset |
| DELETE | `/assets` | Clear all assets and cache |

### Thumbnails
| Method | Path | Description |
|--------|------|-------------|
| GET | `/thumbs/:id` | Get thumbnail `?size=grid\|full` (auto-generates if missing) |

### Scan
| Method | Path | Description |
|--------|------|-------------|
| POST | `/scan` | Trigger async directory scan `{ "path": "..." }` — returns 202 |

## List Assets — Filter Parameters

All query params are optional and combinable:

| Parameter | Type | Example |
|-----------|------|---------|
| `page` | int | `1` |
| `limit` | int | `50` |
| `rating` | int | `3` (>= rating) |
| `color_label` | string | `red`, `blue`, `green`, etc. |
| `camera_model` | string | `ILCE-7CM2` (LIKE match) |
| `file_type` | string | `jpg`, `raw`, `both` |
| `focal_length_min` | float | `24` |
| `focal_length_max` | float | `70` |
| `aperture_min` | float | `2.8` |
| `aperture_max` | float | `5.6` |
| `iso_min` | int | `100` |
| `iso_max` | int | `6400` |
| `captured_after` | RFC3339 | `2025-07-13T00:00:00Z` |
| `captured_before` | RFC3339 | `2025-07-14T00:00:00Z` |
| `search` | string | `DSC01` (LIKE match on filename) |

Example: `GET /assets?camera_model=ILCE-7CM2&focal_length_min=50&aperture_max=2.8&page=1`
