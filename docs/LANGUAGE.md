# Kodae Language Reference

Kodae is designed to be as easy to read as Python or BASIC, but as fast and powerful as C. Everything is public by default, meaning you don't need to worry about complex access modifiers or keywords when writing your code.

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

Sometimes functions can fail (like trying to read a file that doesn't exist). Kodae uses a `result` type for this.
You can return an error using `error()`, and catch errors using `catch`.

```kodae
fn divide(a: int, b: int) -> result[int] {
    if (b == 0) {
        return error("Cannot divide by zero")
    }
    return a / b
}

fn main() {
    ' Use catch to handle the error, providing a default value (like 0)
    let safe_result = catch divide(10, 0) { 0 }
    print(safe_result) ' Prints 0
}
```
