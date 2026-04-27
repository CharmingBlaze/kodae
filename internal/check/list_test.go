package check

import (
	"strings"
	"testing"

	lex "clio/internal/lexer"
	"clio/internal/parser"
)

func TestList_CoreTypingAndMethods(t *testing.T) {
	const src = `fn main() {
  let xs: list[int] = [1, 2, 3]
  xs.push(4)
  let ys: list[int] = [5, 6]
  xs.append(ys)
  let a = xs.pop()
  let b = xs.remove(0)
  let n = len(xs)
  xs[0] = a + b + n
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

func TestList_HeterogeneousLiteralRejected(t *testing.T) {
	const src = `fn main() { let xs = [1, "x"] }`
	p := parser.New(lex.New(src))
	pr := p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
	_, err := Check(pr)
	if err == nil {
		t.Fatal("expected type error for heterogeneous list literal")
	}
}

func TestList_EmptyNeedsAnnotation(t *testing.T) {
	const src = `fn main() { let xs = [] }`
	p := parser.New(lex.New(src))
	pr := p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
	_, err := Check(pr)
	if err == nil {
		t.Fatal("expected type error for empty list without annotation")
	}
	if !strings.Contains(err.Error(), "cannot infer element type") {
		t.Fatalf("unexpected error: %v", err)
	}
}
