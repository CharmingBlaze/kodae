# Kodae Language Reference

Kodae is designed to be as easy to read as Python or BASIC, but as fast and powerful as C. Everything is public by default, meaning you don't need to worry about complex access modifiers or keywords when writing your code.

### Where everything lives

| Topic | Document |
|-------|----------|
| **This page** | Syntax, types, control flow, structs, enums, errors, advanced bits |
| **CLI (`kodae run`, `build`, …)** | [CLI.md](CLI.md) |
| **`#include`, `#link`, installs** | [DIRECTIVES.md](DIRECTIVES.md) |
| **Calling C / Raylib** | [C_LIBRARIES.md](C_LIBRARIES.md) |
| **Bindings from `.h`** | [BINDGEN.md](BINDGEN.md) |
| **Building `.dll` / `.so` libs** | [LIBRARIES.md](LIBRARIES.md) |
| **Shipping the compiler** | [DISTRIBUTION.md](DISTRIBUTION.md) |
| **What is implemented today** | [SUPPORTED.md](../SUPPORTED.md) |
| **Examples** | [examples/README.md](../examples/README.md) |

### Table of contents

1. [The Absolute Basics](#1-the-absolute-basics)  
2. [Variables & Constants](#2-variables--constants)  
3. [Control Flow](#3-control-flow)  
4. [Functions & Tuples](#4-functions--tuples)  
5. [Lists](#5-lists)  
6. [Structs & Enums](#6-custom-types-structs--enums)  
7. [Multi-file Programs](#7-multi-file-programs)  
8. [Error Handling (`result` & `catch`)](#8-error-handling-results)  
9. [Advanced control (`repeat`, `loop`, `defer`)](#9-advanced-control-repeat-loop-defer)  
10. [Operators & bitwise](#10-operators--bitwise)  
11. [Methods, `this`, `with`, lambdas](#11-methods-this-with-and-inline-lambdas)  
12. [Built-ins & standard helpers](#12-built-ins--standard-helpers)  

---

## 1. The Absolute Basics

Every Kodae program consists of statements. You can print text to the screen easily:

```kodae
fn main() {
    print("Hello, World!")
}
```

Line comments start with `'` (apostrophe) or `--`.

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
You can group related constants together (perfect for colors or game states):
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

Functions group code together. You define them using the `fn` keyword. 
If your function returns a value, you specify the return type using the `->` arrow syntax.
They can also have default parameters and return multiple values (Tuples).

```kodae
' b defaults to 0 if not provided
fn add_and_double(a: int, b: int = 0) (int, int) {
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
Structs let you group data together. Methods let you attach functions to that data. Inside a method, `this` refers to the current instance.

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

## 7. Multi-file Programs
You can split your project into multiple files using `#include`. Everything is public, so included functions and structs are available immediately!

**math.kodae**
```kodae
fn double(x: int) -> int {
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

## 8. Error Handling (Results)

Some expressions have type **`result[T]`** — success carries a `T`, failure carries error information handled by Kodae’s **`catch`** form (not Java-style exceptions).

### Syntax

Attach **`catch`** only to an expression whose type is **`result[...]`**:

```kodae
let value = some_call() catch (err) {
  print("failed: " + err)
}
```

The **`catch (name) { ... }`** block runs on failure; `name` is a **`str`** describing the error. The compiler rejects **`catch`** on plain `void` calls or non-`result` values.

Details vary by API; see **`examples/result_minimal.kodae`** and the checklist in **[SUPPORTED.md](../SUPPORTED.md)** (`catch`, `result[T]`). Older **`ok` / `err` / `?`**-style surfaces are not part of Kodae v1 — use **`catch`**.

---

## 9. Advanced control (`repeat`, `loop`, `defer`)

- **`repeat(n) { ... }`** — runs the body exactly **`n`** times (`n` is an `int`). See **`examples/stdlib_v2_test.kodae`** / **SUPPORTED**.
- **`loop { ... }`** — infinite loop; use **`break`** / **`continue`** inside.
- **`while (cond) { ... }`** — condition must be **`bool`**.
- **`for i in a..b`** — half-open integer range (`b` excluded unless your range syntax matches the compiler).
- **`defer expr`** — runs **`expr`** when the function exits (reverse order if multiple defers). In v1, **`defer`** may only appear at the **top level** of a function body (not inside nested blocks). See **SUPPORTED**.

---

## 10. Operators & bitwise

- **Arithmetic:** `+`, `-`, `*`, `/`, `%`; unary `-`, `+`.
- **Compare:** `==`, `!=`, `<`, `>`, `<=`, `>=`.
- **Logic:** **`and`**, **`or`**, **`not`** (spell words; wrap complex tests in parentheses when combining, e.g. `if ((a > 0) and (b > 0))`).
- **Increment:** postfix **`++`** and **`--`** on numeric lvalues.
- **Bitwise:** **`&`**, **`|`**, **`^`**, **`~`** on integers; binary literals **`0b1010`** are supported.

---

## 11. Methods, `this`, `with`, and inline lambdas

### Rules for `this`

1. **`this`** is only valid inside **`fn TypeName.method(...)`** (struct methods).
2. **`this`** always means the receiver instance for that method — including inside **`repeat`**, **`for`**, **`if`**, and nested blocks (lexical binding; unlike JavaScript, it does not “get lost” in nested functions).
3. **Inline lambdas** — **`let cb = fn() { ... }; cb()`** — only **`fn()` with no parameters** and **`void`** body are supported in v1. If the lambda uses **`this`**, it must appear **inside** a method so the compiler can pass the same receiver as C **`self`**. You cannot pass these lambdas to arbitrary C function-pointer parameters yet; call them from Kodae.

### `with` (functional update)

**`expr with { field: value, ... }`** builds a **copy** of a struct value and overrides listed fields:

```kodae
fn Player.clone_named(s: str) -> Player {
  return this with { name: s }
}
```

### Method chaining

Return **`this`** (or a new struct) from methods to chain calls:

```kodae
fn Player.tag(t: str) -> Player {
  this.name = this.name + t
  return this
}
```

See **SUPPORTED** for the precise status of **`this` / `with` / fn lambdas**.

---

## 12. Built-ins & standard helpers

Kodae exposes many **built-in functions** (`print`, `input`, `random`, file helpers, math, timers, JSON helpers, etc.). The exact set evolves; treat **[SUPPORTED.md](../SUPPORTED.md)** as the authoritative checklist, and **`examples/stdlib_v2_test.kodae`** as a broad runtime exercise.

**Linking and C:** use **`extern fn`**, **`# link`**, **`# linkpath`**, and sized types as in **[C_LIBRARIES.md](C_LIBRARIES.md)**.

**Compiler directives** (`#include`, `#library`, …) are described in **[DIRECTIVES.md](DIRECTIVES.md)**.

---

*End of language reference summary — for implementation detail and edge cases, combine this file with **SUPPORTED.md** and the docs linked in the table at the top.*
