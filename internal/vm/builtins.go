package vm

import "github.com/kqnade/gosp/internal/value"

// NewGlobalEnv returns a VM environment seeded with closures for the
// McCarthy primitives. Each primitive is a tiny VM-callable closure
// whose body executes the corresponding opcode, so primitive symbols
// like `car` can be loaded as values via OpLoadVar (e.g. used as the
// callee in a higher-order call) without falling back to an unbound
// symbol error.
func NewGlobalEnv() *value.Env {
	env := value.NewEnv(nil)
	env.Define("car", unaryPrim(OpCar))
	env.Define("cdr", unaryPrim(OpCdr))
	env.Define("atom", unaryPrim(OpAtom))
	env.Define("cons", binaryPrim(OpCons))
	env.Define("eq", binaryPrim(OpEq))
	return env
}

func unaryPrim(op Opcode) *value.Closure {
	code := &Code{
		Instrs: []Instr{
			{Op: OpLoadVar, Arg: 0},
			{Op: op},
			{Op: OpRet},
		},
		Syms: []string{"x"},
	}
	return &value.Closure{
		Params: []value.Symbol{{Name: "x"}},
		Body:   code,
	}
}

func binaryPrim(op Opcode) *value.Closure {
	code := &Code{
		Instrs: []Instr{
			{Op: OpLoadVar, Arg: 0},
			{Op: OpLoadVar, Arg: 1},
			{Op: op},
			{Op: OpRet},
		},
		Syms: []string{"x", "y"},
	}
	return &value.Closure{
		Params: []value.Symbol{{Name: "x"}, {Name: "y"}},
		Body:   code,
	}
}
