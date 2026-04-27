package codegen

import (
	"strings"
	"testing"

	"clio/internal/check"
	lex "clio/internal/lexer"
	"clio/internal/parser"
)

func TestEmit_ListOperations(t *testing.T) {
	const src = `fn main() {
  let xs: list[int] = [1, 2, 3]
  xs.push(4)
  let ys: list[int] = [5, 6]
  xs.append(ys)
  xs[1] = 9
  let x = xs.pop()
  let y = xs.remove(0)
  print(str(x + y + len(xs)))
}`
	p := parser.New(lex.New(src))
	pr := p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
	inf, err := check.Check(pr)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	c, err := EmitC(pr, inf)
	if err != nil {
		t.Fatalf("codegen: %v", err)
	}
	for _, want := range []string{
		"clio_list_new",
		"clio_list_push",
		"clio_list_append",
		"clio_list_pop",
		"clio_list_remove_at",
		"clio_list_at_ptr",
	} {
		if !strings.Contains(c, want) {
			t.Fatalf("expected generated C to contain %q, got:\n%s", want, c)
		}
	}
}
