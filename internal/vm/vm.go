package vm

import (
	"fmt"

	"github.com/kqnade/gosp/value"
)

type frame struct {
	code *Code
	pc   int
	env  *value.Env
}

func Run(code *Code, env *value.Env) (value.Value, error) {
	v, _, err := runWithStats(code, env)
	return v, err
}

// RunWithStats is identical to Run but also reports the maximum
// call-frame depth observed during execution. It is intended for tests
// that want to assert tail-call frame replacement is in effect.
func RunWithStats(code *Code, env *value.Env) (value.Value, int, error) {
	return runWithStats(code, env)
}

func runWithStats(code *Code, env *value.Env) (value.Value, int, error) {
	stack := make([]value.Value, 0, 16)
	frames := make([]frame, 0, 8)
	maxFrames := 0
	pc := 0
	for {
		if pc >= len(code.Instrs) {
			return nil, maxFrames, fmt.Errorf("gosp: vm: missing RET")
		}
		ins := code.Instrs[pc]
		pc++
		switch ins.Op {
		case OpLoadConst:
			if ins.Arg < 0 || ins.Arg >= len(code.Consts) {
				return nil, maxFrames, fmt.Errorf("gosp: vm: invalid %s arg %d", ins.Op, ins.Arg)
			}
			stack = append(stack, code.Consts[ins.Arg])
		case OpLoadVar:
			if ins.Arg < 0 || ins.Arg >= len(code.Syms) {
				return nil, maxFrames, fmt.Errorf("gosp: vm: invalid %s arg %d", ins.Op, ins.Arg)
			}
			name := code.Syms[ins.Arg]
			v, ok := env.Lookup(name)
			if !ok {
				return nil, maxFrames, fmt.Errorf("gosp: vm: unbound symbol: %s", name)
			}
			stack = append(stack, v)
		case OpPop:
			if len(stack) == 0 {
				return nil, maxFrames, fmt.Errorf("gosp: vm: pop on empty stack")
			}
			stack = stack[:len(stack)-1]
		case OpJump:
			if ins.Arg < 0 || ins.Arg >= len(code.Instrs) {
				return nil, maxFrames, fmt.Errorf("gosp: vm: invalid %s target %d", ins.Op, ins.Arg)
			}
			pc = ins.Arg
		case OpJumpIfFalse:
			if len(stack) == 0 {
				return nil, maxFrames, fmt.Errorf("gosp: vm: jump-if-false on empty stack")
			}
			if ins.Arg < 0 || ins.Arg >= len(code.Instrs) {
				return nil, maxFrames, fmt.Errorf("gosp: vm: invalid %s target %d", ins.Op, ins.Arg)
			}
			top := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if value.IsNil(top) {
				pc = ins.Arg
			}
		case OpCar:
			if len(stack) == 0 {
				return nil, maxFrames, fmt.Errorf("gosp: vm: car: empty stack")
			}
			top := stack[len(stack)-1]
			pair, ok := top.(*value.Pair)
			if !ok {
				return nil, maxFrames, fmt.Errorf("gosp: vm: car: argument must be a pair")
			}
			stack[len(stack)-1] = pair.Car
		case OpCdr:
			if len(stack) == 0 {
				return nil, maxFrames, fmt.Errorf("gosp: vm: cdr: empty stack")
			}
			top := stack[len(stack)-1]
			pair, ok := top.(*value.Pair)
			if !ok {
				return nil, maxFrames, fmt.Errorf("gosp: vm: cdr: argument must be a pair")
			}
			stack[len(stack)-1] = pair.Cdr
		case OpCons:
			if len(stack) < 2 {
				return nil, maxFrames, fmt.Errorf("gosp: vm: cons: stack underflow")
			}
			cdr := stack[len(stack)-1]
			car := stack[len(stack)-2]
			stack = stack[:len(stack)-1]
			stack[len(stack)-1] = value.Cons(car, cdr)
		case OpAtom:
			if len(stack) == 0 {
				return nil, maxFrames, fmt.Errorf("gosp: vm: atom: empty stack")
			}
			top := stack[len(stack)-1]
			if value.IsAtom(top) {
				stack[len(stack)-1] = value.Symbol{Name: "t"}
			} else {
				stack[len(stack)-1] = value.NIL
			}
		case OpEq:
			if len(stack) < 2 {
				return nil, maxFrames, fmt.Errorf("gosp: vm: eq: stack underflow")
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
				return nil, maxFrames, fmt.Errorf("gosp: vm: make-closure: bad function index %d", ins.Arg)
			}
			proto := code.Funcs[ins.Arg]
			stack = append(stack, &value.Closure{
				Params: proto.Params,
				Body:   proto.Code,
				Env:    env,
			})
		case OpMakeLabel:
			if len(stack) == 0 {
				return nil, maxFrames, fmt.Errorf("gosp: vm: make-label: empty stack")
			}
			if ins.Arg < 0 || ins.Arg >= len(code.Syms) {
				return nil, maxFrames, fmt.Errorf("gosp: vm: invalid %s arg %d", ins.Op, ins.Arg)
			}
			top := stack[len(stack)-1]
			closure, ok := top.(*value.Closure)
			if !ok {
				return nil, maxFrames, fmt.Errorf("gosp: vm: label: expression must evaluate to a function")
			}
			name := value.Symbol{Name: code.Syms[ins.Arg]}
			bound := *closure
			bound.Self = &name
			stack[len(stack)-1] = &bound
		case OpCall:
			callee, args, err := popCallable(&stack, ins.Arg)
			if err != nil {
				return nil, maxFrames, err
			}
			switch fn := callee.(type) {
			case *value.Closure:
				body, newEnv, err := enterClosureChecked(fn, args)
				if err != nil {
					return nil, maxFrames, err
				}
				frames = append(frames, frame{code: code, pc: pc, env: env})
				if len(frames) > maxFrames {
					maxFrames = len(frames)
				}
				code = body
				pc = 0
				env = newEnv
			case value.Builtin:
				result, err := fn.Fn(args)
				if err != nil {
					return nil, maxFrames, err
				}
				stack = append(stack, result)
			default:
				return nil, maxFrames, fmt.Errorf("gosp: vm: call: not a function: %T", callee)
			}
		case OpTailCall:
			callee, args, err := popCallable(&stack, ins.Arg)
			if err != nil {
				return nil, maxFrames, err
			}
			switch fn := callee.(type) {
			case *value.Closure:
				body, newEnv, err := enterClosureChecked(fn, args)
				if err != nil {
					return nil, maxFrames, err
				}
				code = body
				pc = 0
				env = newEnv
			case value.Builtin:
				// A Builtin has no VM frame to replace; fall through to
				// the surrounding RET, which returns the pushed result.
				result, err := fn.Fn(args)
				if err != nil {
					return nil, maxFrames, err
				}
				stack = append(stack, result)
			default:
				return nil, maxFrames, fmt.Errorf("gosp: vm: call: not a function: %T", callee)
			}
		case OpDefineGlobal:
			if len(stack) == 0 {
				return nil, maxFrames, fmt.Errorf("gosp: vm: define-global: empty stack")
			}
			if ins.Arg < 0 || ins.Arg >= len(code.Syms) {
				return nil, maxFrames, fmt.Errorf("gosp: vm: invalid %s arg %d", ins.Op, ins.Arg)
			}
			env.SetGlobal(code.Syms[ins.Arg], stack[len(stack)-1])
		case OpRet:
			if len(stack) == 0 {
				return nil, maxFrames, fmt.Errorf("gosp: vm: ret on empty stack")
			}
			if len(frames) == 0 {
				return stack[len(stack)-1], maxFrames, nil
			}
			fr := frames[len(frames)-1]
			frames = frames[:len(frames)-1]
			code = fr.code
			pc = fr.pc
			env = fr.env
		default:
			return nil, maxFrames, fmt.Errorf("gosp: vm: unknown opcode %s", ins.Op)
		}
	}
}

// popCallable pops `n` arguments and the callee from the stack. The
// callee may be any value; the caller dispatches on its concrete type.
func popCallable(stack *[]value.Value, n int) (value.Value, []value.Value, error) {
	s := *stack
	if len(s) < n+1 {
		return nil, nil, fmt.Errorf("gosp: vm: call: stack underflow")
	}
	argsBase := len(s) - n
	args := append([]value.Value(nil), s[argsBase:]...)
	callee := s[argsBase-1]
	*stack = s[:argsBase-1]
	return callee, args, nil
}

// enterClosureChecked verifies arity and prepares a new VM frame for the
// closure call. Use this from OpCall/OpTailCall after type-asserting to
// *value.Closure.
func enterClosureChecked(c *value.Closure, args []value.Value) (*Code, *value.Env, error) {
	if len(c.Params) != len(args) {
		return nil, nil, fmt.Errorf("gosp: vm: call: wrong number of arguments")
	}
	return enterClosure(c, args)
}

func enterClosure(c *value.Closure, args []value.Value) (*Code, *value.Env, error) {
	body, ok := c.Body.(*Code)
	if !ok {
		return nil, nil, fmt.Errorf("gosp: vm: call: closure body is not VM code")
	}
	newEnv := value.NewEnv(c.Env)
	if c.Self != nil {
		newEnv.Define(c.Self.Name, c)
	}
	for i, p := range c.Params {
		newEnv.Define(p.Name, args[i])
	}
	return body, newEnv, nil
}
