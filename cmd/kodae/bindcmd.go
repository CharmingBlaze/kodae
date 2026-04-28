package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"kodae/internal/bindgen"
)

func runBind(args []string) error {
	fs := flag.NewFlagSet("bind", flag.ContinueOnError)
	out := fs.String("o", "", "output .kodae path (default: include/<name>/<name>.kodae)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	rest := fs.Args()
	if len(rest) != 2 {
		return fmt.Errorf("usage: kodae bind [-o out.kodae] <name> <path/to/header.h>")
	}
	libName := rest[0]
	headerPath := rest[1]

	outputPath := *out
	if outputPath == "" {
		outputPath = filepath.Join("include", libName, libName+".kodae")
	}

	res, err := bindgen.GenerateBindings(headerPath, libName)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, []byte(res.Content), 0o644); err != nil {
		return err
	}

	// Also update examples copy if it exists
	exampleCopy := filepath.Join("examples", "libs", libName, libName+".kodae")
	_ = os.MkdirAll(filepath.Dir(exampleCopy), 0o755)
	_ = os.WriteFile(exampleCopy, []byte(res.Content), 0o644)

	fmt.Fprintf(os.Stderr, "wrote %s: structs=%d externs=%d skipped=%d\n", outputPath, res.Structs, res.Externs, res.Skipped)
	return nil
}
