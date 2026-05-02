package eval

import (
	"fmt"

	"github.com/kqnade/gosp/value"
)

func NewGlobalEnv() *value.Env {
	env := value.NewEnv(nil)
	env.Define("car", value.Builtin{Name: "car", Fn: builtinCar})
	env.Define("cdr", value.Builtin{Name: "cdr", Fn: builtinCdr})
	env.Define("cons", value.Builtin{Name: "cons", Fn: builtinCons})
	env.Define("atom", value.Builtin{Name: "atom", Fn: builtinAtom})
	env.Define("eq", value.Builtin{Name: "eq", Fn: builtinEq})
	return env
}

func builtinCar(args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("gosp: eval: car: wrong number of arguments")
	}
	pair, ok := args[0].(*value.Pair)
	if !ok {
		return nil, fmt.Errorf("gosp: eval: car: argument must be a pair")
	}
	return pair.Car, nil
}

func builtinCdr(args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("gosp: eval: cdr: wrong number of arguments")
	}
	pair, ok := args[0].(*value.Pair)
	if !ok {
		return nil, fmt.Errorf("gosp: eval: cdr: argument must be a pair")
	}
	return pair.Cdr, nil
}

func builtinCons(args []value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("gosp: eval: cons: wrong number of arguments")
	}
	return value.Cons(args[0], args[1]), nil
}

func builtinAtom(args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("gosp: eval: atom: wrong number of arguments")
	}
	if value.IsAtom(args[0]) {
		return value.Symbol{Name: "t"}, nil
	}
	return value.NIL, nil
}

func builtinEq(args []value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("gosp: eval: eq: wrong number of arguments")
	}
	if value.Eq(args[0], args[1]) {
		return value.Symbol{Name: "t"}, nil
	}
	return value.NIL, nil
}
