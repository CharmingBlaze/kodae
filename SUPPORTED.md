# Clio implementation vs. build-spec

This document tracks what the Go compiler in this repo supports relative to `docs/build-spec.js`. It is a living checklist, not a normative spec.

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
| **result [T]**, `ok` / `err` | Yes (v1) | C layout `value`, `err`, `ok` per `clio_bootstrap` typedefs. **Access**: `.ok` / `.value` / `.err` on a `result` rvalue. **`?`**: unwrapper + propagate (enclosing `-> result[...]`). **`catch`**: as the whole of `let` / `return` / `assign` / expr; success type is the inner `T` (return catch generates `if (!ok) { … } else { return value; }`). **Not** embedded: e.g. not `1 + f() catch { }`. |
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
| `examples/result_minimal.clio` | `result[int]`, `ok` / `err`, field access, postfix `?`, and `catch` (including `return` … `catch`) |
| `examples/textrpg.clio` | Larger sample |
| `examples/multi/lib.clio` + `app.clio` | `use lib` in `app.clio` (build with `clio build app.clio` from that folder) or `clio build lib.clio app.clio` without `use` |

**Syntax note:** `if` always uses a parenthesized condition: `if (cond) { }`. For combined booleans, wrap the full expression, e.g. `if ((a) and (b)) { }` (not `if (a) and (b)`).
