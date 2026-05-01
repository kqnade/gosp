package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kqnade/gosp/internal/compiler"
	"github.com/kqnade/gosp/internal/eval"
	"github.com/kqnade/gosp/internal/printer"
	"github.com/kqnade/gosp/internal/reader"
	"github.com/kqnade/gosp/internal/value"
	"github.com/kqnade/gosp/internal/vm"
)

func Run(path string, out io.Writer) error {
	return runFile(path, out, false)
}

func RunVM(path string, out io.Writer) error {
	return runFile(path, out, true)
}

func runFile(path string, out io.Writer, useVM bool) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("gosp: %w", err)
	}
	forms, err := reader.ReadAll(string(src))
	if err != nil {
		return err
	}
	var result value.Value
	if useVM {
		code, err := compiler.CompileProgram(forms)
		if err != nil {
			return err
		}
		result, err = vm.Run(code, value.NewEnv(nil))
		if err != nil {
			return err
		}
	} else {
		env := eval.NewGlobalEnv()
		result, err = eval.EvalProgram(forms, env)
		if err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(out, printer.Print(result)); err != nil {
		return err
	}
	return nil
}

func RunInteractive(in io.Reader, out io.Writer) error {
	return runInteractive(in, out, false)
}

func RunInteractiveVM(in io.Reader, out io.Writer) error {
	return runInteractive(in, out, true)
}

func runInteractive(in io.Reader, out io.Writer, useVM bool) error {
	treeEnv := eval.NewGlobalEnv()
	vmEnv := value.NewEnv(nil)
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
			var v value.Value
			var err error
			if useVM {
				var code *vm.Code
				code, err = compiler.CompileProgram([]value.Value{form})
				if err == nil {
					v, err = vm.Run(code, vmEnv)
				}
			} else {
				v, err = eval.EvalProgram([]value.Value{form}, treeEnv)
			}
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
	useVM := flag.Bool("vm", false, "use bytecode VM backend instead of the tree-walking evaluator")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: gosp [-vm] [file]")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		var err error
		if *useVM {
			err = RunInteractiveVM(os.Stdin, os.Stdout)
		} else {
			err = RunInteractive(os.Stdin, os.Stdout)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if len(args) != 1 {
		flag.Usage()
		os.Exit(2)
	}
	var err error
	if *useVM {
		err = RunVM(args[0], os.Stdout)
	} else {
		err = Run(args[0], os.Stdout)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
