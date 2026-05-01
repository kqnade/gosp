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
	case *value.Pair:
		return evalPair(x, env)
	default:
		return nil, fmt.Errorf("gosp: eval: cannot evaluate %T", v)
	}
}

func evalPair(p *value.Pair, env *value.Env) (value.Value, error) {
	if head, ok := p.Car.(value.Symbol); ok {
		switch head.Name {
		case "quote":
			return evalQuote(p.Cdr)
		case "cond":
			return evalCond(p.Cdr, env)
		}
	}
	fn, err := Eval(p.Car, env)
	if err != nil {
		return nil, err
	}
	args, err := evalArgs(p.Cdr, env)
	if err != nil {
		return nil, err
	}
	return apply(fn, args)
}

func evalArgs(args value.Value, env *value.Env) ([]value.Value, error) {
	var out []value.Value
	for !value.IsNil(args) {
		pair, ok := args.(*value.Pair)
		if !ok {
			return nil, fmt.Errorf("gosp: eval: improper argument list")
		}
		v, err := Eval(pair.Car, env)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
		args = pair.Cdr
	}
	return out, nil
}

func apply(fn value.Value, args []value.Value) (value.Value, error) {
	switch f := fn.(type) {
	case value.Builtin:
		return f.Fn(args)
	default:
		return nil, fmt.Errorf("gosp: eval: not a function: %T", fn)
	}
}

func evalQuote(args value.Value) (value.Value, error) {
	pair, ok := args.(*value.Pair)
	if !ok {
		return nil, fmt.Errorf("gosp: eval: quote: wrong number of arguments")
	}
	if !value.IsNil(pair.Cdr) {
		return nil, fmt.Errorf("gosp: eval: quote: wrong number of arguments")
	}
	return pair.Car, nil
}

func evalCond(clauses value.Value, env *value.Env) (value.Value, error) {
	for !value.IsNil(clauses) {
		pair, ok := clauses.(*value.Pair)
		if !ok {
			return nil, fmt.Errorf("gosp: eval: cond: malformed clause list")
		}
		clause, ok := pair.Car.(*value.Pair)
		if !ok {
			return nil, fmt.Errorf("gosp: eval: cond: clause must be a list")
		}
		body, ok := clause.Cdr.(*value.Pair)
		if !ok {
			return nil, fmt.Errorf("gosp: eval: cond: clause must have body")
		}
		if !value.IsNil(body.Cdr) {
			return nil, fmt.Errorf("gosp: eval: cond: clause must have exactly one body expression")
		}
		test, err := Eval(clause.Car, env)
		if err != nil {
			return nil, err
		}
		if !value.IsNil(test) {
			return Eval(body.Car, env)
		}
		clauses = pair.Cdr
	}
	return value.NIL, nil
}
