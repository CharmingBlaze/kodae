package codegen

import (
	"strings"
	"testing"

	"kodae/internal/check"
	lex "kodae/internal/lexer"
	"kodae/internal/parser"
)

func TestEmit_StructWithAndLambdaUsesSelf(t *testing.T) {
	const src = `
struct P { name: str x: int }
fn P.example() {
  let q = this with { name: "n", x: 3 }
  print(q.name)
  let cb = fn() {
    this.x += 1
  }
  cb()
}
fn main() {}
`
	pr := parser.New(lex.New(src)).ParseProgram()
	inf, err := check.Check(pr)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	c, err := EmitC(pr, inf)
	if err != nil {
		t.Fatalf("emit: %v", err)
	}
	if !strings.Contains(c, "_uw") {
		t.Fatalf("expected struct-update lowering with _uw:\n%s", c)
	}
	if !strings.Contains(c, "_lam") || !strings.Contains(c, "(self)") {
		t.Fatalf("expected lambda forward/def/call with self:\n%s", c)
	}
}
