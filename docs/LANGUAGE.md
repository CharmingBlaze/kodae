# Kodae: the whole language in one page

Kodae is small on purpose. Line comments start with `'` (apostrophe) or `--`.

**Quick Start:** The easiest way to use Kodae is to download the **Portable Bundle** for your platform from the [Releases](https://github.com/CharmingBlaze/kodae/releases) page. It includes a built-in C toolchain, so you can start coding immediately with no setup.


**New to Kodae?** Start with the [README.md](../README.md) for setup, [CLI.md](CLI.md) for how every terminal command works, and the [examples/README.md](../examples/README.md) for a hands-on learning path. This page serves as the complete language reference.

Below is a **real, self-contained program** (Kodae only allows *statements* like `print` and `let` *inside* a function, so the runnable one-pager uses `fn main() { ... }` after your types and functions are declared at file scope). You can also write `for (i in 0..10)` *or* the shorter `for i in 0..10` — both are supported.

## One page of Kodae

```kodae
' --- Data (file scope: structs, enums, methods, functions)
struct Player {
  name: str
  health: int
  score: int
}

enum Direction { Up, Down, Left, Right }

fn Player.heal(amount: int) {
  ' `this` refers to the current struct instance inside a method
  this.health += amount
  if (this.health > 100) {
    this.health = 100
  }
}

' Functions can have default parameters and return tuples!
fn add_and_double(a: int, b: int = 0) -> (int, int) {
  let sum = a + b
  return sum, sum * 2
}

' --- Program
fn main() {
  ' Variables
  let name = "Ada"
  let score = 0
  let speed = 1.5
  let alive = true
  
  ' Grouped Constants
  const Colors {
      RED    = 0xFF0000FF
      GREEN  = 0x00FF00FF
      BLUE   = 0x0000FFFF
      WHITE  = 0xFFFFFFFF
      BLACK  = 0x000000FF
      YELLOW = 0xFFFF00FF
  }
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

  ' For (range) — prefer the clean syntax without parens
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

  ' Function calling with default params and tuple unpacking
  let sum, double = add_and_double(10)
  print("Sum: $sum, Double: $double")
  
  ' Multiline strings
  let welcome = """
    Welcome to Kodae!
    Easy as BASIC.
    Fast as C.
  """
  print(welcome)
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

1. **Day 1** — `print`, variables, math operators
2. **Day 1** — `if` / `else`, `while`, `for`
3. **Day 2** — functions, tuples, default params
4. **Day 2** — lists and list methods
5. **Day 3** — structs, methods, `this` keyword
6. **Day 3** — enums and `match`
7. **Day 4** — files, JSON, save systems
8. **Day 5** — networking, HTTP, online scores
9. **Day 6** — games with Raylib
10. **Day 7** — C libraries, `pub`, `#include`, `build --lib`

Everything else — portable compiler bundles and `catch` — is **advanced** and can wait.

## The three rules for friendly Kodae

1. **Types are optional. The compiler figures out what it can.**  

   ```kodae
   let x = 10
   let name = "Ada"
   let items: list[str] = []
   ```  

   (Empty `[]` needs a type: `list[str]` or another `list[...]`.)

2. **One obvious way.** No `result[T]` and `?` and `catch` in the *beginner* story—when you add errors, Kodae v1 is **just `catch`**, not a visible result type. There is no separate array type — use **`list[T]`** only.

3. **Errors should read like English.** The compiler gives short, readable messages, for example:

   - `unknown name "scre" — did you mean "score"?`
   - `cannot add int and str (use str(...) on the number, or use "..." for text)`
   - `struct Player has no field "heath" — did you mean "health"?`

   (Exact wording can vary slightly by release.)

## `this` in struct methods (lexical `this`, not JavaScript surprises)

Kodae uses **`this`** only inside **`fn TypeName.method(...)`** bodies. It always means **the receiver instance** — the same rules apply inside loops, nested blocks, and **inline lambdas** (`fn() { ... }`). Unlike JavaScript, **`this` never “breaks”** when you nest a function or pass a callback: the compiler keeps the receiver bound for you (similar to Kotlin, Swift, or C#).

### Three rules

1. **`this` is only valid** in a method declaration (`fn Player.heal(...)`, not in bare `fn foo()`).
2. **`this` always refers to the struct instance** for that method — everywhere in that method’s body.
3. **Nested `fn() { }` lambdas** may use `this` when they appear **inside** a method; they compile to static helpers that receive the same `self` pointer as C-level receiver.

### Useful patterns

- **`return this`** — return the current instance (by value).
- **Method chaining** — return `this` (or a copy) from methods that mutate:

  ```kodae
  fn Player.set_name(name: str) -> Player {
    this.name = name
    return this
  }
  ```

- **Functional update — `expr with { field: value, ... }`** — copy a struct value and override only the listed fields (the base expression is usually `this` or another struct variable):

  ```kodae
  fn Player.renamed(s: str) -> Player {
    return this with { name: s }
  }
  ```

- **Callbacks / lambdas** — assign a zero-argument void lambda, then call it (only **`fn() { }`** with **no parameters** is supported in v1):

  ```kodae
  fn Player.setup() {
    let on_tick = fn() {
      this.x += 1
    }
    on_tick()
  }
  ```

Passing lambdas to **`extern fn`** parameters that expect C function pointers is not fully modeled yet; use normal Kodae calls for now.

### Kodae vs JavaScript (`this`)

| Feature | JavaScript | Kodae |
|--------|------------|--------|
| `this` in a method | Works | Works |
| `this` in a nested non-arrow function | Often **wrong** (`this` is lost) | **Always** the receiver |
| `this` in loops | Same confusion as nested scopes | **Always** the receiver |
| Method chaining | Common pattern | Supported (`return this`) |
| `this` outside any method | Refers to global / `undefined` / strict rules | **Compile error** |

## Optional later topics (not for page one)

| Topic | Where to read |
|--------|----------------|
| `kodae run`, `build`, `check`, and all CLI flags | [CLI.md](CLI.md) |
| Runnable examples in order | [examples/README.md](../examples/README.md) |
| `catch` on calls you define or link | `examples/result_minimal.kodae`, [SUPPORTED.md](../SUPPORTED.md) |
| C interop and linking a game lib (Raylib, etc.) | [C_LIBRARIES.md](C_LIBRARIES.md), [examples/extern_hello.kodae](../examples/extern_hello.kodae) |
| `pub` and C library export | [LIBRARIES.md](LIBRARIES.md) |
| Multi-file, `#include`, and installable `.kodae` | [DIRECTIVES.md](DIRECTIVES.md) |
| Portable `kodae` + bundled toolchain | [DISTRIBUTION.md](DISTRIBUTION.md) |

## Built-ins (quick reference)

### Basics
- `print(a, b, ...)` — prints values separated by spaces
- `input(prompt)`, `input_int(prompt)`, `input_float(prompt)`
- `random(lo, hi)`, `random_float(lo, hi)`, `random_bool()`
- `clear_screen()`
- `len(list_or_str)` — same as `.len`
- **Casts:** `int(x)`, `float(x)`, `str(x)`, `bool(x)`

### String Methods
Strings have many built-in methods:
- `s.upper()`, `s.lower()`, `s.trim()`, `s.reverse()`
- `s.contains("sub")`, `s.starts("sub")`, `s.ends("sub")`
- `s.replace("old", "new")`
- `s.split(",")` (returns a `list[str]`)
- `s.len`, `s.is_empty()`, `s.is_number()`

### Math & Numbers
- `min(a, b)`, `max(a, b)`, `abs(x)`
- `sqrt(x)`, `pow(x, y)`, `log(x)`
- `floor(x)`, `ceil(x)`, `round(x)`
- `sin(x)`, `cos(x)`, `tan(x)`, `atan2(y, x)`
- `format_float(val, decimals)`
- **Game Math:** `distance(x1, y1, x2, y2)`, `angle_to(x1, y1, x2, y2)`
- **Game Math:** `lerp(a, b, t)`, `map(x, in_min, in_max, out_min, out_max)`, `clamp(x, min, max)`

### File Operations
- `read_file("save.txt") -> str`
- `write_file("save.txt", "data")`
- `append_file("log.txt", "line\n")`
- `file_exists("save.dat") -> bool`
- `delete_file("old.txt")`
- `copy_file("a.txt", "b.txt")`, `move_file("a.txt", "b.txt")`
- `make_folder("saves")`, `delete_folder("saves")`, `folder_exists("saves")`
- `list_files("./levels") -> list[str]`

### Networking & Web (via `use net` and `use json`)
Networking is natively built-in (wraps lightweight C libraries).
- `http_get("https://api.example.com") -> result[str]`
- `http_post("https://api.example.com", data) -> result[str]`
- `download("https://example.com/file.dat", "local/file.dat") -> bool`
- `is_online() -> bool`
- `json_parse(text) -> Any`
- `json_build(data) -> str`
- **WebSockets** (Coming in Phase 3!)

### Multi-file Programs
You can split your project into multiple files using `#include`:
- `#include "player"` (includes `player.kodae` in the same folder)
- `#include "libs/mymath"` (includes `libs/mymath.kodae`)
- `#include "raylib"` (includes installed library)

### C Interop Types
For working with C libraries (like Raylib), Kodae provides sized types:
- `i32`, `u32`, `u8`
- `f32` (C float)
- `ptr[byte]` (C pointer)

These types are mainly for data layout in structs and function calls. You can usually pass a standard `int` or `float` to a function expecting these, and the compiler will handle the conversion automatically.

### New in v1.1 (Standard Library & Language Expansion)

#### Time & Timers
- `time()` — seconds since program started (float)
- `time_ms()` — milliseconds since program started (int)
- `wait(seconds)`, `wait_ms(ms)`
- `timer_start()`, `timer_elapsed(timer)`
- `countdown(seconds)`, `countdown_done(cd)`

#### Random & Logic
- `chance(percentage)` — returns true with X% probability
- `random_float(lo, hi)`, `random_pick(list)`
- `in_range(val, min, max)`, `in_rect(px, py, rx, ry, rw, rh)`
- `swap(a, b)` — swap values of two variables

#### List Methods
- `list.sort()`, `list.reverse()`, `list.shuffle()`
- `list.first()`, `list.last()`
- `list.push(item)`, `list.pop()`, `list.remove(index)`

#### Logical Operators
- `and` (Logical AND), `or` (Logical OR), `not` (Logical NOT)
```kodae
if (alive and score > 0) { }
if (dead or quit) { }
if (not done) { }
```

#### Bitwise Operators & Binary Literals
- `&` (AND), `|` (OR), `^` (XOR), `~` (NOT)
- Binary literals: `let b = 0b1010`

#### OS & System
- `run(command)` — execute a system command
- `open_url(url)` — open a link in the browser
- `os_name()` — returns "windows", "macos", or "linux"
- `args()` — returns a list of command-line arguments
- `env(name)` — returns an environment variable

#### Advanced Controls
- `repeat(n) { ... }` — repeat a block exactly `n` times.
- `defer expression` — run an expression when the function ends.
- `json_read(path)`, `json_write(path, value)` — basic JSON persistence.


## Runnable copy

A checked-in version of the one-pager (with `fn main() { ... }`) is [examples/onepage.kodae](../examples/onepage.kodae) — you can `kodae run` or `kodae build` that file to verify your install.

## Implementation status

See [SUPPORTED.md](../SUPPORTED.md) for the feature checklist and details.
