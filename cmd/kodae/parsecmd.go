package main

import (
	"fmt"
	"os"

	"kodae/internal/ast"
)

// runParseFiles loads one or more sources (like build) and prints the merged AST.
func runParseFiles(paths []string) error {
	pr, err := loadProgram(paths)
	if err != nil {
		return err
	}
	ast.Fprint(os.Stdout, pr)
	if pr == nil {
		return fmt.Errorf("no program")
	}
	return nil
}
