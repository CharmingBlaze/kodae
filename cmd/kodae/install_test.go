package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunInstall_CopiesToUserLibDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("KODAE_HOME", home)
	src := filepath.Join(t.TempDir(), "mathlib.kodae")
	if err := os.WriteFile(src, []byte(`#library "mathlib"
pub fn square(n: int) int { return n * n }
`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := runInstall(src); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(home, "libs", "mathlib.kodae"))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) == "" {
		t.Fatal("empty install")
	}
}

func TestRunInstall_BareNameResolvesInCwd(t *testing.T) {
	home := t.TempDir()
	t.Setenv("KODAE_HOME", home)
	dir := t.TempDir()
	old, _ := os.Getwd()
	os.Chdir(dir)
	t.Cleanup(func() { os.Chdir(old) })
	if err := os.WriteFile(filepath.Join(dir, "k.kodae"), []byte("pub fn f() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := runInstall("k"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(home, "libs", "k.kodae")); err != nil {
		t.Fatal(err)
	}
}
