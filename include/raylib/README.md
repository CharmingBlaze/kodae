# Raylib Clio bindings (`raylib.clio`)

This is a **machine-generated** binding with many Raylib 6.x structs + `extern fn` declarations. It is still not 100% of the C API yet (callbacks, function pointers, and some complex signatures are skipped), but it covers a large share of day-to-day APIs.

## Regenerate

1. Download [raylib.h](https://github.com/raysan5/raylib/blob/master/src/raylib.h) (same major version as your installed library).  
2. From the repository root:

```sh
clio bind raylib path/to/raylib.h -o include/raylib/raylib.clio
```

`clio bind` also writes `examples/libs/raylib/raylib.clio` (unless `-o` already points there) so the `examples/raylib_game.clio` sample can `#include "raylib/raylib"`.

## Use in a project

```clio
# include "raylib/raylib"

fn main() {
  InitWindow(800, 600, "Game")
  ' ...
  CloseWindow()
}
```

Set `# link` / `# linkpath` in the generated file, or add your own `-L`/`-l` flags, so the C linker can find the Raylib import library. See [docs/C_LIBRARIES.md](../../docs/C_LIBRARIES.md) and [examples/raylib_game.clio](../../examples/raylib_game.clio).

## Compiler note: `f32`

C `float` in `extern` uses the Clio type `f32` in signatures (returns widen to Clio’s normal `float` / C `double` in the rest of the program). This was added for Raylib-style float parameters.
