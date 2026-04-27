package parser

import (
	"testing"

	lex "clio/internal/lexer"
)

func TestParse_EnumAndMatch(t *testing.T) {
	const src = `
enum S { A, B }
fn f() {
  let x: S? = none
  match (x) {
    none => { return }
  }
}
`
	p := New(lex.New(src))
	_ = p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
}

func TestParse_CompoundAndCast(t *testing.T) {
	const src = `fn f() {
  let a = 1.0
  let b = int(a)
}`
	p := New(lex.New(src))
	_ = p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
}

// Keyword `result` tokenizes as RESULT, not IDENT; `ok`/`err` are OK/ERR in expr position.
func TestParse_ResultTypeAndOkErr(t *testing.T) {
	const src = `fn f() -> result[int] { return ok(0) }
fn g() -> result[int] { return err("e") }`
	p := New(lex.New(src))
	_ = p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
}

func TestParse_TryUnwrap(t *testing.T) {
	const src = `fn f() -> result[int] {
  let a = g()?
  return a
}`
	p := New(lex.New(src))
	_ = p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
}

func TestParse_ResultCatch(t *testing.T) {
	const src = `fn g() -> result[str] { return ok("x") }
fn f() {
  let a = g() catch (e) { print( "e" ) }
}`
	p := New(lex.New(src))
	_ = p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
}
