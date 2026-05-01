package compiler

import (
	"fmt"

	"github.com/kqnade/gosp/internal/value"
	"github.com/kqnade/gosp/internal/vm"
)

func Compile(v value.Value) (*vm.Code, error) {
	c := &builder{code: &vm.Code{}}
	if err := c.compile(v); err != nil {
		return nil, err
	}
	c.emit(vm.OpRet, 0)
	return c.code, nil
}

type builder struct {
	code *vm.Code
}

func (b *builder) compile(v value.Value) error {
	switch x := v.(type) {
	case value.Nil:
		b.emit(vm.OpLoadConst, b.addConst(value.NIL))
		return nil
	case value.Symbol:
		switch x.Name {
		case "t":
			b.emit(vm.OpLoadConst, b.addConst(x))
		case "nil":
			b.emit(vm.OpLoadConst, b.addConst(value.NIL))
		default:
			b.emit(vm.OpLoadVar, b.addSym(x.Name))
		}
		return nil
	case *value.Pair:
		return b.compilePair(x)
	default:
		return fmt.Errorf("gosp: compile: cannot compile %T", v)
	}
}

func (b *builder) compilePair(p *value.Pair) error {
	if head, ok := p.Car.(value.Symbol); ok {
		switch head.Name {
		case "quote":
			return b.compileQuote(p.Cdr)
		case "cond":
			return b.compileCond(p.Cdr)
		case "car":
			return b.compileUnary("car", vm.OpCar, p.Cdr)
		case "cdr":
			return b.compileUnary("cdr", vm.OpCdr, p.Cdr)
		case "cons":
			return b.compileBinary("cons", vm.OpCons, p.Cdr)
		case "atom":
			return b.compileUnary("atom", vm.OpAtom, p.Cdr)
		case "eq":
			return b.compileBinary("eq", vm.OpEq, p.Cdr)
		case "lambda":
			return b.compileLambda(p.Cdr)
		}
	}
	return b.compileCall(p)
}

func (b *builder) compileLambda(form value.Value) error {
	proto, err := buildFuncProto(form)
	if err != nil {
		return err
	}
	idx := len(b.code.Funcs)
	b.code.Funcs = append(b.code.Funcs, proto)
	b.emit(vm.OpMakeClosure, idx)
	return nil
}

func buildFuncProto(form value.Value) (*vm.FuncProto, error) {
	pair, ok := form.(*value.Pair)
	if !ok {
		return nil, fmt.Errorf("gosp: compile: lambda: missing parameter list")
	}
	body, ok := pair.Cdr.(*value.Pair)
	if !ok {
		return nil, fmt.Errorf("gosp: compile: lambda: missing body")
	}
	if !value.IsNil(body.Cdr) {
		return nil, fmt.Errorf("gosp: compile: lambda: body must be a single expression")
	}
	params, err := parseParams(pair.Car)
	if err != nil {
		return nil, err
	}
	bb := &builder{code: &vm.Code{}}
	if err := bb.compile(body.Car); err != nil {
		return nil, err
	}
	bb.emit(vm.OpRet, 0)
	return &vm.FuncProto{Params: params, Code: bb.code}, nil
}

func parseParams(v value.Value) ([]value.Symbol, error) {
	var out []value.Symbol
	for !value.IsNil(v) {
		pair, ok := v.(*value.Pair)
		if !ok {
			return nil, fmt.Errorf("gosp: compile: lambda: parameter list must be proper")
		}
		sym, ok := pair.Car.(value.Symbol)
		if !ok {
			return nil, fmt.Errorf("gosp: compile: lambda: parameter must be a symbol")
		}
		out = append(out, sym)
		v = pair.Cdr
	}
	return out, nil
}

func (b *builder) compileCall(p *value.Pair) error {
	args, err := flatArgs(p.Cdr)
	if err != nil {
		return err
	}
	if err := b.compile(p.Car); err != nil {
		return err
	}
	for _, a := range args {
		if err := b.compile(a); err != nil {
			return err
		}
	}
	b.emit(vm.OpCall, len(args))
	return nil
}

func (b *builder) compileUnary(name string, op vm.Opcode, args value.Value) error {
	xs, err := flatArgs(args)
	if err != nil {
		return err
	}
	if len(xs) != 1 {
		return fmt.Errorf("gosp: compile: %s: wrong number of arguments", name)
	}
	if err := b.compile(xs[0]); err != nil {
		return err
	}
	b.emit(op, 0)
	return nil
}

func (b *builder) compileBinary(name string, op vm.Opcode, args value.Value) error {
	xs, err := flatArgs(args)
	if err != nil {
		return err
	}
	if len(xs) != 2 {
		return fmt.Errorf("gosp: compile: %s: wrong number of arguments", name)
	}
	if err := b.compile(xs[0]); err != nil {
		return err
	}
	if err := b.compile(xs[1]); err != nil {
		return err
	}
	b.emit(op, 0)
	return nil
}

func flatArgs(v value.Value) ([]value.Value, error) {
	var out []value.Value
	for !value.IsNil(v) {
		pair, ok := v.(*value.Pair)
		if !ok {
			return nil, fmt.Errorf("gosp: compile: improper argument list")
		}
		out = append(out, pair.Car)
		v = pair.Cdr
	}
	return out, nil
}

func (b *builder) compileCond(clauses value.Value) error {
	var endJumps []int
	for !value.IsNil(clauses) {
		pair, ok := clauses.(*value.Pair)
		if !ok {
			return fmt.Errorf("gosp: compile: cond: malformed clause list")
		}
		clause, ok := pair.Car.(*value.Pair)
		if !ok {
			return fmt.Errorf("gosp: compile: cond: clause must be a list")
		}
		body, ok := clause.Cdr.(*value.Pair)
		if !ok {
			return fmt.Errorf("gosp: compile: cond: clause must have body")
		}
		if !value.IsNil(body.Cdr) {
			return fmt.Errorf("gosp: compile: cond: clause must have exactly one body expression")
		}
		if err := b.compile(clause.Car); err != nil {
			return err
		}
		jifPos := len(b.code.Instrs)
		b.emit(vm.OpJumpIfFalse, 0)
		if err := b.compile(body.Car); err != nil {
			return err
		}
		jumpPos := len(b.code.Instrs)
		b.emit(vm.OpJump, 0)
		endJumps = append(endJumps, jumpPos)
		b.code.Instrs[jifPos].Arg = len(b.code.Instrs)
		clauses = pair.Cdr
	}
	b.emit(vm.OpLoadConst, b.addConst(value.NIL))
	end := len(b.code.Instrs)
	for _, p := range endJumps {
		b.code.Instrs[p].Arg = end
	}
	return nil
}

func (b *builder) compileQuote(args value.Value) error {
	pair, ok := args.(*value.Pair)
	if !ok {
		return fmt.Errorf("gosp: compile: quote: wrong number of arguments")
	}
	if !value.IsNil(pair.Cdr) {
		return fmt.Errorf("gosp: compile: quote: wrong number of arguments")
	}
	b.emit(vm.OpLoadConst, b.addConst(pair.Car))
	return nil
}

func (b *builder) emit(op vm.Opcode, arg int) {
	b.code.Instrs = append(b.code.Instrs, vm.Instr{Op: op, Arg: arg})
}

func (b *builder) addConst(v value.Value) int {
	b.code.Consts = append(b.code.Consts, v)
	return len(b.code.Consts) - 1
}

func (b *builder) addSym(name string) int {
	b.code.Syms = append(b.code.Syms, name)
	return len(b.code.Syms) - 1
}
