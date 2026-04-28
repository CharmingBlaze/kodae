# Kodae Command-Line Interface (CLI)

This guide explains **every** command available when using `kodae`. It covers what each command is used for, how to use it, and when you should use it.

## Basic Usage

When you open your terminal, you interact with the Kodae compiler by typing `kodae` followed by a command:

```sh
kodae <command> [file.kodae]
```

*(Note: On Windows, depending on your setup, you might need to type `.\bin\kodae.exe` instead of just `kodae`)*

---

## 1. Running & Building Code

These are the commands you will use 99% of the time to build and test your apps and games.

### `kodae run`
- **What it is used for:** Instantly compiling and executing your code in one step.
- **How to use it:** `kodae run my_game.kodae`
- **When to use it:** When you are actively developing and want to see your code work immediately without leaving a compiled `.exe` file cluttering your directory.

### `kodae build`
- **What it is used for:** Compiling your code into a permanent, standalone executable file (`.exe` on Windows, or a binary on Linux/macOS) that you can share with others.
- **How to use it:** `kodae build my_game.kodae`
  - *(Optional)* Set a custom name with `-o`: `kodae build -o game_v1.exe my_game.kodae`
- **When to use it:** When your program is finished and you want to distribute it to users, or when you need a permanent executable file.

### `kodae check`
- **What it is used for:** Scanning your code for syntax and type errors without actually compiling or running it.
- **How to use it:** `kodae check my_game.kodae`
- **When to use it:** When you are writing a lot of code and want a blazing-fast way to check if you made any typos before you actually try to run it.

---

## 2. Code Organization

These commands help you organize your projects and share code across multiple files.

### `kodae install`
- **What it is used for:** Copying a `.kodae` file into your global user library directory, so you can `#include` it from anywhere on your computer without copying the file manually.
- **How to use it:** `kodae install path/to/my_math.kodae`
- **When to use it:** When you've written a useful utility file (like a math library or a game engine wrapper) and want to use it across multiple different projects.

---

## 3. Advanced Tools (Under the Hood)

These commands are mostly for debugging, learning how compilers work, or interacting with C libraries. Beginners rarely need these!

### `kodae cgen` (or `kodae c`)
- **What it is used for:** Kodae translates your code into C code before turning it into an executable. This command prints out the generated C code to the screen.
- **How to use it:** `kodae cgen my_game.kodae`
- **When to use it:** When you are curious about how Kodae works under the hood, or you want to see exactly what C code your script produced.

### `kodae buildc`
- **What it is used for:** Writing the generated C code to a file without building an executable.
- **How to use it:** `kodae buildc my_game.kodae -o generated.c`
- **When to use it:** When you want to take the generated C code and compile it yourself using a custom C compiler or build system (like CMake).

### `kodae bind`
- **What it is used for:** Automatically generating Kodae wrappers from existing C header files (`.h`). *(Requires LLVM/Clang to be installed on your system).*
- **How to use it:** `kodae bind raylib /path/to/raylib.h`
- **When to use it:** When you want to use a massive C library (like Raylib, SDL, or SQLite) in Kodae, and you don't want to type out hundreds of `extern fn` definitions by hand.

### `kodae parse` (or `kodae ast`)
- **What it is used for:** Printing the "Abstract Syntax Tree" (AST) of your program.
- **How to use it:** `kodae parse my_game.kodae`
- **When to use it:** Only used if you are trying to find a bug in the Kodae compiler itself.

### `kodae lex`
- **What it is used for:** Printing the raw text tokens the compiler sees before it even tries to understand the code.
- **How to use it:** `kodae lex my_game.kodae`
- **When to use it:** Only used if you are trying to find a bug in the Kodae compiler itself.

### `kodae bundle`
- **What it is used for:** Packaging the entire Kodae compiler and tools into a clean `dist/` folder. *(Requires Go to be installed).*
- **How to use it:** `kodae bundle`
- **When to use it:** If you are a contributor to the Kodae compiler and want to generate a `.zip` release to give to other people.

---

## 4. Other Utilities

### `kodae version`
- **What it is used for:** Checking which version of the Kodae compiler you have installed.
- **How to use it:** `kodae version`
- **When to use it:** When you are filing a bug report, or you want to make sure you have the latest updates!
