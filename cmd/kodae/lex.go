package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	lex "kodae/internal/lexer"
	"kodae/internal/token"
)

// runLex writes one line per token: type_id, kind, literal, line, column.
func runLex(r io.Reader, w io.Writer) error {
	src, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	l := lex.New(string(src))
	for {
		tk := l.Next()
		_, _ = fmt.Fprintf(w, "%d\t%s\t%q\t%d\t%d\n", int(tk.Type), tokenKindName(tk.Type), tk.Literal, tk.Line, tk.Col)
		if tk.Type == token.ILLEGAL {
			return fmt.Errorf("illegal token %q at %d:%d", tk.Literal, tk.Line, tk.Col)
		}
		if tk.Type == token.EOF {
			return nil
		}
	}
}

func runLexFile(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return runLex(bytes.NewReader(b), os.Stdout)
}
