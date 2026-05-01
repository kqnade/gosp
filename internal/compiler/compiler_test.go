package compiler

import (
	"testing"

	"github.com/kqnade/gosp/internal/value"
	"github.com/kqnade/gosp/internal/vm"
)

func TestCompileT(t *testing.T) {
	code, err := Compile(value.Symbol{Name: "t"})
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	wantInstrs := []vm.Instr{{Op: vm.OpLoadConst, Arg: 0}, {Op: vm.OpRet}}
	if !equalInstrs(code.Instrs, wantInstrs) {
		t.Errorf("Instrs = %v, want %v", code.Instrs, wantInstrs)
	}
	if len(code.Consts) != 1 || !value.Eq(code.Consts[0], value.Symbol{Name: "t"}) {
		t.Errorf("Consts = %v, want [t]", code.Consts)
	}
	if len(code.Syms) != 0 {
		t.Errorf("Syms = %v, want []", code.Syms)
	}
}

func TestCompileNilSymbol(t *testing.T) {
	code, err := Compile(value.Symbol{Name: "nil"})
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	wantInstrs := []vm.Instr{{Op: vm.OpLoadConst, Arg: 0}, {Op: vm.OpRet}}
	if !equalInstrs(code.Instrs, wantInstrs) {
		t.Errorf("Instrs = %v, want %v", code.Instrs, wantInstrs)
	}
	if len(code.Consts) != 1 || !value.IsNil(code.Consts[0]) {
		t.Errorf("Consts = %v, want [Nil]", code.Consts)
	}
	if len(code.Syms) != 0 {
		t.Errorf("Syms = %v, want []", code.Syms)
	}
}

func TestCompileNilLiteral(t *testing.T) {
	code, err := Compile(value.NIL)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	wantInstrs := []vm.Instr{{Op: vm.OpLoadConst, Arg: 0}, {Op: vm.OpRet}}
	if !equalInstrs(code.Instrs, wantInstrs) {
		t.Errorf("Instrs = %v, want %v", code.Instrs, wantInstrs)
	}
	if len(code.Consts) != 1 || !value.IsNil(code.Consts[0]) {
		t.Errorf("Consts = %v, want [Nil]", code.Consts)
	}
}

func TestCompileVarLookup(t *testing.T) {
	code, err := Compile(value.Symbol{Name: "foo"})
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	wantInstrs := []vm.Instr{{Op: vm.OpLoadVar, Arg: 0}, {Op: vm.OpRet}}
	if !equalInstrs(code.Instrs, wantInstrs) {
		t.Errorf("Instrs = %v, want %v", code.Instrs, wantInstrs)
	}
	if len(code.Consts) != 0 {
		t.Errorf("Consts = %v, want []", code.Consts)
	}
	if len(code.Syms) != 1 || code.Syms[0] != "foo" {
		t.Errorf("Syms = %v, want [foo]", code.Syms)
	}
}

func equalInstrs(a, b []vm.Instr) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
