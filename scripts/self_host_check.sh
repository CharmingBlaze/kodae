#!/usr/bin/env sh
# Two-stage smoke check: Go-built kodae compiles the in-tree compiler entrypoint,
# then the produced binary attempts the same build (requires C backend or --backend=llvm).
set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
BOOT="$(mktemp "${TMPDIR:-/tmp}/kodae-bootstrap.XXXXXX")"
ST2="$(mktemp "${TMPDIR:-/tmp}/kodae-stage2.XXXXXX")"
rm -f "$BOOT" "$ST2"
go build -o "$BOOT" ./cmd/kodae
SRC="src/compiler/main.kodae"
"$BOOT" build --backend=c -o "$ST2" "$SRC"
"$ST2" build --backend=c -o "${ST2}.2" "$SRC"
rm -f "$BOOT" "$ST2" "${ST2}.2"
echo "self_host_check: OK (two-stage build of $SRC)"
