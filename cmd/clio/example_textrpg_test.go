package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"clio/internal/ccdriver"
	"clio/internal/check"
	"clio/internal/codegen"
	lexapi "clio/internal/lexer"
	"clio/internal/parser"
)

// Example features.clio: parse, check, and emit (Phase-2 style: defer, continue, and/or, ++/--).
func TestFeaturesClioCompiles(t *testing.T) {
	t.Parallel()
	path := findExample(t, "features.clio")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	p, err := parser.Parse(lexapi.New(string(b)))
	if err != nil {
		t.Fatal("parse:", err)
	}
	inf, err := check.Check(p)
	if err != nil {
		t.Fatal("check:", err)
	}
	if _, err := codegen.EmitC(p, inf); err != nil {
		t.Fatal("codegen:", err)
	}
}

func findExample(t *testing.T, name string) string {
	t.Helper()
	candidates := []string{
		filepath.Join("examples", name),
		filepath.Join("..", "examples", name),
		filepath.Join("..", "..", "examples", name),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	t.Fatalf("find %q: not in %v", name, candidates)
	return ""
}

// Example textrpg.clio: must parse, type-check, and emit C (ensures the showcase game stays buildable).
func TestTextrpgCompiles(t *testing.T) {
	t.Parallel()
	path := findExample(t, "textrpg.clio")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	p, err := parser.Parse(lexapi.New(string(b)))
	if err != nil {
		t.Fatal("parse:", err)
	}
	inf, err := check.Check(p)
	if err != nil {
		t.Fatal("check:", err)
	}
	c, err := codegen.EmitC(p, inf)
	if err != nil {
		t.Fatal("codegen:", err)
	}
	if !strings.Contains(c, "int main(") {
		t.Fatal("generated C missing int main")
	}
}

// If a C compiler is on PATH, link textrpg (skips in CI when no cc).
func TestTextrpgLinks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping link in -short")
	}
	ccc, err := ccdriver.Find("")
	if err != nil {
		t.Skip(err)
	}
	path := findExample(t, "textrpg.clio")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	p, err := parser.Parse(lexapi.New(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	inf, err := check.Check(p)
	if err != nil {
		t.Fatal(err)
	}
	csrc, err := codegen.EmitC(p, inf)
	if err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp("", "clio-textrpg-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	cf := filepath.Join(dir, "out.c")
	if err := os.WriteFile(cf, []byte(csrc), 0644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "trpg")
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	if err := ccdriver.Compile(ccc, cf, out, nil); err != nil {
		t.Fatal(err)
	}
	// No run (needs stdin); link success = enough.
}
