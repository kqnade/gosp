// embed demonstrates using gosp as an embedded Lisp inside a Go
// program: construct a Runtime, register a host-defined builtin, and
// evaluate source code.
package main

import (
	"fmt"
	"os"

	"github.com/kqnade/gosp"
	"github.com/kqnade/gosp/value"
)

func main() {
	rt := gosp.New()

	// Expose a host-defined builtin: (greet 'world) => (hello world).
	rt.Register("greet", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("greet: expected 1 argument, got %d", len(args))
		}
		return value.Cons(value.Symbol{Name: "hello"}, value.Cons(args[0], value.NIL)), nil
	})

	src := `
        (label snd (lambda (x) (car (cdr x))))
        (greet (snd '(a b c)))
    `
	v, err := rt.Eval(src)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(gosp.Print(v))
}
