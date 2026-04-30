# Kodae Distribution Guide

## Goal

Ship one archive per platform containing **`bin/kodae`**, **`toolchain/`** (TinyCC for zero-install `run`/`build`), plus **`include/`**, **`examples/`**, and **`README.md`**.

End users unzip and run **`kodae run examples/hello.kodae`** without installing Go, Clang, or MSVC. Optional **`kodae build --release`** skips sidecar TCC and uses **clang/gcc on `PATH`** for optimized builds.

## Build Kodae Binary

Build per target platform with Go:

- Windows: `GOOS=windows GOARCH=amd64 go build -o kodae.exe ./cmd/kodae`
- Linux: `GOOS=linux GOARCH=amd64 go build -o kodae ./cmd/kodae`
- macOS: `GOOS=darwin GOARCH=arm64 go build -o kodae ./cmd/kodae`

## TinyCC (TCC) for portable bundles

Before creating a bundle, populate **`toolchain/`** at the repo root (see [toolchain/README.md](../toolchain/README.md)):

```sh
chmod +x scripts/fetch-tcc.sh
./scripts/fetch-tcc.sh windows amd64
./scripts/fetch-tcc.sh linux amd64
./scripts/fetch-tcc.sh darwin arm64
```

GitHub Actions **Distribution** workflow runs this step before `go run ./cmd/kodae bundle`.

Release layout:

```text
kodae-<os>-<arch>/
  bin/kodae[.exe]
  toolchain/tcc[.exe]
  include/
  examples/
  README.md
```

`kodae` looks for **`../toolchain/tcc`** relative to the executable (i.e. next to `bin/`). Override behavior:

- **`KODAE_NO_SIDECAR_TCC=1`** — ignore bundled TCC and use `PATH` / **`KODAE_CC`** / **`--cc`**.
- **`kodae build --release`** / **`kodae run --release`** — skip sidecar TCC and prefer **clang**/**gcc** on `PATH`.

## Package Portable Bundle

From the repo root (after optional `scripts/fetch-tcc.sh`):

```sh
go run ./cmd/kodae bundle
```

Optional cross-target arguments: `go run ./cmd/kodae bundle linux amd64`

This creates `dist/kodae-<os>-<arch>/` with `bin/`, `toolchain/` (if present in the repo), `include/`, `examples/`, and `README.md`.

## Running user programs (`kodae run` / `kodae build`)

Default **C backend**: emits C99 then compiles with **sidecar TCC** when shipped, otherwise **clang/gcc/cc** on `PATH`, or **`KODAE_CC`** / **`--cc`**.

### Experimental LLVM IR backend

```sh
kodae build --backend=llvm -o hello examples/hello.kodae
```

This lowers a **supported subset** of the program to LLVM IR, links a runtime bridge with **clang**, and emits an executable. Full AST coverage is still growing.

## Overrides

- **`--cc`** — force a specific C compiler (C backend only).
- **`KODAE_CC`** — same at the environment level.

## Licensing

TinyCC is LGPL. Include appropriate notices with release archives.

## Roadmap

Stronger LLVM integration (in-process LLVM/lld) remains optional; the portable story is **Go compiler + emitted C + bundled TCC**.
