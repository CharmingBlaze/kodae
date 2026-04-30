// Package tccbundle locates a TinyCC (TCC) executable shipped next to kodae (portable zip layout).
// When present under ../toolchain/ relative to the kodae binary (see SidecarPath), ccdriver uses
// it before searching PATH so users need not install a C compiler.
package tccbundle

import (
	"os"
	"path/filepath"
	"runtime"
)

// tccExeName is the TCC executable filename for the current GOOS (host naming).
func tccExeName() string {
	if runtime.GOOS == "windows" {
		return "tcc.exe"
	}
	return "tcc"
}

func evalExeDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", err
	}
	return filepath.Dir(exe), nil
}

// SidecarPath returns an absolute path to a TCC binary shipped beside kodae, if it exists.
// Layout (after unzipping a release bundle):
//
//	bin/kodae[.exe]
//	toolchain/tcc[.exe]   (+ any DLLs shipped next to tcc.exe on Windows)
//
// If kodae is run from PATH without a bundle, we also try <exeDir>/toolchain/.
func SidecarPath() (abs string, ok bool) {
	dir, err := evalExeDir()
	if err != nil {
		return "", false
	}
	name := tccExeName()
	candidates := []string{
		filepath.Join(dir, "toolchain", name),
		filepath.Join(dir, "..", "toolchain", name),
	}
	for _, p := range candidates {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			if runtime.GOOS != "windows" {
				if st.Mode()&0111 == 0 {
					// Not executable — still try; user may fix chmod
				}
			}
			if ap, err := filepath.Abs(p); err == nil {
				return ap, true
			}
			return p, true
		}
	}
	return "", false
}
