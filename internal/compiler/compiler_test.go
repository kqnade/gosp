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

func TestCompileQuoteSymbol(t *testing.T) {
	form := value.List(value.Symbol{Name: "quote"}, value.Symbol{Name: "a"})
	code, err := Compile(form)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	wantInstrs := []vm.Instr{{Op: vm.OpLoadConst, Arg: 0}, {Op: vm.OpRet}}
	if !equalInstrs(code.Instrs, wantInstrs) {
		t.Errorf("Instrs = %v, want %v", code.Instrs, wantInstrs)
	}
	if len(code.Consts) != 1 || !value.Eq(code.Consts[0], value.Symbol{Name: "a"}) {
		t.Errorf("Consts = %v, want [a]", code.Consts)
	}
	if len(code.Syms) != 0 {
		t.Errorf("Syms = %v, want []", code.Syms)
	}
}

func TestCompileQuoteList(t *testing.T) {
	quoted := value.List(value.Symbol{Name: "a"}, value.Symbol{Name: "b"})
	form := value.List(value.Symbol{Name: "quote"}, quoted)
	code, err := Compile(form)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	wantInstrs := []vm.Instr{{Op: vm.OpLoadConst, Arg: 0}, {Op: vm.OpRet}}
	if !equalInstrs(code.Instrs, wantInstrs) {
		t.Errorf("Instrs = %v, want %v", code.Instrs, wantInstrs)
	}
	if len(code.Consts) != 1 {
		t.Fatalf("Consts len = %d, want 1", len(code.Consts))
	}
	got, ok := code.Consts[0].(*value.Pair)
	if !ok {
		t.Fatalf("Consts[0] = %T, want *Pair", code.Consts[0])
	}
	if !value.Eq(got.Car, value.Symbol{Name: "a"}) {
		t.Errorf("Car = %v, want a", got.Car)
	}
}

func TestCompileQuoteWrongArity(t *testing.T) {
	cases := []struct {
		name string
		form value.Value
	}{
		{"no args", value.List(value.Symbol{Name: "quote"})},
		{"too many", value.List(value.Symbol{Name: "quote"}, value.Symbol{Name: "a"}, value.Symbol{Name: "b"})},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Compile(tc.form); err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}

func TestCompileAndRunCond(t *testing.T) {
	t_ := value.Symbol{Name: "t"}
	nilSym := value.Symbol{Name: "nil"}
	quote := value.Symbol{Name: "quote"}
	cond := value.Symbol{Name: "cond"}
	a := value.Symbol{Name: "a"}
	b := value.Symbol{Name: "b"}

	cases := []struct {
		name string
		form value.Value
		env  func() *value.Env
		want value.Value
	}{
		{
			name: "empty cond returns ()",
			form: value.List(cond),
			env:  func() *value.Env { return value.NewEnv(nil) },
			want: value.NIL,
		},
		{
			name: "single t clause",
			form: value.List(cond, value.List(t_, value.List(quote, a))),
			env:  func() *value.Env { return value.NewEnv(nil) },
			want: a,
		},
		{
			name: "first clause false, second wins",
			form: value.List(
				cond,
				value.List(nilSym, value.List(quote, a)),
				value.List(t_, value.List(quote, b)),
			),
			env:  func() *value.Env { return value.NewEnv(nil) },
			want: b,
		},
		{
			name: "no truthy clause returns ()",
			form: value.List(cond, value.List(nilSym, value.List(quote, a))),
			env:  func() *value.Env { return value.NewEnv(nil) },
			want: value.NIL,
		},
		{
			name: "bound symbol is truthy",
			form: value.List(
				cond,
				value.List(value.Symbol{Name: "x"}, value.List(quote, a)),
				value.List(t_, value.List(quote, b)),
			),
			env: func() *value.Env {
				e := value.NewEnv(nil)
				e.Define("x", value.Symbol{Name: "y"})
				return e
			},
			want: a,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, err := Compile(tc.form)
			if err != nil {
				t.Fatalf("Compile: %v", err)
			}
			got, err := vm.Run(code, tc.env())
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if !value.Eq(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCompileCondMalformed(t *testing.T) {
	cond := value.Symbol{Name: "cond"}
	a := value.Symbol{Name: "a"}

	cases := []struct {
		name string
		form value.Value
	}{
		{"non-list clause", value.List(cond, a)},
		{"missing body", value.List(cond, value.List(a))},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Compile(tc.form); err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}

func TestCompileAndRunAtom(t *testing.T) {
	cases := []struct {
		name string
		form value.Value
		env  func() *value.Env
		want value.Value
	}{
		{
			name: "t evaluates to symbol t",
			form: value.Symbol{Name: "t"},
			env:  func() *value.Env { return value.NewEnv(nil) },
			want: value.Symbol{Name: "t"},
		},
		{
			name: "nil symbol evaluates to ()",
			form: value.Symbol{Name: "nil"},
			env:  func() *value.Env { return value.NewEnv(nil) },
			want: value.NIL,
		},
		{
			name: "() literal evaluates to ()",
			form: value.NIL,
			env:  func() *value.Env { return value.NewEnv(nil) },
			want: value.NIL,
		},
		{
			name: "bound symbol resolves via env",
			form: value.Symbol{Name: "x"},
			env: func() *value.Env {
				e := value.NewEnv(nil)
				e.Define("x", value.Symbol{Name: "a"})
				return e
			},
			want: value.Symbol{Name: "a"},
		},
		{
			name: "quoted symbol",
			form: value.List(value.Symbol{Name: "quote"}, value.Symbol{Name: "a"}),
			env:  func() *value.Env { return value.NewEnv(nil) },
			want: value.Symbol{Name: "a"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, err := Compile(tc.form)
			if err != nil {
				t.Fatalf("Compile: %v", err)
			}
			got, err := vm.Run(code, tc.env())
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if !value.Eq(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
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
