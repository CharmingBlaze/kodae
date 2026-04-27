#!/usr/bin/env sh
set -eu

if [ "$#" -lt 3 ]; then
  echo "usage: scripts/package-portable.sh <clio-binary> <zig-binary> <platform> [out-dir]" >&2
  exit 1
fi

CLIO_BIN="$1"
ZIG_BIN="$2"
PLATFORM="$3"
OUT_DIR="${4:-dist}"

BUNDLE_ROOT="$OUT_DIR/clio-$PLATFORM"
BIN_DIR="$BUNDLE_ROOT/bin"
TOOLCHAIN_DIR="$BUNDLE_ROOT/toolchain/zig"

mkdir -p "$BIN_DIR" "$TOOLCHAIN_DIR"
cp "$CLIO_BIN" "$BIN_DIR/clio"
cp "$ZIG_BIN" "$TOOLCHAIN_DIR/zig"
chmod +x "$BIN_DIR/clio" "$TOOLCHAIN_DIR/zig"

ARCHIVE="$OUT_DIR/clio-$PLATFORM.tar.gz"
tar -C "$OUT_DIR" -czf "$ARCHIVE" "clio-$PLATFORM"
echo "portable bundle written: $ARCHIVE"
