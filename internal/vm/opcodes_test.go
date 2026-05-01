package vm

import "testing"

func TestOpcodeString(t *testing.T) {
	tests := []struct {
		op   Opcode
		want string
	}{
		{OpLoadConst, "LOAD_CONST"},
		{OpLoadVar, "LOAD_VAR"},
		{OpMakeClosure, "MAKE_CLOSURE"},
		{OpMakeLabel, "MAKE_LABEL"},
		{OpCall, "CALL"},
		{OpTailCall, "TAIL_CALL"},
		{OpRet, "RET"},
		{OpJump, "JUMP"},
		{OpJumpIfFalse, "JUMP_IF_FALSE"},
		{OpCar, "CAR"},
		{OpCdr, "CDR"},
		{OpCons, "CONS"},
		{OpAtom, "ATOM"},
		{OpEq, "EQ"},
		{OpPop, "POP"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.op.String(); got != tt.want {
				t.Errorf("Opcode(%d).String() = %q, want %q", tt.op, got, tt.want)
			}
		})
	}
}

func TestOpcodeStringUnknown(t *testing.T) {
	const bogus Opcode = 255
	got := bogus.String()
	if got == "" {
		t.Fatalf("expected non-empty string for unknown opcode, got empty")
	}
	// Should clearly mark it as unknown so debug output is not silently misleading.
	if got == "LOAD_CONST" || got == "POP" {
		t.Fatalf("unknown opcode unexpectedly aliased to a real name: %q", got)
	}
}

func TestInstrFields(t *testing.T) {
	in := Instr{Op: OpLoadConst, Arg: 7}
	if in.Op != OpLoadConst {
		t.Errorf("Op = %v, want OpLoadConst", in.Op)
	}
	if in.Arg != 7 {
		t.Errorf("Arg = %d, want 7", in.Arg)
	}
}
