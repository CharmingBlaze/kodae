package check

import (
	"testing"

	lex "kodae/internal/lexer"
	"kodae/internal/parser"
)



func TestCatch_OnVoidRejected(t *testing.T) {
	const src = `fn log_msg() { print("x") }
fn main() {
  let x = log_msg() catch (err) { print(err) }
}`
	p := parser.New(lex.New(src))
	pr := p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
	_, err := Check(pr)
	if err == nil {
		t.Fatalf("expected check error for catch on void expression")
	}
}
