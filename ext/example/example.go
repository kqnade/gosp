// Package example is a tiny extension template that shows how to add
// builtins to a gosp.Runtime. Derivative projects (e.g. ext/numbers,
// ext/strings) should follow the same Register(rt) convention.
package example

import (
	"fmt"

	"github.com/kqnade/gosp"
	"github.com/kqnade/gosp/value"
)

// Register installs the example builtins on rt.
func Register(rt *gosp.Runtime) {
	rt.Register("second", second)
	rt.Register("third", third)
}

func second(args []value.Value) (value.Value, error) {
	return nth("second", args, 1)
}

func third(args []value.Value) (value.Value, error) {
	return nth("third", args, 2)
}

func nth(name string, args []value.Value, idx int) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("gosp/ext/example: %s: expected 1 argument, got %d", name, len(args))
	}
	cur := args[0]
	for i := 0; i < idx; i++ {
		pair, ok := cur.(*value.Pair)
		if !ok {
			return nil, fmt.Errorf("gosp/ext/example: %s: list too short", name)
		}
		cur = pair.Cdr
	}
	pair, ok := cur.(*value.Pair)
	if !ok {
		return nil, fmt.Errorf("gosp/ext/example: %s: list too short", name)
	}
	return pair.Car, nil
}
