# Cross-platform reliability matrix

This checklist tracks release-grade behavior for the primary targets:

- Windows amd64
- Linux amd64
- macOS arm64

## Baseline command matrix

Core commands expected to pass on each target:

- `kodae check <file>`
- `kodae build <file>` (default backend: LLVM)
- `kodae run <file>` (default backend: LLVM)
- `kodae cgen <file>`
- `kodae buildc <file>`
- `kodae build --backend=c <file>`
- `kodae build --backend=llvm <file>`
- `kodae build --lib ... --backend=c` (library mode)
  - with omitted backend in library mode, wrapper auto-selects C backend.

## Current baseline notes

- Windows local baseline passes for `check/build/run/cgen` and full `go test ./...`.
- Windows LLVM wrapper path now suppresses noisy target-triple override warnings during normal `build`/`run`.
- PowerShell device redirection to `NUL` is not portable for all code paths; use `| Out-Null` for smoke scripts.
- Static library archiver lookup now supports `ar`, `llvm-ar`, or `gcc-ar` for better cross-platform toolchain compatibility.

## CI enforcement

Cross-platform workflow:

- [`.github/workflows/cross-platform.yml`](../.github/workflows/cross-platform.yml)

It runs on:

- `ubuntu-latest`
- `windows-latest`
- `macos-latest`

And validates:

- full `go test ./...`
- wrapper smoke commands across `check/build/run/cgen/buildc`, default LLVM, explicit C/LLVM overrides, and library mode.
