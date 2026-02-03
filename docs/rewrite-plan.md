# owlx Go + zellij rewrite plan

## Summary
Rewrite the current Bash CLI to a Go CLI using zellij (not tmux) for sessions.
Maintain the design intent and user workflow, but do not require 1:1 mapping.
Panel is removed. Status is required and should look like a tmux-style status line.

## Goals (MVP)
- Provide equivalent workflows for new/resume/ls/view/del/search/config/notify/status.
- Replace tmux with zellij and remove curl usage.
- Store session metadata outside zellij in a local state directory.
- Keep behavior consistent and predictable; prefer explicit errors over silent failure.

## Non-goals (initial)
- 1:1 behavioral parity with tmux options or session metadata.
- Backwards compatibility with ~/.owlxrc (Bash config).
- Panel feature.
- Robust bridge protocol design (treat as early-stage os.Exec).

## Constraints and decisions
- Zellij: use latest stable version, fixed at release time for product.
- Status: must render like a tmux status line (left fixed fields, right intent).
- Config: redesign using Go best practices (XDG config; env overrides ok).
- Notify: keep; implement via Go net/http (no curl).
- Git/FZF: call through a bridge via os.Exec.
- Work in current branch; no tmux/curl dependency.

## State model
- State dir: $XDG_STATE_HOME/owlx (fallback: ~/.local/state/owlx).
- Session file: state/sessions/<id>.json.
- Optional index: state/index.json for fast list.
- Each session record includes layout, repo, branch, category, intent, worktree,
  repo_dir, worktree_dir, session_name, id, timestamps, and status flags.

## Config model
- Config dir: $XDG_CONFIG_HOME/owlx (fallback: ~/.config/owlx).
- Config file: config.toml (or config.yaml) with:
  - root
  - worktree_dirname
  - notify: host, topic, template, template_file
  - status: on_new, right_len
- Env vars override config; defaults are applied if unset.

## Zellij integration
- Use zellij CLI for create/attach/list/kill sessions.
- Generate a dedicated layout file (owlx.kdl) to wire up tabs and the status bar.
- Status line is implemented as a custom status-bar plugin (WASM), reading from
  owlx state on a timer and rendering left/right fields.

## POC checklist
- Install fixed zellij version on a dev machine.
- Verify zellij can load a custom layout file (KDL).
- Verify status bar plugin can be overridden via layout or config.
- Build a minimal status plugin that renders static text.
- Extend plugin to read a local file and refresh the status line.
- Confirm permissions and file access behavior for plugins.

## Phases
1. POC: validate zellij layout + status plugin behavior and file access.
2. MVP core: implement state + core commands + zellij session control.
3. UX polish: status formatting, search, notify templates, docs update.
4. Hardening: error cases, cleanup, migration notes, packaging.

## Risks and mitigations
- Plugin access limitations: test early; fall back to env-based rendering.
- Zellij CLI changes: pin version; add compatibility checks.
- State drift: add health checks (missing sessions, missing worktrees).
- User config collisions: write only an owlx-owned layout file.
