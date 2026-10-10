# AGENTS.md

## IMPORTANT: Keep This File Updated

> After making any changes to the project (code, structure, dependencies, build commands, etc.), you MUST update this AGENTS.md file to reflect the current state of the project.

## Project Overview

ReSo Backup is a Reed-Solomon encrypted backup tool built with Go 1.27.1+ and mygo v0.3.7 Native UI. The desktop UI is entirely Go, drawn by mygo; no HTML, JavaScript, WebView, frontend build or generated bindings.

- No arguments: launch the native desktop GUI.
- CLI: `rsbackup backup/restore` with the existing flags.
- `-tags cli`: build/test without importing the GUI framework or loading desktop libraries; no arguments print help.

## Architecture

| Layer | Path | Purpose |
|-------|------|---------|
| Entry | `main.go` | CLI parsing and GUI dispatch |
| CLI interaction | `cli.go` | Terminal warning confirmation and password prompting |
| Native GUI | `gui.go` | State, backup/restore views, dialogs, async work and app lifecycle |
| Native geometry | `gui_geometry_linux.go`, `gui_geometry_other.go` | GTK frame extents via existing purego; native bounds on other platforms |
| CLI-only entry | `gui_cli.go` | `cli` build-tag stub |
| Theme | `theme.go` | Xianii dark/light token mapping, platform sans-serif font, system text scaling and contrast |
| i18n | `i18n.go` | English/Chinese UI strings |
| Business logic | `internal/app/app.go` | Backup/restore, validation, warnings, metadata discovery |
| Crypto | `internal/cryptox/crypto.go` | AES-256-GCM and base64 helpers |
| RS | `internal/rs/rs.go` | Reed-Solomon encode/reconstruct |
| Format | `internal/format/metadata.go` | Metadata JSON structs |
| Packaging | `mygo.json`, `resources/icon.png` | Native app identity, version, icon and package configuration |
| Theme source | `third_party/xianii-theme/` | Pinned upstream tokens, provenance and MIT license |

## Build & Dev Commands

```bash
make                      # current-platform GUI + CLI executable in dist/
make build-all            # Linux amd64, Windows amd64, macOS arm64 raw executables
make cli                  # CLI-only executable in dist/
task build                # current-platform executable
go tool mygo dev          # rebuild/restart native GUI on Go changes
go tool mygo build        # current-platform desktop packages in dist/packages/
go tool mygo build -platform darwin/universal,windows/amd64,linux/amd64
go tool mygo vet .         # Go vet plus native UI lifetime checks
make test                 # Go vet, all tests, CLI-only tests
go test -race ./...        # race checks (requires CGO for the race runtime)
```

Normal builds use `CGO_ENABLED=0`. Linux GUI runtime requires GTK 3, Cairo, GDK-Pixbuf and a display (X11 or Wayland); native-only windows do not load WebKitGTK, but mygo 0.3.7 still hardcodes it in `.deb` dependencies and the install script checks. Raw executables need no WebKitGTK. macOS packaging/signing/DMGs require macOS. Raw Windows executables retain the console for CLI use; mygo desktop packages use its GUI packaging defaults. Release CI tests on all three platforms, builds raw executables and desktop packages, and sets the package version from the release tag.

## Key Conventions

- Go standard library style; private business packages under `internal/`. No extra abstractions or dependencies for UI helpers.
- GUI files/tests have `//go:build !cli`; CLI entry stub has `//go:build cli`.
- Views build on the main thread. Store model state, never `ui.Element` or `*ui.Context`; use stable `c.Key` identities for conditional controls.
- Snapshot operation inputs before starting goroutines. Publish results with `Window.Update`; no worker mutates view state directly.
- Forms and tabs are disabled during operations; window close and app quit are prevented until work finishes. No cancellation/progress API is implemented in the core.
- Risk confirmation belongs to callers: CLI uses terminal confirmation, GUI uses a themed modal. `app.BackupWarnings` supplies the shared warning policy; `app.ValidateBackupOptions` validates both entry points. Existing warning/error text from the core remains English.
- Core backup/restore never prompt on stdin; encrypted operations reject empty passwords. CLI obtains a password before calling the core.
- The window starts at 900×500 content DIP; it is created hidden and shown after the window reference is assigned; after its first native layout, it measures the actual form, scroll viewport and footer/header bounds (independent of initial viewport height) and requests bounds that fit and vertically center within the current display work area, including the actual GTK titlebar/borders on Linux (queried through GDK using the existing purego dependency) or native window decoration on other platforms, and preserving horizontal position. Height and position are submitted together with SetBounds. GTK/Wayland global positioning is compositor-controlled, so application requests cannot guarantee vertical centering. If GTK frame measurement is unavailable, it logs a library-load failure and falls back to mygo bounds; decoration fitting then cannot be guaranteed. Startup sizing makes one request on the main thread; mygo v0.3.7 includes the upstream Wayland early-resize fix (PR #151), so no application timer, retry or rendered-height tracking is needed. Later form changes and manual resizing do not trigger auto-sizing. Backup/Restore modes use centered, equal-width, filled tabs with keyboard navigation; keep scrolling for smaller windows, larger system text and long messages.
- GUI defaults to the Xianii dark theme, with a light-theme toggle. Source tokens are pinned under `third_party/`; update Go mappings and provenance together. Use system font fallback and preserve keyboard labels/focus and system accessibility preferences.
- Metadata version remains v2 with `encrypted` / `encrypt_file_name`; v1 restore stays backward compatible (always encrypted).

## Tests

`gui_test.go` uses mygo's headless tester for navigation, language/theme changes, password visibility and disabled tabs, plus slider mapping/ratio and risk-modal cancellation tests. Layout checks ensure measured startup heights fit both plaintext and expanded encrypted backup forms in both languages without excess bottom space, are independent of the initial viewport height, remain unchanged after manual resizing and the mode tabs are symmetric, large enough, and support left/right keys. Startup geometry tests cover decoration, height caps, small work areas, minimum height, negative monitor origins and preserved horizontal bounds. Core tests include encrypted/plaintext roundtrips without terminal interaction and empty-password rejection and legacy v1 metadata recovery. CLI tests cover noninteractive risk rejection. Native rendering screenshots can be produced with `ui.NewTester(...).Image()`; these do not replace real desktop input-method and screen-reader checks.

## Migration Verification

Known defect retained by user decision: on native Linux Wayland, startup height adjustment can leave the window vertically off-center because the compositor ignores application positioning requests. Keep the native Wayland backend; do not force XWayland or change desktop settings to work around this. Documented in both READMEs. Height fitting remains active, but placement and titlebar accessibility ultimately depend on the compositor.

The Native UI migration was checked with Go vet, mygo lifetime checks, headless UI tests, core/CLI tests, race tests, and CGO-free cross-compilation for Linux amd64, Windows amd64 and macOS arm64. Linux mygo packaging was exercised. Startup sizing was additionally reproduced and checked on native Wayland using rendered surface dimensions rather than requested window bounds. An actual Linux X11 native-window event loop also completed an asynchronous encrypted backup/restore roundtrip. Windows/macOS desktop interaction, input methods, screen readers, signing and release CI still need their platform checks.

The mygo v0.3.7 upgrade and removal of the startup resize workaround were checked with Go vet, mygo lifetime checks, GUI/core/CLI tests, race tests and CGO-free Linux amd64, Windows amd64 and macOS arm64 builds. Native Wayland checks covered English/Chinese plaintext/encrypted forms, verified actual rendered dimensions against the measured height and display cap, and confirmed later resizing remains in effect. Desktop packaging and Windows/macOS interaction were not rechecked for this upgrade.

Startup vertical centering was checked with geometry regression tests, Go vet, mygo lifetime checks, GUI/core/CLI and race tests, plus CGO-free Linux/Windows/macOS builds. Native X11 checks confirmed the decorated frame is centered and fits the work area for both a normal form and an oversized form capped to the work-area height; later manual resizing remained in effect. Native Wayland checks confirmed sizing and later resizing, but global placement remains controlled by the compositor. Windows/macOS desktop positioning was not tested on hardware.
