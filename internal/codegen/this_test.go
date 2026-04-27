package codegen

import (
	"strings"
	"testing"

	"clio/internal/check"
	lex "clio/internal/lexer"
	"clio/internal/parser"
)

func TestEmit_ImplicitThisMapsToSelf(t *testing.T) {
	const src = `struct Player { health: int }
fn Player.move(dx: int) {
  this.health += dx
}
fn Player.hurt(amount: int) {
  this.move(amount)
}
fn main() {}`
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
	if !strings.Contains(c, "static void f_Player_move(S_Player* self, int64_t dx)") {
		t.Fatalf("expected hidden self param in C method signature:\n%s", c)
	}
	if !strings.Contains(c, "((self)->u_health) += dx;") {
		t.Fatalf("expected this.health to emit as self->field:\n%s", c)
	}
	if !strings.Contains(c, "f_Player_move(self, amount)") {
		t.Fatalf("expected this.move(...) to pass self as receiver:\n%s", c)
	}
}
