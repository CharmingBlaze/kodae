# Clio C Libraries

This guide covers exporting Clio code as a C library.

## 1) Mark exports with `pub`

```clio
#mode "library"
#library "mymath"
#version "1.0.0"
#author "Ada"

pub fn add(a: int, b: int) -> int {
  return a + b
}

pub struct Vec2 {
  x: float
  y: float
}
```

Only `pub` symbols are exported in the generated header.

## 2) Build as a library

```bash
clio build --lib mymath.clio
```

Produces:

- `mymath.c`
- `mymath.h`
- `mymath.a`
- shared library (`mymath.so` / `mymath.dll` / `mymath.dylib`)

## 3) ABI mapping

Public Clio -> C mapping:

- `int` -> `int64_t`
- `float` -> `double`
- `bool` -> `bool`
- `str` -> `const char*`
- `struct` -> generated C struct

Not exportable in `pub` API:

- `list[T]`
- `ptr[...]`
- optional/`none` types

## 4) String boundary

Exported `str` parameters/returns become `const char*` in generated headers.
Wrapper code converts between ABI strings and internal Clio string values.

## 5) C consumer example

```c
#include "mymath.h"
#include <stdio.h>

int main(void) {
  printf("%lld\n", (long long)add(2, 3));
  return 0;
}
```
