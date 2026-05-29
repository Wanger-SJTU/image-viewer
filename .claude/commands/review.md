---
name: review
description: Review current changes for quality, security, and Go/Vue conventions
---

Review all uncommitted changes in this repository. Check for:

1. **Go code** — idiomatic patterns, error handling, concurrency safety, SQLite usage
2. **Vue 3 code** — Composition API correctness, reactivity, performance (virtual scroll)
3. **Security** — no hardcoded secrets, input validation, SQL injection prevention
4. **Architecture** — adherence to handler → service → repository layering, shared type contracts

Use the **code-reviewer** agent for general review and **security-reviewer** agent if auth, DB, or file I/O code is touched.
Run `go vet ./...` and `go build ./...` to verify no compilation issues.
