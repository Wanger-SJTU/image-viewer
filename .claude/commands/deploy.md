---
name: deploy
description: Build the single-binary distribution for the current platform
---

Build the project for distribution:

1. Build the Vue 3 frontend: `cd web && npm run build`
2. Build the Go binary with embedded frontend: `go build -o viewer ./cmd/viewer`
3. Verify the binary: `./viewer --help` or `./viewer --version`
4. The resulting `viewer` binary is self-contained — copy it anywhere and run.
