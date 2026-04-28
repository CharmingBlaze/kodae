# Binding generator (`kodae bind` / `kodae-bind`)

Kodae can generate **`extern fn`**, **`pub struct`**, and **`pub enum`** declarations from a C header using Clang’s JSON AST dump (`clang -Xclang -ast-dump=json`).

## Prerequisites

- **Clang** on your `PATH` (LLVM). The generator does not parse headers textually; it uses Clang.

## Usage

```text
kodae bind [-o output.kodae] <shortName> <path/to/header.h>
```

The standalone tool **`kodae-bind`** accepts the same arguments:

```text
kodae-bind [-o output.kodae] <shortName> <path/to/header.h>
```

- **`<shortName>`** — Library label used in generated metadata (e.g. `# link "shortName"`).
- **`header.h`** — Path to the main include file.
- **`-o`** — Optional output path. Default: `include/<shortName>/<shortName>.kodae`.

Put **`-o` before** the two positional arguments (Go flag parsing stops at the first non-flag).

## Example

```sh
kodae bind -o include/raylib/raylib.kodae raylib "C:\raylib\include\raylib.h"
```

## What gets generated

- C **`struct`** → Kodae `pub struct` (layout-oriented types like `f32` / `i32` where needed).
- C **`enum`** → Kodae `pub enum` where mappable.
- C functions → Kodae **`extern fn`**.
- Pointers often map to **`ptr[byte]`**; see table in older commits or generated files for type mapping details.

After generation, use **`# include "shortName"`** (or your chosen path), **`# link "..."`**, and **`# linkpath`** as described in [C_LIBRARIES.md](C_LIBRARIES.md).

## Limitations

Very large headers may skip some declarations; callbacks and complex macros may not translate. Raylib’s checked-in binding is the practical reference.
