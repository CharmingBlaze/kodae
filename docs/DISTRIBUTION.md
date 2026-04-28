# Kodae Distribution Guide

This guide shows how to ship Kodae so end users do not install toolchains manually.

## Goal

Distribute one archive per platform:

- Windows: zip containing `bin/kodae.exe` and `toolchain/zig/zig.exe`
- Linux/macOS: tar.gz containing `bin/kodae` and `toolchain/zig/zig`

End users extract and run `bin/kodae`.

## Build Kodae Binary

Build per target platform with Go:

- Windows: `GOOS=windows GOARCH=amd64 go build -o kodae.exe ./cmd/kodae`
- Linux: `GOOS=linux GOARCH=amd64 go build -o kodae ./cmd/kodae`
- macOS: `GOOS=darwin GOARCH=arm64 go build -o kodae ./cmd/kodae`

## Package Portable Bundle

Use the helper scripts after building `kodae` and obtaining a Zig binary for the same platform.

### Windows

```powershell
scripts/package-portable.ps1 -KodaeBinary .\kodae.exe -ZigBinary .\zig.exe -Platform windows-amd64
```

### Linux/macOS

```sh
scripts/package-portable.sh ./kodae ./zig linux-amd64
scripts/package-portable.sh ./kodae ./zig darwin-arm64
```

## Runtime Behavior

`kodae` automatically detects bundled Zig at:

- `toolchain/zig/zig.exe` (Windows)
- `toolchain/zig/zig` (Linux/macOS)

This means `kodae build`, `kodae run`, and `kodae build --lib` work with no global compiler installation.

## Overrides

- Use `--cc` to force a specific compiler.
- Use `KODAE_CC` for environment-level override.

Both overrides take precedence over bundled Zig.
