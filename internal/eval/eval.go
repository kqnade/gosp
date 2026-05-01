package eval

import (
	"fmt"

	"github.com/kqnade/gosp/internal/value"
)

func Eval(v value.Value, env *value.Env) (value.Value, error) {
	switch x := v.(type) {
	case value.Nil:
		return value.NIL, nil
	case value.Symbol:
		switch x.Name {
		case "t":
			return x, nil
		case "nil":
			return value.NIL, nil
		}
		got, ok := env.Lookup(x.Name)
		if !ok {
			return nil, fmt.Errorf("gosp: eval: unbound symbol: %s", x.Name)
		}
		return got, nil
	default:
		return nil, fmt.Errorf("gosp: eval: cannot evaluate %T", v)
	}
}
