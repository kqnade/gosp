package vm

import "github.com/kqnade/gosp/value"

type Code struct {
	Instrs []Instr
	Consts []value.Value
	Syms   []string
	Funcs  []*FuncProto
}

// FuncProto is a compiled lambda body together with its parameter list.
// It is referenced by MAKE_CLOSURE via its index in the enclosing Code.Funcs.
type FuncProto struct {
	Params []value.Symbol
	Code   *Code
}
