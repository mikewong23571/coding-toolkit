# Repository Guidelines

## Project Structure & Module Organization

- `bin/owlx` is the main Bash CLI entrypoint (tmux + git-worktree workflow).
- `bin/lint-md` runs Markdown linting for docs.
- `completions/` holds shell completion scripts (e.g., `completions/owlx.bash`).
- `scripts/` contains developer tooling (e.g., `scripts/shellcheck.sh`).
- No dedicated `tests/` directory exists today; validation is via linting and manual checks.

## Build, Test, and Development Commands

- `bin/owlx <command>`: run the CLI locally from the repo.
- `scripts/shellcheck.sh`: run ShellCheck over the Bash scripts.
- `bin/lint-md`: lint Markdown using `markdownlint-cli2` (via `npx`).

Dependencies: `bash`, `git`, and `tmux` are required.
`fzf` is required for `owlx search`, and `npx` is needed for Markdown linting.

## Coding Style & Naming Conventions

- Language: Bash. Keep `set -euo pipefail` at script tops.
- Indentation: 2 spaces; use `local` for function scope.
- Naming: `OXL_*` for constants and env vars, lowercase `snake_case` for locals and functions.
- Prefer small, single-purpose functions and explicit error handling via `die()`.

## Testing Guidelines

- No automated test suite is present.
- Run linters before changes:
  - `scripts/shellcheck.sh` for shell scripts.
  - `bin/lint-md` for Markdown.
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
