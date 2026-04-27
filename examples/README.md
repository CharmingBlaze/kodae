# Clio examples

All paths are relative to the **repository root**. If `clio` is on your `PATH`, you can `cd` to this folder and use shorter names.

```sh
# From repo root
go run ./cmd/clio run examples/hello.clio
```

## Suggested order (learning)

1. **hello.clio** — one `print`, minimal `main`.  
2. **onepage.clio** — variables, `if` / `while` / `for`, `list`, structs, enums, `match` (same content as the big example in [LANGUAGE.md](../docs/LANGUAGE.md)).  
3. **list_basic.clio** — `list[T]`, `push` / `pop` / `append` / `remove`, `len`.  
4. **features.clio** — `defer`, `continue`, `and` / `or`, `++` / `--`, exhaustive `match` on an enum.  
5. **strings.clio** — `+` vs `$` interpolation in string literals.  
6. **include/main.clio** — `#include "helpers"` and **`pub`** for anything used across files.  
7. **multi/app.clio** (run from `examples/multi/`) — the legacy **`use name`** form loading `lib.clio` in the same directory.  
8. **extern_hello.clio** — `extern fn` and a C `printf` call.  
9. **raylib_minimal.clio** — few hand-written `extern` lines; needs native Raylib.  
10. **raylib_game.clio** — `#include` the large generated binding in `include/raylib/raylib.clio` (also under `examples/libs/raylib/` for the include path).  
11. **result_minimal.clio** — `catch` style.  
12. **textrpg.clio** — larger sample.  

## Full index

| File / folder | What it shows |
|---------------|---------------|
| [hello.clio](hello.clio) | “Hello” + variable |
| [onepage.clio](onepage.clio) | Single-file language tour |
| [features.clio](features.clio) | Control flow and enum `match` extras |
| [list_basic.clio](list_basic.clio) | `list[int]` and methods |
| [structtest.clio](structtest.clio) | Struct literals, field update, `==` on structs |
| [strings.clio](strings.clio) | String usage |
| [include/](include/) | **`#include`**, `pub fn` / `pub struct` across two `.clio` files |
| [multi/](multi/) | **`use lib`**, `pub` on the shared `double` in `lib.clio` |
| [extern_hello.clio](extern_hello.clio) | C interop (`extern fn` + `printf`) |
| [raylib_minimal.clio](raylib_minimal.clio) | Minimal Raylib `extern` set |
| [raylib_game.clio](raylib_game.clio) | Many Raylib functions via generated [include/raylib/raylib.clio](../include/raylib/raylib.clio) |
| [result_minimal.clio](result_minimal.clio) | `catch` |
| [textrpg.clio](textrpg.clio) | Larger game-style script |

## Multi-file commands

`#include` (see [DIRECTIVES.md](../docs/DIRECTIVES.md)):

```sh
clio run examples/include/main.clio
clio check examples/include/main.clio
```

`use` (same-directory only; see `examples/multi`):

```sh
cd examples/multi
clio run app.clio
# or, from repo root, pass both files in order:
clio build lib.clio app.clio
```

## Installing a reusable `.clio` library (optional)

To put a `mathlib.clio` into the user lib directory for `#include "mathlib"` from any project:

```sh
clio install path/to/mathlib.clio
# uses $CLIO_HOME/libs, or ~/.clio/libs/ by default
```

Documented in [DIRECTIVES.md](../docs/DIRECTIVES.md).
