package parser

import (
	"testing"

	"clio/internal/ast"
	lex "clio/internal/lexer"
)

func TestParse_EnumAndMatch(t *testing.T) {
	const src = `
enum S { A, B }
fn f() {
  let x: S = S.A
  match (x) {
    S.A => { return }
    S.B => { return }
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

func TestParse_V1RejectsResultType(t *testing.T) {
	const src = `fn f() -> result[int] { return 0 }`
	p := New(lex.New(src))
	_ = p.ParseProgram()
	if p.Err() == nil {
		t.Fatalf("expected parse error for result type")
	}
}

func TestParse_V1RejectsTryUnwrap(t *testing.T) {
	const src = `fn f() { g()? }`
	p := New(lex.New(src))
	_ = p.ParseProgram()
	if p.Err() == nil {
		t.Fatalf("expected parse error for ? operator")
	}
}

func TestParse_V1RejectsOptionalSyntax(t *testing.T) {
	const src = `fn f() { let x: int? = none }`
	p := New(lex.New(src))
	_ = p.ParseProgram()
	if p.Err() == nil {
		t.Fatalf("expected parse error for T? syntax")
	}
}

func TestParse_MethodImplicitThisAddsHiddenSelf(t *testing.T) {
	const src = `struct Player { health: int }
fn Player.hurt(amount: int) {
  this.health -= amount
}`
	p := New(lex.New(src))
	pr := p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
	var fn *ast.FnDecl
	for _, d := range pr.Decls {
		if f, ok := d.(*ast.FnDecl); ok && f.Name == "Player_hurt" {
			fn = f
			break
		}
	}
	if fn == nil {
		t.Fatalf("expected method decl Player_hurt")
	}
	if len(fn.Params) != 2 {
		t.Fatalf("expected hidden self + amount params, got %d", len(fn.Params))
	}
	if fn.Params[0].Name != "self" || fn.Params[0].T == nil || fn.Params[0].T.Name != "Player" {
		t.Fatalf("first param must be hidden self: Player, got %+v", fn.Params[0])
	}
}

func TestParse_ListTypeLiteralAndIndex(t *testing.T) {
	const src = `fn main() {
  let xs: list[int] = [1, 2, 3]
  xs[1] = 9
  let y = xs[1]
  xs.push(10)
}`
	p := New(lex.New(src))
	_ = p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
}
