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

func TestCompileAndRunLabel(t *testing.T) {
	q := value.Symbol{Name: "quote"}
	lambda := value.Symbol{Name: "lambda"}
	label := value.Symbol{Name: "label"}
	cond := value.Symbol{Name: "cond"}
	eq := value.Symbol{Name: "eq"}
	cdr := value.Symbol{Name: "cdr"}
	drop := value.Symbol{Name: "drop"}
	xs := value.Symbol{Name: "xs"}
	x := value.Symbol{Name: "x"}
	a := value.Symbol{Name: "a"}
	b := value.Symbol{Name: "b"}
	c := value.Symbol{Name: "c"}
	done := value.Symbol{Name: "done"}
	okSym := value.Symbol{Name: "ok"}
	tSym := value.Symbol{Name: "t"}
	ff := value.Symbol{Name: "ff"}

	t.Run("label of identity is callable", func(t *testing.T) {
		// ((label F (lambda (x) x)) 'a) → a
		form := value.List(
			value.List(label, value.Symbol{Name: "F"}, value.List(lambda, value.List(x), x)),
			value.List(q, a),
		)
		code, err := Compile(form)
		if err != nil {
			t.Fatalf("Compile: %v", err)
		}
		got, err := vm.Run(code, value.NewEnv(nil))
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if !value.Eq(got, a) {
			t.Errorf("got %v, want a", got)
		}
	})

	t.Run("recursive drop via label", func(t *testing.T) {
		// ((label drop (lambda (xs) (cond ((eq xs 'done) 'ok) (t (drop (cdr xs)))))) '(a b c done))
		body := value.List(
			cond,
			value.List(value.List(eq, xs, value.List(q, done)), value.List(q, okSym)),
			value.List(tSym, value.List(drop, value.List(cdr, xs))),
		)
		labFn := value.List(label, drop, value.List(lambda, value.List(xs), body))
		// Build (a . (b . (c . done))) so traversal terminates at `done`.
		listArg := value.Cons(a, value.Cons(b, value.Cons(c, done)))
		form := value.List(labFn, value.List(q, listArg))
		code, err := Compile(form)
		if err != nil {
			t.Fatalf("Compile: %v", err)
		}
		got, err := vm.Run(code, value.NewEnv(nil))
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if !value.Eq(got, okSym) {
			t.Errorf("got %v, want ok", got)
		}
	})

	t.Run("ff walker visits leaves", func(t *testing.T) {
		// (label ff (lambda (x) (cond ((atom x) x) (t (ff (car x))))))
		// applied to ((a b) c) → a (descend leftmost)
		body := value.List(
			cond,
			value.List(value.List(value.Symbol{Name: "atom"}, x), x),
			value.List(tSym, value.List(ff, value.List(value.Symbol{Name: "car"}, x))),
		)
		labFn := value.List(label, ff, value.List(lambda, value.List(x), body))
		form := value.List(labFn, value.List(q, value.List(value.List(a, b), c)))
		code, err := Compile(form)
		if err != nil {
			t.Fatalf("Compile: %v", err)
		}
		got, err := vm.Run(code, value.NewEnv(nil))
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if !value.Eq(got, a) {
			t.Errorf("got %v, want a", got)
		}
	})

	t.Run("label expression that is not a function fails at runtime", func(t *testing.T) {
		form := value.List(label, value.Symbol{Name: "F"}, value.List(q, a))
		code, err := Compile(form)
		if err != nil {
			t.Fatalf("Compile: %v", err)
		}
		if _, err := vm.Run(code, value.NewEnv(nil)); err == nil {
			t.Fatal("expected runtime error for non-function label, got nil")
		}
	})

	t.Run("label malformed", func(t *testing.T) {
		cases := []struct {
			name string
			form value.Value
		}{
			{"missing name", value.List(label)},
			{"name not symbol", value.List(label, value.List(q, a), value.List(q, a))},
			{"too many", value.List(label, value.Symbol{Name: "F"}, value.List(q, a), value.List(q, a))},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				if _, err := Compile(tc.form); err == nil {
					t.Errorf("expected compile error, got nil")
				}
			})
		}
	})
}

func TestCompileProgramTopLevelLabel(t *testing.T) {
	q := value.Symbol{Name: "quote"}
	lambda := value.Symbol{Name: "lambda"}
	label := value.Symbol{Name: "label"}
	cadr := value.Symbol{Name: "cadr"}
	x := value.Symbol{Name: "x"}
	a := value.Symbol{Name: "a"}
	b := value.Symbol{Name: "b"}
	c := value.Symbol{Name: "c"}

	t.Run("(label cadr ...) then (cadr '(a b c)) returns b", func(t *testing.T) {
		labelForm := value.List(
			label, cadr,
			value.List(
				lambda, value.List(x),
				value.List(value.Symbol{Name: "car"}, value.List(value.Symbol{Name: "cdr"}, x)),
			),
		)
		callForm := value.List(cadr, value.List(q, value.List(a, b, c)))
		code, err := CompileProgram([]value.Value{labelForm, callForm})
		if err != nil {
			t.Fatalf("CompileProgram: %v", err)
		}
		got, err := vm.Run(code, value.NewEnv(nil))
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if !value.Eq(got, b) {
			t.Errorf("got %v, want b", got)
		}
	})

	t.Run("single top-level label returns the bound value", func(t *testing.T) {
		form := value.List(
			label,
			value.Symbol{Name: "ident"},
			value.List(lambda, value.List(x), x),
		)
		code, err := CompileProgram([]value.Value{form})
		if err != nil {
			t.Fatalf("CompileProgram: %v", err)
		}
		env := value.NewEnv(nil)
		got, err := vm.Run(code, env)
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if _, ok := got.(*value.Closure); !ok {
			t.Errorf("got %T, want *Closure", got)
		}
		if _, ok := env.Lookup("ident"); !ok {
			t.Errorf("ident not bound in global env")
		}
	})

	t.Run("empty program returns ()", func(t *testing.T) {
		code, err := CompileProgram(nil)
		if err != nil {
			t.Fatalf("CompileProgram: %v", err)
		}
		got, err := vm.Run(code, value.NewEnv(nil))
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if !value.IsNil(got) {
			t.Errorf("got %v, want ()", got)
		}
	})

	t.Run("multiple plain forms — last value wins", func(t *testing.T) {
		forms := []value.Value{
			value.List(q, a),
			value.List(q, b),
			value.List(q, c),
		}
		code, err := CompileProgram(forms)
		if err != nil {
			t.Fatalf("CompileProgram: %v", err)
		}
		got, err := vm.Run(code, value.NewEnv(nil))
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if !value.Eq(got, c) {
			t.Errorf("got %v, want c", got)
		}
	})

	t.Run("inner-expression label still works inside a top-level form", func(t *testing.T) {
		// Apply an in-expr label closure.
		form := value.List(
			value.List(label, value.Symbol{Name: "F"}, value.List(lambda, value.List(x), x)),
			value.List(q, a),
		)
		code, err := CompileProgram([]value.Value{form})
		if err != nil {
			t.Fatalf("CompileProgram: %v", err)
		}
		got, err := vm.Run(code, value.NewEnv(nil))
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if !value.Eq(got, a) {
			t.Errorf("got %v, want a", got)
		}
	})
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

func TestTopLevelRebindingPrimitive(t *testing.T) {
	q := value.Symbol{Name: "quote"}
	lambda := value.Symbol{Name: "lambda"}
	label := value.Symbol{Name: "label"}
	cons := value.Symbol{Name: "cons"}
	x := value.Symbol{Name: "x"}
	a := value.Symbol{Name: "a"}

	t.Run("(label car ...) makes later (car ...) use the rebound car", func(t *testing.T) {
		// Forms:
		//   (label car (lambda (x) (cons x '())))
		//   (car 'a)
		// Expected: (a). With the bug, (car 'a) is OpCar applied to symbol
		// 'a — a runtime error.
		car := value.Symbol{Name: "car"}
		labelForm := value.List(
			label, car,
			value.List(lambda, value.List(x), value.List(cons, x, value.List(q, value.NIL))),
		)
		callForm := value.List(car, value.List(q, a))

		code, err := CompileProgram([]value.Value{labelForm, callForm})
		if err != nil {
			t.Fatalf("CompileProgram: %v", err)
		}
		// The top-level call to `car` after the rebinding must NOT lower
		// to OpCar; it should emit a regular CALL through env lookup.
		if countOp(code.Instrs, vm.OpCar) != 0 {
			t.Errorf("rebound (car ...) at top level must not lower to OpCar; got %v", code.Instrs)
		}

		got, err := vm.Run(code, vm.NewGlobalEnv())
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		want := value.List(a)
		if !valueEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("rebinding propagates into a later lambda body", func(t *testing.T) {
		// Forms:
		//   (label cdr (lambda (xs) xs))   ; rebinds cdr to identity
		//   (label foo (lambda (xs) (cdr xs)))
		//   (foo '(a b))
		// Expected: (a b). foo's body is compiled AFTER cdr is rebound,
		// so (cdr xs) inside foo must NOT lower to OpCdr.
		cdr := value.Symbol{Name: "cdr"}
		foo := value.Symbol{Name: "foo"}
		xs := value.Symbol{Name: "xs"}
		b := value.Symbol{Name: "b"}

		labelCdr := value.List(label, cdr, value.List(lambda, value.List(xs), xs))
		labelFoo := value.List(label, foo, value.List(lambda, value.List(xs), value.List(cdr, xs)))
		callFoo := value.List(foo, value.List(q, value.List(a, b)))

		code, err := CompileProgram([]value.Value{labelCdr, labelFoo, callFoo})
		if err != nil {
			t.Fatalf("CompileProgram: %v", err)
		}
		// foo is the second FuncProto registered (cdr's lambda is the
		// first). Both must be free of OpCdr — cdr's because its body
		// uses no cdr; foo's because cdr is rebound at compile time.
		for i, fp := range code.Funcs {
			if countOp(fp.Code.Instrs, vm.OpCdr) != 0 {
				t.Errorf("FuncProto[%d] must not lower (cdr ...) to OpCdr after rebinding: %v",
					i, fp.Code.Instrs)
			}
		}

		got, err := vm.Run(code, vm.NewGlobalEnv())
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		want := value.List(a, b)
		if !valueEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("compile-time arity check is preserved for unrebound primitives", func(t *testing.T) {
		// (car) at top level (no prior rebinding) must still be a
		// compile-time error. Strategy B retains the optimization for
		// the unrebound case so arity is validated at compile time.
		form := value.List(value.Symbol{Name: "car"})
		if _, err := CompileProgram([]value.Value{form}); err == nil {
			t.Error("expected compile-time arity error for (car), got nil")
		}
	})

	t.Run("after rebinding, primitive arity is no longer compile-time enforced", func(t *testing.T) {
		// Once `car` is rebound, the compiler can no longer assume the
		// primitive's arity. (car a b) must compile cleanly and dispatch
		// to the rebound function at runtime.
		car := value.Symbol{Name: "car"}
		labelForm := value.List(
			label, car,
			value.List(lambda, value.List(x, value.Symbol{Name: "y"}), x),
		)
		callForm := value.List(car, value.List(q, a), value.List(q, value.Symbol{Name: "b"}))
		code, err := CompileProgram([]value.Value{labelForm, callForm})
		if err != nil {
			t.Fatalf("CompileProgram: %v", err)
		}
		got, err := vm.Run(code, vm.NewGlobalEnv())
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if !value.Eq(got, a) {
			t.Errorf("got %v, want a", got)
		}
	})
}

func TestLocalShadowingPrimitives(t *testing.T) {
	q := value.Symbol{Name: "quote"}
	lambda := value.Symbol{Name: "lambda"}
	cond := value.Symbol{Name: "cond"}
	tSym := value.Symbol{Name: "t"}
	a := value.Symbol{Name: "a"}
	b := value.Symbol{Name: "b"}
	x := value.Symbol{Name: "x"}
	y := value.Symbol{Name: "y"}
	z := value.Symbol{Name: "z"}

	t.Run("lambda parameter shadows car", func(t *testing.T) {
		// ((lambda (car) (car 'x))
		//  (lambda (z) (cons z '())))
		car := value.Symbol{Name: "car"}
		inner := value.List(
			lambda, value.List(z),
			value.List(value.Symbol{Name: "cons"}, z, value.List(q, value.NIL)),
		)
		outer := value.List(
			lambda, value.List(car),
			value.List(car, value.List(q, x)),
		)
		form := value.List(outer, inner)

		code, err := Compile(form)
		if err != nil {
			t.Fatalf("Compile: %v", err)
		}
		// The shadowed call must NOT lower to OpCar; the body must use
		// CALL or TAIL_CALL via env lookup.
		if len(code.Funcs) < 1 {
			t.Fatalf("expected at least one FuncProto")
		}
		outerBody := code.Funcs[0].Code.Instrs
		if countOp(outerBody, vm.OpCar) != 0 {
			t.Errorf("shadowed (car ...) must not lower to OpCar; got body %v", outerBody)
		}
		if countOp(outerBody, vm.OpCall)+countOp(outerBody, vm.OpTailCall) != 1 {
			t.Errorf("shadowed (car ...) must compile to CALL/TAIL_CALL; got %v", outerBody)
		}

		got, err := vm.Run(code, value.NewEnv(nil))
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		want := value.List(x)
		if !valueEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("lambda parameter shadows each primitive", func(t *testing.T) {
		// For each primitive name P, build:
		//   ((lambda (P) (P 'a [arg2])) (lambda (...) BODY))
		// where BODY produces a witness symbol (`shadow`), and assert
		// the witness flows out instead of the primitive being applied.
		shadow := value.Symbol{Name: "shadow"}
		quoteShadow := value.List(q, shadow)
		cases := []struct {
			name   string
			prim   string
			arity  int
			callee value.Value // (lambda (...) shadow)
		}{
			{"shadow car", "car", 1, value.List(lambda, value.List(x), quoteShadow)},
			{"shadow cdr", "cdr", 1, value.List(lambda, value.List(x), quoteShadow)},
			{"shadow atom", "atom", 1, value.List(lambda, value.List(x), quoteShadow)},
			{"shadow cons", "cons", 2, value.List(lambda, value.List(x, y), quoteShadow)},
			{"shadow eq", "eq", 2, value.List(lambda, value.List(x, y), quoteShadow)},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				p := value.Symbol{Name: tc.prim}
				var call value.Value
				if tc.arity == 1 {
					call = value.List(p, value.List(q, a))
				} else {
					call = value.List(p, value.List(q, a), value.List(q, b))
				}
				outer := value.List(lambda, value.List(p), call)
				form := value.List(outer, tc.callee)

				code, err := Compile(form)
				if err != nil {
					t.Fatalf("Compile: %v", err)
				}
				got, err := vm.Run(code, value.NewEnv(nil))
				if err != nil {
					t.Fatalf("Run: %v", err)
				}
				if !value.Eq(got, shadow) {
					t.Errorf("%s: got %v, want shadow", tc.prim, got)
				}
			})
		}
	})

	t.Run("label name shadows primitive", func(t *testing.T) {
		// (label car (lambda (n) (cond ((eq n '()) 'done) (t (car (cdr n))))))
		// applied to (a) should recurse via the labeled `car` and reach 'done.
		// The body's `(car (cdr n))` must NOT lower to OpCar.
		car := value.Symbol{Name: "car"}
		n := value.Symbol{Name: "n"}
		cdr := value.Symbol{Name: "cdr"}
		eq := value.Symbol{Name: "eq"}
		done := value.Symbol{Name: "done"}
		body := value.List(
			cond,
			value.List(value.List(eq, n, value.List(q, value.NIL)), value.List(q, done)),
			value.List(tSym, value.List(car, value.List(cdr, n))),
		)
		labFn := value.List(value.Symbol{Name: "label"}, car, value.List(lambda, value.List(n), body))
		form := value.List(labFn, value.List(q, value.List(a)))

		code, err := Compile(form)
		if err != nil {
			t.Fatalf("Compile: %v", err)
		}
		// FuncProto[0] is the lambda body. The shadowed (car (cdr n))
		// must not lower to OpCar. (cdr n) is the only OpCdr.
		if len(code.Funcs) < 1 {
			t.Fatalf("expected at least one FuncProto")
		}
		bodyInstrs := code.Funcs[0].Code.Instrs
		if countOp(bodyInstrs, vm.OpCar) != 0 {
			t.Errorf("label-shadowed (car ...) must not lower to OpCar; got %v", bodyInstrs)
		}

		got, err := vm.Run(code, value.NewEnv(nil))
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if !value.Eq(got, done) {
			t.Errorf("got %v, want done", got)
		}
	})

	t.Run("outer lambda binding shadows primitive in inner lambda body", func(t *testing.T) {
		// ((lambda (car)
		//    ((lambda (x) (car x)) (quote a)))
		//  (lambda (z) (cons z '())))
		// Inner `(car x)` is shadowed by the outer parameter `car`.
		car := value.Symbol{Name: "car"}
		inner := value.List(
			lambda, value.List(z),
			value.List(value.Symbol{Name: "cons"}, z, value.List(q, value.NIL)),
		)
		nested := value.List(
			lambda, value.List(car),
			value.List(
				value.List(lambda, value.List(x), value.List(car, x)),
				value.List(q, a),
			),
		)
		form := value.List(nested, inner)

		code, err := Compile(form)
		if err != nil {
			t.Fatalf("Compile: %v", err)
		}
		// Inner lambda is the second FuncProto registered in the outer
		// lambda's body (callee is the first).
		// Check that none of the FuncProtos use OpCar.
		for i, fp := range code.Funcs {
			if countOp(fp.Code.Instrs, vm.OpCar) != 0 {
				t.Errorf("FuncProto[%d] must not use OpCar (shadowed): %v", i, fp.Code.Instrs)
			}
		}

		got, err := vm.Run(code, value.NewEnv(nil))
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		want := value.List(a)
		if !valueEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
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
