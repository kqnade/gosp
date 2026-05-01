package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kqnade/gosp/internal/eval"
	"github.com/kqnade/gosp/internal/printer"
	"github.com/kqnade/gosp/internal/reader"
	"github.com/kqnade/gosp/internal/value"
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

func RunInteractive(in io.Reader, out io.Writer) error {
	env := eval.NewGlobalEnv()
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	var buf strings.Builder
	for {
		if buf.Len() == 0 {
			fmt.Fprint(out, "> ")
		} else {
			fmt.Fprint(out, "  ")
		}
		if !scanner.Scan() {
			break
		}
		buf.WriteString(scanner.Text())
		buf.WriteByte('\n')
		if !parensBalanced(buf.String()) {
			continue
		}
		src := buf.String()
		buf.Reset()
		forms, err := reader.ReadAll(src)
		if err != nil {
			fmt.Fprintln(out, err)
			continue
		}
		for _, form := range forms {
			v, err := eval.EvalProgram([]value.Value{form}, env)
			if err != nil {
				fmt.Fprintln(out, err)
				break
			}
			fmt.Fprintln(out, printer.Print(v))
		}
	}
	return scanner.Err()
}

func parensBalanced(src string) bool {
	depth := 0
	inComment := false
	for _, r := range src {
		if inComment {
			if r == '\n' {
				inComment = false
			}
			continue
		}
		switch r {
		case ';':
			inComment = true
		case '(':
			depth++
		case ')':
			depth--
		}
	}
	return depth <= 0
}

func main() {
	if len(os.Args) == 1 {
		if err := RunInteractive(os.Stdin, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: gosp [file]")
		os.Exit(2)
	}
	if err := Run(os.Args[1], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
