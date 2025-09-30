#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
cd "$ROOT_DIR"

PREFIX=${PREFIX:-/usr/local}
BINDIR=${BINDIR:-"$PREFIX/bin"}
TARGET=${TARGET:-"$BINDIR/jcli"}
GOMODCACHE_PATH=${GOMODCACHE_PATH:-"$ROOT_DIR/.gocache"}

OS="$(uname -s || echo unknown)"
case "$OS" in
  Linux|Darwin)
    ;;
  *)
    echo "[warn] Unsupported OS '$OS'. Script is tested on Linux and macOS; continuing..." >&2
    ;;
esac

TMP_DIR=$(mktemp -d)
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

mkdir -p "$TMP_DIR" "$BINDIR"

OUTPUT="$TMP_DIR/jcli" GOMODCACHE_PATH="$GOMODCACHE_PATH" ./scripts/build.sh

if command -v install >/dev/null 2>&1; then
  install -m 0755 "$TMP_DIR/jcli" "$TARGET"
else
  cp "$TMP_DIR/jcli" "$TARGET"
  chmod 755 "$TARGET"
fi

echo "jcli installed to $TARGET"

echo
echo "Make sure '$BINDIR' is on your PATH."
