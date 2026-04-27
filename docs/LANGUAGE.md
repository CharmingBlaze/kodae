# Clio: the whole language in one page

Clio is small on purpose. Line comments start with `'` (apostrophe) or `--`.

**If you are new to the repo:** start with the root [README.md](../README.md) (how to run the compiler) and the [examples/README.md](../examples/README.md) (runnable files in order). This page is the **language** itself; a map of *all* docs is in [docs/README.md](README.md).

Below is a **real, self-contained program** (Clio only allows *statements* like `print` and `let` *inside* a function, so the runnable one-pager uses `fn main() { ... }` after your types and functions are declared at file scope). You can also write `for (i in 0..10)` *or* the shorter `for i in 0..10` — both are supported.

## One page of Clio

```clio
' --- Data (file scope: structs, enums, methods, functions)
struct Player {
  name: str
  health: int
  score: int
}

enum Direction { Up, Down, Left, Right }

fn Player.heal(amount: int) {
  this.health += amount
  if (this.health > 100) {
    this.health = 100
  }
}

fn add(a: int, b: int) -> int {
  return a + b
}

' --- Program
fn main() {
  ' Variables
  let name = "Ada"
  let score = 0
  let speed = 1.5
  let alive = true
  const MAX = 100

  ' Print
  print("Hello $name!")
  print("Score: $score")

  ' Math
  score += 10
  score++
  let total = score + 100

  ' If / else
  if (score > 50) {
    print("winning!")
  } else {
    print("keep going")
  }

  ' While
  while (alive) {
    score += 1
    if (score >= MAX) {
      break
    }
  }

  ' For (range) — with or without outer ( ): same meaning
  for i in 0..10 {
    print("$i")
  }

  ' List
  let items: list[str] = ["sword", "shield", "potion"]
  items.push("bow")
  print(items[0])
  print(items.len)

  for item in items {
    print("$item")
  }

  ' Struct
  let p = Player { name: "Hero", health: 100, score: 0 }
  print(p.name)
  p.health -= 10

  p.heal(25)

  ' Enum + match
  let dir = Direction.Up

  match (dir) {
    Direction.Up    => { print("going up") }
    Direction.Down  => { print("going down") }
    Direction.Left  => { print("going left") }
    Direction.Right => { print("going right") }
  }

  ' Function
  let result = add(10, 20)
  print("$result")
}
```

That is the full beginner surface. Everything else in the compiler and repo is **optional** until you need it (see the table below).

## What we skip at first (keep the language easy)

| Cut / hide | Why |
|------------|-----|
| `result[T]`, `ok()`, `err()`, `?` | Beginners do not need error *types* — you can use `catch` later. |
| `ptr[...]` in your own code | C interop only: shows up in `extern fn` (e.g. games / Raylib). |
| `module` / `use` | One file is enough to start. |
| `pub`, `#library`, `build --lib` | “Ship a C library” is advanced. |
| `extern fn` | Advanced; day 4 / games. |
| `#mode library` | C export pipeline — not for lesson one. |
| Opaque / low-level C concepts | Not part of the beginner path. |
| A type on every `let` | Let the compiler infer: `let x = 10` is enough. |

## Beginner learning path

1. **Day 1** — `print` and variables  
2. **Day 1** — `if`, `while`, `for`  
3. **Day 2** — functions  
4. **Day 2** — `list`  
5. **Day 3** — structs, methods, `this`  
6. **Day 3** — enums and `match`  
7. **Any time after functions (optional)** — **multi-file** programs: `#include` and `pub` in [examples/include/](../examples/include/); the older `use` style is [examples/multi/](../examples/multi/) (see [DIRECTIVES.md](DIRECTIVES.md)).  
8. **Day 4** — games: link a C library and call it with `extern fn` (see [C_LIBRARIES.md](C_LIBRARIES.md) for the full Raylib-style flow: `# link`, `# linkpath`, and a small sample in [examples/raylib_minimal.clio](../examples/raylib_minimal.clio); [examples/extern_hello.clio](../examples/extern_hello.clio) is the minimal `printf` interop test)

Everything else — portable compiler bundles and `catch` — is **advanced** and can wait.

## The three rules for friendly Clio

1. **Types are optional. The compiler figures out what it can.**  

   ```clio
   let x = 10
   let name = "Ada"
   let items: list[str] = []
   ```  

   (Empty `[]` needs a type: `list[str]` or another `list[...]`.)

2. **One obvious way.** No `result[T]` and `?` and `catch` in the *beginner* story—when you add errors, Clio v1 is **just `catch`**, not a visible result type. There is no separate array type — use **`list[T]`** only.

3. **Errors should read like English.** The compiler gives short, readable messages, for example:

   - `unknown name "scre" — did you mean "score"?`
   - `cannot add int and str (use str(...) on the number, or use "..." for text)`
   - `struct Player has no field "heath" — did you mean "health"?`

   (Exact wording can vary slightly by release.)

## Optional later topics (not for page one)

| Topic | Where to read |
|--------|----------------|
| Runnable examples in order | [examples/README.md](../examples/README.md) |
| `catch` on calls you define or link | `examples/result_minimal.clio`, [SUPPORTED.md](../SUPPORTED.md) |
| C interop and linking a game lib (Raylib, etc.) | [C_LIBRARIES.md](C_LIBRARIES.md), [examples/extern_hello.clio](../examples/extern_hello.clio) |
| `pub` and C library export | [LIBRARIES.md](LIBRARIES.md) |
| Multi-file, `#include`, and installable `.clio` | [DIRECTIVES.md](DIRECTIVES.md) |
| Portable `clio` + bundled toolchain | [DISTRIBUTION.md](DISTRIBUTION.md) |

## Built-ins (quick reference)

- `print(...)` — strings can use `"Hello $name"` and `"$i"`-style **simple** names inside the quotes where the compiler supports it  
- `input(prompt)`, `random(lo, hi)`, `clear_screen()`  
- `len(list)` — same as `list.len` on a list value  
- Casts: `int(x)`, `float(x)`, `str(x)`, `bool(x)`  
- Helpers: `min`, `max`, `abs`

## Runnable copy

A checked-in version of the one-pager (with `fn main() { ... }`) is [examples/onepage.clio](../examples/onepage.clio) — you can `clio run` or `clio build` that file to verify your install.

## Implementation status

See [SUPPORTED.md](../SUPPORTED.md) for the feature checklist and details.
