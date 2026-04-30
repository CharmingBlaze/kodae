package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"kodae/internal/ccdriver"
)

func TestBuildLib_GeneratesArtifactsAndCConsumerBuilds(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "mymath.kodae")
	code := `#mode "library"
#library "mymath"
fn add(a: int, b: int) -> int { return a + b }`
	if err := os.WriteFile(src, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(old) }()
	if err := runBuild([]string{src}, "", false, "", nil, buildOptions{LibMode: true, Static: true, Backend: "c"}); err != nil {
		t.Fatalf("runBuild --lib: %v", err)
	}
	for _, f := range []string{"mymath.c", "mymath.h", "mymath.a"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Fatalf("missing %s: %v", f, err)
		}
	}
	csrc := filepath.Join(dir, "consumer.c")
	cprog := `#include "mymath.h"
#include <stdio.h>
int main(void){ printf("%lld\n", (long long)add(2,3)); return 0; }`
	if err := os.WriteFile(csrc, []byte(cprog), 0644); err != nil {
		t.Fatal(err)
	}
	cc, err := ccdriver.Find(ccdriver.FindConfig{})
	if err != nil {
		t.Skipf("no C compiler found for smoke test: %v", err)
	}
	out := filepath.Join(dir, "consumer")
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	args := append([]string{}, cc.Prefix...)
	args = append(args, "-std=gnu99", "-O2", "-o", out, csrc, filepath.Join(dir, "mymath.a"))
	if runtime.GOOS == "windows" {
		args = append(args, "-lws2_32")
	} else {
		args = append(args, "-lm")
	}
	cmd := exec.Command(cc.Prog, args...)
	cmd.Dir = dir
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("c consumer compile failed: %v\n%s", err, string(b))
	}
}
