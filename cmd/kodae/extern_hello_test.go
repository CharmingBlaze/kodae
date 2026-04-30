package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"kodae/internal/ccdriver"
	"kodae/internal/check"
	"kodae/internal/codegen"
	lexapi "kodae/internal/lexer"
	"kodae/internal/parser"
)

func TestExternPrintfExampleLinks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping link in -short")
	}
	ccc, err := ccdriver.Find(ccdriver.FindConfig{})
	if err != nil {
		t.Skip(err)
	}
	path := findExample(t, "extern_hello.kodae")
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
	csrc, err := codegen.EmitC(p, inf)
	if err != nil {
		t.Fatal("codegen:", err)
	}
	if !strings.Contains(csrc, "printf(") {
		t.Fatal("expected direct printf call in C")
	}
	dir, err := os.MkdirTemp("", "kodae-extern-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	cf := filepath.Join(dir, "out.c")
	if err := os.WriteFile(cf, []byte(csrc), 0644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "ex")
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	if err := ccdriver.Compile(ccc, cf, out, nil, false); err != nil {
		t.Fatal("link:", err)
	}
}
