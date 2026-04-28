# Kodae Language Reference

Kodae is designed to be as easy to read as BASIC or Python, but as fast and powerful as C. 
This document covers **everything** you need to know to write programs in Kodae, from your first `print` statement to building networked multiplayer games.

Line comments start with `'` (apostrophe) or `--`.

---

## 1. The Absolute Basics

Every Kodae program consists of statements. You can print text to the screen easily:

```kodae
fn main() {
    print("Hello, World!")
}
```

## 2. Variables & Constants

You don't need to specify types for variables; the compiler figures it out for you!

```kodae
fn main() {
    let name = "Ada"      ' a string
    let score = 100       ' an integer
    let speed = 1.5       ' a float
    let alive = true      ' a boolean

    ' You can inject variables directly into strings:
    print("Player $name has $score points.")
    
    ' Math is standard
    score += 10
    score++
}
```

### Grouped Constants
You can group related constants together (perfect for colors or states):
```kodae
const Colors {
    RED    = 0xFF0000FF
    GREEN  = 0x00FF00FF
    BLUE   = 0x0000FFFF
    WHITE  = 0xFFFFFFFF
    BLACK  = 0x000000FF
}

fn main() {
    print(Colors.RED)
}
```

## 3. Control Flow

### If / Else
```kodae
if (score > 50) {
    print("Winning!")
} else {
    print("Keep going!")
}
```

### Loops (While & For)
```kodae
' While loops run as long as a condition is true
while (alive) {
    score += 1
    if (score >= 200) { break }
}

' For loops let you count easily
for i in 1..10 {
    print("Counting: $i")
}
```

### Logical Operators
Use `and`, `or`, and `not` to combine conditions:
```kodae
if (alive and score > 0) { print("Playing") }
if (not alive) { print("Game Over") }
```

## 4. Functions & Tuples

Functions group code together. They can have default parameters and return multiple values (Tuples).

```kodae
' b defaults to 0 if not provided
fn add_and_double(a: int, b: int = 0) -> (int, int) {
    let sum = a + b
    return sum, sum * 2
}

fn main() {
    ' Unpacking a tuple return value
    let sum, double = add_and_double(10)
    print("Sum: $sum, Double: $double")
}
```

## 5. Lists

Kodae has a built-in `list` type that grows automatically.

```kodae
fn main() {
    let items: list[str] = ["sword", "shield"]
    
    items.push("bow")       ' Add to the end
    print(items.len)        ' Prints 3
    print(items[0])         ' Prints "sword"
    
    items.remove(1)         ' Removes "shield"
    
    for item in items {
        print(item)
    }
}
```

**List Methods:** `push(item)`, `pop()`, `remove(index)`, `sort()`, `reverse()`, `shuffle()`, `first()`, `last()`.

## 6. Custom Types (Structs & Enums)

### Structs & Methods
Structs let you group data, and Methods let you attach functions to that data. Inside a method, `this` refers to the current instance.

```kodae
struct Player {
    name: str
    health: int
}

fn Player.heal(amount: int) {
    this.health += amount
    if (this.health > 100) {
        this.health = 100
    }
}

fn main() {
    let p = Player { name: "Hero", health: 50 }
    p.heal(25)
    print(p.health) ' Prints 75
}
```

### Enums & Match
Enums represent a fixed set of choices. `match` is a cleaner way to check them.

```kodae
enum State { Menu, Playing, GameOver }

fn main() {
    let current = State.Playing
    
    match (current) {
        State.Menu     => { print("In Menu") }
        State.Playing  => { print("Game On!") }
        State.GameOver => { print("You died") }
    }
}
```

## 7. Standard Library (Built-ins)

Kodae includes many built-in functions out of the box to help you build games and apps.

### Strings
- `s.upper()`, `s.lower()`, `s.trim()`, `s.reverse()`
- `s.contains("sub")`, `s.starts("sub")`, `s.ends("sub")`
- `s.replace("old", "new")`
- `s.split(",")` (returns a `list[str]`)
- `s.len`, `s.is_empty()`, `s.is_number()`

*(Note: Multiline strings can be created using `"""`)*

### Math & Numbers
- `min(a, b)`, `max(a, b)`, `abs(x)`
- `sqrt(x)`, `pow(x, y)`, `log(x)`
- `floor(x)`, `ceil(x)`, `round(x)`
- `sin(x)`, `cos(x)`, `tan(x)`, `atan2(y, x)`
- `format_float(val, decimals)`
- **Game Math:** `distance(x1, y1, x2, y2)`, `angle_to(x1, y1, x2, y2)`, `lerp(a, b, t)`, `map(x, in_min, in_max, out_min, out_max)`

### File Operations
- `read_file("save.txt") -> str`
- `write_file("save.txt", "data")`
- `append_file("log.txt", "line\n")`
- `file_exists("save.dat") -> bool`
- `delete_file("old.txt")`
- `copy_file("a.txt", "b.txt")`, `move_file("a.txt", "b.txt")`
- `make_folder("saves")`, `delete_folder("saves")`, `folder_exists("saves")`
- `list_files("./levels") -> list[str]`

### Networking, JSON, and WebSockets (via `use net` or `#include "libs/net"`)
- `http_get(url) -> result[str]`
- `http_post(url, data) -> result[str]`
- `download(url, dest) -> bool`
- `is_online() -> bool`
- `json_parse(text) -> Any`
- `json_build(data) -> str`

**WebSockets:**
```kodae
#include "libs/net"

fn main() {
    let ws = socket_connect("ws://localhost:8080")
    if (ws.is_valid()) {
        ws.send("Hello!")
        print(ws.receive())
        ws.close()
    }
}
```

### OS & System
- `run(command)` — execute a system command
- `open_url(url)` — open a link in the browser
- `os_name()` — returns "windows", "macos", or "linux"

## 8. Multi-file Programs
You can split your project into multiple files using `#include`. Prefix things with `pub` if they need to be accessed from other files!

**math.kodae**
```kodae
pub fn double(x: int) -> int {
    return x * 2
}
```

**main.kodae**
```kodae
#include "math"

fn main() {
    print(double(10))
}
```

## 9. C Interop (Advanced)
For working with C libraries (like Raylib), Kodae provides sized types (`i32`, `u32`, `u8`, `f32`, `ptr[byte]`). These are used in `extern fn` signatures to bind directly to C libraries with zero overhead.

You can usually pass a standard Kodae `int` or `float` to a function expecting these, and the compiler handles the conversion automatically!
