#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/.." && pwd)"
tools_dir="$repo_root/tools/zellij"
version_file="$tools_dir/VERSION"

if [ ! -f "$version_file" ]; then
  echo "missing version file: $version_file" >&2
  exit 1
fi

version="$(cat "$version_file")"
version="${version#v}"
if [ -z "$version" ]; then
  echo "empty zellij version in: $version_file" >&2
  exit 1
fi

asset="${ZELLIJ_ASSET:-zellij-no-web-x86_64-unknown-linux-musl.tar.gz}"
base_url="https://github.com/zellij-org/zellij/releases/download/v${version}"
url="${ZELLIJ_URL:-${base_url}/${asset}}"

tar_path="$tools_dir/$asset"
bin_path="$tools_dir/zellij"

mkdir -p "$tools_dir"

if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$url" -o "$tar_path"
elif command -v wget >/dev/null 2>&1; then
  wget -qO "$tar_path" "$url"
else
  echo "missing dependency: curl or wget" >&2
  exit 1
fi

tar -xzf "$tar_path" -C "$tools_dir"
chmod +x "$bin_path"

echo "installed: $bin_path"
