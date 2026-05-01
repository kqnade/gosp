package compiler

import (
	"fmt"

	"github.com/kqnade/gosp/internal/value"
	"github.com/kqnade/gosp/internal/vm"
)

type Code struct {
	Instrs []vm.Instr
	Consts []value.Value
	Syms   []string
}

func Compile(v value.Value) (*Code, error) {
	c := &Code{}
	if err := c.compile(v); err != nil {
		return nil, err
	}
	c.Instrs = append(c.Instrs, vm.Instr{Op: vm.OpRet})
	return c, nil
}

func (c *Code) compile(v value.Value) error {
	switch x := v.(type) {
	case value.Nil:
		c.emit(vm.OpLoadConst, c.addConst(value.NIL))
		return nil
	case value.Symbol:
		switch x.Name {
		case "t":
			c.emit(vm.OpLoadConst, c.addConst(x))
		case "nil":
			c.emit(vm.OpLoadConst, c.addConst(value.NIL))
		default:
			c.emit(vm.OpLoadVar, c.addSym(x.Name))
		}
		return nil
	default:
		return fmt.Errorf("gosp: compile: cannot compile %T", v)
	}
}

func (c *Code) emit(op vm.Opcode, arg int) {
	c.Instrs = append(c.Instrs, vm.Instr{Op: op, Arg: arg})
}

func (c *Code) addConst(v value.Value) int {
	c.Consts = append(c.Consts, v)
	return len(c.Consts) - 1
}

func (c *Code) addSym(name string) int {
	c.Syms = append(c.Syms, name)
	return len(c.Syms) - 1
}
