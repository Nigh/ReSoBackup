# AGENTS.md

## IMPORTANT: Keep This File Updated

> **After making any changes to the project (code, structure, dependencies, build commands, etc.), you MUST update this AGENTS.md file to reflect the current state of the project.** This file serves as the single source of truth for agents working on this codebase. Keeping it accurate ensures future sessions can quickly understand the project without re-reading all source files.

## Project Overview

**ReSo Backup** is a Reed-Solomon encrypted backup tool built with Go + Wails v3 + Svelte 5.

- CLI mode: `rsbackup backup/restore` with flags
- GUI mode: Wails v3 desktop app with Svelte 5 + DaisyUI + Tailwind CSS frontend

## Architecture

| Layer | Path | Purpose |
|-------|------|---------|
| Entry | `main.go` | CLI arg parsing, dispatches to CLI or Wails GUI |
| GUI Service | `gui.go` | Wails v3 `BackupService` exposing file dialogs and backup/restore APIs to JS |
| Business Logic | `internal/app/app.go` | `RunBackup()` / `RunRestore()` orchestration |
| Crypto | `internal/cryptox/crypto.go` | AES-256-GCM encrypt/decrypt, base64 string helpers |
| RS | `internal/rs/rs.go` | Thin wrapper around `klauspost/reedsolomon` |
| Format | `internal/format/metadata.go` | `Metadata` and `KDFInfo` JSON structs |
| Frontend | `frontend/src/` | Svelte 5 SPA with `BackupView`, `RestoreView`, i18n |

## Build & Dev Commands

```bash
# Dev mode (hot reload)
wails3 dev -config ./build/config.yml -port 9245

# Frontend only
cd frontend && npm run dev

# Build frontend
cd frontend && npm run build

# Build all platforms
make

# Build current platform
task build

# Regenerate Wails JS bindings (after changing gui.go service methods)
wails3 generate bindings

# Run Go tests
go test ./...
```

## Key Conventions

- **Go**: Standard library style, no comments unless asked, `internal/` for private packages
- **Frontend**: Svelte 5 runes (`$state`, `$effect`), DaisyUI components, Tailwind utility classes
- **i18n**: English + Chinese in `frontend/src/lib/i18n.svelte.js`, use `getT().keyName`
- **Bindings**: Auto-generated in `frontend/bindings/` (gitignored), regenerated via `wails3 generate bindings`
- **Metadata version**: v2 format with `encrypted` and `encrypt_file_name` fields; v1 backward compatible (always encrypted)

## File Structure

```
├── main.go                          # CLI entry + Wails app launch
├── gui.go                           # Wails GUI service (BackupService)
├── internal/
│   ├── app/app.go                   # Core backup/restore logic
│   ├── cryptox/crypto.go            # AES-256-GCM helpers
│   ├── format/metadata.go           # Metadata JSON structs
│   └── rs/rs.go                     # Reed-Solomon encode/reconstruct
├── frontend/
│   ├── src/
│   │   ├── App.svelte               # Root layout with tabs
│   │   ├── lib/BackupView.svelte    # Backup form UI
│   │   ├── lib/RestoreView.svelte   # Restore form UI
│   │   └── lib/i18n.svelte.js       # Translations (en/zh)
│   ├── bindings/                    # Auto-generated Wails JS bindings
│   └── package.json                 # Frontend deps (Svelte 5, DaisyUI, Tailwind)
├── build/                           # Build configs per platform
├── Taskfile.yml                     # Task runner config
└── Makefile                         # Cross-platform Go build
```

## Lint & Typecheck

```bash
# Go
go vet ./...

# Frontend
cd frontend && npm run build
```
