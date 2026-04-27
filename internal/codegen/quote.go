package codegen

import (
	"strconv"
	"strings"
)

// cStringLit returns a C string literal, including surrounding quotes, escaped for C
func cStringLit(s string) string {
	if !strings.ContainsAny(s, "\\\"\n\r\t") && strings.IndexByte(s, 0) < 0 {
		return `"` + s + `"`
	}
	return strconv.Quote(s) // Produces double-quoted Go; same escaping works for C for common cases
}
