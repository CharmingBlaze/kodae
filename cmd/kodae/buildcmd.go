package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"kodae/internal/ast"
	"kodae/internal/ccdriver"
	"kodae/internal/check"
	"kodae/internal/codegen"
	"kodae/internal/headergen"
	lexapi "kodae/internal/lexer"
	"kodae/internal/loader"
	"kodae/internal/parser"
)

type buildOptions struct {
	LibMode bool
	Static  bool
	Shared  bool
}

func runCheck(paths ...string) error {
	p, err := loadProgram(paths)
	if err != nil {
		return err
	}
	_, err = check.Check(p)
	return err
}

func runBuild(paths []string, out string, cOnly bool, cc string, ldExtra []string, opt buildOptions) error {
	p, err := loadProgram(paths)
	if err != nil {
		return err
	}
	inf, err := check.Check(p)
	if err != nil {
		return err
	}
	libMode := opt.LibMode || strings.EqualFold(strings.TrimSpace(inf.Meta["mode"]), "library")
	csrc, err := codegen.EmitCWithOptions(p, inf, codegen.EmitOptions{LibraryMode: libMode})
	if err != nil {
		return err
	}
	libName := strings.TrimSpace(inf.Meta["library"])
	if libName == "" {
		libName = strings.TrimSuffix(filepath.Base(paths[0]), filepath.Ext(paths[0]))
	}
	if libName == "" {
		libName = "kodae_lib"
	}
	if libMode {
		cOut := libName + ".c"
		hOut := libName + ".h"
		if out != "" {
			cOut = out
		}
		if err := os.WriteFile(cOut, []byte(csrc), 0644); err != nil {
			return err
		}
		hdr, err := headergen.Generate(p, inf, headergen.Options{LibraryName: libName})
		if err != nil {
			return err
		}
		if err := os.WriteFile(hOut, []byte(hdr), 0644); err != nil {
			return err
		}
		ccc, err := ccdriver.Find(cc)
		if err != nil {
			return err
		}
		link := append(append([]string{}, inf.LinkFlags...), ldExtra...)
		obj := libName + ".o"
		if err := ccdriver.CompileObject(ccc, cOut, obj, link); err != nil {
			return err
		}
		if opt.Static || (!opt.Static && !opt.Shared) {
			if err := ccdriver.ArchiveStatic(obj, libName+".a"); err != nil {
				return err
			}
		}
		if opt.Shared || (!opt.Static && !opt.Shared) {
			sharedOut := libName + ".so"
			switch runtime.GOOS {
			case "windows":
				sharedOut = libName + ".dll"
			case "darwin":
				sharedOut = libName + ".dylib"
			}
			if err := ccdriver.LinkShared(ccc, cOut, sharedOut, link, !inf.UsesConsole); err != nil {
				return err
			}
		}
		return nil
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
	d, err := os.MkdirTemp("", "kodae-build-*")
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
	if err := ccdriver.Compile(ccc, cf, out, link, !inf.UsesConsole); err != nil {
		return err
	}
	return nil
}

// defaultOutBin is the executable name for `kodae build` / `kodae run` when -o is omitted: the
// .kodae file’s basename with .exe (Windows) or the bare stem (Unix). This avoids clobbering
// a single fixed a.out/a.exe so a runaway process or debugger lock on that file does not
// make every rebuild fail to link.
func defaultOutBin(kodaePath string) string {
	base := strings.TrimSuffix(filepath.Base(kodaePath), filepath.Ext(kodaePath))
	if base == "" {
		base = "a"
	}
	if runtime.GOOS == "windows" {
		return base + ".exe"
	}
	return base
}

func runBuildAndRun(paths []string, cc string, ldExtra []string) error {
	if err := runBuild(paths, "", false, cc, ldExtra, buildOptions{}); err != nil {
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
// one or more .kodae files (concatenated in order into one program).
func parseBuildFlags(in []string) (files []string, out, cc, ld string, err error) {
	if len(in) < 1 {
		return nil, "", "", "", fmt.Errorf("need a .kodae file")
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
				return nil, "", "", "", fmt.Errorf("unknown flag %q (use -o, --cc, --ldflags, or .kodae paths)", a)
			}
			if a != "" {
				pos = append(pos, a)
			}
		}
	}
	if len(pos) < 1 {
		return nil, "", "", "", fmt.Errorf("need a .kodae file")
	}
	var withKodae []string
	for _, p := range pos {
		if strings.EqualFold(filepath.Ext(p), ".kodae") {
			withKodae = append(withKodae, p)
		}
	}
	switch {
	case len(withKodae) >= 1:
		// All positional args with .kodae; reject mixed non-kodae
		if len(withKodae) != len(pos) {
			return nil, "", "", "", fmt.Errorf("mix of .kodae and other paths: %q", pos)
		}
		files = withKodae
	case len(pos) == 1:
		files = pos
	default:
		return nil, "", "", "", fmt.Errorf("no .kodae in arguments: got %q", pos)
	}
	return files, out, strings.TrimSpace(cc), strings.TrimSpace(ld), nil
}

func parseBuildFlagsExt(in []string) (files []string, out, cc, ld string, opt buildOptions, err error) {
	raw := make([]string, 0, len(in))
	for _, a := range in {
		switch a {
		case "--lib":
			opt.LibMode = true
		case "--static":
			opt.Static = true
		case "--shared":
			opt.Shared = true
		default:
			raw = append(raw, a)
		}
	}
	files, out, cc, ld, err = parseBuildFlags(raw)
	return
}

// programLoader follows `# include "p"` and `use name` (same-dir + libs + ~/.kodae/libs), depth-first, dedup, cycle check.
// Include and Use are resolved and omitted from the merged program.
type programLoader struct {
	loaded  map[string]struct{}
	loading map[string]struct{}
	out     *ast.Program
}

func setDeclFile(d ast.Decl, abs string) {
	switch t := d.(type) {
	case *ast.FnDecl:
		t.File = abs
	case *ast.StructDecl:
		t.File = abs
	case *ast.EnumDecl:
		t.File = abs
	case *ast.ExternDecl:
		t.File = abs
	case *ast.LetDecl:
		t.File = abs
	}
}

// loadProgram parses one or more .kodae file paths in order. Each file’s #include and `use` load
// dependencies first. Absolute paths are deduplicated.
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
		if !strings.EqualFold(filepath.Ext(path), ".kodae") {
			return nil, fmt.Errorf("source must be a .kodae file: %q", path)
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

// loadFile reads and merges one translation unit, loading #include and use targets first.
func (l *programLoader) loadFile(path string) error {
	abs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return err
	}
	if _, ok := l.loaded[abs]; ok {
		return nil
	}
	if _, ok := l.loading[abs]; ok {
		return fmt.Errorf("include/use cycle: re-entering %q", abs)
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
	for _, d := range p.Decls {
		switch t := d.(type) {
		case *ast.IncludeDecl:
			inc, e := loader.ResolveKodaeInclude(filepath.Dir(abs), t.Path)
			if e != nil {
				return e
			}
			if e2 := l.loadFile(inc); e2 != nil {
				return fmt.Errorf("%q: # include %q: %w", abs, t.Path, e2)
			}
		case *ast.UseDecl:
			dep, e := loader.ResolveKodaeInclude(filepath.Dir(abs), t.Name)
			if e != nil {
				return fmt.Errorf("%q: use %q: %w", abs, t.Name, e)
			}
			if e2 := l.loadFile(dep); e2 != nil {
				return fmt.Errorf("%q: use %q: %w", abs, t.Name, e2)
			}
		}
	}
	for _, d := range p.Decls {
		if _, ok := d.(*ast.IncludeDecl); ok {
			continue
		}
		if _, ok := d.(*ast.UseDecl); ok {
			continue
		}
		setDeclFile(d, abs)
		l.out.Decls = append(l.out.Decls, d)
	}
	l.out.Statements = append(l.out.Statements, p.Statements...)
	l.loaded[abs] = struct{}{}
	return nil
}
