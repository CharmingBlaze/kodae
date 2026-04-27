package check

import (
	"testing"

	lex "clio/internal/lexer"
	"clio/internal/parser"
)

func TestTryUnwrap_RejectedInIf(t *testing.T) {
	const src = `fn one() -> result[int] { return ok(1) }
fn f() -> result[int] {
  if (one()?) { return ok(0) }
  return err("x")
}`
	p := parser.New(lex.New(src))
	pr := p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
	_, err := Check(pr)
	if err == nil {
		t.Fatal("expected type error: ? in if condition")
	}
	_ = pr
}

func TestTryUnwrap_OKInLet(t *testing.T) {
	const src = `fn one() -> result[int] { return ok(1) }
fn f() -> result[int] {
  let x = one()?
  return ok(x)
}`
	p := parser.New(lex.New(src))
	pr := p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
	_, err := Check(pr)
	if err != nil {
		t.Fatalf("typecheck: %v", err)
	}
}
