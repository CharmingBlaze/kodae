# Clio

A small language that compiles to **C99** — one obvious way to do most things, with a Go-based compiler in this repository.

## Quick Start

### 1. Download & Run (No Toolchain Needed)
The easiest way to use Clio is to download a **Portable Bundle** which includes the compiler and a built-in C toolchain (Zig).

1. **Download** the latest release for Windows, Linux, or macOS from the [Releases](https://github.com/CharmingBlaze/clio/releases) page.
2. **Extract** the archive to any folder.
3. **Run** an example:
   ```sh
   # Windows
   .\bin\clio.exe run examples\hello.clio

   # Linux / macOS
   ./bin/clio run examples/hello.clio
   ```

### 2. From Source (For Developers)
If you have Go 1.21+ installed, you can run directly from the repository:

```sh
go run ./cmd/clio run examples/hello.clio
```

Or build and install to your PATH:

```sh
go build -o clio ./cmd/clio
./clio run examples/onepage.clio
```

**Note:** When running from source, you need a C compiler (`clang`, `gcc`, or `zig`) on your PATH. The Portable Bundles above handle this for you automatically.


## Documentation

| Doc | What it is |
|-----|------------|
| [docs/LANGUAGE.md](docs/LANGUAGE.md) | **Start here** — the language on one page, with a learning path. |
| [docs/DIRECTIVES.md](docs/DIRECTIVES.md) | `#include`, `#library`, `#link`, and where files are found. |
| [docs/C_LIBRARIES.md](docs/C_LIBRARIES.md) | `extern fn`, `f32` (C `float`) in `extern`, Raylib linking, generated [include/raylib/raylib.clio](include/raylib/raylib.clio). |
| [docs/BINDGEN.md](docs/BINDGEN.md) | **C binding generator** — `clio bind <lib> <header.h>`. |
| [docs/LIBRARIES.md](docs/LIBRARIES.md) | `clio build --lib`, headers and static libraries. |
| [docs/DISTRIBUTION.md](docs/DISTRIBUTION.md) | Portable bundles and the compiler driver. |
| [SUPPORTED.md](SUPPORTED.md) | Feature checklist vs the internal build-spec. |
| [examples/README.md](examples/README.md) | **Runnable programs** and suggested order. |

## Examples (short)

| Run | Notes |
|-----|--------|
| `clio run examples/hello.clio` | Smallest program |
| `clio run examples/onepage.clio` | One-file tour of the language |
| `clio run examples/include/main.clio` | `#include` and `pub` across files |
| `clio run examples/multi/app.clio` | The older `use lib` form (same idea) |

See [examples/README.md](examples/README.md) for the full list.

## Repository layout

- `cmd/clio` — the compiler driver (`build`, `check`, `run`, `install`, etc.)
- `internal/*` — lexer, parser, type checker, C codegen
- `docs/` — language and tooling docs; `build-spec.js` is a reference for implementers
- `examples/` — small programs you can run locally

