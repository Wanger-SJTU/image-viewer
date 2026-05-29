---
name: fix-issue
description: Diagnose and fix a bug or build failure
---

Analyze the reported issue or build failure:

1. Read error output and trace to the relevant source files
2. Check `internal/` for backend issues, `web/src/` for frontend issues
3. For build errors, use **build-error-resolver** agent
4. For Go-specific issues, check concurrency patterns (worker pool, channels), SQLite queries, and handler logic
5. Apply minimal fix and verify with `go build ./...` and `go vet ./...`
6. If applicable, add a regression test
