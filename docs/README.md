# Clio documentation

| Document | Read this when… |
|----------|-----------------|
| [LANGUAGE.md](LANGUAGE.md) | You are **learning the language** — one-page syntax and a day-by-day path. |
| [DIRECTIVES.md](DIRECTIVES.md) | You need **`#include`**, `#link`, library metadata, or `clio install`. |
| [C_LIBRARIES.md](C_LIBRARIES.md) | You call **C** from Clio: `extern fn`, `# link`, Raylib, SDL, etc. |
| [LIBRARIES.md](LIBRARIES.md) | You want **`clio build --lib`**, `.h` + `.a` / shared libs. |
| [DISTRIBUTION.md](DISTRIBUTION.md) | You package **clio + a compiler** (e.g. portable tree with `zig cc`). |

## Elsewhere in the repo

- **[README.md](../README.md)** — quick start, doc map, and example one-liners.  
- **[examples/README.md](../examples/README.md)** — every runnable example and a suggested order.  
- **[SUPPORTED.md](../SUPPORTED.md)** — what the current Go compiler actually implements.  
- **build-spec.js** — large reference for compiler authors; the normative beginner story is [LANGUAGE.md](LANGUAGE.md), not the build-spec.

## Typographic convention

- Code samples use **apostrophe** `'` or double-dash `--` for line comments, matching the compiler.  
- Shell snippets assume a Unix-style shell; on Windows, use the same `clio` / `go run` commands from PowerShell or `cmd`.
