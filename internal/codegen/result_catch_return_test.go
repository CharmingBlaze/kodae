package codegen

import (
	"strings"
	"testing"

	"clio/internal/check"
	lex "clio/internal/lexer"
	"clio/internal/parser"
)

func TestEmit_ReturnResultCatch(t *testing.T) {
	const src = `fn g() -> result[int] { return ok(1) }
fn f() -> int { return g() catch (e) { return 0 } }
fn main() { }`
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
	if !strings.Contains(c, "c_rc_") || !strings.Contains(c, "return (") {
		t.Fatalf("expected c_rc_ temp and return in else, got:\n%s", c)
	}
}
