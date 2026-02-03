#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/.." && pwd)"
plugin_dir="$repo_root/poc/owlx-status-plugin"

if ! command -v cargo >/dev/null 2>&1; then
  echo "missing dependency: cargo" >&2
  exit 1
fi
if ! command -v rustup >/dev/null 2>&1; then
  echo "missing dependency: rustup" >&2
  exit 1
fi

target="wasm32-wasip1"

rustup target add "$target" >/dev/null

cd "$plugin_dir"
cargo build --target "$target" --release

out="$repo_root/tools/zellij/owlx-status.wasm"
mkdir -p "$(dirname "$out")"
cp "target/$target/release/owlx_status_plugin.wasm" "$out"
chmod 755 "$out"

echo "built: $out"
