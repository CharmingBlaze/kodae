# Kodae command-line reference (`kodae`)

The `kodae` program is the compiler driver: it loads `.kodae` sources, can dump tokens or the AST, type-checks, emits **C99**, and invokes a C toolchain to produce executables or libraries.

Run `kodae` with no arguments (or an unknown command) to print built-in help.

**Related:** [LANGUAGE.md](LANGUAGE.md) (language), [DISTRIBUTION.md](DISTRIBUTION.md) (bundled toolchain), [BINDGEN.md](BINDGEN.md) (`bind`), [SUPPORTED.md](../SUPPORTED.md) (feature status).

---

## Environment variables

| Variable | Meaning |
|----------|---------|
| `KODAE_CC` | Default C compiler (`clang`, `gcc`, full path, or `zig` / `zig cc`). Overridden by `kodae --cc ...`. |
| `KODAE_HOME` | If set, user libraries live under `$KODAE_HOME/libs` (see [DIRECTIVES.md](DIRECTIVES.md)). |

---

## Shared flags (several commands)

Many commands accept one or more `.kodae` files merged in order (same as `build`). Optional flags:

- **`-o`** — output path (meaning depends on the command).
- **`--cc`** — C compiler for compile/link steps.
- **`--ldflags "..."`** — extra linker tokens (split on spaces), e.g. `--ldflags "-lraylib -lm"`.

---

## Commands

### `version` (aliases: `-v`, `-version`)

Prints the `kodae` version string.

---

### `lex` (aliases: `tokenize`, `lexdump`)

**Usage:** `kodae lex <file.kodae>`

Writes one line per token (debugging). Not needed for normal development.

---

### `parse` / `ast`

**Usage:** `kodae parse <a.kodae> [b.kodae] …`

Parses and prints the merged AST (same file merge rules as `build`). Ignores useless `-o` / `--cc` / `--ldflags` for this command.

---

### `check` (alias: `typecheck`)

**Usage:** `kodae check <a.kodae> [b.kodae] …`

Type-checks the merged program. Prints `ok` on success.

---

### `cgen` / `emit` / `c`

**Usage:** `kodae cgen <a.kodae> [b.kodae] …`

Type-checks and prints generated C to **stdout**.

---

### `build`

**Usage:** `kodae build [--lib] [--static] [--shared] <files.kodae> … [-o out] [--cc cc] [--ldflags "..."]`

- Default: link an **executable**. Without `-o`, the output name is derived from the first input file (e.g. `hello.exe` on Windows).
- **`--lib`**: emit C library artifacts (`.c`, `.h`, plus static/shared where applicable); see [LIBRARIES.md](LIBRARIES.md).
- **`--static`** / **`--shared`**: control library outputs in library mode.

---

### `buildc`

**Usage:** `kodae buildc <file.kodae> [-o out.c]`

Writes generated C only (no link). Default output stem matches the source file.

---

### `run`

**Usage:** `kodae run <file.kodae> [--cc cc] [--ldflags "..."]`

Builds then runs the resulting binary with stdin/stdout/stderr connected.

---

### `install`

**Usage:** `kodae install <path/to/file.kodae>` or `kodae install name` (looks for `name.kodae` in the current directory)

Copies the file into the user library directory so `#include "name"` can resolve it from any project.

---

### `bind`

**Usage:** `kodae bind [-o out.kodae] <shortName> <path/to/header.h>`

Generates Kodae bindings from a C header (needs **Clang** on `PATH`). Default output: `include/<shortName>/<shortName>.kodae`. Put `-o` **before** the positional arguments.

Same binding logic exists as the **`kodae-bind`** helper executable.

See [BINDGEN.md](BINDGEN.md).

---

### `bundle`

**Usage:** `kodae bundle [os] [arch]`

Maintainer tool: builds `kodae` with Go, optionally bundles Zig from `PATH`, copies `include/` and `examples/` into `dist/`. Requires Go to run.

---

## Typical workflows

- **Develop:** `kodae run src/main.kodae`
- **Ship:** `kodae build -o game.exe src/main.kodae`
- **CI:** `kodae check …`
- **Inspect C:** `kodae cgen main.kodae > out.c`
- **Headers → Kodae:** `kodae bind raylib path/to/raylib.h`

---

## See also

- [README.md](../README.md) — download and quick start  
- [examples/README.md](../examples/README.md) — runnable samples
