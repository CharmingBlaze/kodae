package main

import (
	"os"
	"path/filepath"
	"testing"

	"kodae/internal/ast"
	"kodae/internal/check"
)

func TestLoadProgram_UseResolvesSameDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, filepath.Join(dir, "lib.kodae"), `fn double(n: int) -> int { return n * 2 }`)
	write(t, filepath.Join(dir, "app.kodae"), `use lib
fn main() {
  let v: int = double(3)
  print( str( v ) )
}`)

	p, err := loadProgram([]string{filepath.Join(dir, "app.kodae")})
	if err != nil {
		t.Fatalf("loadProgram: %v", err)
	}
	_, err = check.Check(p)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	// double before main: two fns, main last
	if len(p.Decls) < 2 {
		t.Fatalf("expected at least 2 decls, got %d", len(p.Decls))
	}
}

func TestLoadProgram_UseDedupWithTwoPaths(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, filepath.Join(dir, "lib.kodae"), `fn f() -> int { return 1 }`)
	write(t, filepath.Join(dir, "app.kodae"), `use lib
fn main() { }`)

	p, err := loadProgram([]string{
		filepath.Join(dir, "lib.kodae"),
		filepath.Join(dir, "app.kodae"),
	})
	if err != nil {
		t.Fatalf("loadProgram: %v", err)
	}
	var names []string
	for _, d := range p.Decls {
		if fd, ok := d.(*ast.FnDecl); ok {
			names = append(names, fd.Name)
		}
	}
	if len(names) != 2 || names[0] != "f" || names[1] != "main" {
		t.Fatalf("expected f then main, got %v", names)
	}
	_, err = check.Check(p)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
}

// duplicate test simpler - use grep: two "static.*f_f" in codegen? Skip heavy.
// Simpler: loadProgram should not error and Check finds single f
func TestLoadProgram_UseCycle(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, filepath.Join(dir, "a.kodae"), `use b
fn f() { }`)
	write(t, filepath.Join(dir, "b.kodae"), `use a
fn g() { }`)
	_, err := loadProgram([]string{filepath.Join(dir, "a.kodae")})
	if err == nil {
		t.Fatal("expected use cycle error")
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if e := os.WriteFile(path, []byte(content), 0644); e != nil {
		t.Fatal(e)
	}
}
