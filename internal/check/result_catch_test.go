package check

import (
	"testing"

	lex "clio/internal/lexer"
	"clio/internal/parser"
)

func TestResultCatch_OK(t *testing.T) {
	const src = `fn okf() -> result[str] { return ok("hi") }
fn with_c() {
  let data = okf() catch (e) {
    print("fail")
  }
}`
	p := parser.New(lex.New(src))
	pr := p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
	_, err := Check(pr)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
}

func TestResultCatch_ReturnValue(t *testing.T) {
	const src = `fn g() -> result[int] { return ok(1) }
fn f() -> int { return g() catch (e) { return 0 } }
fn main() { }`
	p := parser.New(lex.New(src))
	pr := p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
	_, err := Check(pr)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
}

func TestResultCatch_NestedInAdd(t *testing.T) {
	const src = `fn okf() -> result[str] { return ok("a") }
fn f() {
  let x = 1 + okf() catch (e) { }
}`
	p := parser.New(lex.New(src))
	pr := p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
	_, err := Check(pr)
	if err == nil {
		t.Fatal("expected error: catch not full expression")
	}
}
