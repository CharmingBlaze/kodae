# Interacting with C Libraries

Kodae is designed to easily connect with any existing C code or libraries (like Raylib for games, or standard math libraries).

## 1. Calling a C Function (`extern fn`)

If you want to call a C function in Kodae, you tell Kodae about it using the `extern fn` keyword. You do not need to write the function body; you are just giving Kodae the signature.

```kodae
#link "m"  ' Link the math library (libm.a or libm.so)

' Tell Kodae that the C function `sin` exists
extern fn sin(x: float) -> float

fn main() {
    let result = sin(3.14)
    print(result)
}
```

## 2. Using `#link`
To use C functions, you must tell the compiler to "link" the actual C library during the build.
You do this using the `#link` directive.

- `#link "raylib"` links the Raylib library (`-lraylib`)
- `#link "opengl32"` links the OpenGL library (`-lopengl32`)
- `#link "ws2_32"` links the Windows sockets library (`-lws2_32`)

## 3. Sized Types for C (Advanced)

C uses specific sizes for numbers (like 8-bit integers or 32-bit floats). Since Kodae usually figures this out for you with `int` and `float`, when you write an `extern fn` or a `struct` that needs to match C exactly, you use special sized types:

- `i32`: 32-bit integer (C `int`)
- `u32`: 32-bit unsigned integer (C `uint32_t`)
- `u8`: 8-bit unsigned integer (C `uint8_t` or `unsigned char`)
- `f32`: 32-bit float (C `float`)
- `ptr[byte]`: A raw pointer to bytes (C `void*` or `char*`)

**Example linking to a C Game Engine:**

```kodae
#link "raylib"

struct Color {
    r: u8
    g: u8
    b: u8
    a: u8
}

extern fn InitWindow(width: int, height: int, title: str)
extern fn ClearBackground(color: Color)
extern fn CloseWindow()

fn main() {
    InitWindow(800, 600, "My Game")
    
    let red = Color { r: 255, g: 0, b: 0, a: 255 }
    ClearBackground(red)
    
    CloseWindow()
}
```

Notice that even if the C library expects an `i32` or `f32`, you can still pass standard Kodae `int` and `float` variables to the `extern fn`! Kodae handles the conversion automatically behind the scenes.
