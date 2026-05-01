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

func TestVMTailCallBounded(t *testing.T) {
	const n = 1000

	xs := value.Symbol{Name: "xs"}
	loop := value.Symbol{Name: "loop"}
	self_ := value.Symbol{Name: "self"}
	q := value.Symbol{Name: "quote"}
	done := value.Symbol{Name: "done"}
	tSym := value.Symbol{Name: "t"}
	cond := value.Symbol{Name: "cond"}
	eq := value.Symbol{Name: "eq"}
	cdr := value.Symbol{Name: "cdr"}
	lambda := value.Symbol{Name: "lambda"}
	okSym := value.Symbol{Name: "ok"}

	list := value.Value(done)
	for range n {
		list = value.Cons(value.Symbol{Name: "x"}, list)
	}

	body := value.List(
		cond,
		value.List(value.List(eq, xs, value.List(q, done)), value.List(q, okSym)),
		value.List(tSym, value.List(loop, loop, value.List(cdr, xs))),
	)
	inner := value.List(lambda, value.List(loop, xs), body)
	outer := value.List(lambda, value.List(self_), value.List(self_, self_, value.List(q, list)))
	form := value.List(outer, inner)

	code, err := Compile(form)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	got, max, err := vm.RunWithStats(code, value.NewEnv(nil))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !value.Eq(got, okSym) {
		t.Errorf("got %v, want ok", got)
	}
	if max > 16 {
		t.Errorf("max frame depth = %d for %d tail-iterations, want bounded by ~16", max, n)
	}
}

func TestTailPositionEmission(t *testing.T) {
	q := value.Symbol{Name: "quote"}
	lambda := value.Symbol{Name: "lambda"}
	cond := value.Symbol{Name: "cond"}
	tSym := value.Symbol{Name: "t"}
	x := value.Symbol{Name: "x"}
	f := value.Symbol{Name: "f"}
	a := value.Symbol{Name: "a"}

	t.Run("tail call in lambda body emits TAIL_CALL", func(t *testing.T) {
		// (lambda (x) (f x))
		form := value.List(lambda, value.List(x), value.List(f, x))
		code := compileForTest(t, form)
		// Top level: MAKE_CLOSURE + RET; no CALL/TAIL_CALL.
		if countOp(code.Instrs, vm.OpTailCall) != 0 {
			t.Errorf("top level should not emit TAIL_CALL")
		}
		if countOp(code.Instrs, vm.OpCall) != 0 {
			t.Errorf("top level should not emit CALL")
		}
		if len(code.Funcs) != 1 {
			t.Fatalf("want 1 FuncProto, got %d", len(code.Funcs))
		}
		body := code.Funcs[0].Code.Instrs
		if countOp(body, vm.OpTailCall) != 1 {
			t.Errorf("lambda body should emit exactly one TAIL_CALL, got %d", countOp(body, vm.OpTailCall))
		}
		if countOp(body, vm.OpCall) != 0 {
			t.Errorf("lambda body should not emit CALL, got %d", countOp(body, vm.OpCall))
		}
	})

	t.Run("non-tail call in lambda body emits CALL", func(t *testing.T) {
		// (lambda (x) (cons (f x) 'a)) — (f x) is the first arg of cons, not tail.
		form := value.List(
			lambda,
			value.List(x),
			value.List(value.Symbol{Name: "cons"}, value.List(f, x), value.List(q, a)),
		)
		code := compileForTest(t, form)
		body := code.Funcs[0].Code.Instrs
		if countOp(body, vm.OpCall) != 1 {
			t.Errorf("non-tail (f x) should emit CALL, got %d", countOp(body, vm.OpCall))
		}
		if countOp(body, vm.OpTailCall) != 0 {
			t.Errorf("non-tail (f x) should not emit TAIL_CALL, got %d", countOp(body, vm.OpTailCall))
		}
	})

	t.Run("cond clause body inherits tail position", func(t *testing.T) {
		// (lambda (x) (cond ((eq x 'a) (f x)) (t 'a)))
		form := value.List(
			lambda,
			value.List(x),
			value.List(
				cond,
				value.List(
					value.List(value.Symbol{Name: "eq"}, x, value.List(q, a)),
					value.List(f, x),
				),
				value.List(tSym, value.List(q, a)),
			),
		)
		code := compileForTest(t, form)
		body := code.Funcs[0].Code.Instrs
		if countOp(body, vm.OpTailCall) != 1 {
			t.Errorf("clause body call should emit TAIL_CALL, got %d", countOp(body, vm.OpTailCall))
		}
		if countOp(body, vm.OpCall) != 0 {
			t.Errorf("no CALL expected, got %d", countOp(body, vm.OpCall))
		}
	})

	t.Run("cond predicate is not tail", func(t *testing.T) {
		// (lambda (x) (cond ((f x) 'a) (t 'a))) — (f x) is the predicate, not tail.
		form := value.List(
			lambda,
			value.List(x),
			value.List(
				cond,
				value.List(value.List(f, x), value.List(q, a)),
				value.List(tSym, value.List(q, a)),
			),
		)
		code := compileForTest(t, form)
		body := code.Funcs[0].Code.Instrs
		if countOp(body, vm.OpCall) != 1 {
			t.Errorf("predicate (f x) should emit CALL, got %d", countOp(body, vm.OpCall))
		}
		if countOp(body, vm.OpTailCall) != 0 {
			t.Errorf("predicate (f x) should not emit TAIL_CALL, got %d", countOp(body, vm.OpTailCall))
		}
	})

	t.Run("top-level call emits CALL not TAIL_CALL", func(t *testing.T) {
		// ((lambda (x) x) 'a) — outer call is at top level.
		form := value.List(
			value.List(lambda, value.List(x), x),
			value.List(q, a),
		)
		code := compileForTest(t, form)
		if countOp(code.Instrs, vm.OpCall) != 1 {
			t.Errorf("top-level call should emit CALL, got %d", countOp(code.Instrs, vm.OpCall))
		}
		if countOp(code.Instrs, vm.OpTailCall) != 0 {
			t.Errorf("top-level call should not emit TAIL_CALL, got %d", countOp(code.Instrs, vm.OpTailCall))
		}
	})
}

func compileForTest(t *testing.T, form value.Value) *vm.Code {
	t.Helper()
	code, err := Compile(form)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	return code
}

func countOp(instrs []vm.Instr, op vm.Opcode) int {
	n := 0
	for _, ins := range instrs {
		if ins.Op == op {
			n++
		}
	}
	return n
}

func TestCompileAndRunLambdaApply(t *testing.T) {
	q := value.Symbol{Name: "quote"}
	lambda := value.Symbol{Name: "lambda"}
	a := value.Symbol{Name: "a"}
	b := value.Symbol{Name: "b"}
	x := value.Symbol{Name: "x"}
	y := value.Symbol{Name: "y"}

	cases := []struct {
		name string
		form value.Value
		env  func() *value.Env
		want value.Value
	}{
		{
			name: "((lambda (x) x) 'a) — identity",
			form: value.List(
				value.List(lambda, value.List(x), x),
				value.List(q, a),
			),
			env:  func() *value.Env { return value.NewEnv(nil) },
			want: a,
		},
		{
			name: "((lambda () (quote ok))) — nullary",
			form: value.List(
				value.List(lambda, value.NIL, value.List(q, value.Symbol{Name: "ok"})),
			),
			env:  func() *value.Env { return value.NewEnv(nil) },
			want: value.Symbol{Name: "ok"},
		},
		{
			name: "two-arg lambda over cons",
			form: value.List(
				value.List(
					lambda,
					value.List(x, y),
					value.List(value.Symbol{Name: "cons"}, x, y),
				),
				value.List(q, a),
				value.List(q, b),
			),
			env:  func() *value.Env { return value.NewEnv(nil) },
			want: value.Cons(a, b),
		},
		{
			name: "lambda bound in env then called",
			form: value.List(
				value.Symbol{Name: "f"},
				value.List(q, a),
			),
			env: func() *value.Env {
				e := value.NewEnv(nil)
				code, err := Compile(value.List(lambda, value.List(x), x))
				if err != nil {
					t.Fatalf("Compile lambda: %v", err)
				}
				v, err := vm.Run(code, e)
				if err != nil {
					t.Fatalf("Run lambda: %v", err)
				}
				e.Define("f", v)
				return e
			},
			want: a,
		},
		{
			name: "lexical capture: ((lambda (x) ((lambda (y) x) 'b)) 'a)",
			form: value.List(
				value.List(
					lambda,
					value.List(x),
					value.List(
						value.List(lambda, value.List(y), x),
						value.List(q, b),
					),
				),
				value.List(q, a),
			),
			env:  func() *value.Env { return value.NewEnv(nil) },
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
			if !valueEqual(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCompileLambdaMalformed(t *testing.T) {
	lambda := value.Symbol{Name: "lambda"}
	x := value.Symbol{Name: "x"}

	cases := []struct {
		name string
		form value.Value
	}{
		{"missing param list", value.List(lambda)},
		{"missing body", value.List(lambda, value.List(x))},
		{"non-symbol param", value.List(lambda, value.List(value.List(x)), x)},
		{"too many bodies", value.List(lambda, value.List(x), x, x)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Compile(tc.form); err == nil {
				t.Errorf("expected compile error, got nil")
			}
		})
	}
}

func TestRunWrongArity(t *testing.T) {
	lambda := value.Symbol{Name: "lambda"}
	x := value.Symbol{Name: "x"}
	q := value.Symbol{Name: "quote"}
	a := value.Symbol{Name: "a"}
	b := value.Symbol{Name: "b"}

	form := value.List(
		value.List(lambda, value.List(x), x),
		value.List(q, a),
		value.List(q, b),
	)
	code, err := Compile(form)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if _, err := vm.Run(code, value.NewEnv(nil)); err == nil {
		t.Fatal("expected runtime error for wrong arity, got nil")
	}
}

func TestCompileAndRunAtomEq(t *testing.T) {
	q := value.Symbol{Name: "quote"}
	a := value.Symbol{Name: "a"}
	b := value.Symbol{Name: "b"}
	tSym := value.Symbol{Name: "t"}

	cases := []struct {
		name string
		form value.Value
		want value.Value
	}{
		{
			name: "atom of symbol",
			form: value.List(value.Symbol{Name: "atom"}, value.List(q, a)),
			want: tSym,
		},
		{
			name: "atom of () is t",
			form: value.List(value.Symbol{Name: "atom"}, value.List(q, value.NIL)),
			want: tSym,
		},
		{
			name: "atom of pair is ()",
			form: value.List(value.Symbol{Name: "atom"}, value.List(q, value.List(a, b))),
			want: value.NIL,
		},
		{
			name: "eq same symbol",
			form: value.List(value.Symbol{Name: "eq"}, value.List(q, a), value.List(q, a)),
			want: tSym,
		},
		{
			name: "eq different symbols",
			form: value.List(value.Symbol{Name: "eq"}, value.List(q, a), value.List(q, b)),
			want: value.NIL,
		},
		{
			name: "eq nil and nil",
			form: value.List(value.Symbol{Name: "eq"}, value.List(q, value.NIL), value.List(q, value.NIL)),
			want: tSym,
		},
		{
			name: "eq pair and pair is ()",
			form: value.List(
				value.Symbol{Name: "eq"},
				value.List(q, value.List(a)),
				value.List(q, value.List(a)),
			),
			want: value.NIL,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, err := Compile(tc.form)
			if err != nil {
				t.Fatalf("Compile: %v", err)
			}
			got, err := vm.Run(code, value.NewEnv(nil))
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if !value.Eq(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCompileAtomEqArity(t *testing.T) {
	q := value.Symbol{Name: "quote"}
	a := value.Symbol{Name: "a"}
	cases := []struct {
		name string
		form value.Value
	}{
		{"atom no args", value.List(value.Symbol{Name: "atom"})},
		{"atom too many", value.List(value.Symbol{Name: "atom"}, value.List(q, a), value.List(q, a))},
		{"eq one arg", value.List(value.Symbol{Name: "eq"}, value.List(q, a))},
		{"eq three args", value.List(value.Symbol{Name: "eq"}, value.List(q, a), value.List(q, a), value.List(q, a))},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Compile(tc.form); err == nil {
				t.Errorf("expected compile error, got nil")
			}
		})
	}
}

func TestCompileAndRunCarCdrCons(t *testing.T) {
	q := value.Symbol{Name: "quote"}
	a := value.Symbol{Name: "a"}
	b := value.Symbol{Name: "b"}
	c := value.Symbol{Name: "c"}
	abc := value.List(a, b, c)

	cases := []struct {
		name string
		form value.Value
		want value.Value
	}{
		{
			name: "car of (a b c)",
			form: value.List(value.Symbol{Name: "car"}, value.List(q, abc)),
			want: a,
		},
		{
			name: "cdr of (a b c) is (b c)",
			form: value.List(value.Symbol{Name: "cdr"}, value.List(q, abc)),
			want: value.List(b, c),
		},
		{
			name: "cons a onto (b c)",
			form: value.List(
				value.Symbol{Name: "cons"},
				value.List(q, a),
				value.List(q, value.List(b, c)),
			),
			want: value.List(a, b, c),
		},
		{
			name: "car of cdr — composition",
			form: value.List(
				value.Symbol{Name: "car"},
				value.List(value.Symbol{Name: "cdr"}, value.List(q, abc)),
			),
			want: b,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, err := Compile(tc.form)
			if err != nil {
				t.Fatalf("Compile: %v", err)
			}
			got, err := vm.Run(code, value.NewEnv(nil))
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if !valueEqual(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCompileCarCdrConsErrors(t *testing.T) {
	q := value.Symbol{Name: "quote"}
	a := value.Symbol{Name: "a"}

	cases := []struct {
		name string
		form value.Value
	}{
		{"car wrong arity", value.List(value.Symbol{Name: "car"})},
		{"car too many", value.List(value.Symbol{Name: "car"}, value.List(q, a), value.List(q, a))},
		{"cdr wrong arity", value.List(value.Symbol{Name: "cdr"})},
		{"cons one arg", value.List(value.Symbol{Name: "cons"}, value.List(q, a))},
		{"cons three args", value.List(value.Symbol{Name: "cons"}, value.List(q, a), value.List(q, a), value.List(q, a))},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Compile(tc.form); err == nil {
				t.Errorf("expected compile error, got nil")
			}
		})
	}
}

func TestRunCarOfAtomFails(t *testing.T) {
	form := value.List(value.Symbol{Name: "car"}, value.List(value.Symbol{Name: "quote"}, value.Symbol{Name: "a"}))
	code, err := Compile(form)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if _, err := vm.Run(code, value.NewEnv(nil)); err == nil {
		t.Fatal("expected runtime error for car of symbol, got nil")
	}
}

// valueEqual compares two values for structural equality (symbols + nested pairs).
func valueEqual(a, b value.Value) bool {
	if value.IsAtom(a) || value.IsAtom(b) {
		return value.Eq(a, b)
	}
	pa, oka := a.(*value.Pair)
	pb, okb := b.(*value.Pair)
	if !oka || !okb {
		return false
	}
	return valueEqual(pa.Car, pb.Car) && valueEqual(pa.Cdr, pb.Cdr)
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
