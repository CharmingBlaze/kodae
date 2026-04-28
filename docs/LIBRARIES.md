# Building Shared Libraries in Kodae

Kodae allows you to build shared libraries (`.a`, `.h`, `.c`) that you can give to C or C++ developers to use in their projects!

## 1. Creating a Library

To create a library, you use `#mode "library"` and `#library "name"`. Every function and struct you write will automatically be exported to the C header file, as long as it uses C-compatible types!

**mymath.kodae**
```kodae
#mode "library"
#library "mymath"
#version "1.0.0"
#author "Ada"

struct Vec2 {
    x: float
    y: float
}

fn add_vectors(a: Vec2, b: Vec2) -> Vec2 {
    return Vec2 {
        x: a.x + b.x,
        y: a.y + b.y
    }
}
```

## 2. Building the Library

Run the following command in your terminal:
`kodae build --lib mymath.kodae`

Kodae will generate 3 files for you:
1. `mymath.a`: The static C library file.
2. `mymath.h`: The C header file containing the definitions.
3. `mymath.c`: The C source code.

## 3. Using it in C

You can now use `mymath.h` and `mymath.a` in a standard C program.

**main.c**
```c
#include "mymath.h"
#include <stdio.h>

int main() {
    S_Vec2 a = { 1.5, 2.0 };
    S_Vec2 b = { 0.5, 3.0 };
    S_Vec2 c = add_vectors(a, b);
    
    printf("Result: %f, %f\n", c.x, c.y);
    return 0;
}
```

Compile it with GCC:
`gcc main.c mymath.a -o main.exe`

## Unsupported Types

Note: Kodae automatically skips functions that use complex Kodae-specific types that C cannot easily understand. 
If your function uses `list`, `ptr`, `none`, or `Any`, it will simply not appear in the generated `.h` file!
