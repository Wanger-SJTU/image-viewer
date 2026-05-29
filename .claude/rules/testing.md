# Testing — Image Viewer

## Requirements

- Go: table-driven tests, test files alongside source (`*_test.go` in same package or `_test` package)
- Vue: Vitest for unit tests, Playwright for E2E
- Minimum 80% coverage for new code

## Key Test Scenarios

- **Scanner**: concurrent walk + dual-track matching, orphan detection, batch insert with transaction rollback
- **Thumbnail**: cache hit/miss, RAW embedded preview extraction, WebP generation
- **Handler**: HTTP status codes, JSON response shape, error handling
- **Repository**: CRUD operations, query filtering, pagination

## Running Tests

```bash
go test ./...                    # All Go tests
go test -v -run TestScanner ./internal/service/  # Single test
cd web && npx vitest             # Frontend unit tests
```
