package codegen

import (
	"strings"
	"testing"

	"kodae/internal/check"
	lex "kodae/internal/lexer"
	"kodae/internal/parser"
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
  let w = xs.len
  print(str(x + y + len(xs) + w))
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
		"kodae_list_new",
		"kodae_list_push",
		"kodae_list_append",
		"kodae_list_pop",
		"kodae_list_remove_at",
		"kodae_list_at_ptr",
	} {
		if !strings.Contains(c, want) {
			t.Fatalf("expected generated C to contain %q, got:\n%s", want, c)
		}
	}
}
