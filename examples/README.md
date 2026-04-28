To run these examples, use the `kodae` command from your portable bundle. The full command reference (every subcommand and flag) is in **[docs/CLI.md](../docs/CLI.md)**.

```sh
# Example:
./bin/kodae run examples/hello.kodae
```


## Suggested order (learning)

1. **Variables & Math** — `print`, variables, math operators
2. **Control Flow** — `if` / `else`, `while`, `for`
3. **Functions** — functions, tuples, default params
4. **Data Structures** — lists and list methods
5. **Custom Types** — structs, methods, `this` keyword
6. **Enums** — enums and `match`
7. **File I/O** — files, JSON, save systems
8. **Networking** — HTTP, WebSockets, online scores
9. **Game Dev** — games with Raylib
10. **C Interop** — C libraries, `pub`, `#include`, `build --lib`

## Full index

| File / folder | What it shows |
|---------------|---------------|
| [hello.kodae](hello.kodae) | “Hello” + variable |
| [onepage.kodae](onepage.kodae) | Single-file language tour |
| [features.kodae](features.kodae) | Control flow and enum `match` extras |
| [list_basic.kodae](list_basic.kodae) | `list[int]` and methods |
| [structtest.kodae](structtest.kodae) | Struct literals, field update, `==` on structs |
| [strings.kodae](strings.kodae) | String usage |
| [include/](include/) | **`#include`**, `pub fn` / `pub struct` across two `.kodae` files |
| [multi/](multi/) | **`use lib`**, `pub` on the shared `double` in `lib.kodae` |
| [extern_hello.kodae](extern_hello.kodae) | C interop (`extern fn` + `printf`) |
| [raylib_minimal.kodae](raylib_minimal.kodae) | Minimal Raylib `extern` set |
| [raylib_game.kodae](raylib_game.kodae) | Many Raylib functions via generated [include/raylib/raylib.kodae](../include/raylib/raylib.kodae) |
| [result_minimal.kodae](result_minimal.kodae) | `catch` |
| [textrpg.kodae](textrpg.kodae) | Larger game-style script |
| [stdlib_v2_test.kodae](stdlib_v2_test.kodae) | **New built-ins**: bitwise, `repeat`, `sort`, `reverse`, `swap`, etc. |

## Multi-file commands

`#include` (see [DIRECTIVES.md](../docs/DIRECTIVES.md)):

```sh
kodae run examples/include/main.kodae
kodae check examples/include/main.kodae
```

`use` (same-directory only; see `examples/multi`):

```sh
cd examples/multi
kodae run app.kodae
# or, from repo root, pass both files in order:
kodae build lib.kodae app.kodae
```

## Installing a reusable `.kodae` library (optional)

To put a `mathlib.kodae` into the user lib directory for `#include "mathlib"` from any project:

```sh
kodae install path/to/mathlib.kodae
# uses $KODAE_HOME/libs, or ~/.kodae/libs/ by default
```

Documented in [DIRECTIVES.md](../docs/DIRECTIVES.md).
