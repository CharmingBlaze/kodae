package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestParseBuildFlagsExt_Backend(t *testing.T) {
	t.Parallel()
	files, _, _, _, opt, err := parseBuildFlagsExt([]string{"--backend=llvm", "a.kodae"})
	if err != nil {
		t.Fatal(err)
	}
	if opt.Backend != "llvm" {
		t.Fatalf("backend: got %q", opt.Backend)
	}
	if len(files) != 1 || files[0] != "a.kodae" {
		t.Fatalf("files: %#v", files)
	}
}

func TestParseBuildFlagsExt_BackendDefaultEmpty(t *testing.T) {
	t.Parallel()
	_, _, _, _, opt, err := parseBuildFlagsExt([]string{"a.kodae"})
	if err != nil {
		t.Fatal(err)
	}
	if opt.Backend != "" {
		t.Fatalf("backend default should be empty before runBuild normalization, got %q", opt.Backend)
	}
}

func TestRunBuildLLVMSmoke(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang not on PATH")
	}
	src := filepath.Join("..", "..", "examples", "hello.kodae") // cwd is cmd/kodae when tests run
	if _, err := os.Stat(src); err != nil {
		t.Skip("examples/hello.kodae not found")
	}
	dir := t.TempDir()
	out := filepath.Join(dir, "hi")
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	if err := runBuild([]string{src}, out, false, "", nil, buildOptions{Backend: "llvm"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("missing output: %v", err)
	}
}

func TestRunBuildDefaultBackendUsesLLVM(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang not on PATH")
	}
	src := filepath.Join("..", "..", "examples", "hello.kodae")
	if _, err := os.Stat(src); err != nil {
		t.Skip("examples/hello.kodae not found")
	}
	dir := t.TempDir()
	out := filepath.Join(dir, "hi-default")
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	// Empty buildOptions exercises default backend normalization in runBuild.
	if err := runBuild([]string{src}, out, false, "", nil, buildOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("missing output: %v", err)
	}
}
