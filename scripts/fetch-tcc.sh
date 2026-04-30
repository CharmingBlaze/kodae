#!/usr/bin/env sh
# Download TinyCC prebuilt into ./toolchain/ for the given GOOS/GOARCH pair (for bundle CI).
# Usage: ./scripts/fetch-tcc.sh windows amd64
set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
OS="${1:?need os (windows|linux|darwin)}"
ARCH="${2:?need arch (amd64|arm64)}"
mkdir -p toolchain
case "$OS-$ARCH" in
  windows-amd64)
    ZIP="$ROOT/toolchain/.tcc-win64.zip"
    curl -fsSL -o "$ZIP" "http://download.savannah.gnu.org/releases/tinycc/tcc-0.9.27-win64-bin.zip"
    rm -rf toolchain/.tcc-win-extract
    mkdir -p toolchain/.tcc-win-extract
    unzip -q -o "$ZIP" -d toolchain/.tcc-win-extract
    if test -f toolchain/.tcc-win-extract/tcc/tcc.exe; then
      cp toolchain/.tcc-win-extract/tcc/tcc.exe toolchain/tcc.exe
      cp toolchain/.tcc-win-extract/tcc/*.dll toolchain/ 2>/dev/null || true
    else
      echo "unexpected zip layout under toolchain/.tcc-win-extract" >&2
      find toolchain/.tcc-win-extract -maxdepth 3 -type f >&2 || true
      exit 1
    fi
    rm -rf toolchain/.tcc-win-extract "$ZIP"
    ;;
  linux-amd64)
    if test -n "${GITHUB_ACTIONS:-}" && command -v sudo >/dev/null 2>&1; then
      sudo apt-get update -qq
      sudo apt-get install -y tcc
      cp "$(command -v tcc)" toolchain/tcc
      chmod +x toolchain/tcc
    else
      echo "On Linux install TinyCC (e.g. apt install tcc) then: cp \"\$(command -v tcc)\" toolchain/tcc" >&2
      exit 2
    fi
    ;;
  darwin-arm64)
    curl -fsSL -o toolchain/tcc "https://github.com/vlang/tccbin/raw/master/tcc-darwin-arm64"
    chmod +x toolchain/tcc
    ;;
  darwin-amd64)
    curl -fsSL -o toolchain/tcc "https://github.com/vlang/tccbin/raw/master/tcc-darwin-amd64"
    chmod +x toolchain/tcc
    ;;
  *)
    echo "unsupported pair: $OS-$ARCH" >&2
    exit 2
    ;;
esac
echo "fetch-tcc: OK ($OS-$ARCH) -> toolchain/"
