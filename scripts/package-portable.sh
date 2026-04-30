#!/usr/bin/env sh
set -eu

if [ "$#" -lt 2 ]; then
  echo "usage: scripts/package-portable.sh <kodae-binary> <platform> [out-dir]" >&2
  echo "  Creates dist/kodae-<platform>/ with bin/kodae only (no bundled compiler)." >&2
  exit 1
fi

KODAE_BIN="$1"
PLATFORM="$2"
OUT_DIR="${3:-dist}"

BUNDLE_ROOT="$OUT_DIR/kodae-$PLATFORM"
BIN_DIR="$BUNDLE_ROOT/bin"

mkdir -p "$BIN_DIR"
cp "$KODAE_BIN" "$BIN_DIR/kodae"
chmod +x "$BIN_DIR/kodae"

ARCHIVE="$OUT_DIR/kodae-$PLATFORM.tar.gz"
tar -C "$OUT_DIR" -czf "$ARCHIVE" "kodae-$PLATFORM"
echo "portable bundle written: $ARCHIVE"
