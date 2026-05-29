# Code Style — Image Viewer

## Go

- Follow standard Go project layout: `cmd/`, `internal/`, `shared/`
- Use bounded concurrency: worker pool pattern with buffered channels for I/O operations
- Repository layer isolates all raw SQL — never write SQL in handlers or services
- Always use transactions for batch writes
- Error handling: return wrapped errors with context, never silently swallow

## Vue 3

- Composition API with `<script setup>` syntax only
- Composables for reusable logic (e.g., `useKeyboardShortcut`, `useVirtualScroll`)
- Pinia stores for shared state (current asset, filters, scan progress)
- Component props/emits must have TypeScript type declarations

## Shared Types

- `shared/types/` is the single source of truth for Asset, MediaFile, ExifMeta
- Frontend mirrors Go types as TypeScript interfaces manually or via codegen
- Changes to shared types require updating both Go and TS definitions
