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
		}
	}
	return fmt.Errorf("gosp: compile: cannot compile call form")
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
