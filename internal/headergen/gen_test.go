package headergen

import (
	"strings"
	"testing"

	"kodae/internal/check"
	lex "kodae/internal/lexer"
	"kodae/internal/parser"
)

func TestGenerate_HeaderForPubAPI(t *testing.T) {
	const src = `struct Vec2 { x: float, y: float }
fn add(a: int, b: int) -> int { return a + b }
fn greet(name: str) -> str { return "hi " + name }`
	p := parser.New(lex.New(src))
	pr := p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
	inf, err := check.Check(pr)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	h, err := Generate(pr, inf, Options{LibraryName: "mylib"})
	if err != nil {
		t.Fatalf("headergen: %v", err)
	}
	for _, want := range []string{"typedef struct S_Vec2", "int64_t add(", "const char* greet("} {
		if !strings.Contains(h, want) {
			t.Fatalf("missing %q in header:\n%s", want, h)
		}
	}
}
