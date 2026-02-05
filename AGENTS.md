# Repository Guidelines

## Project Structure & Module Organization

- `owlx` is the main Go CLI binary (tmux + git-worktree workflow).
- `scripts/lint-md.sh` runs Markdown linting for docs.
- Shell completion scripts are generated via `owlx completion`.
- `scripts/` contains developer tooling (e.g., `scripts/shellcheck.sh`).
- No dedicated `tests/` directory exists today; validation is via linting and manual checks.

## Build, Test, and Development Commands

- `./owlx <command>`: run the CLI locally from the repo (after `make build`).
- `scripts/shellcheck.sh`: run ShellCheck over the Bash scripts.
- `scripts/lint-md.sh`: lint Markdown using `markdownlint-cli2` (via `npx`).

Dependencies: `bash` (scripts), `go` (build), and `git` (worktrees) are required.
Embedded tmux is mandatory; system tmux is not used (Linux amd64 only).
`fzf` is required for `owlx search`, and `npx` is needed for Markdown linting.

## Coding Style & Naming Conventions

- Scripts: Bash. Keep `set -euo pipefail` at script tops.
- Go: standard gofmt formatting; keep helpers small and focused.
- Indentation: 2 spaces; use `local` for function scope.
- Naming: `OXL_*` for constants and env vars, lowercase `snake_case` for locals and functions.
- Prefer small, single-purpose functions and explicit error handling via `die()`.

## Testing Guidelines

- No automated test suite is present.
- Run linters before changes:
  - `make lint` (sh + md + go).
  - `make lint-sh` (ShellCheck).
  - `make lint-md` (Markdown).
  - `make lint-go` (gofmt check + `go vet`).
- Lint workflow: run `make lint` for full checks, or run the specific target
  if you are only touching shell/markdown/go files.
- Do a quick manual smoke test for common flows (e.g., `owlx new`, `owlx ls`, `owlx resume`).

## Commit & Pull Request Guidelines

- Commit history uses short, imperative messages.
  Follow that style when possible, including conventional prefixes (e.g., `chore:`).
- PRs should include a brief summary, rationale, and the commands you ran
  (e.g., `scripts/shellcheck.sh`).
- UI screenshots are generally not required for this CLI tool.

## Configuration Notes

- User overrides live in `~/.owlxrc` (shell-style assignments).
- Key env vars: `OXL_ROOT` (default `~/projs`) and `OXL_WT_DIRNAME`
  (default `.worktrees`).
