# API Conventions — Image Viewer

## Base URL

`http://localhost:8080/api/v1`

## Response Envelope

All JSON responses use a consistent envelope:

```json
{
  "success": true,
  "data": {},
  "error": null,
  "meta": { "total": 100, "page": 1, "limit": 50 }
}
```

## Endpoints (planned)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/health` | Health check |
| GET | `/api/v1/assets` | List assets (paginated, filterable) |
| GET | `/api/v1/assets/:id` | Single asset detail |
| POST | `/api/v1/assets/:id/rate` | Set rating (body: `{ "rating": 0-5 }`) |
| POST | `/api/v1/assets/:id/label` | Set color label |
| DELETE | `/api/v1/assets/:id` | Delete asset |
| GET | `/api/v1/thumbs/:id` | Get thumbnail (query: `?size=grid\|full`) |
| POST | `/api/v1/scan` | Trigger directory scan (body: `{ "path": "..." }`) |
| GET | `/api/v1/scan/status` | Scan progress (SSE/WebSocket) |

## Naming

- Plural nouns for collections: `/assets`, `/thumbs`
- Actions via POST on sub-resources: `/assets/:id/rate`
- Query params for filtering: `?rating=5&camera=A7M4&page=1&limit=50`
