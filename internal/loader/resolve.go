// Package loader resolves #include "path" to an absolute .kodae file path.
package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveKodaeInclude finds a .kodae file to satisfy #include in order:
//  1) path next to the current file
//  2) <currentDir>/libs/<path>
//  3) ~/.kodae/libs/<path>
func ResolveKodaeInclude(currentFileDir, spec string) (string, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return "", fmt.Errorf("# include: path is empty")
	}
	// Absolute .kodae path
	if filepath.IsAbs(spec) {
		abs := filepath.Clean(spec)
		if st, e := os.Stat(abs); e == nil && !st.IsDir() {
			return abs, nil
		}
		return "", fmt.Errorf("# include: not found: %q", spec)
	}
	rel := spec
	if !strings.EqualFold(filepath.Ext(rel), ".kodae") {
		rel = rel + ".kodae"
	}

	candidates := []string{
		filepath.Clean(filepath.Join(currentFileDir, rel)),
		filepath.Clean(filepath.Join(currentFileDir, "libs", rel)),
	}
	if lib, err := UserLibDir(); err == nil {
		candidates = append(candidates, filepath.Clean(filepath.Join(lib, rel)))
	}
	var tried []string
	for _, c := range candidates {
		tried = append(tried, c)
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c, nil
		}
	}
	return "", fmt.Errorf("# include %q: not found (tried: %s); for a C API use # link and extern fn, or add a .kodae file in the project or ~/.kodae/libs",
		strings.TrimSuffix(spec, ".kodae"), strings.Join(tried, ", "))
}

// UserLibDir returns the directory for installed .kodae libraries: $KODAE_HOME/libs, or
// ~/.kodae/libs (or the platform equivalent under the user home).
func UserLibDir() (string, error) {
	if d := strings.TrimSpace(os.Getenv("KODAE_HOME")); d != "" {
		return filepath.Clean(filepath.Join(d, "libs")), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".kodae", "libs"), nil
}
