# Kodae Directives

Directives are special commands that you place at the top of your Kodae files to tell the compiler how to handle your code. They always start with the `#` symbol.

## 1. `#include`
**Usage:** `#include "filename"`

This is the most important directive. It allows you to split your project across multiple files. When you `#include` another file, all of its functions, structs, and variables become available to use in your current file automatically.

When you use `#include`, Kodae looks for the file in this order:
1. The **same directory** as your current file.
2. A **`libs/`** folder inside your current directory.
3. Your **global user library folder** (where `kodae install` puts files).

**Example:**
```kodae
#include "mathlib"

fn main() {
    let result = square(5)
    print(result)
}
```

## 2. `#link`
**Usage:** `#link "library_name"`

If you are using external C libraries (like Raylib for games or cURL for networking), you use `#link` to tell the compiler to link them during the build process.

For example, `#link "raylib"` tells the underlying C compiler to add `-lraylib` when it compiles your code.

**Example:**
```kodae
#link "raylib"
#link "opengl32"
#link "gdi32"
#link "winmm"

extern fn InitWindow(width: int, height: int, title: str)
```

## 3. `#library`
**Usage:** `#library "name"`

If you are building a library that you want to share with other people (especially if you are compiling it to a C library with `kodae build --lib`), you use this directive to give your library a formal name. This name will be used to generate the output files (like `name.h` and `name.a`).

## 4. `#version` and `#author`
**Usage:** `#version "1.0"` / `#author "Name"`

These optional directives allow you to attach metadata to your file. This is useful when you are writing reusable libraries.

**Example of a Library File:**
```kodae
#library "MyMath"
#version "1.0.0"
#author "Ada"

fn add(a: int, b: int) -> int {
    return a + b
}
```
