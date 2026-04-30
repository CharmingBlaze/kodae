package llir

import (
	"strings"
	"testing"

	"kodae/internal/check"
	lexapi "kodae/internal/lexer"
	"kodae/internal/parser"
)

func TestLowerToLLVM_HelloShape(t *testing.T) {
	t.Parallel()
	const src = `fn main() {
  let x: int = 1
  print("Hello")
  print("x = " + str(x))
}
`
	p, err := parser.Parse(lexapi.New(src))
	if err != nil {
		t.Fatal(err)
	}
	inf, err := check.Check(p)
	if err != nil {
		t.Fatal(err)
	}
	ir, err := LowerToLLVM(p, inf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(ir, "define i32 @main()") {
		t.Fatalf("missing main: %s", ir)
	}
	if !strings.Contains(ir, "@rt_sb_append_lit") {
		t.Fatalf("missing str append: %s", ir)
	}
	if !strings.Contains(ir, "@rt_print_int64") {
		t.Fatalf("missing print int path: %s", ir)
	}
}
