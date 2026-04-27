package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"clio/internal/loader"
)

// runInstall copies a .clio file into the user library directory so #include "name" can find it
// (same path as internal/loader.ResolveClioInclude step 3). The argument is either a path to
// a .clio file or a bare name; the latter is resolved as name.clio in the current directory.
func runInstall(arg string) error {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return fmt.Errorf("install: need a .clio file or library name (e.g. clio install mathlib.clio)")
	}
	var src string
	if strings.EqualFold(filepath.Ext(arg), ".clio") {
		var err error
		src, err = filepath.Abs(arg)
		if err != nil {
			return err
		}
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		src = filepath.Join(cwd, arg+".clio")
	}
	st, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("install: %w", err)
	}
	if st.IsDir() {
		return fmt.Errorf("install: %q is a directory, expected a .clio file", src)
	}
	libDir, err := loader.UserLibDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(libDir, 0755); err != nil {
		return err
	}
	dest := filepath.Join(libDir, filepath.Base(src))
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	ce := out.Close()
	if err != nil {
		return err
	}
	if ce != nil {
		return ce
	}
	fmt.Println(dest)
	return nil
}
