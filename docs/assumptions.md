# Runtime assumptions

This document lists the external assumptions that owlx relies on at runtime.
The Go implementation is aligned with these assumptions and returns explicit
errors when they are not met.

## OS and filesystem
- Linux with standard XDG directories.
- Writable directories:
  - Config: `~/.config/owlx/`
  - State: `~/.local/state/owlx/`
  - Data: `~/.local/share/owlx/`

## Required executables
- `zellij` (configurable with `OXL_ZELLIJ_BIN` or `zellij_bin` in config).
- `git` (for worktrees).
- `fzf` (only for `owlx search`).

## Layout and repo structure
- Repos live under `OXL_ROOT/<layout>/<repo>` (default `~/projs`).
- Valid layouts: `main`, `fork`, `playground`, `archive`.
- Worktrees stored under `OXL_WT_DIRNAME` (default `.worktrees`).

## Zellij integration
- Zellij CLI supports `--new-session-with-layout`, `attach`, `list-sessions`,
  and `kill-session`.
- Status bar is provided by a WASM plugin and loaded by per-session layout
  files stored in the owlx state directory.
- If `zellij_layout` / `OXL_ZELLIJ_LAYOUT` is set, owlx assumes the layout file
  exists and includes the status bar plugin. Otherwise owlx generates a
  per-session layout file in `~/.local/state/owlx/layouts/<id>.kdl`.

## Network
- `owlx notify` requires HTTP access to the configured `notify.host/topic`
  or an explicit `--url`.

## Build-time (developer)
- Rust toolchain (`cargo`, `rustup`) is required to build the status plugin.
- Use:
  - `scripts/fetch-zellij.sh`
  - `scripts/build-status-plugin.sh`
