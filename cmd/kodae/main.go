package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"kodae/internal/check"
	"kodae/internal/codegen"
)

const version = "0.1.0"

func main() {
	flag.Usage = func() {
		name := "kodae"
		if len(os.Args) > 0 {
			name = os.Args[0]
		}
		fmt.Fprintf(os.Stderr, "kodae — Kodae to C (C99) compiler\n\n")
		fmt.Fprintf(os.Stderr, "usage: %s <command> [args]\n\n", name)
		fmt.Fprintf(os.Stderr, "commands:\n")
		fmt.Fprintf(os.Stderr, "  lex <file>         tokenize a .kodae file\n")
		fmt.Fprintf(os.Stderr, "  parse|ast <a.kodae> [b.kodae]  parse and print AST (merged like build; optional -o/--cc/--ldflags ignored)\n")
		fmt.Fprintf(os.Stderr, "  check <a.kodae> [b.kodae]     type-check (merges; optional -o/--cc/--ldflags ignored)\n")
		fmt.Fprintf(os.Stderr, "  cgen|emit <a.kodae> [b.kodae]  print generated C to stdout (merged program)\n")
		fmt.Fprintf(os.Stderr, "  build [--lib] [--static] [--shared] [--release] [--backend=llvm|c] <a.kodae> [b.kodae] [-o out] [--cc c] [--ldflags \"-lfoo\"]\n")
		fmt.Fprintf(os.Stderr, "  buildc <file.kodae> [-o f.c]  write C only\n")
		fmt.Fprintf(os.Stderr, "  run <file.kodae> [--release] [--backend=llvm|c] [--cc c]  build and run (default backend: llvm; --release skips sidecar TCC)\n")
		fmt.Fprintf(os.Stderr, "  install <name.kodae|name>  copy a .kodae into the user lib dir (see $KODAE_HOME or ~/.kodae/libs) for #include\n")
		fmt.Fprintf(os.Stderr, "  bind <name> <header.h> [-o out.kodae]  generate Kodae bindings from a C header (needs Clang; see docs/BINDGEN.md)\n")
		fmt.Fprintf(os.Stderr, "  bundle [os] [arch]  create dist/ bundle (bin + include + examples; copies toolchain/ if present)\n")
		fmt.Fprintf(os.Stderr, "  version\n")
		os.Exit(1)
	}
	if len(os.Args) < 2 {
		flag.Usage()
	}
	switch os.Args[1] {
	case "version", "-v", "-version":
		fmt.Println("kodae", version)
		return
	case "lex", "tokenize", "lexdump":
		need(3)
		if err := runLexFile(os.Args[2]); err != nil {
			fatal(err)
		}
	case "parse", "ast":
		files, _, _, _, err := parseBuildFlags(argsAfterCmd())
		if err != nil {
			fatal(err)
		}
		if len(files) < 1 {
			fatal(fmt.Errorf("parse: need at least one .kodae file (same flags as build)"))
		}
		if err := runParseFiles(files); err != nil {
			fatal(err)
		}
	case "check", "typecheck":
		files, _, _, _, err := parseBuildFlags(argsAfterCmd())
		if err != nil {
			fatal(err)
		}
		if len(files) < 1 {
			fatal(fmt.Errorf("check: need at least one .kodae file (same flags as build)"))
		}
		if err := runCheck(files...); err != nil {
			fatal(err)
		}
		fmt.Println("ok")
	case "cgen", "emit", "c":
		files, _, _, _, err := parseBuildFlags(argsAfterCmd())
		if err != nil {
			fatal(err)
		}
		if len(files) < 1 {
			fatal(fmt.Errorf("cgen: need at least one .kodae file (same flags as build)"))
		}
		if err := runCgen(files...); err != nil {
			fatal(err)
		}
	case "build":
		a := os.Args[2:]
		in, out, cc, ld, bopt, err := parseBuildFlagsExt(a)
		if err != nil {
			fatal(err)
		}
		if len(in) == 0 {
			fatal(fmt.Errorf("build: kodae build [--lib] [--static] [--shared] [--release] [--backend=llvm|c] <file>  or  kodae build -o <out> [--cc clang] [--ldflags \"-lfoo\"] <a.kodae> [b.kodae]"))
		}
		ldx := []string{}
		if ld != "" {
			ldx = strings.Fields(ld)
		}
		if err := runBuild(in, out, false, cc, ldx, bopt); err != nil {
			fatal(err)
		}
	case "buildc":
		a := os.Args[2:]
		in, out, cc, ld, err := parseBuildFlags(a)
		if err != nil {
			fatal(err)
		}
		if len(in) == 0 {
			fatal(fmt.Errorf("buildc: need a .kodae file (last arg)"))
		}
		ldx := []string{}
		if ld != "" {
			ldx = strings.Fields(ld)
		}
		if err := runBuild(in, out, true, cc, ldx, buildOptions{}); err != nil {
			fatal(err)
		}
	case "run":
		a := os.Args[2:]
		in, _, cc, ld, bopt, err := parseBuildFlagsExt(a)
		if err != nil {
			fatal(err)
		}
		if len(in) == 0 {
			fatal(fmt.Errorf("run: kodae run <file>  or  kodae run [--backend=llvm|c] [--cc clang] <file>"))
		}
		ldx := []string{}
		if ld != "" {
			ldx = strings.Fields(ld)
		}
		if err := runBuildAndRun(in, cc, ldx, bopt); err != nil {
			fatal(err)
		}
	case "install":
		rest := argsAfterCmd()
		if len(rest) != 1 {
			fatal(fmt.Errorf("install: kodae install <name.kodae>  or  kodae install <name> (looks for name.kodae in the current directory)"))
		}
		if err := runInstall(rest[0]); err != nil {
			fatal(err)
		}
	case "bind":
		if err := runBind(argsAfterCmd()); err != nil {
			fatal(err)
		}
	case "bundle":
		if err := runBundle(argsAfterCmd()); err != nil {
			fatal(err)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		flag.Usage()
	}
}

func need(n int) {
	if len(os.Args) < n {
		fmt.Fprintln(os.Stderr, "kodae: missing argument")
		os.Exit(1)
	}
}
func fatal(e error) {
	fmt.Fprintln(os.Stderr, "kodae:", e)
	os.Exit(1)
}

// argsAfterCmd returns subcommand arguments (everything after "kodae <command>").
func argsAfterCmd() []string {
	if len(os.Args) < 2 {
		return nil
	}
	return os.Args[2:]
}

func runCgen(paths ...string) error {
	p, err := loadProgram(paths)
	if err != nil {
		return err
	}
	inf, err := check.Check(p)
	if err != nil {
		return err
	}
	s, err := codegen.EmitC(p, inf)
	if err != nil {
		return err
	}
	_, err = os.Stdout.WriteString(s)
	return err
}
