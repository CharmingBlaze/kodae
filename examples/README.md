To run these examples, use the `kodae` command from your portable bundle:

```sh
# Example:
./bin/kodae run examples/hello.kodae
```


## Suggested order (learning)

1. **hello.kodae** — one `print`, minimal `main`.  
2. **onepage.kodae** — variables, `if` / `while` / `for`, `list`, structs, enums, `match` (same content as the big example in [LANGUAGE.md](../docs/LANGUAGE.md)).  
3. **list_basic.kodae** — `list[T]`, `push` / `pop` / `append` / `remove`, `len`.  
4. **features.kodae** — `defer`, `continue`, `and` / `or`, `++` / `--`, exhaustive `match` on an enum.  
5. **strings.kodae** — `+` vs `$` interpolation in string literals.  
6. **include/main.kodae** — `#include "helpers"` and **`pub`** for anything used across files.  
7. **multi/app.kodae** (run from `examples/multi/`) — the legacy **`use name`** form loading `lib.kodae` in the same directory.  
8. **extern_hello.kodae** — `extern fn` and a C `printf` call.  
9. **raylib_minimal.kodae** — few hand-written `extern` lines; needs native Raylib.  
10. **raylib_game.kodae** — `#include` the large generated binding in `include/raylib/raylib.kodae` (also under `examples/libs/raylib/` for the include path).  
11. **result_minimal.kodae** — `catch` style.  
12. **textrpg.kodae** — larger sample.  
13. **stdlib_v2_test.kodae** — full test of the new standard library features (bitwise, repeat, sort, etc).

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
