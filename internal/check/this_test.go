package check

import (
	"strings"
	"testing"

	lex "kodae/internal/lexer"
	"kodae/internal/parser"
)

func TestThis_OutsideMethodRejected(t *testing.T) {
	const src = `fn f() int { return this }`
	p := parser.New(lex.New(src))
	pr := p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
	_, err := Check(pr)
	if err == nil {
		t.Fatal("expected type error for this outside method")
	}
	if !strings.Contains(err.Error(), "this can only be used inside a method") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestThis_InsideMethodOK(t *testing.T) {
	const src = `struct Player { health: int }
fn Player.hurt(amount: int) {
  this.health -= amount
}`
	p := parser.New(lex.New(src))
	pr := p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
	if _, err := Check(pr); err != nil {
		t.Fatalf("check: %v", err)
	}
}
