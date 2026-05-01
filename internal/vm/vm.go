package vm

import (
	"fmt"

	"github.com/kqnade/gosp/internal/value"
)

func Run(code *Code, env *value.Env) (value.Value, error) {
	stack := make([]value.Value, 0, 16)
	pc := 0
	for pc < len(code.Instrs) {
		ins := code.Instrs[pc]
		pc++
		switch ins.Op {
		case OpLoadConst:
			stack = append(stack, code.Consts[ins.Arg])
		case OpLoadVar:
			name := code.Syms[ins.Arg]
			v, ok := env.Lookup(name)
			if !ok {
				return nil, fmt.Errorf("gosp: vm: unbound symbol: %s", name)
			}
			stack = append(stack, v)
		case OpPop:
			if len(stack) == 0 {
				return nil, fmt.Errorf("gosp: vm: pop on empty stack")
			}
			stack = stack[:len(stack)-1]
		case OpJump:
			pc = ins.Arg
		case OpJumpIfFalse:
			if len(stack) == 0 {
				return nil, fmt.Errorf("gosp: vm: jump-if-false on empty stack")
			}
			top := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if value.IsNil(top) {
				pc = ins.Arg
			}
		case OpCar:
			if len(stack) == 0 {
				return nil, fmt.Errorf("gosp: vm: car: empty stack")
			}
			top := stack[len(stack)-1]
			pair, ok := top.(*value.Pair)
			if !ok {
				return nil, fmt.Errorf("gosp: vm: car: argument must be a pair")
			}
			stack[len(stack)-1] = pair.Car
		case OpCdr:
			if len(stack) == 0 {
				return nil, fmt.Errorf("gosp: vm: cdr: empty stack")
			}
			top := stack[len(stack)-1]
			pair, ok := top.(*value.Pair)
			if !ok {
				return nil, fmt.Errorf("gosp: vm: cdr: argument must be a pair")
			}
			stack[len(stack)-1] = pair.Cdr
		case OpCons:
			if len(stack) < 2 {
				return nil, fmt.Errorf("gosp: vm: cons: stack underflow")
			}
			cdr := stack[len(stack)-1]
			car := stack[len(stack)-2]
			stack = stack[:len(stack)-1]
			stack[len(stack)-1] = value.Cons(car, cdr)
		case OpRet:
			if len(stack) == 0 {
				return nil, fmt.Errorf("gosp: vm: ret on empty stack")
			}
			return stack[len(stack)-1], nil
		default:
			return nil, fmt.Errorf("gosp: vm: unknown opcode %s", ins.Op)
		}
	}
	return nil, fmt.Errorf("gosp: vm: missing RET")
}
