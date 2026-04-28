# Feature Support Status

This document tracks the implementation status of Kodae features. 

For a beginner-friendly introduction, see `docs/LANGUAGE.md`. Deeper topics: `docs/DIRECTIVES.md` (includes and libraries), and `docs/C_LIBRARIES.md` (C interop).

| Area | Status | Notes |
|------|--------|--------|
| **Single / multi file** | Partial | `kodae build|check|cgen|parse` all accept `a.kodae b.kodae` in order. **`#include "p"`** resolves: same directory → `./libs/p` → `$KODAE_HOME/libs` or `~/.kodae/libs`; deduplicated; cycles error. See [docs/DIRECTIVES.md](docs/DIRECTIVES.md). |
| **kodae install** | Yes | Copies a `.kodae` into the user lib dir (same place as “step 3” in `#include` search) so any project can `#include` by name. |
| **module / use / pub** | Partial | `module` / `pub` are parsed. **`use name`**: loads `name.kodae` from the **same directory** as the current file, before that file’s other top-level decls. Cycles are an error. Cross-file references require `pub` on the defining decl. `module` is not used to resolve paths (v1). |
| **extern fn** | Yes | C signatures; calls use the real C name (no `f_` prefix). stdio `printf` etc. skip redecl (see `<stdio.h>`). `str` → `ptr[byte]` uses `(s).data`. C `float` in `extern` is spelled **`f32`**; return values widen to Kodae’s normal `float` (C `double`). |
| **# link "…"** | Yes | Appended to the C link line after `-lm`. Bare names (e.g. `"raylib"`) become `-lraylib`; tokens starting with `-` pass through. CLI: `--ldflags` / `--ldflags=...`. |
| **# linkpath "dir"** | Yes | Appends `-Ldir` so the linker finds `lib*.a` / import libs. |
| **Int literals** | Yes | Decimal, `0x` hex, and **`0b`** binary literals. |
| **Bitwise** | Yes | `&`, `|`, `^`, `~` operators. |
| **repeat** | Yes | `repeat(n) { ... }` for exactly N iterations. |
| **Link driver** | Yes | `kodae` passes extra argv to the same C compiler line as the object file. |
| **continue** | Yes | In loops. |
| **defer** | Partial | Typechecker: `defer` may only sit at the **top** of a function body (v1). Emitted in reverse on return and at fallthrough end. |
| **and / or** | Yes | Spelled `and` / `or` (same precedence as in the Pratt table). |
| **`catch`** | Yes (simplified) | `catch` is supported as a full `let` init / return value / assignment RHS / expression statement. |
| **`result[T]`, `ok/err`, `.ok/.value/.err`, `?`** | No (removed) | Not part of Kodae v1 surface syntax. Use `catch`. |
| **`T?` optional syntax** | No (removed) | V1 uses implicit nullable behavior with `none`; explicit `T?` is rejected. |
| **`ptr[T]`** | Restricted | Only allowed in `extern fn` signatures. |
| **`pub` exports** | Yes (v1) | `pub fn` / `pub struct` define exported C library API surface for `build --lib`. |
| **`build --lib`** | Yes (v1) | Emits `.c`, `.h`, static (`.a`) and shared (`.so`/`.dll`/`.dylib`) artifacts. |
| **Portable bundle (no global C install)** | Yes | If `toolchain/zig/zig(.exe)` exists next to `kodae`, compiler auto-uses bundled `zig cc`. |
| **`list[T]`** | Yes (v1) | Type syntax `list[T]`, literals `[a, b]`, index read/write `xs[i]`, `len(xs)`, methods `push`, `pop`, `append`, `remove`, and `for (x in list)` iteration. |
| **Standard Library** | Yes (v1) | ~50 built-in functions: Time (`time`, `wait`, `timer`), Random (`random`, `chance`, `random_pick`), Files (`read_file`, `write_file`, `file_exists`), Strings (`.upper`, `.lower`, `.trim`, `.contains`, `.replace`, `.split`), Math (`sqrt`, `abs`, `clamp`, `lerp`, `map`), IO (`print`, `input`, `clear_screen`), OS (`os_name`, `args`, `env`), and JSON placeholders. |
| **`this` / `with` / `fn(){ }` lambdas** | Yes (partial) | **`this`** is lexical in methods (loops, nested blocks, inline **`fn() { }`**). **`base with { field: expr }`** copies a struct and overrides fields. Lambdas are **`fn()` void only**, **no parameters**, assigned with `let` then called by name — no general function-pointer / higher-order typing yet. |
| **Match on enums** | Yes | **Exhaustiveness** is checked (all variants or error). |
| **-- (decrement)** | Yes | Postfix, like `++`. |
| **`kodae bind` (Generic)** | Yes | Uses LLVM/Clang (`-ast-dump=json`) to generate robust `pub struct`, `pub enum`, and `extern fn` bindings for any C library header. Handles nested types and complex C signatures. |

Environment: `KODAE_CC` and `kodae build --cc` select the C toolchain (see `internal/ccdriver`). For packaging details, see `docs/DISTRIBUTION.md`.

### Example programs (repo)

| File | Exercises |
|------|-----------|
| `examples/hello.kodae` | Minimal program |
| `examples/onepage.kodae` | Full beginner one-pager (see `docs/LANGUAGE.md`) |
| `examples/raylib_minimal.kodae` | Template: Raylib + `# link` / `# linkpath` (needs Raylib installed) |
| `docs/C_LIBRARIES.md` | How to use C libraries (e.g. Raylib) from Kodae |
| `examples/extern_hello.kodae` | `extern fn` + `printf` + varargs |
| `examples/features.kodae` | `defer`, `continue`, `and` / `or`, `++` / `--`, enums + `match` |
| `examples/list_basic.kodae` | `list[T]`, literals, index, `len`, `push`/`pop`/`append`/`remove` |
| `examples/result_minimal.kodae` | Simplified `catch` usage in v1 style |
| `examples/textrpg.kodae` | Larger sample |
| `examples/stdlib_v2_test.kodae` | Bitwise, `repeat`, `swap`, `sort`, `reverse`, etc. |
| `examples/include/main.kodae` + `helpers.kodae` | `#include "helpers"` and `pub` for cross-file names (preferred for resolving `./libs` and `~/.kodae/libs` too) |
| `examples/multi/lib.kodae` + `app.kodae` | `use lib` in `app.kodae` (run `kodae run app.kodae` from `examples/multi/`) or `kodae build lib.kodae app.kodae` in order; `lib.kodae` must `pub` export `double` |

**Syntax note:** `if` always uses a parenthesized condition: `if (cond) { }`. For combined booleans, wrap the full expression, e.g. `if ((a) and (b)) { }` (not `if (a) and (b)`).
