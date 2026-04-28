# Kodae Binding Generator (`kodae bind`)

The `kodae bind` command generates Kodae wrappers for C libraries by parsing their header files using LLVM/Clang.

## Prerequisites

- **Clang**: You must have `clang` installed and available on your `PATH`.
- The binding generator uses `clang -Xclang -ast-dump=json` to accurately understand C types and structures.

## Usage

```bash
kodae bind [-o output.kodae] <name> <path/to/header.h>
```

Put `-o` **before** `<name>` if you use it (Go’s flag parser stops at the first non-flag argument).

- `<name>`: The short name of the library (e.g., `raylib`, `sqlite3`). This will be used in the generated `# link` directive.
- `<path/to/header.h>`: The path to the main C header file.
- `-o output.kodae`: (Optional) The output path for the generated Kodae file. Defaults to `include/<name>/<name>.kodae`.

## What is generated?

### 1. Structs
C `struct` definitions are converted to Kodae `pub struct`.
- Fields are mapped to the closest Kodae type.
- Nested structs and complex types are supported if they are defined in the same header or can be resolved by Clang.

### 2. Enums
C `enum` definitions are converted to Kodae `pub enum`.

### 3. Extern Functions
C functions are converted to Kodae `extern fn`.
- Return types and parameters are mapped.
- `void` return type becomes `-> void`.
- Pointers (e.g., `const char*`, `void*`) are mapped to `ptr[byte]`.

## Type Mapping Reference

| C Type | Kodae Type | Notes |
|--------|-----------|-------|
| `int`, `long`, `long long` | `int` | Kodae `int` is 64-bit. |
| `float` | `f32` | Mapped to C `float`. Restricted to interop. |
| `double` | `float` | Kodae `float` is C `double`. Standard logic type. |
| `int32_t` | `i32` | Restricted to interop. |
| `unsigned char`, `uint8_t` | `u8` | Restricted to interop. |
| `char*`, `void*`, `T*` | `ptr[byte]` | Generic pointer type in Kodae. |

### Note on Sized Types (`i32`, `f32`, `u32`, `u8`)
These types are only allowed in `extern fn` signatures and `struct` fields. However, the Kodae compiler automatically coerces standard `int` and `float` values to these sized types when passing them to functions or assigning them to struct fields, making interop seamless for beginners.

## Example

Generating bindings for a simple header `math_utils.h`:

```c
// math_utils.h
struct Vector2 { float x, y; };
float Vector2Length(struct Vector2 v);
```

Run:
```bash
kodae bind math_utils math_utils.h
```

Generated `include/math_utils/math_utils.kodae`:
```kodae
' AUTO-GENERATED bindings for math_utils
# link "math_utils"

pub struct Vector2 {
  x: f32
  y: f32
}

extern fn Vector2Length(v: Vector2) -> f32
```
