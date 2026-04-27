// Package loader resolves #include "path" to an absolute .clio file path.
package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveClioInclude finds a .clio file to satisfy #include in order:
//  1) path next to the current file
//  2) <currentDir>/libs/<path>
//  3) ~/.clio/libs/<path>
func ResolveClioInclude(currentFileDir, spec string) (string, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return "", fmt.Errorf("# include: path is empty")
	}
	// Absolute .clio path
	if filepath.IsAbs(spec) {
		abs := filepath.Clean(spec)
		if st, e := os.Stat(abs); e == nil && !st.IsDir() {
			return abs, nil
		}
		return "", fmt.Errorf("# include: not found: %q", spec)
	}
	rel := spec
	if !strings.EqualFold(filepath.Ext(rel), ".clio") {
		rel = rel + ".clio"
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
	return "", fmt.Errorf("# include %q: not found (tried: %s); for a C API use # link and extern fn, or add a .clio file in the project or ~/.clio/libs",
		strings.TrimSuffix(spec, ".clio"), strings.Join(tried, ", "))
}

// UserLibDir returns the directory for installed .clio libraries: $CLIO_HOME/libs, or
// ~/.clio/libs (or the platform equivalent under the user home).
func UserLibDir() (string, error) {
	if d := strings.TrimSpace(os.Getenv("CLIO_HOME")); d != "" {
		return filepath.Clean(filepath.Join(d, "libs")), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".clio", "libs"), nil
}
