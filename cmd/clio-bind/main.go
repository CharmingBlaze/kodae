// Command clio-bind: planned binder for C headers (see build-spec) — not yet implemented.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "clio-bind: not implemented; use extern fn and # link in Clio for now")
	os.Exit(1)
}
