package vm

import (
	"fmt"

	"github.com/kqnade/gosp/internal/value"
)

type frame struct {
	code *Code
	pc   int
	env  *value.Env
}

func Run(code *Code, env *value.Env) (value.Value, error) {
	stack := make([]value.Value, 0, 16)
	frames := make([]frame, 0, 8)
	pc := 0
	for {
		if pc >= len(code.Instrs) {
			return nil, fmt.Errorf("gosp: vm: missing RET")
		}
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
		case OpAtom:
			if len(stack) == 0 {
				return nil, fmt.Errorf("gosp: vm: atom: empty stack")
			}
			top := stack[len(stack)-1]
			if value.IsAtom(top) {
				stack[len(stack)-1] = value.Symbol{Name: "t"}
			} else {
				stack[len(stack)-1] = value.NIL
			}
		case OpEq:
			if len(stack) < 2 {
				return nil, fmt.Errorf("gosp: vm: eq: stack underflow")
			}
			rhs := stack[len(stack)-1]
			lhs := stack[len(stack)-2]
			stack = stack[:len(stack)-1]
			if value.Eq(lhs, rhs) {
				stack[len(stack)-1] = value.Symbol{Name: "t"}
			} else {
				stack[len(stack)-1] = value.NIL
			}
		case OpMakeClosure:
			if ins.Arg < 0 || ins.Arg >= len(code.Funcs) {
				return nil, fmt.Errorf("gosp: vm: make-closure: bad function index %d", ins.Arg)
			}
			proto := code.Funcs[ins.Arg]
			stack = append(stack, &value.Closure{
				Params: proto.Params,
				Body:   proto.Code,
				Env:    env,
			})
		case OpCall, OpTailCall:
			n := ins.Arg
			if len(stack) < n+1 {
				return nil, fmt.Errorf("gosp: vm: call: stack underflow")
			}
			argsBase := len(stack) - n
			args := append([]value.Value(nil), stack[argsBase:]...)
			callee := stack[argsBase-1]
			stack = stack[:argsBase-1]
			closure, ok := callee.(*value.Closure)
			if !ok {
				return nil, fmt.Errorf("gosp: vm: call: not a function: %T", callee)
			}
			if len(closure.Params) != n {
				return nil, fmt.Errorf("gosp: vm: call: wrong number of arguments")
			}
			body, ok := closure.Body.(*Code)
			if !ok {
				return nil, fmt.Errorf("gosp: vm: call: closure body is not VM code")
			}
			newEnv := value.NewEnv(closure.Env)
			if closure.Self != nil {
				newEnv.Define(closure.Self.Name, closure)
			}
			for i, p := range closure.Params {
				newEnv.Define(p.Name, args[i])
			}
			frames = append(frames, frame{code: code, pc: pc, env: env})
			code = body
			pc = 0
			env = newEnv
		case OpRet:
			if len(stack) == 0 {
				return nil, fmt.Errorf("gosp: vm: ret on empty stack")
			}
			if len(frames) == 0 {
				return stack[len(stack)-1], nil
			}
			fr := frames[len(frames)-1]
			frames = frames[:len(frames)-1]
			code = fr.code
			pc = fr.pc
			env = fr.env
		default:
			return nil, fmt.Errorf("gosp: vm: unknown opcode %s", ins.Op)
		}
	}
}
