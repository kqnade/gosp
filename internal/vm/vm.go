package vm

import (
	"fmt"

	"github.com/kqnade/gosp/internal/value"
)

func Run(code *Code, env *value.Env) (value.Value, error) {
	stack := make([]value.Value, 0, 16)
	for pc := 0; pc < len(code.Instrs); pc++ {
		ins := code.Instrs[pc]
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
