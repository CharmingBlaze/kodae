// Package ccdriver picks a C driver (clang/llvm, gcc, cc, sidecar TCC, or zig) for linking generated C.
package ccdriver

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"kodae/internal/tccbundle"
)

// CCmd is a concrete compiler invocation: Prog and optional args before our flags.
// Example: (clang, nil), (zig, ["cc"]), (path/to/tcc.exe, nil) with TCC true.
type CCmd struct {
	Prog   string
	Prefix []string
	// TCC is true when Prog is TinyCC (different CLI from GNU clang/gcc).
	TCC bool
}

// FindConfig configures Find.
type FindConfig struct {
	Override string
	// Release skips the sidecar TCC and prefers a full optimizing toolchain on PATH
	// (clang, gcc, …) for `kodae build --release`.
	Release bool
}

// Find resolves a C compiler, with optional user override (flag) and $KODAE_CC.
// override and env take precedence: override > $KODAE_CC > (optional sidecar TCC) > PATH search.
func Find(cfg FindConfig) (CCmd, error) {
	if s := strings.TrimSpace(cfg.Override); s != "" {
		return parseOne(s)
	}
	if s := strings.TrimSpace(os.Getenv("KODAE_CC")); s != "" {
		return parseOne(s)
	}
	if os.Getenv("KODAE_NO_SIDECAR_TCC") == "" && !cfg.Release {
		if p, ok := tccbundle.SidecarPath(); ok {
			return CCmd{Prog: p, TCC: true}, nil
		}
	}
	// Order: prefer LLVM/Clang, then common GCC, then "cc", then PATH zig (no bundled toolchain).
	chain := []struct{ name string; prefix []string }{
		{"clang", nil},
		{"gcc", nil},
		{"cc", nil},
	}
	for _, c := range chain {
		if p, err := exec.LookPath(c.name); err == nil {
			return CCmd{Prog: p, Prefix: c.prefix}, nil
		}
	}
	if p, err := exec.LookPath("zig"); err == nil {
		return CCmd{Prog: p, Prefix: []string{"cc"}}, nil
	}
	return CCmd{}, fmt.Errorf("%s", hintText())
}

func parseOne(s string) (CCmd, error) {
	s = strings.TrimSpace(s)
	if s == "zig" || s == "zig cc" {
		return zigCC()
	}
	if p := strings.Fields(s); len(p) >= 2 && p[0] == "zig" && p[1] == "cc" {
		return zigCC()
	}
	if strings.ContainsRune(s, ' ') {
		return CCmd{}, fmt.Errorf("KODAE_CC: use a path without spaces, a PATH name (clang, zig), or a symlink")
	}
	if st, e := os.Stat(s); e == nil && !st.IsDir() {
		cc := CCmd{Prog: s, Prefix: nil}
		if isTCCPath(s) {
			cc.TCC = true
		}
		return cc, nil
	}
	p, err := exec.LookPath(s)
	if err != nil {
		return CCmd{}, fmt.Errorf("KODAE_CC %q: not a file and not on PATH", s)
	}
	if strings.EqualFold(filepath.Base(p), "zig") {
		return CCmd{Prog: p, Prefix: []string{"cc"}}, nil
	}
	cc := CCmd{Prog: p, Prefix: nil}
	if isTCCPath(p) {
		cc.TCC = true
	}
	return cc, nil
}

func isTCCPath(p string) bool {
	b := strings.ToLower(filepath.Base(p))
	return b == "tcc" || b == "tcc.exe"
}

func zigCC() (CCmd, error) {
	p, err := exec.LookPath("zig")
	if err != nil {
		return CCmd{}, fmt.Errorf("KODAE_CC=zig: zig not found on PATH: %v", err)
	}
	return CCmd{Prog: p, Prefix: []string{"cc"}}, nil
}

func hintText() string {
	if runtime.GOOS == "windows" {
		return "no C compiler (clang, gcc, or cc) on PATH and no sidecar TCC.\n" +
			"  Portable zip: place tcc.exe under toolchain/ next to bin/kodae.exe (see docs/DISTRIBUTION.md).\n" +
			"  LLVM/Clang: https://github.com/llvm/llvm-project/releases or winget install LLVM.LLVM\n" +
			"  Or set KODAE_CC to the full path to clang.exe.\n" +
			"  Experimental: `kodae build --backend=llvm` uses clang on LLVM IR and does not need a C compiler for that path."
	}
	return "no C compiler (clang, gcc, or cc) on PATH and no sidecar TCC.\n" +
		"  Portable tarball: ship tcc under toolchain/ next to bin/kodae (see docs/DISTRIBUTION.md).\n" +
		"  macOS: xcode-select --install   (Apple clang)\n" +
		"  Linux: sudo apt install clang  (or gcc)\n" +
		"  Or set KODAE_CC. To ignore a broken sidecar TCC: KODAE_NO_SIDECAR_TCC=1\n" +
		"  Experimental: `kodae build --backend=llvm` compiles LLVM IR with clang."
}
