package vm

import (
	"fmt"
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

func TestRunCallBuiltin(t *testing.T) {
	// Use a Builtin (the tree-walker representation) directly from the
	// VM to ensure cross-evaluator interop. Program:
	//   LOAD_VAR "id"; LOAD_CONST 'a; CALL 1; RET
	id := value.Builtin{
		Name: "id",
		Fn: func(args []value.Value) (value.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("id: arity")
			}
			return args[0], nil
		},
	}
	env := value.NewEnv(nil)
	env.Define("id", id)
	code := &Code{
		Instrs: []Instr{
			{Op: OpLoadVar, Arg: 0},
			{Op: OpLoadConst, Arg: 0},
			{Op: OpCall, Arg: 1},
			{Op: OpRet},
		},
		Consts: []value.Value{value.Symbol{Name: "a"}},
		Syms:   []string{"id"},
	}
	got, err := Run(code, env)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !value.Eq(got, value.Symbol{Name: "a"}) {
		t.Errorf("got %v, want a", got)
	}
}

func TestRunTailCallBuiltinReturnsValue(t *testing.T) {
	// Tail-call to a Builtin should fall through to the surrounding RET
	// and return the Builtin's result.
	id := value.Builtin{
		Name: "id",
		Fn: func(args []value.Value) (value.Value, error) {
			return args[0], nil
		},
	}
	env := value.NewEnv(nil)
	env.Define("id", id)
	code := &Code{
		Instrs: []Instr{
			{Op: OpLoadVar, Arg: 0},
			{Op: OpLoadConst, Arg: 0},
			{Op: OpTailCall, Arg: 1},
			{Op: OpRet},
		},
		Consts: []value.Value{value.Symbol{Name: "z"}},
		Syms:   []string{"id"},
	}
	got, err := Run(code, env)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !value.Eq(got, value.Symbol{Name: "z"}) {
		t.Errorf("got %v, want z", got)
	}
}

func TestRunCallBuiltinPropagatesError(t *testing.T) {
	bad := value.Builtin{
		Name: "bad",
		Fn: func(args []value.Value) (value.Value, error) {
			return nil, fmt.Errorf("bad: boom")
		},
	}
	env := value.NewEnv(nil)
	env.Define("bad", bad)
	code := &Code{
		Instrs: []Instr{
			{Op: OpLoadVar, Arg: 0},
			{Op: OpCall, Arg: 0},
			{Op: OpRet},
		},
		Syms: []string{"bad"},
	}
	if _, err := Run(code, env); err == nil {
		t.Fatal("expected Builtin error to surface, got nil")
	}
}

func TestRunMalformedBytecodeReturnsError(t *testing.T) {
	cases := []struct {
		name string
		code *Code
	}{
		{
			name: "OpLoadConst out of range",
			code: &Code{Instrs: []Instr{{Op: OpLoadConst, Arg: 5}, {Op: OpRet}}},
		},
		{
			name: "OpLoadConst negative",
			code: &Code{Instrs: []Instr{{Op: OpLoadConst, Arg: -1}, {Op: OpRet}}},
		},
		{
			name: "OpLoadVar out of range",
			code: &Code{Instrs: []Instr{{Op: OpLoadVar, Arg: 0}, {Op: OpRet}}},
		},
		{
			name: "OpJump negative",
			code: &Code{Instrs: []Instr{{Op: OpJump, Arg: -1}, {Op: OpRet}}},
		},
		{
			name: "OpJump past end",
			code: &Code{Instrs: []Instr{{Op: OpJump, Arg: 99}, {Op: OpRet}}},
		},
		{
			name: "OpJumpIfFalse out of range",
			code: &Code{
				Instrs: []Instr{
					{Op: OpLoadConst, Arg: 0},
					{Op: OpJumpIfFalse, Arg: 99},
					{Op: OpRet},
				},
				Consts: []value.Value{value.NIL},
			},
		},
		{
			name: "OpMakeLabel sym out of range",
			code: &Code{
				Instrs: []Instr{
					{Op: OpMakeClosure, Arg: 0},
					{Op: OpMakeLabel, Arg: 5},
					{Op: OpRet},
				},
				Funcs: []*FuncProto{{
					Code: &Code{Instrs: []Instr{{Op: OpRet}}},
				}},
			},
		},
		{
			name: "OpDefineGlobal sym out of range",
			code: &Code{
				Instrs: []Instr{
					{Op: OpLoadConst, Arg: 0},
					{Op: OpDefineGlobal, Arg: 5},
					{Op: OpRet},
				},
				Consts: []value.Value{value.NIL},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Run(tc.code, value.NewEnv(nil))
			if err == nil {
				t.Fatal("expected error for malformed bytecode, got nil")
			}
			if !strings.Contains(err.Error(), "gosp: vm:") {
				t.Errorf("err = %v, want gosp: vm: prefix", err)
			}
		})
	}
}
