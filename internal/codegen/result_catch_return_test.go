package codegen

import (
	"strings"
	"testing"

	"clio/internal/check"
	lex "clio/internal/lexer"
	"clio/internal/parser"
)

func TestEmit_CatchOnPlainValueLowersToValue(t *testing.T) {
	const src = `fn read_file(path: str) -> str { return path }
fn main() {
  let data = read_file("save.dat") catch (err) {
    print("Failed: " + err)
  }
  print(data)
}`
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
	if !strings.Contains(c, "read_file") || !strings.Contains(c, "save.dat") {
		t.Fatalf("expected C emission for catch expression subject, got:\n%s", c)
	}
}
