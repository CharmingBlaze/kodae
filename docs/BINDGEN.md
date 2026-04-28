# Automatic C Bindings (Bindgen)

Kodae includes an advanced tool called `kodae-bind` that automatically reads C header files (`.h`) and generates Kodae `extern fn` bindings for you!

This means you don't have to manually type out hundreds of `extern fn` signatures for big libraries like Raylib or SDL.

## How to use `kodae-bind`

You run `kodae-bind` from the terminal, pointing it to the C header file you want to parse:

`kodae-bind -i raylib.h -o libs/raylib.kodae -lib raylib`

- `-i`: The input C header file.
- `-o`: The output Kodae file.
- `-lib`: The name of the library (this will add `#link "raylib"` to the top of the generated file).

## What it does

- C `struct` definitions are converted to Kodae `struct`.
- C `enum` definitions are converted to Kodae `enum`.
- C `#define` constants are converted to Kodae `const`.
- C function prototypes are converted to Kodae `extern fn`.

Once generated, you can simply `#include "libs/raylib"` in your project and start making games!
