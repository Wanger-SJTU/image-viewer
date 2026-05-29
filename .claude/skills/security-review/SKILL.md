---
name: security-review
description: Security audit for the Image Viewer codebase
---

# Security Review Skill

Automated security review workflow for this project.

## Steps

1. Scan changed files for hardcoded secrets, API keys, tokens
2. Check HTTP handlers for input validation (path params, query params, JSON body)
3. Verify SQLite queries use parameterized statements (no string concatenation)
4. Check file system operations: path traversal prevention on scan paths and thumbnail access
5. Verify CORS middleware is configured and restrictive
6. Report findings with severity: CRITICAL / HIGH / MEDIUM / LOW

## Context

This is a local-first desktop app, but it serves HTTP on LAN. Attack surface:
- Malicious scan paths (path traversal)
- Thumbnail path injection
- Unauthenticated API access on LAN
