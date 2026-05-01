package main

import (
	"fmt"
	"io"
	"os"

	"github.com/kqnade/gosp/internal/eval"
	"github.com/kqnade/gosp/internal/printer"
	"github.com/kqnade/gosp/internal/reader"
)

func Run(path string, out io.Writer) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("gosp: %w", err)
	}
	forms, err := reader.ReadAll(string(src))
	if err != nil {
		return err
	}
	env := eval.NewGlobalEnv()
	result, err := eval.EvalProgram(forms, env)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, printer.Print(result)); err != nil {
		return err
	}
	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: gosp <file>")
		os.Exit(2)
	}
	if err := Run(os.Args[1], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
