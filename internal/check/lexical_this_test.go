package check

import (
	"strings"
	"testing"

	lex "kodae/internal/lexer"
	"kodae/internal/parser"
)

func TestLexicalThis_RepeatAndForInsideMethod(t *testing.T) {
	const src = `
struct P { x: int }
fn P.bump() {
  repeat(2) {
    this.x += 1
  }
  for i in 0..2 {
    this.x += i
  }
}
fn main() {}
`
	pr := parser.New(lex.New(src)).ParseProgram()
	_, err := Check(pr)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
}

func TestLexicalThis_ReturnThisAndWith(t *testing.T) {
	const src = `
struct P { name: str health: int }
fn P.me() -> P {
  return this
}
fn P.copy() -> P {
  return this with { name: "copy" }
}
fn main() {}
`
	pr := parser.New(lex.New(src)).ParseProgram()
	_, err := Check(pr)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
}

func TestLexicalThis_LambdaCapturesThis(t *testing.T) {
	const src = `
struct P { x: int }
fn P.run() {
  let cb = fn() {
    this.x += 1
  }
  cb()
}
fn main() {}
`
	pr := parser.New(lex.New(src)).ParseProgram()
	inf, err := Check(pr)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if inf.Closures == nil || len(inf.Closures) != 1 {
		t.Fatalf("expected one closure metadata, got %v", inf.Closures)
	}
}

func TestLexicalThis_LambdaThisOutsideMethodRejected(t *testing.T) {
	const src = `
fn g() {
  let cb = fn() {
    print(this)
  }
}
fn main() {}
`
	pr := parser.New(lex.New(src)).ParseProgram()
	_, err := Check(pr)
	if err == nil {
		t.Fatal("expected error for this in lambda outside method")
	}
	if !strings.Contains(err.Error(), "this") {
		t.Fatalf("unexpected: %v", err)
	}
}
