package check

import (
	"testing"

	lex "clio/internal/lexer"
	"clio/internal/parser"
)

func TestCheck_LinkAndLinkpathToLdFlags(t *testing.T) {
	t.Parallel()
	const src = `# link "raylib"
# linkpath "./raylib"
extern fn f() -> void
fn main() { }`
	p := parser.New(lex.New(src))
	pr := p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
	inf, err := Check(pr)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	want := []string{"-lraylib", "-L./raylib"}
	if len(inf.LinkFlags) != len(want) {
		t.Fatalf("LinkFlags=%v", inf.LinkFlags)
	}
	for i, w := range want {
		if i >= len(inf.LinkFlags) || inf.LinkFlags[i] != w {
			t.Fatalf("LinkFlags: got %v want %v", inf.LinkFlags, want)
		}
	}
}

func TestCheck_LinkPreformattedPassesThrough(t *testing.T) {
	t.Parallel()
	const src = `# link "-lfoo -L/bar"
extern fn f() -> void
fn main() { }`
	p := parser.New(lex.New(src))
	pr := p.ParseProgram()
	if p.Err() != nil {
		t.Fatalf("parse: %v", p.Err())
	}
	inf, err := Check(pr)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if len(inf.LinkFlags) < 2 || inf.LinkFlags[0] != "-lfoo" || inf.LinkFlags[1] != "-L/bar" {
		t.Fatalf("LinkFlags=%v", inf.LinkFlags)
	}
}
