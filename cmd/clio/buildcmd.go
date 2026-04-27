package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"clio/internal/ast"
	"clio/internal/ccdriver"
	"clio/internal/check"
	"clio/internal/codegen"
	lexapi "clio/internal/lexer"
	"clio/internal/parser"
)

func runCheck(paths ...string) error {
	p, err := loadProgram(paths)
	if err != nil {
		return err
	}
	_, err = check.Check(p)
	return err
}

func runBuild(paths []string, out string, cOnly bool, cc string, ldExtra []string) error {
	p, err := loadProgram(paths)
	if err != nil {
		return err
	}
	inf, err := check.Check(p)
	if err != nil {
		return err
	}
	csrc, err := codegen.EmitC(p, inf)
	if err != nil {
		return err
	}
	if cOnly {
		if out == "" {
			out = strings.TrimSuffix(paths[0], filepath.Ext(paths[0])) + ".c"
		}
		return os.WriteFile(out, []byte(csrc), 0644)
	}
	if out == "" {
		out = defaultOutBin(paths[0])
	}
	d, err := os.MkdirTemp("", "clio-build-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(d)
	cf := filepath.Join(d, "out.c")
	if err := os.WriteFile(cf, []byte(csrc), 0644); err != nil {
		return err
	}
	ccc, err := ccdriver.Find(cc)
	if err != nil {
		return err
	}
	link := append(append([]string{}, inf.LinkFlags...), ldExtra...)
	if err := ccdriver.Compile(ccc, cf, out, link); err != nil {
		return err
	}
	return nil
}

// defaultOutBin is the executable name for `clio build` / `clio run` when -o is omitted: the
// .clio file’s basename with .exe (Windows) or the bare stem (Unix). This avoids clobbering
// a single fixed a.out/a.exe so a runaway process or debugger lock on that file does not
// make every rebuild fail to link.
func defaultOutBin(clioPath string) string {
	base := strings.TrimSuffix(filepath.Base(clioPath), filepath.Ext(clioPath))
	if base == "" {
		base = "a"
	}
	if runtime.GOOS == "windows" {
		return base + ".exe"
	}
	return base
}

func runBuildAndRun(paths []string, cc string, ldExtra []string) error {
	if err := runBuild(paths, "", false, cc, ldExtra); err != nil {
		return err
	}
	abs, _ := filepath.Abs(defaultOutBin(paths[0]))
	c := exec.Command(abs)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// parseBuildFlags accepts optional -o, --cc, --ldflags in any order. Non-flag args are
// one or more .clio files (concatenated in order into one program).
func parseBuildFlags(in []string) (files []string, out, cc, ld string, err error) {
	if len(in) < 1 {
		return nil, "", "", "", fmt.Errorf("need a .clio file")
	}
	var pos []string
	for i := 0; i < len(in); i++ {
		a := in[i]
		switch {
		case a == "-o":
			if i+1 >= len(in) {
				return nil, "", "", "", fmt.Errorf("-o needs an output path")
			}
			out = in[i+1]
			i++
		case a == "--cc":
			if i+1 >= len(in) {
				return nil, "", "", "", fmt.Errorf("--cc needs a command")
			}
			cc = in[i+1]
			i++
		case a == "--ldflags":
			if i+1 >= len(in) {
				return nil, "", "", "", fmt.Errorf("--ldflags needs a string")
			}
			ld = in[i+1]
			i++
		case strings.HasPrefix(a, "--cc="):
			cc = strings.TrimPrefix(a, "--cc=")
		case strings.HasPrefix(a, "--ldflags="):
			ld = strings.TrimPrefix(a, "--ldflags=")
		default:
			if strings.HasPrefix(a, "-") {
				return nil, "", "", "", fmt.Errorf("unknown flag %q (use -o, --cc, --ldflags, or .clio paths)", a)
			}
			if a != "" {
				pos = append(pos, a)
			}
		}
	}
	if len(pos) < 1 {
		return nil, "", "", "", fmt.Errorf("need a .clio file")
	}
	var withClio []string
	for _, p := range pos {
		if strings.EqualFold(filepath.Ext(p), ".clio") {
			withClio = append(withClio, p)
		}
	}
	switch {
	case len(withClio) >= 1:
		// All positional args with .clio; reject mixed non-clio
		if len(withClio) != len(pos) {
			return nil, "", "", "", fmt.Errorf("mix of .clio and other paths: %q", pos)
		}
		files = withClio
	case len(pos) == 1:
		files = pos
	default:
		return nil, "", "", "", fmt.Errorf("no .clio in arguments: got %q", pos)
	}
	return files, out, strings.TrimSpace(cc), strings.TrimSpace(ld), nil
}

// programLoader follows `use name` to same-dir `name.clio` (depth-first, dedup, cycle check).
// UseDecl nodes are resolved and omitted from the merged program.
type programLoader struct {
	loaded  map[string]struct{}
	loading map[string]struct{}
	out     *ast.Program
}

// loadProgram parses one or more .clio file paths in order. Each file’s `use` names load
// sibling `name.clio` before the rest of that file’s decls. Absolute paths are deduplicated.
func loadProgram(paths []string) (*ast.Program, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("no source files")
	}
	l := &programLoader{
		loaded:  make(map[string]struct{}),
		loading: make(map[string]struct{}),
		out:     &ast.Program{},
	}
	for _, path := range paths {
		if !strings.EqualFold(filepath.Ext(path), ".clio") {
			return nil, fmt.Errorf("source must be a .clio file: %q", path)
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		if err := l.loadFile(abs); err != nil {
			return nil, err
		}
	}
	return l.out, nil
}

// loadFile reads and merges one translation unit, recursively loading `use` targets first.
func (l *programLoader) loadFile(path string) error {
	abs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return err
	}
	if _, ok := l.loaded[abs]; ok {
		return nil
	}
	if _, ok := l.loading[abs]; ok {
		return fmt.Errorf("use cycle: re-entering %q", abs)
	}
	l.loading[abs] = struct{}{}
	defer delete(l.loading, abs)

	b, err := os.ReadFile(abs)
	if err != nil {
		return err
	}
	p, err := parser.Parse(lexapi.New(string(b)))
	if err != nil {
		return err
	}

	// `use` first in source order (so deps load before this file’s non-use decls).
	for _, d := range p.Decls {
		u, ok := d.(*ast.UseDecl)
		if !ok {
			continue
		}
		dep := filepath.Join(filepath.Dir(abs), u.Name+".clio")
		if err := l.loadFile(dep); err != nil {
			return fmt.Errorf("%q: use %q: %w", abs, u.Name, err)
		}
	}
	for _, d := range p.Decls {
		if _, ok := d.(*ast.UseDecl); ok {
			continue
		}
		l.out.Decls = append(l.out.Decls, d)
	}
	l.out.Statements = append(l.out.Statements, p.Statements...)
	l.loaded[abs] = struct{}{}
	return nil
}
