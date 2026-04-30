package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func buildInTreeCompilerForTests(t *testing.T) string {
	t.Helper()
	src := filepath.Join("..", "..", "src", "compiler", "main.kodae")
	if _, err := os.Stat(src); err != nil {
		t.Skip("in-tree compiler source not found")
	}
	out := filepath.Join(t.TempDir(), "kcomp")
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	if err := runBuild([]string{src}, out, false, "", nil, buildOptions{Backend: "c"}); err != nil {
		t.Fatalf("build in-tree compiler: %v", err)
	}
	return out
}

func TestCrossWrapper_CheckParityCurated(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang not on PATH")
	}
	kcomp := buildInTreeCompilerForTests(t)
	samples := []string{
		filepath.Join("..", "..", "examples", "hello.kodae"),
		filepath.Join("..", "..", "examples", "features.kodae"),
		filepath.Join("..", "..", "examples", "list_basic.kodae"),
		filepath.Join("..", "..", "examples", "brick_breaker.kodae"),
	}
	for _, s := range samples {
		if err := runCheck(s); err != nil {
			t.Fatalf("go wrapper check failed for %s: %v", s, err)
		}
		cmd := exec.Command(kcomp, "check", s)
		b, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("in-tree check failed for %s: %v\n%s", s, err, string(b))
		}
		if !strings.Contains(string(b), "check OK") {
			t.Fatalf("in-tree check unexpected output for %s:\n%s", s, string(b))
		}
	}
}

func TestCrossWrapper_CgenNoCriticalStubs(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang not on PATH")
	}
	kcomp := buildInTreeCompilerForTests(t)

	features := filepath.Join("..", "..", "examples", "features.kodae")
	fcmd := exec.Command(kcomp, "cgen", features)
	fout, err := fcmd.CombinedOutput()
	if err != nil {
		t.Fatalf("features cgen failed: %v\n%s", err, string(fout))
	}
	ftext := string(fout)
	if !strings.Contains(ftext, "switch ((int64_t)") {
		t.Fatalf("features cgen missing match lowering:\n%s", ftext)
	}
	if strings.Contains(ftext, "/* match */") {
		t.Fatalf("features cgen still has match stub:\n%s", ftext)
	}
	if strings.Contains(ftext, "/* defer */") {
		t.Fatalf("features cgen still has defer stub:\n%s", ftext)
	}
	if !strings.Contains(ftext, "kodae_print_str(kodae_str_from(\"features done\"))") {
		t.Fatalf("features cgen missing defer lowering payload:\n%s", ftext)
	}

	brick := filepath.Join("..", "..", "examples", "brick_breaker.kodae")
	bcmd := exec.Command(kcomp, "cgen", brick)
	bout, err := bcmd.CombinedOutput()
	if err != nil {
		t.Fatalf("brick_breaker cgen failed: %v\n%s", err, string(bout))
	}
	btext := string(bout)
	if strings.Contains(btext, "/* for-in */") {
		t.Fatalf("brick_breaker cgen still has for-in stub:\n%s", btext)
	}
}

func TestGoWrapper_BuildC_CompilesCurated(t *testing.T) {
	clang, err := exec.LookPath("clang")
	if err != nil {
		t.Skip("clang not on PATH")
	}
	samples := []string{
		filepath.Join("..", "..", "examples", "hello.kodae"),
		filepath.Join("..", "..", "examples", "features.kodae"),
		filepath.Join("..", "..", "examples", "list_basic.kodae"),
	}
	for _, s := range samples {
		base := strings.TrimSuffix(filepath.Base(s), filepath.Ext(s))
		cOut := filepath.Join(t.TempDir(), base+".c")
		if err := runBuild([]string{s}, cOut, true, "", nil, buildOptions{Backend: "c"}); err != nil {
			t.Fatalf("buildc failed for %s: %v", s, err)
		}
		obj := filepath.Join(t.TempDir(), base+".o")
		cmd := exec.Command(clang, "-std=c99", "-Wno-override-module", "-c", "-o", obj, cOut)
		b, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("clang compile failed for %s: %v\n%s", s, err, string(b))
		}
	}
}

