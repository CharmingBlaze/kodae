# Kodae command-line reference

This page explains **every** `kodae` subcommand: what it does, **when to use it**, and **how** to type the commands. If you are new, read [Quick start for beginners](#quick-start-for-beginners) first, then skim the [command index](#command-index).

**Related docs**

| You want to… | Read |
|--------------|------|
| Learn the language | [LANGUAGE.md](LANGUAGE.md) |
| Work through examples | [examples/README.md](../examples/README.md) |
| `#include`, `use`, `kodae install` | [DIRECTIVES.md](DIRECTIVES.md) |
| Call C (Raylib, `extern fn`) | [C_LIBRARIES.md](C_LIBRARIES.md) |
| Build a `.a` / `.so` / `.dll` from Kodae | [LIBRARIES.md](LIBRARIES.md) |
| Package the compiler for others | [DISTRIBUTION.md](DISTRIBUTION.md) |
| Generate bindings from C headers | [BINDGEN.md](BINDGEN.md) |

---

## What is `kodae`?

`kodae` is the **Kodae compiler driver**. It reads `.kodae` source files, can show you the tokenizer output or the syntax tree, type-check your program, emit C99, and **invoke a C compiler** to produce an executable or a library.

- You run it in a **terminal** (PowerShell, `cmd`, or a Unix shell).
- On Windows, the program is often named `kodae.exe`. Examples below use `kodae`; add `.exe` on Windows if your shell requires it.
- If you use the [portable bundle](https://github.com/CharmingBlaze/kodae/releases), put `bin` on your `PATH` or call the binary with a path, e.g. `.\bin\kodae.exe` (Windows) or `./bin/kodae` (Linux/macOS).

**Usage pattern**

```text
kodae <command> [arguments and flags]
```

Run `kodae` with **no** command (or an unknown command) to print a short built-in help list.

---

## Quick start for beginners

1. **Run a program in one step** — compiles to a temporary build, then runs the result:

   ```sh
   kodae run path/to/hello.kodae
   ```

2. **Build an executable** without running it — writes a binary next to your file (name = base name of the first `.kodae` file, e.g. `hello.exe` on Windows, `hello` on Linux/macOS):

   ```sh
   kodae build path/to/hello.kodae
   ```

3. **Custom output name**:

   ```sh
   kodae build -o mygame path/to/main.kodae
   ```

4. **Check that your code type-checks** (no binary produced):

   ```sh
   kodae check path/to/main.kodae
   ```

5. **Link a C library** (e.g. games): add `# link` / `# linkpath` in your source (see [C_LIBRARIES.md](C_LIBRARIES.md)) and pass extra linker flags if needed:

   ```sh
   kodae run examples/raylib_minimal.kodae --ldflags "-lraylib"
   ```

If `build` or `run` fails with “no C compiler”, install LLVM/Clang, GCC, or [Zig](https://ziglang.org/download/), **or** use a release bundle that includes a C toolchain under `toolchain/zig/`. See [Environment variables](#environment-variables) and [DISTRIBUTION.md](DISTRIBUTION.md).

---

## Environment variables

| Variable | Purpose |
|----------|---------|
| `KODAE_CC` | Choose which C compiler `kodae` uses for `build`, `run`, `buildc`, and `build --lib`. Can be a full path, or a name on `PATH` such as `clang` or `zig`. `zig` is treated as `zig cc`. **Precedence:** `kodae --cc ...` wins over `KODAE_CC`, which wins over automatic discovery. |
| `KODAE_HOME` | If set, user-installed Kodae libraries go to `$KODAE_HOME/libs`. Otherwise the default is `~/.kodae/libs` (or the platform equivalent). Used by `kodae install` and `#include` resolution. See [DIRECTIVES.md](DIRECTIVES.md). |

---

## Shared flags: multiple files, output, and linking

Several commands share the same idea of “**build flags**” as `kodae build`:

- **Input files:** one or more `.kodae` paths. They are merged into **one** program in the order given (and dependencies from `#include` / `use` are loaded first, as in a normal build).
- **`-o <path>`** — output path. Meaning depends on the command (executable, `.c` file, or library artifact).
- **`--cc <command>`** — C compiler to use (overrides `KODAE_CC` when provided).
- **`--ldflags "..."` —** extra arguments for the linker (split on spaces), e.g. `--ldflags "-lraylib -lm"`.

**Only `kodae build`** also accepts:

- **`--lib`** — build a Kodae **library** (emits `.c` + `.h`, compiles to `.o`, and produces static and/or shared library artifacts; see [LIBRARIES.md](LIBRARIES.md)).
- **`--static`** / **`--shared`** — control which library kinds are produced in library mode (if neither is given, the tool may still produce both; see implementation in the repo).

Commands **`parse`**, **`check`**, and **`cgen`** accept the same file list and `-o` / `--cc` / `--ldflags` **parsing** as `build`, but **ignore** output and linker details where they do not apply (parsing/printing only).

---

## Command index

| Command | One-line purpose |
|---------|------------------|
| [`version`](#kodae-version) | Print compiler version. |
| [`lex`](#kodae-lex) | Print every token (debugging lexer). |
| [`parse` / `ast`](#kodae-parse--ast) | Parse and print the merged AST. |
| [`check`](#kodae-check) | Type-check; prints `ok` on success. |
| [`cgen` / `emit`](#kodae-cgen--emit) | Print generated C to stdout. |
| [`build`](#kodae-build) | Compile to an executable or (with `--lib`) a C library. |
| [`buildc`](#kodae-buildc) | Write generated C to a file only (no link step). |
| [`run`](#kodae-run) | Build, then run the default binary. |
| [`install`](#kodae-install) | Copy a `.kodae` file into the user lib dir for `#include`. |
| [`bind`](#kodae-bind) | Generate Kodae bindings from a C header (needs Clang). |
| [`bundle`](#kodae-bundle) | Maintainer: package `kodae` + examples into `dist/` (requires Go). |

There is also a standalone tool **`kodae-bind`** with the same binding behavior as `kodae bind`.

---

### `kodae version`

**Aliases:** `-v`, `-version`  
**What it does:** Prints the Kodae version string (for bug reports and tutorials).

```sh
kodae version
```

---

### `kodae lex`

**Aliases:** `tokenize`, `lexdump`  
**What it does:** Reads a **single** `.kodae` file and prints **one line per token**: numeric type id, kind name, quoted literal, line, column.

**Why:** Debugging the lexer or learning how the source is split into tokens. **Not** needed for normal development.

```sh
kodae lex myfile.kodae
```

---

### `kodae parse` / `ast`

**What it does:** Loads the same merged program as `build` (respecting `#include` / `use`), then prints the **abstract syntax tree** to stdout.

**Why:** Compiler hacking, debugging syntax issues, or understanding how declarations are represented.

```sh
kodae parse src/main.kodae
kodae ast src/a.kodae src/b.kodae
```

---

### `kodae check`

**Aliases:** `typecheck`  
**What it does:** Runs the full pipeline **through type-checking** only. Prints `ok` and exits **0** if there are no errors.

**Why:** Fast feedback in editors or CI without generating C or binaries.

```sh
kodae check game.kodae
kodae check part1.kodae part2.kodae
```

---

### `kodae cgen` / `emit`

**Aliases:** `c`, `emit` (see `kodae` help; `c` is shorthand)

**What it does:** Type-checks, then prints the **generated C99** for the merged program to **stdout**.

**Why:** Inspect what Kodae emits, pipe to a file, or feed another tool.

```sh
kodae cgen main.kodae > out.c
```

---

### `kodae build`

**What it does:**

1. Loads and merges sources (like other commands).
2. Type-checks.
3. Emits C and invokes the C toolchain:
   - **Default:** link an **executable**. If `-o` is omitted, the output name is the **basename** of the first `.kodae` file, with `.exe` on Windows (avoids a fixed `a.out` that can get locked by a running process).
   - **`--lib`:** library mode — writes `.c`/`.h`, compiles, and builds archive/shared artifacts. Library metadata can also come from `#library` / `#mode library` in source (see [LIBRARIES.md](LIBRARIES.md)).

**Typical examples:**

```sh
kodae build app.kodae
kodae build -o bin/game main.kodae
kodae build main.kodae extras.kodae --ldflags "-lfoo"
kodae build --lib mylib.kodae
```

---

### `kodae buildc`

**What it does:** Type-checks and writes **only** the generated C to a file. **Does not** run the linker.

- If `-o` is omitted, the output defaults to the first file’s path with the extension replaced by `.c` (e.g. `main.kodae` → `main.c`).

**Why:** You want to hand the C to another build system, audit the output, or compile with custom flags outside `kodae`.

```sh
kodae buildc sketch.kodae -o sketch_generated.c
```

---

### `kodae run`

**What it does:** Builds an executable using the **default output name** (same rule as `build` when `-o` is omitted), then **runs** that executable with stdin/stdout/stderr connected to your terminal.

**Why:** The fastest way to try a program while learning.

```sh
kodae run examples/hello.kodae
kodae run --cc clang myapp.kodae
```

---

### `kodae install`

**What it does:** Copies a **library** `.kodae` file into the [user library directory](DIRECTIVES.md) so any project can resolve `#include "name"` (by stem) without copying files by hand.

**Arguments:**

- `kodae install /path/to/mathlib.kodae`, or
- `kodae install mathlib` — looks for `mathlib.kodae` in the **current working directory**.

On success, it prints the destination path.

**Why:** Share small Kodae-only libraries across projects. See [DIRECTIVES.md](DIRECTIVES.md) for include search order.

```sh
kodae install mylib.kodae
kodae install coolutils
```

---

### `kodae bind`

**What it does:** Calls the **binding generator** to read a C header and write a `.kodae` file with `struct`/`enum`/`extern fn` declarations. **Requires Clang** (LLVM) on your system; see [BINDGEN.md](BINDGEN.md).

**Syntax:**

```text
kodae bind [-o output.kodae] <name> <path/to/header.h>
```

- **`<name>`** — Short name for the library; used in generated metadata (e.g. `# link`).
- **`<path/to/header.h>`** — Path to the main header.
- **`-o`** — Optional output path. Default: `include/<name>/<name>.kodae` (relative to the current directory when you run the command).

**Why:** Avoid hand-writing hundreds of `extern fn` lines for large C APIs (games, SDL, etc.).

```sh
kodae bind raylib "C:\raylib\include\raylib.h"
kodae bind -o third_party/sqlite3.kodae sqlite3 /usr/include/sqlite3.h
```

**Flag order:** Put `-o` **before** `<name>` and the header path. Go’s flag parser stops at the first non-flag argument, so `-o` after the positional paths may not apply.

The repository also provides **`kodae-bind`**, a separate executable with the same interface (useful when packaging or scripting).

---

### `kodae bundle`

**What it does:** **For maintainers** building from source: cross-compiles the `kodae` binary with `go build`, optionally copies **Zig** from your `PATH` if the target OS/arch matches the **host**, copies `include/` and `examples/`, and writes a tree under `dist/kodae-<os>-<arch>/`.

**Arguments (optional):** `kodae bundle [os] [arch]` — e.g. `kodae bundle linux amd64`. Defaults to the current machine’s `GOOS` / `GOARCH`.

**Why:** Produce a directory you can zip and ship; end-user story is in [DISTRIBUTION.md](DISTRIBUTION.md). You need the **Go** toolchain in `PATH` to run this.

```sh
kodae bundle
kodae bundle windows amd64
```

---

## Which command should I use? (flowchart in words)

- **Learning / running small programs** → `kodae run file.kodae`
- **Shipping a game or app** → `kodae build -o ...` (plus `# link` in source as needed)
- **CI or “is my code valid?”** → `kodae check`
- **Curious about the C output** → `kodae cgen` or `kodae buildc`
- **New C bindings** → `kodae bind` + [BINDGEN.md](BINDGEN.md)
- **Puzzling parse errors** → `kodae parse` (AST) or `kodae lex` (tokens)

---

## Common errors (beginners)

| Message / situation | What to do |
|---------------------|------------|
| No C compiler found | Install Clang/GCC/Zig, set `KODAE_CC`, or use a bundle with `toolchain/zig/`. |
| `bind` / headers fail | Install Clang; `kodae bind` uses the Clang JSON AST dump. |
| Include not found | Check paths in [DIRECTIVES.md](DIRECTIVES.md); use `kodae install` for user-wide `.kodae` libs. |
| Mixed non-`.kodae` and `.kodae` args | Positional args should be all `.kodae` paths (or a single non-kodae path in some legacy cases — prefer listing `.kodae` files explicitly). |

---

## See also

- Root [README.md](../README.md) — download and first run  
- [SUPPORTED.md](../SUPPORTED.md) — what the compiler implements today
