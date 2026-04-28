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

**Note:** You do **not** need to install any C toolchain or other dependencies. Everything required is included in the bundle.

## Learning Kodae

| Resource | What it is |
|-----|------------|
| [docs/CLI.md](docs/CLI.md) | **Command-line** — every `kodae` subcommand (`run`, `build`, `check`, `bind`, …), flags, and beginner workflows. |
| [docs/LANGUAGE.md](docs/LANGUAGE.md) | **Full Reference** — the whole language on one page. |
| [examples/README.md](examples/README.md) | **Examples** — a suggested path to learn by doing. |
| [docs/DIRECTIVES.md](docs/DIRECTIVES.md) | **Modules** — how to use `#include` and libraries. |
| [docs/C_LIBRARIES.md](docs/C_LIBRARIES.md) | **C Interop** — calling C functions (like Raylib) from Kodae. |

## Why Kodae?

- **Zero Setup**: Download, extract, and start coding immediately.
- **Lightning Fast**: Compiles directly to optimized C code.
- **Simple Syntax**: A modern, readable syntax that stays out of your way.
- **Truly Portable**: Works exactly the same on Windows, Linux, and Mac.


