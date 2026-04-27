package parser

import (
	"strings"
	"testing"

	"clio/internal/ast"
)

func containsAll(s string, sub ...string) bool {
	for _, x := range sub {
		if !strings.Contains(s, x) {
			return false
		}
	}
	return true
}

func TestExpandStringInterpolation_Plain(t *testing.T) {
	e, err := ExpandStringInterpolation("hello")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := e.(*ast.StringLit); !ok {
		t.Fatalf("want StringLit, got %T", e)
	}
}

func TestExpandStringInterpolation_Interp(t *testing.T) {
	e, err := ExpandStringInterpolation("Hi $name")
	if err != nil {
		t.Fatal(err)
	}
	b, ok := e.(*ast.BinaryExpr)
	if !ok || b.Op != "+" {
		t.Fatalf("want a+b, got %T", e)
	}
}

func TestExpandStringInterpolation_DollarPath(t *testing.T) {
	e, err := ExpandStringInterpolation("G $self.hp $player.gold")
	if err != nil {
		t.Fatal(err)
	}
	// $self.hp and $player.gold become member exprs; dump should show ".hp" and ".gold"
	s := ast.ExprString(e)
	if s == "" {
		t.Fatal("empty expr string")
	}
	if !containsAll(s, "self.hp", "player.gold") {
		t.Fatalf("expected member paths in dump, got %q", s)
	}
}

func TestExpandStringInterpolation_DoubleDollar(t *testing.T) {
	e, err := ExpandStringInterpolation("a$$b")
	if err != nil {
		t.Fatal(err)
	}
	// a + $ + b  => String "a" + String "$" + Ident b
	b, ok := e.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("want + chain, got %T: %s", e, ast.ExprString(e))
	}
	if b.Op != "+" {
		t.Fatal(b.Op)
	}
}

func TestParseExpressionFragment_Simple(t *testing.T) {
	e, err := ParseExpressionFragment("str(score)")
	if err != nil {
		t.Fatal(err)
	}
	if ast.ExprString(e) == "" {
		t.Fatal("empty string")
	}
}
