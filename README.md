# Kodae

Kodae is a small, fast language that compiles to **C99**. It is designed to be simple, with one obvious way to do things, making it perfect for game development and learning.

## Quick Start

### 1. Download
Download the **Portable Bundle** for your platform from the [Releases](https://github.com/CharmingBlaze/kodae/releases) page.

- **Windows**: `kodae-windows-amd64.zip`
- **Linux**: `kodae-linux-amd64.tar.gz`
- **macOS**: `kodae-darwin-arm64.tar.gz` (Apple Silicon)

### 2. Extract
Extract the archive to a folder of your choice.

### 3. Run
Open your terminal in that folder and run your first program:

```sh
# Windows
.\bin\kodae.exe run examples\hello.kodae

# Linux / macOS
./bin/kodae run examples/hello.kodae
```

**Note:** Release bundles can include **`toolchain/tcc`** so the compatibility **C backend** (`--backend=c`) works with no system compiler. The default backend is **LLVM** for `kodae build` / `kodae run`, which requires **clang** on `PATH`. Use **`kodae build --backend=c`** to force the C path; library mode (`--lib` or `#mode \"library\"`) auto-selects C backend when `--backend` is omitted. Use **`kodae build --release`** to skip sidecar TCC and prefer system toolchains. See [docs/DISTRIBUTION.md](docs/DISTRIBUTION.md).

## Learning Kodae

| Resource | What it is |
|-----|------------|
| [docs/CLI.md](docs/CLI.md) | **Command-line** — every `kodae` subcommand (`run`, `build`, `check`, `bind`, …), flags, and beginner workflows. |
| [docs/LANGUAGE.md](docs/LANGUAGE.md) | **Full Reference** — the whole language on one page. |
| [examples/README.md](examples/README.md) | **Examples** — a suggested path to learn by doing. |
| [docs/DIRECTIVES.md](docs/DIRECTIVES.md) | **Modules** — how to use `#include` and libraries. |
| [docs/C_LIBRARIES.md](docs/C_LIBRARIES.md) | **C Interop** — calling C functions (like Raylib) from Kodae. |
| [SUPPORTED.md](SUPPORTED.md) | **Implementation status** — what the compiler supports today. |
| [docs/CROSS_PLATFORM.md](docs/CROSS_PLATFORM.md) | **Cross-platform reliability** — matrix and CI checks for Windows/Linux/macOS. |
| [docs/README.md](docs/README.md) | **Documentation index** — all guides in one table. |

## Why Kodae?

- **Small footprint**: Download and extract; add a system C compiler (or use the experimental LLVM IR path).
- **Lightning Fast**: Compiles directly to optimized C code.
- **Simple Syntax**: A modern, readable syntax that stays out of your way.
- **Truly Portable**: Works exactly the same on Windows, Linux, and Mac.


