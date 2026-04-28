package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"kodae/internal/bindgen"
)

func main() {
	fs := flag.NewFlagSet("kodae-bind", flag.ContinueOnError)
	out := fs.String("o", "", "output .kodae path (default: include/<name>/<name>.kodae)")
	if err := fs.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}
	rest := fs.Args()
	if len(rest) != 2 {
		fmt.Fprintln(os.Stderr, "usage: kodae-bind [-o out.kodae] <name> <path/to/header.h>")
		os.Exit(2)
	}
	libName := rest[0]
	headerPath := rest[1]

	outputPath := *out
	if outputPath == "" {
		outputPath = filepath.Join("include", libName, libName+".kodae")
	}

	res, err := bindgen.GenerateBindings(headerPath, libName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "kodae-bind:", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "kodae-bind:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(outputPath, []byte(res.Content), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "kodae-bind:", err)
		os.Exit(1)
	}

	// Also update examples copy if it exists
	exampleCopy := filepath.Join("examples", "libs", libName, libName+".kodae")
	_ = os.MkdirAll(filepath.Dir(exampleCopy), 0o755)
	_ = os.WriteFile(exampleCopy, []byte(res.Content), 0o644)

	fmt.Fprintf(os.Stderr, "wrote %s: structs=%d externs=%d skipped=%d\n", outputPath, res.Structs, res.Externs, res.Skipped)
}
