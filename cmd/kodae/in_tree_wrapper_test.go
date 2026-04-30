package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInTreeWrapper_CheckAndCgenContracts(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang not on PATH")
	}

	inTreeSrc := filepath.Join("..", "..", "src", "compiler", "main.kodae")
	if _, err := os.Stat(inTreeSrc); err != nil {
		t.Skip("in-tree compiler source not found")
	}

	dir := t.TempDir()
	kcomp := filepath.Join(dir, "kcomp")
	if runtime.GOOS == "windows" {
		kcomp += ".exe"
	}

	// Build in-tree compiler using explicit C backend (LLVM MVP doesn't cover full compiler source yet).
	if err := runBuild([]string{inTreeSrc}, kcomp, false, "", nil, buildOptions{Backend: "c"}); err != nil {
		t.Fatalf("build in-tree compiler: %v", err)
	}

	hello := filepath.Join("..", "..", "examples", "hello.kodae")

	// check contract
	checkCmd := exec.Command(kcomp, "check", hello)
	checkOut, err := checkCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("kcomp check failed: %v\n%s", err, string(checkOut))
	}
	if !strings.Contains(string(checkOut), "check OK") {
		t.Fatalf("unexpected check output:\n%s", string(checkOut))
	}

	// cgen contract: should emit C-like output with main.
	cgenCmd := exec.Command(kcomp, "cgen", hello)
	cgenOut, err := cgenCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("kcomp cgen failed: %v\n%s", err, string(cgenOut))
	}
	text := string(cgenOut)
	if !strings.Contains(text, "int main(") {
		t.Fatalf("cgen missing main:\n%s", text)
	}
}

