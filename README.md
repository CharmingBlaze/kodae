# Clio

Clio is a small, fast language that compiles to **C99**. It is designed to be simple, with one obvious way to do things, making it perfect for game development and learning.

## Quick Start

### 1. Download
Download the **Portable Bundle** for your platform from the [Releases](https://github.com/CharmingBlaze/clio/releases) page.

- **Windows**: `clio-windows-amd64.zip`
- **Linux**: `clio-linux-amd64.tar.gz`
- **macOS**: `clio-darwin-arm64.tar.gz` (Apple Silicon)

### 2. Extract
Extract the archive to a folder of your choice.

### 3. Run
Open your terminal in that folder and run your first program:

```sh
# Windows
.\bin\clio.exe run examples\hello.clio

# Linux / macOS
./bin/clio run examples/hello.clio
```

**Note:** You do **not** need to install any C toolchain or other dependencies. Everything required is included in the bundle.

## Learning Clio

| Resource | What it is |
|-----|------------|
| [docs/LANGUAGE.md](docs/LANGUAGE.md) | **Full Reference** — the whole language on one page. |
| [examples/README.md](examples/README.md) | **Examples** — a suggested path to learn by doing. |
| [docs/DIRECTIVES.md](docs/DIRECTIVES.md) | **Modules** — how to use `#include` and libraries. |
| [docs/C_LIBRARIES.md](docs/C_LIBRARIES.md) | **C Interop** — calling C functions (like Raylib) from Clio. |

## Why Clio?

- **Zero Setup**: Download, extract, and start coding immediately.
- **Lightning Fast**: Compiles directly to optimized C code.
- **Simple Syntax**: A modern, readable syntax that stays out of your way.
- **Truly Portable**: Works exactly the same on Windows, Linux, and Mac.


