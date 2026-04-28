package main

import (
	"os"
	"path/filepath"
	"testing"

	"kodae/internal/check"
)

func TestParseBuildFlags_KodaeAnyOrder(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in       []string
		files   []string
		o        string
		err      bool
	}{
		{[]string{"-o", "out.exe", "src/x.kodae"}, []string{"src/x.kodae"}, "out.exe", false},
		{[]string{"src/x.kodae", "-o", "out.exe"}, []string{"src/x.kodae"}, "out.exe", false},
		{[]string{"-o", "out.c", "prog.kodae", "--cc", "gcc"}, []string{"prog.kodae"}, "out.c", false},
		{[]string{"a.kodae", "b.kodae"}, []string{"a.kodae", "b.kodae"}, "", false},
	}
	for i, c := range cases {
		f, o, _, _, err := parseBuildFlags(c.in)
		if c.err && err == nil {
			t.Fatalf("%d: want error, got file=%q out=%q", i, f, o)
		}
		if !c.err {
			if err != nil {
				t.Fatalf("%d: %v", i, err)
			}
			if len(f) != len(c.files) {
				t.Fatalf("%d: got files=%#v want %#v", i, f, c.files)
			}
			for j := range c.files {
				if f[j] != c.files[j] {
					t.Fatalf("%d: got files=%#v want %#v", i, f, c.files)
					break
				}
			}
			if o != c.o {
				t.Fatalf("%d: got out=%q want %q", i, o, c.o)
			}
		}
	}
}

// Two-file merge: lib defines fn; app calls it (order matches kodae build lib.kodae app.kodae)
func TestLoadProgramTwoFileMergeTypechecks(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	lib := filepath.Join(dir, "lib.kodae")
	app := filepath.Join(dir, "app.kodae")
	if err := os.WriteFile(lib, []byte("pub fn double(n: int) -> int { return n * 2 }\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(app, []byte("fn main() {\n  let v: int = double(10)\n  print( str( v ) )\n}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// call double from a file that appears *after* its definition in merge order
	p, err := loadProgram([]string{lib, app})
	if err != nil {
		t.Fatalf("loadProgram: %v", err)
	}
	if _, err := check.Check(p); err != nil {
		t.Fatalf("check: %v", err)
	}
	if len(p.Decls) < 2 {
		t.Fatalf("want at least 2 top-level decls, got %d", len(p.Decls))
	}
}
