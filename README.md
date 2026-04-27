# Clio

A small language that compiles to **C99** — one obvious way to do most things, with a Go-based compiler in this repository.

## Quick start

From a clone of this repo:

```sh
go run ./cmd/clio run examples/hello.clio
```

Build the `clio` tool and put it on your `PATH` if you prefer:

```sh
go build -o clio ./cmd/clio
./clio run examples/onepage.clio
./clio check .
```

(Use a concrete `.clio` path with `clio check`; the compiler merges multiple files when you list them.)

**Environment:** you need a C toolchain (`gcc`, `clang`, or a bundled `zig` — see [docs/DISTRIBUTION.md](docs/DISTRIBUTION.md)). Set `CLIO_CC` to choose a compiler, or use `clio build --cc <cmd> …`.

## Documentation

| Doc | What it is |
|-----|------------|
| [docs/LANGUAGE.md](docs/LANGUAGE.md) | **Start here** — the language on one page, with a learning path. |
| [docs/DIRECTIVES.md](docs/DIRECTIVES.md) | `#include`, `#library`, `#link`, and where files are found. |
| [docs/C_LIBRARIES.md](docs/C_LIBRARIES.md) | `extern fn`, Raylib-style linking. |
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

