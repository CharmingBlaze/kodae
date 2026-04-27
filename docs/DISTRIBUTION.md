# Clio Distribution Guide

This guide shows how to ship Clio so end users do not install toolchains manually.

## Goal

Distribute one archive per platform:

- Windows: zip containing `bin/clio.exe` and `toolchain/zig/zig.exe`
- Linux/macOS: tar.gz containing `bin/clio` and `toolchain/zig/zig`

End users extract and run `bin/clio`.

## Build Clio Binary

Build per target platform with Go:

- Windows: `GOOS=windows GOARCH=amd64 go build -o clio.exe ./cmd/clio`
- Linux: `GOOS=linux GOARCH=amd64 go build -o clio ./cmd/clio`
- macOS: `GOOS=darwin GOARCH=arm64 go build -o clio ./cmd/clio`

## Package Portable Bundle

Use the helper scripts after building `clio` and obtaining a Zig binary for the same platform.

### Windows

```powershell
scripts/package-portable.ps1 -ClioBinary .\clio.exe -ZigBinary .\zig.exe -Platform windows-amd64
```

### Linux/macOS

```sh
scripts/package-portable.sh ./clio ./zig linux-amd64
scripts/package-portable.sh ./clio ./zig darwin-arm64
```

## Runtime Behavior

`clio` automatically detects bundled Zig at:

- `toolchain/zig/zig.exe` (Windows)
- `toolchain/zig/zig` (Linux/macOS)

This means `clio build`, `clio run`, and `clio build --lib` work with no global compiler installation.

## Overrides

- Use `--cc` to force a specific compiler.
- Use `CLIO_CC` for environment-level override.

Both overrides take precedence over bundled Zig.
