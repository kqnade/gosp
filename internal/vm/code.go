package vm

import "github.com/kqnade/gosp/internal/value"

type Code struct {
	Instrs []Instr
	Consts []value.Value
	Syms   []string
}
