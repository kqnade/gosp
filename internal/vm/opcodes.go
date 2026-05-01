package vm

import "fmt"

type Opcode uint8

const (
	OpLoadConst Opcode = iota
	OpLoadVar
	OpMakeClosure
	OpMakeLabel
	OpCall
	OpTailCall
	OpRet
	OpJump
	OpJumpIfFalse
	OpCar
	OpCdr
	OpCons
	OpAtom
	OpEq
	OpPop
)

func (o Opcode) String() string {
	switch o {
	case OpLoadConst:
		return "LOAD_CONST"
	case OpLoadVar:
		return "LOAD_VAR"
	case OpMakeClosure:
		return "MAKE_CLOSURE"
	case OpMakeLabel:
		return "MAKE_LABEL"
	case OpCall:
		return "CALL"
	case OpTailCall:
		return "TAIL_CALL"
	case OpRet:
		return "RET"
	case OpJump:
		return "JUMP"
	case OpJumpIfFalse:
		return "JUMP_IF_FALSE"
	case OpCar:
		return "CAR"
	case OpCdr:
		return "CDR"
	case OpCons:
		return "CONS"
	case OpAtom:
		return "ATOM"
	case OpEq:
		return "EQ"
	case OpPop:
		return "POP"
	default:
		return fmt.Sprintf("OP(%d)", uint8(o))
	}
}

type Instr struct {
	Op  Opcode
	Arg int
}
