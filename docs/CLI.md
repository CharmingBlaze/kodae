# Kodae command-line reference (`kodae`)

The `kodae` program is the compiler driver: it loads `.kodae` sources, can dump tokens or the AST, type-checks, then by default uses **`--backend=llvm`** (lower subset to LLVM IR and invoke **`clang`** to link a runtime bridge). The compatibility **C backend** remains available with **`--backend=c`** (emits **C99** and compiles via sidecar TinyCC or a system C compiler).

Run `kodae` with no arguments (or an unknown command) to print built-in help.

**Related:** [LANGUAGE.md](LANGUAGE.md) (language), [DISTRIBUTION.md](DISTRIBUTION.md) (releases and backends), [BINDGEN.md](BINDGEN.md) (`bind`), [SUPPORTED.md](../SUPPORTED.md) (feature status).

---

## Environment variables

| Variable | Meaning |
|----------|---------|
| `KODAE_CC` | Default C compiler (`clang`, `gcc`, full path, or `zig` / `zig cc` if installed). Overridden by `kodae --cc ...`. Not used by **`--backend=llvm`** (that path looks up **`clang`** on `PATH` only). |
| `KODAE_NO_SIDECAR_TCC` | If set, do not use **`toolchain/tcc`** next to the `kodae` binary; use `PATH` / `KODAE_CC` instead. |
| `KODAE_HOME` | If set, user libraries live under `$KODAE_HOME/libs` (see [DIRECTIVES.md](DIRECTIVES.md)). |

---

## Shared flags (several commands)

Many commands accept one or more `.kodae` files merged in order (same as `build`). Optional flags:

- **`-o`** — output path (meaning depends on the command).
- **`--cc`** — C compiler for compile/link steps.
- **`--ldflags "..."`** — extra linker tokens (split on spaces), e.g. `--ldflags "-lraylib -lm"`.
- **`--backend=llvm`** (default) — type-check, lower supported code to LLVM IR, compile and link with **`clang`** (no C emission for this path).
- **`--release`** — skip sidecar **TCC** and use **clang/gcc/cc** on `PATH` (optimizing toolchain).
- **`--backend=c`** — compatibility path: emit C99 and compile with a C compiler (sidecar **TCC** when available, otherwise `--cc` / `KODAE_CC` / `PATH`).

---

## Commands

### `version` (aliases: `-v`, `-version`)

Prints the `kodae` version string.

**Examples**

```sh
kodae version
kodae -v
```

---

### `lex` (aliases: `tokenize`, `lexdump`)

**Usage:** `kodae lex <file.kodae>`

Writes one line per token (debugging). Not needed for normal development.

**Examples**

```sh
kodae lex examples/hello.kodae
kodae tokenize examples/onepage.kodae
```

---

### `parse` / `ast`

**Usage:** `kodae parse <a.kodae> [b.kodae] …`

Parses and prints the merged AST (same file merge rules as `build`). Ignores useless `-o` / `--cc` / `--ldflags` for this command.

**Examples**

```sh
kodae parse examples/hello.kodae
kodae ast examples/include/helpers.kodae examples/include/main.kodae
```

---

### `check` (alias: `typecheck`)

**Usage:** `kodae check <a.kodae> [b.kodae] …`

Type-checks the merged program. Prints `ok` on success.

**Examples**

```sh
kodae check examples/hello.kodae
kodae typecheck examples/include/helpers.kodae examples/include/main.kodae
```

---

### `cgen` / `emit` / `c`

**Usage:** `kodae cgen <a.kodae> [b.kodae] …`

Type-checks and prints generated C to **stdout**.

**Examples**

```sh
kodae cgen examples/hello.kodae
kodae emit examples/hello.kodae > hello.c
```

---

### `build`

**Usage:** `kodae build [--lib] [--static] [--shared] [--release] [--backend=llvm|c] <files.kodae> … [-o out] [--cc cc] [--ldflags "..."]`

- Default: link an **executable** with **LLVM backend**. Without `-o`, the output name is derived from the first input file (e.g. `hello.exe` on Windows).
- **`--lib`**: emit C library artifacts (`.c`, `.h`, plus static/shared where applicable); see [LIBRARIES.md](LIBRARIES.md). Not supported with **`--backend=llvm`** yet.
  - If `--backend` is omitted and library mode is active (`--lib` or `#mode "library"`), the driver auto-selects the **C backend** for compatibility.
- **`--static`** / **`--shared`**: control library outputs in library mode.
- **`--release`**: prefer **PATH** **clang/gcc** over bundled **TCC** (see [DISTRIBUTION.md](DISTRIBUTION.md)).

**Examples**

```sh
kodae build examples/hello.kodae
kodae build --release -o mygame.exe examples/hello.kodae
kodae build -o dist/mygame examples/textrpg.kodae
kodae build -o hello examples/hello.kodae
kodae build --backend=c -o hello-c examples/hello.kodae
kodae build --cc clang examples/hello.kodae
kodae build --ldflags "-lraylib -lopengl32 -lgdi32 -lwinmm" examples/raylib_minimal.kodae
kodae build --lib examples/include/helpers.kodae
kodae build --lib --shared examples/include/helpers.kodae
```

---

### `buildc`

**Usage:** `kodae buildc <file.kodae> [-o out.c]`

Writes generated C only (no link). Default output stem matches the source file.

**Examples**

```sh
kodae buildc examples/hello.kodae
kodae buildc examples/hello.kodae -o out/hello.generated.c
```

---

### `run`

**Usage:** `kodae run <file.kodae> [--release] [--backend=llvm|c] [--cc cc] [--ldflags "..."]`

Builds then runs the resulting binary with stdin/stdout/stderr connected.

**Examples**

```sh
kodae run examples/hello.kodae
kodae run --backend=c examples/hello.kodae
kodae run --cc clang examples/onepage.kodae
kodae run --ldflags "-lraylib -lopengl32 -lgdi32 -lwinmm" examples/raylib_minimal.kodae
```

---

### `install`

**Usage:** `kodae install <path/to/file.kodae>` or `kodae install name` (looks for `name.kodae` in the current directory)

Copies the file into the user library directory so `#include "name"` can resolve it from any project.

**Examples**

```sh
kodae install libs/net.kodae
kodae install mathlib
```

---

### `bind`

**Usage:** `kodae bind [-o out.kodae] <shortName> <path/to/header.h>`

Generates Kodae bindings from a C header (needs **Clang** on `PATH`). Default output: `include/<shortName>/<shortName>.kodae`. Put `-o` **before** the positional arguments.

Same binding logic exists as the **`kodae-bind`** helper executable.

See [BINDGEN.md](BINDGEN.md).

**Examples**

```sh
kodae bind raylib "C:/raylib/include/raylib.h"
kodae bind -o include/sqlite/sqlite.kodae sqlite "C:/sqlite3/sqlite3.h"
kodae-bind -o include/raylib/raylib.kodae raylib "C:/raylib/include/raylib.h"
```

---

### `bundle`

**Usage:** `kodae bundle [os] [arch]`

Maintainer tool: builds `kodae` with Go and copies `include/`, `examples/`, and **`toolchain/`** (if present — use [scripts/fetch-tcc.sh](../scripts/fetch-tcc.sh)) into `dist/`. Requires Go to run.

**Examples**

```sh
kodae bundle
kodae bundle windows amd64
kodae bundle linux amd64
kodae bundle darwin arm64
```

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
