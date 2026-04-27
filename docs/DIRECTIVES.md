# Directives

All of these are **file-level** lines starting with `#` followed by a name and a string (or, for `link`, a linker argument string). Line comments use `'` or `--`.

**Runnable sample:** a two-file program using `#include` and `pub` is in the repo at `examples/include/` (run `clio run examples/include/main.clio` from the project root). The older `use name` form is in `examples/multi/`.

| Directive | Meaning |
|-----------|---------|
| `#library "name"` | Declares that this translation unit is a **named library** (used with `clio build --lib` and metadata like generated headers). |
| `#version "1.0.0"` | Optional: version string (stored in compiler metadata for `--lib` builds). |
| `#author "name"` | Optional: author string (metadata). |
| `#include "path"` | **Merge** another Clio file into the program. Resolution order is below. Each file is **included at most once** (repeated includes are ignored). |
| `#link "…"` | Pass **linker** flags: bare names (e.g. `"raylib"`) become `-lraylib`, paths and `-L` pass through. See [C_LIBRARIES.md](C_LIBRARIES.md). |

A minimal shareable library source file often starts like:

```clio
#library "mathlib"
#version "1.0.0"
#author "Ada"

pub fn square(x: int) -> int {
  return x * x
}
```

`#mode "library"` is an alternative way to set library code generation; `#library` also supplies the public name for artifacts and headers.

## Where `#include` looks (in order)

1. The **same directory** as the file that contains the `#include`.
2. **`./libs/`** under that directory (e.g. `…/src/libs/foo.clio` for `#include "foo"`).
3. The **user library directory**: `$CLIO_HOME/libs` if `CLIO_HOME` is set, otherwise `~/.clio/libs/` (or the platform’s user-home equivalent). Use `clio install` to copy a `.clio` file there so any project can `#include "name"` by stem.

If none of the candidates exist, you get a **file not found** error. If you meant a **C** API only (no `.clio` wrapper), use `extern` + `# link` and a small Clio file as needed.

## Public vs private (multi-file)

Symbols from another file are only visible in your file if they are **exported** with `pub` (`pub fn`, `pub struct`, `pub enum`). The same name in the same file is always in scope. Private `fn` / `struct` / `enum` are limited to the **defining** `.clio` file. `#library` is optional for ordinary project files; it matters mainly when building a shareable C library with `clio build --lib`.

## Installing a source library for other projects

After you build a distributable (or you only ship `.clio` sources), consumers can add the file to the user lib dir:

```sh
clio install mathlib.clio
# or, from a directory that contains mathlib.clio:
clio install mathlib
```

Then in their project:

```clio
#include "mathlib"

fn main() {
  let n = square(5)
  print("$n")
}
```

For producing static/shared C artifacts and headers, see [LIBRARIES.md](LIBRARIES.md) (`clio build --lib`).

## Related

- [LANGUAGE.md](LANGUAGE.md) — language overview.  
- [C_LIBRARIES.md](C_LIBRARIES.md) — `# link`, `# linkpath`, `extern fn`.  
- [LIBRARIES.md](LIBRARIES.md) — `--lib` output and C interop.  

The legacy **`use name`** form still loads `name.clio` from the **same directory** only; prefer `#include` for resolution across `./libs` and the user lib directory.
