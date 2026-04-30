package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestParseBuildFlagsExt_BackendCExplicit(t *testing.T) {
	t.Parallel()
	files, _, _, _, opt, err := parseBuildFlagsExt([]string{"--backend=c", "a.kodae"})
	if err != nil {
		t.Fatal(err)
	}
	if opt.Backend != "c" {
		t.Fatalf("backend: got %q", opt.Backend)
	}
	if len(files) != 1 || files[0] != "a.kodae" {
		t.Fatalf("files: %#v", files)
	}
}

func TestRunBuild_UnknownBackendError(t *testing.T) {
	t.Parallel()
	src := filepath.Join("..", "..", "examples", "hello.kodae")
	err := runBuild([]string{src}, "", false, "", nil, buildOptions{Backend: "nope"})
	if err == nil {
		t.Fatal("expected unknown backend error")
	}
	if !strings.Contains(err.Error(), "unknown --backend") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunBuild_LLVMRejectsLibMode(t *testing.T) {
	t.Parallel()
	src := filepath.Join("..", "..", "examples", "hello.kodae")
	err := runBuild([]string{src}, "", false, "", nil, buildOptions{Backend: "llvm", LibMode: true})
	if err == nil {
		t.Fatal("expected llvm library mode rejection")
	}
	if !strings.Contains(err.Error(), "library mode not supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunBuild_LibModeDefaultsToCBackend(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "mylib.kodae")
	code := `#mode "library"
#library "mylib"
fn add(a: int, b: int) -> int { return a + b }`
	if err := os.WriteFile(src, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(old) }()

	// Empty backend should auto-select C for library mode.
	if err := runBuild([]string{src}, "", false, "", nil, buildOptions{LibMode: true, Static: true}); err != nil {
		t.Fatalf("lib mode default backend failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "mylib.a")); err != nil {
		t.Fatalf("expected static library artifact: %v", err)
	}
}

func TestRunBuild_COnlyDefaultsToCBackend(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src, err := filepath.Abs(filepath.Join("..", "..", "examples", "hello.kodae"))
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "hello.c")
	if err := runBuild([]string{src}, out, true, "", nil, buildOptions{}); err != nil {
		t.Fatalf("buildc default backend failed: %v", err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("expected generated C output: %v", err)
	}
}

func TestRunBuildAndRun_DefaultLLVMPath(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang not on PATH")
	}
	src := filepath.Join("..", "..", "examples", "hello.kodae")
	// Ensures empty backend in options still runs through default (LLVM now).
	if err := runBuildAndRun([]string{src}, "", nil, buildOptions{}); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			t.Fatalf("build+run failed: %s", string(ee.Stderr))
		}
		t.Fatalf("build+run failed: %v", err)
	}
}

func TestDefaultOutBinName(t *testing.T) {
	t.Parallel()
	got := defaultOutBin("some/path/myprog.kodae")
	if runtime.GOOS == "windows" {
		if got != "myprog.exe" {
			t.Fatalf("got %q want myprog.exe", got)
		}
		return
	}
	if got != "myprog" {
		t.Fatalf("got %q want myprog", got)
	}
}

