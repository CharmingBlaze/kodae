# Clio Language Guide (V1)

Clio V1 is intentionally small and beginner-friendly.

## Data Type Quick Map

- `int`, `float`, `str`, `bool`: plain values
- `struct`: named-field data (config, entities, settings)
- `enum`: named states
- `list[T]`: ordered homogeneous collection

## Core Syntax

```clio
let x = 10
const MAX = 100

fn add(a: int, b: int) -> int {
  return a + b
}
```

## Structs and Nested Structs

```clio
struct Config {
  speed: int
  lives: int
  name: str
}

struct GameSettings {
  volume: int
  fullscreen: bool
}

struct Game {
  title: str
  settings: GameSettings
}
```

Usage:

```clio
let config = Config { speed: 5, lives: 3, name: "Ada" }
config.speed = 10
config.lives -= 1
print(config.name)

let game = Game {
  title: "Dragon Cave",
  settings: GameSettings { volume: 80, fullscreen: false }
}
game.settings.volume = 50
```

### Methods (`this`)

```clio
fn Config.boost(delta: int) {
  this.speed += delta
}
```

- `this` is automatically in scope inside methods.
- `this` outside methods is a compile error.

## Enums and Match

```clio
enum State { Menu, Playing, Dead }

match (state) {
  State.Menu => { show_menu() }
  State.Playing => { update() }
  State.Dead => { game_over() }
}
```

Enum matches are exhaustive.

## Lists

List type syntax:

```clio
let xs: list[int] = [1, 2, 3]
```

Supported list operations:

```clio
xs.push(4)
let ys: list[int] = [5, 6]
xs.append(ys)

let a = xs.pop()
let b = xs.remove(0)

xs[0] = a + b
print(str(len(xs)))
```

- Index read/write: `xs[i]`, `xs[i] = v`
- Methods: `push`, `pop`, `append`, `remove`
- Built-in length: `len(xs)`
- Lists are homogeneous (`list[T]`)

## Control Flow

```clio
if ((x > 10) and (not dead)) { }
while (alive) { }
for (i in 0..10) { }
loop { if (done) { break } }
```

Boolean operators:

- `and`
- `or`
- `not` / `NOT`

## Errors (`catch`)

Clio V1 uses `catch` as the visible error-flow construct:

```clio
let data = read_file("save.dat") catch (err) {
  print("Failed: " + err)
}
```

`catch` is valid as the full expression in:

- `let` initializer
- assignment right-hand side
- `return` value
- expression statement

Not in V1:

- `result[T]`
- `.ok`, `.value`, `.err`
- `ok(...)`, `err(...)`
- `?` propagation
- `T?` syntax (use `none` with implicit nullable behavior)

## C Interop

Use `extern fn` for C calls:

```clio
extern fn InitWindow(w: int, h: int, title: ptr[byte]) -> void
```

`ptr[...]` is only allowed in `extern fn` signatures.

## Built-ins

- `print(...)`
- `input(prompt)`
- `random(lo, hi)`
- `clear_screen()`
- `len(listValue)`
- casts: `int(x)`, `float(x)`, `str(x)`, `bool(x)`
- numeric helpers: `min`, `max`, `abs`

## Out of Scope (Current)

- arrays
- completed `clio-bind`

See `SUPPORTED.md` for implementation status details.
