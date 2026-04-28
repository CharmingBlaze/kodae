// Package ccdriver picks a C driver (clang/llvm, gcc, or zig cc) for linking generated C.
package ccdriver

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// CCmd is a concrete compiler invocation: Prog and optional args before our flags.
// Example: (clang, nil), (zig, ["cc"]).
type CCmd struct {
	Prog   string
	Prefix []string
}

// Find resolves a C compiler, with optional user override (flag) and $KODAE_CC.
// override and env take precedence: override > $KODAE_CC > PATH search.
func Find(override string) (CCmd, error) {
	if s := strings.TrimSpace(override); s != "" {
		return parseOne(s)
	}
	if s := strings.TrimSpace(os.Getenv("KODAE_CC")); s != "" {
		return parseOne(s)
	}
	if b, ok := bundledZigCC(); ok {
		return b, nil
	}
	// Order: prefer LLVM/Clang, then common GCC, then "cc", then Zig (bundles libc + lld for easy distribution).
	//
	// Users who install the official LLVM/Clang build get `clang` on PATH — that is the usual "use LLVM" setup.
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
		return CCmd{Prog: s, Prefix: nil}, nil
	}
	p, err := exec.LookPath(s)
	if err != nil {
		return CCmd{}, fmt.Errorf("KODAE_CC %q: not a file and not on PATH", s)
	}
	if strings.EqualFold(filepath.Base(p), "zig") {
		return CCmd{Prog: p, Prefix: []string{"cc"}}, nil
	}
	return CCmd{Prog: p, Prefix: nil}, nil
}

func zigCC() (CCmd, error) {
	p, err := exec.LookPath("zig")
	if err != nil {
		return CCmd{}, fmt.Errorf("KODAE_CC=zig: zig not found on PATH: %v", err)
	}
	return CCmd{Prog: p, Prefix: []string{"cc"}}, nil
}

func bundledZigCC() (CCmd, bool) {
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	candidates := bundledZigCandidates(exe, cwd)
	for _, c := range candidates {
		if st, e := os.Stat(c); e == nil && !st.IsDir() {
			return CCmd{Prog: c, Prefix: []string{"cc"}}, true
		}
	}
	return CCmd{}, false
}

func bundledZigCandidates(exePath, cwd string) []string {
	zigName := zigExeName()
	var roots []string
	if exePath != "" {
		roots = append(roots, filepath.Dir(exePath))
	}
	if cwd != "" {
		roots = append(roots, cwd)
	}

	seenRoots := map[string]struct{}{}
	var uniqRoots []string
	for _, r := range roots {
		clean := filepath.Clean(r)
		if _, ok := seenRoots[clean]; ok {
			continue
		}
		seenRoots[clean] = struct{}{}
		uniqRoots = append(uniqRoots, clean)
	}

	seenPaths := map[string]struct{}{}
	var out []string
	for _, root := range uniqRoots {
		cur := root
		for i := 0; i < 4; i++ {
			p := filepath.Join(cur, "toolchain", "zig", zigName)
			p = filepath.Clean(p)
			if _, ok := seenPaths[p]; !ok {
				seenPaths[p] = struct{}{}
				out = append(out, p)
			}
			parent := filepath.Dir(cur)
			if parent == cur {
				break
			}
			cur = parent
		}
	}
	return out
}

func zigExeName() string {
	if runtime.GOOS == "windows" {
		return "zig.exe"
	}
	return "zig"
}

func hintText() string {
	if runtime.GOOS == "windows" {
		return "no C compiler (clang, gcc, cc, or zig) on PATH.\n" +
			"  LLVM/Clang (recommended for Windows): https://github.com/llvm/llvm-project/releases\n" +
			"  or:  winget install LLVM.LLVM\n" +
			"  Zig (portable, includes a C driver):  https://ziglang.org/download/\n" +
			"  Or set KODAE_CC to the full path to clang.exe or to \"zig\" to use `zig cc`."
	}
	return "no C compiler (clang, gcc, cc, or zig) on PATH.\n" +
		"  macOS: xcode-select --install   (gives Apple clang, LLVM-based)\n" +
		"  Linux: sudo apt install clang  (or gcc)\n" +
		"  Or install Zig: https://ziglang.org/download/\n" +
		"  Or set KODAE_CC, e.g.  export KODAE_CC=clang  or  export KODAE_CC=/opt/llvm/bin/clang"
}
