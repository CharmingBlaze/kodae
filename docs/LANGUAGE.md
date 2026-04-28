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

```kodae
fn countdown_demo() {
    defer print("leaving countdown_demo")
    defer print("cleanup runs second")

    repeat(3) {
        print("repeat tick")
    }

    let n = 0
    while (n < 2) {
        print("while n=$n")
        n++
    }

    for i in 0..3 {
        print("for i=$i")
    }

    let spins = 0
    loop {
        spins++
        if (spins == 2) { continue }
        if (spins >= 4) { break }
        print("loop spins=$spins")
    }
}
```

---

## 10. Operators & bitwise

- **Arithmetic:** `+`, `-`, `*`, `/`, `%`; unary `-`, `+`.
- **Compare:** `==`, `!=`, `<`, `>`, `<=`, `>=`.
- **Logic:** **`and`**, **`or`**, **`not`** (spell words; wrap complex tests in parentheses when combining, e.g. `if ((a > 0) and (b > 0))`).
- **Increment:** postfix **`++`** and **`--`** on numeric lvalues.
- **Bitwise:** **`&`**, **`|`**, **`^`**, **`~`** on integers; binary literals **`0b1010`** are supported.

```kodae
fn operators_demo() {
    let a = 10
    let b = 3

    print(a + b)   ' 13
    print(a - b)   ' 7
    print(a * b)   ' 30
    print(a / b)   ' 3
    print(a % b)   ' 1
    print(-a)      ' unary minus
    print(+b)      ' unary plus

    if (a == 10) { print("eq") }
    if (a != b)  { print("neq") }
    if (a > b)   { print("gt") }
    if (a >= b)  { print("gte") }
    if (b < a)   { print("lt") }
    if (b <= a)  { print("lte") }

    if ((a > 0) and (b > 0)) { print("both positive") }
    if ((a < 0) or (b < 0))  { print("any negative") }
    if (not (a < 0))         { print("a not negative") }

    let c = 5
    c++
    c--

    let bits = 0b1010
    print(bits & 0b1100)
    print(bits | 0b0101)
    print(bits ^ 0b1111)
    print(~bits)
}
```

---

## 11. Methods, `this`, `with`, and inline lambdas

### Rules for `this`

1. **`this`** is only valid inside **`fn TypeName.method(...)`** (struct methods).
2. **`this`** always means the receiver instance for that method — including inside **`repeat`**, **`for`**, **`if`**, and nested blocks (lexical binding; unlike JavaScript, it does not “get lost” in nested functions).
3. **Inline lambdas** — **`let cb = fn() { ... }; cb()`** — only **`fn()` with no parameters** and **`void`** body are supported in v1. If the lambda uses **`this`**, it must appear **inside** a method so the compiler can pass the same receiver as C **`self`**. You cannot pass these lambdas to arbitrary C function-pointer parameters yet; call them from Kodae.

```kodae
struct Player {
    name: str
    hp: int
}

fn Player.tick() {
    repeat(2) {
        this.hp -= 1
    }

    let announce = fn() {
        print("$this.name now has $this.hp hp")
    }
    announce()
}
```

### `with` (functional update)

**`expr with { field: value, ... }`** builds a **copy** of a struct value and overrides listed fields:

```kodae
fn Player.clone_named(s: str) -> Player {
  return this with { name: s }
}
```

```kodae
fn with_demo() {
    let p1 = Player { name: "Ari", hp: 100 }
    let p2 = p1 with { hp: 80 }
    print("$p1.name hp=$p1.hp") ' original unchanged
    print("$p2.name hp=$p2.hp") ' copied and updated
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

```kodae
fn Player.heal(amount: int) -> Player {
    this.hp += amount
    return this
}

fn chain_demo() {
    let p = Player { name: "Neo", hp: 50 }
    p.tag("_pro").heal(25).tick()
}
```

See **SUPPORTED** for the precise status of **`this` / `with` / fn lambdas**.

---

## 12. Built-ins & standard helpers

### No module imports needed

For core utilities, **you do not need `module` or `#include`**. These functions are globally available in normal Kodae programs.

This does **not** remove directives: `#include` and `#library` remain supported and are the right tools for multi-file projects and library packaging.

### Printing & input

- `print(...)`, `printn(...)`
- `input(prompt)`, `input_int(prompt)`, `input_float(prompt)`
- `clear_screen()`

```kodae
fn io_demo() {
    clear_screen()
    print("Welcome")
    printn("Type your name: ")
    let name = input("")
    let age = input_int("Age: ")
    let speed = input_float("Speed: ")
    print("Hello $name, age $age, speed $speed")
}
```

### Math & geometry

- `abs`, `min`, `max`, `clamp`
- `floor`, `ceil`, `round`, `sqrt`, `pow`
- `sin`, `cos`, `tan`, `atan2`, `log`
- `lerp`, `map`, `distance`, `angle_to`
- helpers: `in_range`, `in_rect`

```kodae
fn math_demo() {
    print(abs(-7))
    print(min(2, 9))
    print(max(2, 9))
    print(clamp(15, 0, 10))

    print(floor(3.9))
    print(ceil(3.1))
    print(round(3.6))
    print(sqrt(25.0))
    print(pow(2.0, 8.0))

    print(sin(1.57))
    print(cos(0.0))
    print(tan(0.5))
    print(atan2(1.0, 1.0))
    print(log(10.0))

    print(lerp(0.0, 10.0, 0.25))
    print(map(5.0, 0.0, 10.0, 0.0, 100.0))
    print(distance(0.0, 0.0, 3.0, 4.0))
    print(angle_to(0.0, 0.0, 1.0, 1.0))

    print(in_range(8, 1, 10))
    print(in_rect(5, 5, 0, 0, 10, 10))
}
```

### Random

- `random`, `random_float`, `random_bool`, `chance`
- `random_pick(list)`
- `list.shuffle()`

```kodae
fn random_demo() {
    let n = random(1, 6)
    let f = random_float(0.0, 1.0)
    let coin = random_bool()
    if (chance(0.2)) {
        print("critical hit")
    }

    let loot = ["potion", "shield", "sword"]
    print(random_pick(loot))
    loot.shuffle()
    print(loot)
    print("n=$n f=$f coin=$coin")
}
```

### Time

- `time()`, `time_ms()`, `delta_time()`
- `wait(seconds)`, `wait_ms(ms)`
- `timer_start()`, `timer_elapsed(t)`
- `countdown(seconds)`, `countdown_done(t)`

```kodae
fn time_demo() {
    print(time())
    print(time_ms())
    print(delta_time())

    let t = timer_start()
    wait(0.1)
    wait_ms(50)
    print("elapsed=" + str(timer_elapsed(t)))

    let cd = countdown(1.0)
    while (not countdown_done(cd)) {
        print("countdown running...")
        wait(0.2)
    }
    print("countdown done")
}
```

### Strings

- `s.len`
- `s.upper()`, `s.lower()`, `s.trim()`
- `s.contains(x)`, `s.starts(x)`, `s.ends(x)`
- `s.replace(old, new)`, `s.split(delim)`, `s.slice(start, end)`
- `s.reverse()`, `s.repeat(n)`, `s.is_empty()`
- casts: `str(x)`, `int(x)`, `float(x)`, `bool(x)`

```kodae
fn string_demo() {
    let s = "  Kodae Rocks  "
    print(s.len)
    print(s.upper())
    print(s.lower())
    print(s.trim())

    let t = "kodae-language"
    print(t.contains("dae"))
    print(t.starts("kod"))
    print(t.ends("age"))
    print(t.replace("-", "_"))
    print(t.split("-"))
    print(t.slice(0, 5))
    print(t.reverse())
    print("ha".repeat(3))
    print("".is_empty())

    print(str(123))
    print(int("42"))
    print(float("3.14"))
    print(bool(1))
}
```

### Lists

- `push`, `pop`, `append`, `remove`
- `first`, `last`, `clear`, `contains`, `is_empty`
- `sort`, `reverse`, `shuffle`
- `list.len` and `len(list)`

```kodae
fn list_demo() {
    let nums: list[int] = [3, 1, 2]
    nums.push(4)
    nums.append([5, 6])
    print(nums.pop())
    nums.remove(0)

    print(nums.first())
    print(nums.last())
    print(nums.contains(2))
    print(nums.is_empty())

    nums.sort()
    nums.reverse()
    nums.shuffle()
    print(nums.len)
    print(len(nums))

    nums.clear()
    print(nums.is_empty())
}
```

### Files & JSON

- files: `read_file`, `write_file`, `append_file`, `copy_file`, `move_file`, `delete_file`
- folders: `make_folder`, `delete_folder`, `folder_exists`, `list_files`
- checks: `file_exists`
- JSON convenience:
  - text: `json_read(path)`, `json_write(path, value)`
  - parsed/Any API: `json_parse`, `json_build`, `json_get`, `json_at`, `json_len`, `json_as_int`, `json_as_float`, `json_as_str`, `json_as_bool`

```kodae
fn file_json_demo() {
    make_folder("save")
    write_file("save/data.txt", "hello")
    append_file("save/data.txt", "\nworld")
    print(read_file("save/data.txt"))

    copy_file("save/data.txt", "save/data_copy.txt")
    move_file("save/data_copy.txt", "save/data_moved.txt")
    print(file_exists("save/data_moved.txt"))
    print(folder_exists("save"))
    print(list_files("save"))

    delete_file("save/data_moved.txt")

    let text_json = json_read("save/player.json")
    print(text_json)
    json_write("save/player_out.json", "{\"score\":100}")

    let parsed = json_parse("{\"hp\":80,\"name\":\"Ari\",\"alive\":true}")
    let hp_any = json_get(parsed, "hp")
    print(json_as_int(hp_any))
    print(json_as_str(json_get(parsed, "name")))
    print(json_as_bool(json_get(parsed, "alive")))

    let arr = json_parse("[10, 20, 30]")
    print(json_len(arr))
    print(json_as_float(json_at(arr, 1)))
    print(json_build(parsed))

    delete_file("save/data.txt")
    delete_file("save/player_out.json")
    delete_folder("save")
}
```

### Save helpers (key-value)

- `save_set(key, value)` (value is stringified)
- `save_get_int(key)`, `save_get_str(key)`
- `save_exists(key)`, `save_delete(key)`, `save_clear()`

```kodae
fn save_demo() {
    save_set("coins", 120)
    save_set("player_name", "Ari")

    if (save_exists("coins")) {
        print(save_get_int("coins"))
    }
    print(save_get_str("player_name"))

    save_delete("coins")
    save_clear()
}
```

### Networking & system

- net: `is_online()`, `ping(host)`, `http_get(url)`, `http_post(url, body)`, `download(url, dest)`
- os: `os_name()`, `is_windows()`, `is_mac()`, `is_linux()`, `args()`, `env(name)`
- shell/browser: `run(cmd)`, `open_url(url)`, `exit(code)`

```kodae
fn net_system_demo() {
    print(is_online())
    print(ping("example.com"))
    print(http_get("https://example.com"))
    print(http_post("https://httpbin.org/post", "{\"ok\":true}"))
    print(download("https://example.com/file.txt", "file.txt"))

    print(os_name())
    print(is_windows())
    print(is_mac())
    print(is_linux())
    print(args())
    print(env("PATH"))

    run("echo hello-from-kodae")
    open_url("https://kodae.dev")
    ' exit(0) ' uncomment when you really want to terminate the app
}
```

### Debug helpers

```kodae
debug(player)               ' pretty-print a value
assert(health >= 0, "health cant be negative")
todo("finish this later")
log("player moved")
log_info("game started")
log_warn("low health: $health")
log_error("file not found: $path")
```

---

*End of language reference summary — for implementation detail and edge cases, combine this file with **SUPPORTED.md** and the docs linked in the table at the top.*
