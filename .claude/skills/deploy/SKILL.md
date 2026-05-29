---
name: deploy
description: Build and verify the single-binary distribution
---

# Deploy Skill

Build the self-contained viewer binary.

## Steps

1. **Frontend build**: `cd web && npm run build` — outputs to `web/dist/`
2. **Go build**: `CGO_ENABLED=1 go build -ldflags="-s -w" -o viewer ./cmd/viewer`
3. **Verify**: `./viewer --version` prints version info
4. **Package**: The `viewer` binary + optional `storage/` directory is the full distribution

## Notes

- CGO is required for SQLite (mattn/go-sqlite3)
- Cross-compilation needs a C cross-compiler for the target platform
- `-ldflags="-s -w"` strips debug info to reduce binary size
