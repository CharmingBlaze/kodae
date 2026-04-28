# Kodae documentation index

Read these in any order; the table shows what each document is for.

| Document | Read when you need… |
|----------|---------------------|
| **[LANGUAGE.md](LANGUAGE.md)** | **Syntax and semantics** — types, control flow, structs, `this`, `catch`, directives overview. |
| **[CLI.md](CLI.md)** | **`kodae` commands** — `run`, `build`, `check`, `bind`, flags, env vars. |
| **[DIRECTIVES.md](DIRECTIVES.md)** | **`#include`**, `#link`, `#library`, metadata, `kodae install`, search paths. |
| **[C_LIBRARIES.md](C_LIBRARIES.md)** | **C interop** — `extern fn`, sized types, linking game libs (e.g. Raylib). |
| **[BINDGEN.md](BINDGEN.md)** | **Auto bindings** from `.h` files (`kodae bind` / `kodae-bind`). |
| **[LIBRARIES.md](LIBRARIES.md)** | **`kodae build --lib`** — shipping `.c` / `.h` / `.a` / `.dll` / `.so`. |
| **[DISTRIBUTION.md](DISTRIBUTION.md)** | **Portable bundles** — shipping `kodae` + toolchain (e.g. Zig). |

**Also in the repo**

| Path | Purpose |
|------|---------|
| [README.md](../README.md) | Quick start, bundle download, doc links. |
| [examples/README.md](../examples/README.md) | Runnable examples and suggested learning order. |
| [SUPPORTED.md](../SUPPORTED.md) | **Implementation checklist** — what the compiler implements today. |

Large machine-readable specs (e.g. `build-spec.js`) exist for tooling authors; **learn the language from [LANGUAGE.md](LANGUAGE.md), not from the build-spec.**
