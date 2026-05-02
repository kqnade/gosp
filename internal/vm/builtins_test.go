package vm

import (
	"testing"

	"github.com/kqnade/gosp/internal/value"
)

func TestNewGlobalEnvDefinesPrimitives(t *testing.T) {
	env := NewGlobalEnv()
	for _, name := range []string{"car", "cdr", "atom", "cons", "eq"} {
		v, ok := env.Lookup(name)
		if !ok {
			t.Errorf("%s: not bound in global env", name)
			continue
		}
		if _, ok := v.(*value.Closure); !ok {
			t.Errorf("%s: bound value is %T, want *value.Closure", name, v)
		}
	}
}

// runWithLoadedCallee builds a tiny program that loads the named primitive
// from env and applies it to the supplied constants.
func runWithLoadedCallee(t *testing.T, name string, args []value.Value) value.Value {
	t.Helper()
	code := &Code{
		Syms: []string{name},
	}
	code.Instrs = append(code.Instrs, Instr{Op: OpLoadVar, Arg: 0})
	for _, arg := range args {
		code.Consts = append(code.Consts, arg)
		code.Instrs = append(code.Instrs, Instr{Op: OpLoadConst, Arg: len(code.Consts) - 1})
	}
	code.Instrs = append(code.Instrs,
		Instr{Op: OpCall, Arg: len(args)},
		Instr{Op: OpRet},
	)
	got, err := Run(code, NewGlobalEnv())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return got
}

func TestPrimitiveClosuresExecute(t *testing.T) {
	a := value.Symbol{Name: "a"}
	b := value.Symbol{Name: "b"}
	tSym := value.Symbol{Name: "t"}
	abc := value.List(a, b, value.Symbol{Name: "c"})

	t.Run("car", func(t *testing.T) {
		got := runWithLoadedCallee(t, "car", []value.Value{abc})
		if !value.Eq(got, a) {
			t.Errorf("car: got %v, want a", got)
		}
	})
	t.Run("cdr", func(t *testing.T) {
		got := runWithLoadedCallee(t, "cdr", []value.Value{abc})
		pair, ok := got.(*value.Pair)
		if !ok {
			t.Fatalf("cdr: got %T, want *Pair", got)
		}
		if !value.Eq(pair.Car, b) {
			t.Errorf("cdr: car = %v, want b", pair.Car)
		}
	})
	t.Run("atom of symbol", func(t *testing.T) {
		got := runWithLoadedCallee(t, "atom", []value.Value{a})
		if !value.Eq(got, tSym) {
			t.Errorf("atom: got %v, want t", got)
		}
	})
	t.Run("atom of pair", func(t *testing.T) {
		got := runWithLoadedCallee(t, "atom", []value.Value{abc})
		if !value.IsNil(got) {
			t.Errorf("atom of pair: got %v, want ()", got)
		}
	})
	t.Run("cons", func(t *testing.T) {
		got := runWithLoadedCallee(t, "cons", []value.Value{a, value.NIL})
		pair, ok := got.(*value.Pair)
		if !ok {
			t.Fatalf("cons: got %T, want *Pair", got)
		}
		if !value.Eq(pair.Car, a) || !value.IsNil(pair.Cdr) {
			t.Errorf("cons: got %v, want (a)", got)
		}
	})
	t.Run("eq same", func(t *testing.T) {
		got := runWithLoadedCallee(t, "eq", []value.Value{a, a})
		if !value.Eq(got, tSym) {
			t.Errorf("eq same: got %v, want t", got)
		}
	})
	t.Run("eq different", func(t *testing.T) {
		got := runWithLoadedCallee(t, "eq", []value.Value{a, b})
		if !value.IsNil(got) {
			t.Errorf("eq different: got %v, want ()", got)
		}
	})
}

func TestLoadPrimitiveAsValue(t *testing.T) {
	// Program: simply LOAD_VAR "car" and RET. Should yield a closure.
	code := &Code{
		Instrs: []Instr{
			{Op: OpLoadVar, Arg: 0},
			{Op: OpRet},
		},
		Syms: []string{"car"},
	}
	got, err := Run(code, NewGlobalEnv())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if _, ok := got.(*value.Closure); !ok {
		t.Errorf("got %T, want *value.Closure", got)
	}
}
