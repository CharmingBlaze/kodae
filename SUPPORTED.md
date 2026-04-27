# Clio implementation vs. build-spec

This document tracks what the Go compiler in this repo supports relative to `docs/build-spec.js`. It is a living checklist, not a normative spec.

For user-facing language syntax and examples, see `docs/LANGUAGE.md`.

| Area | Status | Notes |
|------|--------|--------|
| **Single / multi file** | Partial | `clio build|check|cgen|parse` all accept `a.clio b.clio` in order (same `parseBuildFlags` as build; extra `-o` / `--cc` / `--ldflags` are ignored for check/cgen/parse). |
| **module / use / pub** | Partial | `module` / `pub` are parsed. **`use name`**: loads `name.clio` from the **same directory** as the current file, before that file’s other top-level decls. Cycles are an error. `module` is not used to resolve paths (v1). |
| **extern fn** | Yes | C signatures; calls use the real C name (no `f_` prefix). stdio `printf` etc. skip redecl (see `<stdio.h>`). `str` → `ptr[byte]` uses `(s).data`. |
| **# link "flags"** | Yes | Split on whitespace; appended to the C link line after `-lm`. CLI: `--ldflags` / `--ldflags=...`. |
| **Link driver** | Yes | `clio` passes extra argv to the same C compiler line as the object file. |
| **continue** | Yes | In loops. |
| **defer** | Partial | Typechecker: `defer` may only sit at the **top** of a function body (v1). Emitted in reverse on return and at fallthrough end. |
| **and / or** | Yes | Spelled `and` / `or` (same precedence as in the Pratt table). |
| **`catch`** | Yes (simplified) | `catch` is supported as a full `let` init / return value / assignment RHS / expression statement. |
| **`result[T]`, `ok/err`, `.ok/.value/.err`, `?`** | No (removed) | Not part of Clio v1 surface syntax. Use `catch`. |
| **`T?` optional syntax** | No (removed) | V1 uses implicit nullable behavior with `none`; explicit `T?` is rejected. |
| **`ptr[T]`** | Restricted | Only allowed in `extern fn` signatures. |
| **`pub` exports** | Yes (v1) | `pub fn` / `pub struct` define exported C library API surface for `build --lib`. |
| **`build --lib`** | Yes (v1) | Emits `.c`, `.h`, static (`.a`) and shared (`.so`/`.dll`/`.dylib`) artifacts. |
| **`list[T]`** | Yes (v1) | Type syntax `list[T]`, literals `[a, b]`, index read/write `xs[i]`, `len(xs)`, methods `push`, `pop`, `append`, `remove`, and `for (x in list)` iteration. |
| **Match on enums** | Yes | **Exhaustiveness** is checked (all variants or error). |
| **-- (decrement)** | Yes | Postfix, like `++`. |
| **Arrays, clio bind** | No / stub | `cmd/clio-bind` exits 1 with a message. |

Environment: `CLIO_CC` and `clio build --cc` select the C toolchain (see `internal/ccdriver`).

### Example programs (repo)

| File | Exercises |
|------|-----------|
| `examples/hello.clio` | Minimal program |
| `examples/extern_hello.clio` | `extern fn` + `printf` + varargs |
| `examples/features.clio` | `defer`, `continue`, `and` / `or`, `++` / `--`, enums + `match` |
| `examples/list_basic.clio` | `list[T]`, literals, index, `len`, `push`/`pop`/`append`/`remove` |
| `examples/result_minimal.clio` | Simplified `catch` usage in v1 style |
| `examples/textrpg.clio` | Larger sample |
| `examples/multi/lib.clio` + `app.clio` | `use lib` in `app.clio` (build with `clio build app.clio` from that folder) or `clio build lib.clio app.clio` without `use` |

**Syntax note:** `if` always uses a parenthesized condition: `if (cond) { }`. For combined booleans, wrap the full expression, e.g. `if ((a) and (b)) { }` (not `if (a) and (b)`).
