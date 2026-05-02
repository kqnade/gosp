package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kqnade/gosp"
)

func Run(path string, out io.Writer) error {
	return runFile(path, out, false)
}

func RunVM(path string, out io.Writer) error {
	return runFile(path, out, true)
}

func RunInteractive(in io.Reader, out io.Writer) error {
	return newRuntime(false).REPL(in, out)
}

func RunInteractiveVM(in io.Reader, out io.Writer) error {
	return newRuntime(true).REPL(in, out)
}

func runFile(path string, out io.Writer, useVM bool) error {
	rt := newRuntime(useVM)
	v, err := rt.RunFile(path)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, gosp.Print(v))
	return err
}

func newRuntime(useVM bool) *gosp.Runtime {
	if useVM {
		return gosp.New(gosp.WithBackend(gosp.BackendVM))
	}
	return gosp.New()
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
