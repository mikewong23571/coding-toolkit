#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

# Derived from build-static-tmux (MIT License).
# Source: https://github.com/mjakob-gh/build-static-tmux
# See scripts/build-embedded-tmux.LICENSE.

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

usage() {
  cat <<'USAGE'
Usage: scripts/build-embedded-tmux.sh

Build or install the embedded static tmux binary at:
  assets/tmux/linux_amd64/tmux

Inputs:
  TMUX_STRIPPED          Path to a prebuilt tmux.*.stripped.gz (skips build).
  TMUX_META_FILE         Metadata file (default: assets/tmux/buildinfo.env).
  TMUX_STATIC_HOME       Build output root (default: ~/.cache/owlx/tmux-static).
  TMUX_DEFAULT_DIR       Compile-time tmux default dir (default: from TMUX_META_FILE).
  TMUX_DEFAULT_CONF      Compile-time tmux default config path (overrides TMUX_DEFAULT_DIR).
  TMUX_DEFAULT_SOCK      Compile-time tmux default socket path (overrides TMUX_DEFAULT_DIR).
  TMUX_VERSION           Tmux version (default: from TMUX_META_FILE).
  MUSL_VERSION           Musl version (default: from TMUX_META_FILE).
  NCURSES_VERSION        Ncurses version (default: from TMUX_META_FILE).
  LIBEVENT_VERSION       Libevent version (default: from TMUX_META_FILE).
  CLEAN_BUILD=1          Remove TMUX_STATIC_HOME after a successful build.
USAGE
}

die() {
  echo "error: $*" >&2
  exit 1
}

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    die "missing tool: $1"
  fi
}

fetch() {
  local url="$1"
  local dest="$2"

  if [ -f "$dest" ]; then
    return 0
  fi

  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$dest"
    return 0
  fi

  if command -v wget >/dev/null 2>&1; then
    wget -qO "$dest" "$url"
    return 0
  fi

  die "need curl or wget to download $url"
}

jobs() {
  local count

  if command -v getconf >/dev/null 2>&1; then
    count="$(getconf _NPROCESSORS_ONLN 2>/dev/null || echo 1)"
  else
    count=1
  fi

  if [ -z "$count" ]; then
    count=1
  fi

  echo "$count"
}

trim() {
  local s="$1"
  s="${s#"${s%%[![:space:]]*}"}"
  s="${s%"${s##*[![:space:]]}"}"
  printf '%s' "$s"
}

load_meta() {
  local file="$1"
  local line
  local key
  local val

  if [ ! -f "$file" ]; then
    return 0
  fi

  while IFS= read -r line || [ -n "$line" ]; do
    line="${line%%#*}"
    line="$(trim "$line")"
    if [ -z "$line" ] || [ "${line#*=}" = "$line" ]; then
      continue
    fi
    key="$(trim "${line%%=*}")"
    val="$(trim "${line#*=}")"
    case "$key" in
      TMUX_VERSION) META_TMUX_VERSION="$val" ;;
      MUSL_VERSION) META_MUSL_VERSION="$val" ;;
      NCURSES_VERSION) META_NCURSES_VERSION="$val" ;;
      LIBEVENT_VERSION) META_LIBEVENT_VERSION="$val" ;;
      TMUX_DEFAULT_DIR) META_TMUX_DEFAULT_DIR="$val" ;;
      TMUX_DEFAULT_CONF) META_TMUX_DEFAULT_CONF="$val" ;;
      TMUX_DEFAULT_SOCK) META_TMUX_DEFAULT_SOCK="$val" ;;
      *)
        ;;
    esac
  done < "$file"
}

platform() {
  local os
  local arch

  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  arch="$(uname -m)"
  case "$arch" in
    x86_64) arch="amd64" ;;
    aarch64) arch="arm64" ;;
    *) ;;
  esac

  printf '%s %s\n' "$os" "$arch"
}

ensure_linux_amd64() {
  local os
  local arch

  IFS=' ' read -r os arch < <(platform)

  if [ "$os" != "linux" ] || [ "$arch" != "amd64" ]; then
    die "only linux/amd64 builds are supported (got ${os}/${arch})"
  fi
}

ensure_tools() {
  need_cmd tar
  need_cmd make
  need_cmd patch
  need_cmd gzip
  need_cmd strip
  need_cmd cc
}

prepare_dirs() {
  rm -rf "${TMUX_STATIC_HOME:?}/src" \
    "${TMUX_STATIC_HOME:?}/bin" \
    "${TMUX_STATIC_HOME:?}/lib" \
    "${TMUX_STATIC_HOME:?}/include"

  mkdir -p "${TMUX_STATIC_HOME}/src" \
    "${TMUX_STATIC_HOME}/bin" \
    "${TMUX_STATIC_HOME}/lib" \
    "${TMUX_STATIC_HOME}/include"
}

build_musl() {
  local src="${TMUX_STATIC_HOME}/src"
  local j

  j="$(jobs)"

  cd "$src"
  fetch "$MUSL_URL" "$MUSL_ARCHIVE"
  rm -rf "musl-${MUSL_VERSION}"
  tar xzf "$MUSL_ARCHIVE"
  cd "musl-${MUSL_VERSION}"
  ./configure --prefix="${TMUX_STATIC_HOME}"
  make -j "$j"
  make install
}

build_libevent() {
  local src="${TMUX_STATIC_HOME}/src"
  local j

  j="$(jobs)"

  cd "$src"
  fetch "$LIBEVENT_URL" "$LIBEVENT_ARCHIVE"
  rm -rf "libevent-${LIBEVENT_VERSION}-stable"
  tar xzf "$LIBEVENT_ARCHIVE"
  cd "libevent-${LIBEVENT_VERSION}-stable"
  ./configure \
    --prefix="${TMUX_STATIC_HOME}" \
    --disable-shared \
    --disable-openssl \
    --disable-samples \
    --disable-tests
  make -j "$j"
  make install
}

build_ncurses() {
  local src="${TMUX_STATIC_HOME}/src"
  local j

  j="$(jobs)"

  cd "$src"
  fetch "$NCURSES_URL" "$NCURSES_ARCHIVE"
  rm -rf "ncurses-${NCURSES_VERSION}"
  tar xzf "$NCURSES_ARCHIVE"
  cd "ncurses-${NCURSES_VERSION}"
  ./configure \
    --prefix="${TMUX_STATIC_HOME}" \
    --includedir="${TMUX_STATIC_HOME}/include" \
    --libdir="${TMUX_STATIC_HOME}/lib" \
    --enable-pc-files \
    --with-pkg-config="${TMUX_STATIC_HOME}/lib/pkgconfig" \
    --with-pkg-config-libdir="${TMUX_STATIC_HOME}/lib/pkgconfig" \
    --without-ada \
    --without-cxx \
    --without-cxx-binding \
    --without-tests \
    --without-manpages \
    --without-debug \
    --disable-lib-suffixes \
    --with-ticlib \
    --with-termlib \
    --with-default-terminfo-dir=/usr/share/terminfo \
    --with-terminfo-dirs=/etc/terminfo:/lib/terminfo:/usr/share/terminfo
  make -j "$j"
  make install.libs
}

patch_tmux_sock_path() {
  if grep -q "TMUX_SOCK_PATH" tmux.c; then
    return 0
  fi

  patch -p1 <<'PATCH'
--- a/tmux.c
+++ b/tmux.c
@@ -517,8 +517,23 @@ main(int argc, char **argv)
 			path = xstrdup(s);
 			path[strcspn(path, ",")] = '\0';
 		}
 	}
+#ifdef TMUX_SOCK_PATH
+	if (path == NULL && label == NULL) {
+		char	**paths;
+		u_int	  n;
+
+		expand_paths(TMUX_SOCK_PATH, &paths, &n, 1);
+		if (n != 0) {
+			path = paths[0];
+			for (i = 1; i < n; i++)
+				free(paths[i]);
+			free(paths);
+			flags |= CLIENT_DEFAULTSOCKET;
+		}
+	}
+#endif
 	if (path == NULL) {
 		if ((path = make_label(label, &cause)) == NULL) {
 			if (cause != NULL) {
 				fprintf(stderr, "%s\n", cause);
PATCH
}

build_tmux() {
  local src="${TMUX_STATIC_HOME}/src"
  local j
  local cppflags

  j="$(jobs)"

  cd "$src"
  fetch "$TMUX_URL" "$TMUX_ARCHIVE"
  rm -rf "tmux-${TMUX_VERSION}"
  tar xzf "$TMUX_ARCHIVE"
  cd "tmux-${TMUX_VERSION}"

  patch_tmux_sock_path

  cppflags="-I${TMUX_STATIC_HOME}/include -DTMUX_CONF=\\\"${TMUX_DEFAULT_CONF}\\\" -DTMUX_SOCK_PATH=\\\"${TMUX_DEFAULT_SOCK}\\\""

  ./configure \
    --prefix="${TMUX_STATIC_HOME}" \
    --enable-static \
    --includedir="${TMUX_STATIC_HOME}/include" \
    --libdir="${TMUX_STATIC_HOME}/lib" \
    CFLAGS="-I${TMUX_STATIC_HOME}/include" \
    LDFLAGS="-L${TMUX_STATIC_HOME}/lib" \
    CPPFLAGS="$cppflags" \
    LIBEVENT_LIBS="-L${TMUX_STATIC_HOME}/lib -levent" \
    LIBNCURSES_CFLAGS="-I${TMUX_STATIC_HOME}/include/ncurses" \
    LIBNCURSES_LIBS="-L${TMUX_STATIC_HOME}/lib -lncurses" \
    LIBTINFO_CFLAGS="-I${TMUX_STATIC_HOME}/include/ncurses" \
    LIBTINFO_LIBS="-L${TMUX_STATIC_HOME}/lib -ltinfo"

  make -j "$j"
  make install
}

package_tmux() {
  local os
  local arch
  local base

  IFS=' ' read -r os arch < <(platform)

  base="${TMUX_STATIC_HOME}/bin/tmux.${os}-${arch}"
  cp "${TMUX_STATIC_HOME}/bin/tmux" "$base"
  cp "$base" "${base}.stripped"
  strip "${base}.stripped"
  gzip -f "$base"
  gzip -f "${base}.stripped"

  echo "Standard tmux binary:   ${base}.gz"
  echo "Stripped tmux binary:   ${base}.stripped.gz"
}

install_embedded() {
  local stripped_gz="$1"
  local out_dir="$ROOT/assets/tmux/linux_amd64"
  local out="$out_dir/tmux"

  mkdir -p "$out_dir"
  gzip -dc "$stripped_gz" > "$out"
  chmod 0755 "$out"

  if command -v file >/dev/null 2>&1; then
    if ! file "$out" | grep -q "ELF 64-bit"; then
      die "embedded tmux is not ELF 64-bit"
    fi
    if ! file "$out" | grep -q "statically linked"; then
      die "embedded tmux is not statically linked"
    fi
  fi

  echo "installed embedded tmux: $out"
}

cleanup_build() {
  if [ "${CLEAN_BUILD:-0}" = "1" ]; then
    rm -rf "${TMUX_STATIC_HOME}"
  fi
}

main() {
  if [ "${1:-}" = "-h" ] || [ "${1:-}" = "--help" ]; then
    usage
    exit 0
  fi

  TMUX_STATIC_HOME="${TMUX_STATIC_HOME:-$HOME/.cache/owlx/tmux-static}"
  TMUX_META_FILE="${TMUX_META_FILE:-$ROOT/assets/tmux/buildinfo.env}"
  load_meta "$TMUX_META_FILE"

  TMUX_VERSION="${TMUX_VERSION:-${META_TMUX_VERSION:-3.6a}}"
  MUSL_VERSION="${MUSL_VERSION:-${META_MUSL_VERSION:-1.2.5}}"
  NCURSES_VERSION="${NCURSES_VERSION:-${META_NCURSES_VERSION:-6.5}}"
  LIBEVENT_VERSION="${LIBEVENT_VERSION:-${META_LIBEVENT_VERSION:-2.1.12}}"

  TMUX_DEFAULT_DIR="${TMUX_DEFAULT_DIR:-${META_TMUX_DEFAULT_DIR:-~/.owlx/tmux/default}}"
  TMUX_DEFAULT_CONF="${TMUX_DEFAULT_CONF:-${META_TMUX_DEFAULT_CONF:-${TMUX_DEFAULT_DIR}/tmux.conf}}"
  TMUX_DEFAULT_SOCK="${TMUX_DEFAULT_SOCK:-${META_TMUX_DEFAULT_SOCK:-${TMUX_DEFAULT_DIR}/tmux.sock}}"

  MUSL_ARCHIVE="musl-${MUSL_VERSION}.tar.gz"
  MUSL_URL="https://musl.libc.org/releases/${MUSL_ARCHIVE}"

  NCURSES_ARCHIVE="ncurses-${NCURSES_VERSION}.tar.gz"
  NCURSES_URL="https://invisible-island.net/archives/ncurses/${NCURSES_ARCHIVE}"

  LIBEVENT_ARCHIVE="libevent-${LIBEVENT_VERSION}-stable.tar.gz"
  LIBEVENT_URL="https://github.com/libevent/libevent/releases/download/release-${LIBEVENT_VERSION}-stable/${LIBEVENT_ARCHIVE}"

  TMUX_ARCHIVE="tmux-${TMUX_VERSION}.tar.gz"
  TMUX_URL="https://github.com/tmux/tmux/releases/download/${TMUX_VERSION}/${TMUX_ARCHIVE}"

  if [ -z "${TMUX_STRIPPED:-}" ]; then
    ensure_linux_amd64
    ensure_tools
    prepare_dirs
    build_musl

    export CC="${TMUX_STATIC_HOME}/bin/musl-gcc -static"

    build_libevent
    build_ncurses
    build_tmux
    package_tmux

    TMUX_STRIPPED="${TMUX_STATIC_HOME}/bin/tmux.$(platform | tr ' ' '-').stripped.gz"
  fi

  if [ ! -f "$TMUX_STRIPPED" ]; then
    die "tmux binary not found: $TMUX_STRIPPED"
  fi

  install_embedded "$TMUX_STRIPPED"
  cleanup_build
}

main "$@"
