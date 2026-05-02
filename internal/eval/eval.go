package eval

import (
	"fmt"

	"github.com/kqnade/gosp/value"
)

func EvalProgram(forms []value.Value, env *value.Env) (value.Value, error) {
	var last value.Value = value.NIL
	for _, form := range forms {
		v, err := evalTopLevel(form, env)
		if err != nil {
			return nil, err
		}
		last = v
	}
	return last, nil
}

func evalTopLevel(form value.Value, env *value.Env) (value.Value, error) {
	if pair, ok := form.(*value.Pair); ok {
		if head, ok := pair.Car.(value.Symbol); ok && head.Name == "label" {
			return evalTopLevelLabel(pair.Cdr, env)
		}
	}
	return Eval(form, env)
}

func evalTopLevelLabel(form value.Value, env *value.Env) (value.Value, error) {
	pair, ok := form.(*value.Pair)
	if !ok {
		return nil, fmt.Errorf("gosp: eval: label: missing name")
	}
	name, ok := pair.Car.(value.Symbol)
	if !ok {
		return nil, fmt.Errorf("gosp: eval: label: name must be a symbol")
	}
	rest, ok := pair.Cdr.(*value.Pair)
	if !ok {
		return nil, fmt.Errorf("gosp: eval: label: missing expression")
	}
	if !value.IsNil(rest.Cdr) {
		return nil, fmt.Errorf("gosp: eval: label: too many arguments")
	}
	v, err := Eval(rest.Car, env)
	if err != nil {
		return nil, err
	}
	env.SetGlobal(name.Name, v)
	return v, nil
}

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
		case "lambda":
			return evalLambda(p.Cdr, env)
		case "label":
			return evalLabel(p.Cdr, env)
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

func evalLambda(form value.Value, env *value.Env) (value.Value, error) {
	pair, ok := form.(*value.Pair)
	if !ok {
		return nil, fmt.Errorf("gosp: eval: lambda: missing parameter list")
	}
	body, ok := pair.Cdr.(*value.Pair)
	if !ok {
		return nil, fmt.Errorf("gosp: eval: lambda: missing body")
	}
	if !value.IsNil(body.Cdr) {
		return nil, fmt.Errorf("gosp: eval: lambda: body must be a single expression")
	}
	params, err := parseParams(pair.Car)
	if err != nil {
		return nil, err
	}
	return &value.Func{Params: params, Body: body.Car, Env: env}, nil
}

func evalLabel(form value.Value, env *value.Env) (value.Value, error) {
	pair, ok := form.(*value.Pair)
	if !ok {
		return nil, fmt.Errorf("gosp: eval: label: missing name")
	}
	name, ok := pair.Car.(value.Symbol)
	if !ok {
		return nil, fmt.Errorf("gosp: eval: label: name must be a symbol")
	}
	rest, ok := pair.Cdr.(*value.Pair)
	if !ok {
		return nil, fmt.Errorf("gosp: eval: label: missing expression")
	}
	if !value.IsNil(rest.Cdr) {
		return nil, fmt.Errorf("gosp: eval: label: too many arguments")
	}
	v, err := Eval(rest.Car, env)
	if err != nil {
		return nil, err
	}
	fn, ok := v.(*value.Func)
	if !ok {
		return nil, fmt.Errorf("gosp: eval: label: expression must evaluate to a function")
	}
	bound := *fn
	self := name
	bound.Self = &self
	return &bound, nil
}

func parseParams(v value.Value) ([]value.Symbol, error) {
	var out []value.Symbol
	for !value.IsNil(v) {
		pair, ok := v.(*value.Pair)
		if !ok {
			return nil, fmt.Errorf("gosp: eval: lambda: parameter list must be proper")
		}
		sym, ok := pair.Car.(value.Symbol)
		if !ok {
			return nil, fmt.Errorf("gosp: eval: lambda: parameter must be a symbol")
		}
		out = append(out, sym)
		v = pair.Cdr
	}
	return out, nil
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
	case *value.Func:
		if len(args) != len(f.Params) {
			return nil, fmt.Errorf("gosp: eval: wrong number of arguments")
		}
		frame := value.NewEnv(f.Env)
		if f.Self != nil {
			frame.Define(f.Self.Name, f)
		}
		for i, p := range f.Params {
			frame.Define(p.Name, args[i])
		}
		return Eval(f.Body, frame)
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
