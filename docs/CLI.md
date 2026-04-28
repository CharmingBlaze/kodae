# Kodae Command Line Interface (CLI)

The Kodae CLI is your main tool for building and running your Kodae programs.

## Overview of Commands

You can interact with Kodae through several main commands:

- `kodae run <file>`: Compile and run a Kodae program immediately.
- `kodae build <file>`: Compile a Kodae program into a standalone executable.
- `kodae test <file_or_dir>`: Run tests in your Kodae files.
- `kodae install <file>`: Install a Kodae library globally on your computer so you can include it anywhere.
- `kodae c <file>`: Generate C code from your Kodae program (advanced).

---

## Command Details

### 1. `run`
**Usage:** `kodae run <filename.kodae>`

This is the command you'll use most often during development. It takes your Kodae source file, compiles it in the background, and immediately runs it. It's the fastest way to test your code.

**Example:**
`kodae run main.kodae`

### 2. `build`
**Usage:** `kodae build [options] <filename.kodae>`

When you are ready to share your application or game, use the `build` command. It compiles your code into a permanent executable file (like a `.exe` on Windows).

**Options:**
- `-o <output_name>`: Specify the name of the final executable file. (Example: `kodae build -o game.exe main.kodae`)
- `--cc <compiler>`: Specify a custom C compiler (like `clang` or `gcc`) instead of the default.
- `--lib`: Build a C-compatible shared library (`.a` / `.h` / `.c`) instead of an executable. This is useful if you are writing a library in Kodae that you want C/C++ developers to use.

**Example:**
`kodae build -o my_app main.kodae`

### 3. `install`
**Usage:** `kodae install <filename.kodae>`

If you write a helpful library of functions (like math tools or game logic) and want to use it in all of your projects, you can "install" it. This copies the `.kodae` file to a global library folder on your computer.

Once installed, you can use `#include "filename"` in any project to access that library!

**Example:**
`kodae install mathlib.kodae`

### 4. `test`
**Usage:** `kodae test [filename_or_directory]`

This command searches for any function in your code that starts with the word `test` (like `fn test_math()`) and runs it automatically. It will report whether the tests passed or failed.

**Example:**
`kodae test .` (Runs all tests in the current folder)

### 5. `c`
**Usage:** `kodae c <filename.kodae>`

This command transpiles your Kodae code into readable C code and stops without compiling it into an executable. This is very useful if you want to see exactly how Kodae translates your code, or if you want to manually compile the C code yourself.

**Example:**
`kodae c main.kodae`
