#!/usr/bin/env bash
set -euo pipefail

if ! command -v npx >/dev/null 2>&1; then
  echo "npx is required to run markdownlint (install Node.js)." >&2
  exit 1
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONFIG_FILE="${MARKDOWNLINT_CONFIG:-$ROOT_DIR/.markdownlint.jsonc}"
GLOB_PATTERN="${MARKDOWNLINT_GLOB:-**/*.md}"
IGNORE_PATTERN="${MARKDOWNLINT_IGNORE_GLOB:-!.worktrees/**}"
cd "$ROOT_DIR"

args=(--config "$CONFIG_FILE" "$GLOB_PATTERN")
if [ -n "$IGNORE_PATTERN" ]; then
  args+=("$IGNORE_PATTERN")
fi

npx --yes markdownlint-cli2 "${args[@]}"
