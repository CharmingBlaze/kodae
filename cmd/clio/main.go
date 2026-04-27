package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"clio/internal/check"
	"clio/internal/codegen"
)

const version = "0.1.0"

func main() {
	flag.Usage = func() {
		name := "clio"
		if len(os.Args) > 0 {
			name = os.Args[0]
		}
		fmt.Fprintf(os.Stderr, "clio — Clio to C (C99) compiler\n\n")
		fmt.Fprintf(os.Stderr, "usage: %s <command> [args]\n\n", name)
		fmt.Fprintf(os.Stderr, "commands:\n")
		fmt.Fprintf(os.Stderr, "  lex <file>         tokenize a .clio file\n")
		fmt.Fprintf(os.Stderr, "  parse|ast <a.clio> [b.clio]  parse and print AST (merged like build; optional -o/--cc/--ldflags ignored)\n")
		fmt.Fprintf(os.Stderr, "  check <a.clio> [b.clio]     type-check (merges; optional -o/--cc/--ldflags ignored)\n")
		fmt.Fprintf(os.Stderr, "  cgen|emit <a.clio> [b.clio]  print generated C to stdout (merged program)\n")
		fmt.Fprintf(os.Stderr, "  build [--lib] [--static] [--shared] <a.clio> [b.clio] [-o out] [--cc c] [--ldflags \"-lfoo\"]\n")
		fmt.Fprintf(os.Stderr, "  buildc <file.clio> [-o f.c]  write C only\n")
		fmt.Fprintf(os.Stderr, "  run <file.clio> [--cc c]  build and run the same binary (set CLIO_CC to pick the C compiler)\n")
		fmt.Fprintf(os.Stderr, "  install <name.clio|name>  copy a .clio into the user lib dir (see $CLIO_HOME or ~/.clio/libs) for #include\n")
		fmt.Fprintf(os.Stderr, "  bind raylib <raylib.h> [-o include/raylib/raylib.clio]  generate Clio externs from raylib.h\n")
		fmt.Fprintf(os.Stderr, "  version\n")
		os.Exit(1)
	}
	if len(os.Args) < 2 {
		flag.Usage()
	}
	switch os.Args[1] {
	case "version", "-v", "-version":
		fmt.Println("clio", version)
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
			fatal(fmt.Errorf("parse: need at least one .clio file (same flags as build)"))
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
			fatal(fmt.Errorf("check: need at least one .clio file (same flags as build)"))
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
			fatal(fmt.Errorf("cgen: need at least one .clio file (same flags as build)"))
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
			fatal(fmt.Errorf("build: clio build [--lib] [--static] [--shared] <file>  or  clio build -o <out> [--cc clang] [--ldflags \"-lfoo\"] <a.clio> [b.clio]"))
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
			fatal(fmt.Errorf("buildc: need a .clio file (last arg)"))
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
		in, _, cc, ld, err := parseBuildFlags(a)
		if err != nil {
			fatal(err)
		}
		if len(in) == 0 {
			fatal(fmt.Errorf("run: clio run <file>  or  clio run [--cc zig] <file>"))
		}
		ldx := []string{}
		if ld != "" {
			ldx = strings.Fields(ld)
		}
		if err := runBuildAndRun(in, cc, ldx); err != nil {
			fatal(err)
		}
	case "install":
		rest := argsAfterCmd()
		if len(rest) != 1 {
			fatal(fmt.Errorf("install: clio install <name.clio>  or  clio install <name> (looks for name.clio in the current directory)"))
		}
		if err := runInstall(rest[0]); err != nil {
			fatal(err)
		}
	case "bind":
		if err := runBind(argsAfterCmd()); err != nil {
			fatal(err)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		flag.Usage()
	}
}

func need(n int) {
	if len(os.Args) < n {
		fmt.Fprintln(os.Stderr, "clio: missing argument")
		os.Exit(1)
	}
}
func fatal(e error) {
	fmt.Fprintln(os.Stderr, "clio:", e)
	os.Exit(1)
}

// argsAfterCmd returns subcommand arguments (everything after "clio <command>").
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
