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
