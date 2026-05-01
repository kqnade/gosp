package vm

import (
	"strings"
	"testing"

	"github.com/kqnade/gosp/internal/value"
)

func TestRunLoadConst(t *testing.T) {
	code := &Code{
		Instrs: []Instr{{Op: OpLoadConst, Arg: 0}, {Op: OpRet}},
		Consts: []value.Value{value.Symbol{Name: "t"}},
	}
	got, err := Run(code, value.NewEnv(nil))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !value.Eq(got, value.Symbol{Name: "t"}) {
		t.Errorf("got %v, want t", got)
	}
}

func TestRunLoadConstNil(t *testing.T) {
	code := &Code{
		Instrs: []Instr{{Op: OpLoadConst, Arg: 0}, {Op: OpRet}},
		Consts: []value.Value{value.NIL},
	}
	got, err := Run(code, value.NewEnv(nil))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !value.IsNil(got) {
		t.Errorf("got %v, want ()", got)
	}
}

func TestRunLoadVar(t *testing.T) {
	env := value.NewEnv(nil)
	env.Define("x", value.Symbol{Name: "a"})
	code := &Code{
		Instrs: []Instr{{Op: OpLoadVar, Arg: 0}, {Op: OpRet}},
		Syms:   []string{"x"},
	}
	got, err := Run(code, env)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !value.Eq(got, value.Symbol{Name: "a"}) {
		t.Errorf("got %v, want a", got)
	}
}

func TestRunLoadVarUnbound(t *testing.T) {
	code := &Code{
		Instrs: []Instr{{Op: OpLoadVar, Arg: 0}, {Op: OpRet}},
		Syms:   []string{"missing"},
	}
	_, err := Run(code, value.NewEnv(nil))
	if err == nil {
		t.Fatal("expected error for unbound symbol, got nil")
	}
	if !strings.Contains(err.Error(), "missing") {
		t.Errorf("error %q does not mention symbol name", err)
	}
}

func TestRunPop(t *testing.T) {
	code := &Code{
		Instrs: []Instr{
			{Op: OpLoadConst, Arg: 0},
			{Op: OpLoadConst, Arg: 1},
			{Op: OpPop},
			{Op: OpRet},
		},
		Consts: []value.Value{
			value.Symbol{Name: "a"},
			value.Symbol{Name: "b"},
		},
	}
	got, err := Run(code, value.NewEnv(nil))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !value.Eq(got, value.Symbol{Name: "a"}) {
		t.Errorf("got %v, want a (b should have been popped)", got)
	}
}

func TestRunRetEmptyStack(t *testing.T) {
	code := &Code{
		Instrs: []Instr{{Op: OpRet}},
	}
	_, err := Run(code, value.NewEnv(nil))
	if err == nil {
		t.Fatal("expected error for RET on empty stack, got nil")
	}
}
