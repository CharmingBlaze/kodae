# Using C libraries (e.g. Raylib) from Clio

This is the full process, in five steps, as simple as possible.

## Step 1 — Get the library

**Example: Raylib.** You download a folder that usually contains things like:

```
raylib/
  raylib.h
  libraylib.a   /  raylib.a
  libraylib.lib
```

* **macOS / Linux** — a static or shared library (often `libraylib.a` and/or `.so` / `.dylib`).
* **Windows** — often a `.lib` for linking (and sometimes a separate `.dll` to ship next to your game).

The header **`.h`** is your reference: it lists function names, parameters, and C types. You only *declare* in Clio the functions you actually use.

## Step 2 — Point Clio at the library (link line)

At the top of your `.clio` file:

```clio
# link "raylib"
# linkpath "./raylib"
```

- `# link "raylib"` — a **short name**: Clio turns this into a linker flag **`-lraylib`** (same as typing `# link "-lraylib"`).
- `# linkpath "./raylib"` — adds **`-L./raylib`**, i.e. “search this folder for `libraylib.a` / the platform’s import library”.

For full control you can pass raw flags instead, for example: `# link "-lraylib -L./raylib"` (see [SUPPORTED.md](../SUPPORTED.md) for the general `# link` behavior).

**Also:** the CLI can still add extra link flags: `clio run game.clio --ldflags "..."` if you need something special for your platform.

## Step 3 — Declare the functions you use

You do **not** have to import the whole header. Open `raylib.h`, find the C declaration, and write a matching `extern fn` in Clio.

C → Clio type cheat sheet:

| C type | In Clio |
|--------|---------|
| `int` | `int` |
| `float` / `double` | `float` (Clio `float` maps to C `double` in the compiler today) |
| `bool` | `bool` |
| `void` in return | `-> void` in Clio |
| `const char *` / `char *` (string pointer) | `ptr[byte]` in `extern fn` parameters |
| `void *` | `ptr[byte]` (or another `ptr[...]`) in **extern** signatures only |
| `unsigned int` | `int` for many cases, or be explicit in comments |
| `Color` in Raylib | *often* a **packed `int` colour** in practice; many samples use a hex `int` literal, e.g. `0xFFFFFFFF` (Clio supports `0x…` for integers) |

**Example (from a typical `raylib.h`):**

```clio
' C: void InitWindow(int width, int height, const char *title);
extern fn InitWindow(w: int, h: int, title: ptr[byte]) -> void

' C: bool WindowShouldClose(void);
extern fn WindowShouldClose() -> bool

' C: void BeginDrawing(void);
extern fn BeginDrawing() -> void

' C: void EndDrawing(void);
extern fn EndDrawing() -> void

' C: void CloseWindow(void);
extern fn CloseWindow() -> void

' C: void ClearBackground(Color color);
extern fn ClearBackground(color: int) -> void

' C: void DrawText(const char *text, int x, int y, int fontSize, Color color);
extern fn DrawText(text: ptr[byte], x: int, y: int, size: int, color: int) -> void
```

Clio requires a return type: use **`-> void`** for C `void` functions, not “nothing” after the closing `)`.

**Strings:** pass Clio `str` where the signature is `ptr[byte]`; the C backend is set up to pass a pointer to the string’s bytes.

**Vararg C functions** (e.g. `printf`‑style) are a special case; the repository already has a small `printf` example — see [examples/extern_hello.clio](../examples/extern_hello.clio). Raylib’s `DrawText` in the form above is **fixed** arity, which is easy.

## Step 4 — Call them from `fn main` (or any function)

```clio
# link "raylib"
# linkpath "./raylib"

extern fn InitWindow(w: int, h: int, title: ptr[byte]) -> void
extern fn WindowShouldClose() -> bool
extern fn BeginDrawing() -> void
extern fn EndDrawing() -> void
extern fn CloseWindow() -> void
extern fn ClearBackground(color: int) -> void
extern fn DrawText(text: ptr[byte], x: int, y: int, size: int, color: int) -> void

fn main() {
  InitWindow(800, 600, "My Game")

  loop {
    if (WindowShouldClose()) { break }

    BeginDrawing()
    ClearBackground(0x181818FF)
    DrawText("Hello from Clio!", 300, 280, 24, 0xFFFFFFFF)
    EndDrawing()
  }

  CloseWindow()
}
```

(Adjust `# linkpath` to the real folder that contains the Raylib import library for your OS/toolchain.)

## Step 5 — Build and run

With a C compiler on `PATH` (or `CLIO_CC` / `--cc`):

```text
clio run game.clio
```

Clio: **Clio source → C →** your C driver **links** your `game` with **Raylib** and produces an executable. If the linker cannot find the library, fix the path in `# linkpath` or the library name in `# link` for your platform.

**Rough flow:**

```text
game.clio
   │  clio run
   v
C source + Clio runtime
   │  C compiler
   v
game.exe  ←── linked with raylib (as configured)
```

## Future: `clio bind`

A future tool could generate the `extern fn` lines from a `.h` (for example: `clio bind raylib.h`). For now, **hand‑written `extern` only for the symbols you use** is normal — many small games need only a few dozen functions.

A minimal Raylib-style skeleton is in [examples/raylib_minimal.clio](../examples/raylib_minimal.clio) (you must install Raylib and fix `# linkpath` yourself; CI does not require it to be present).
