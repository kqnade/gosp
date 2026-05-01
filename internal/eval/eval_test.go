package eval

import (
	"testing"

	"github.com/kqnade/gosp/internal/value"
)

func TestEvalT(t *testing.T) {
	env := value.NewEnv(nil)
	got, err := Eval(value.Symbol{Name: "t"}, env)
	if err != nil {
		t.Fatalf("Eval(t) error: %v", err)
	}
	sym, ok := got.(value.Symbol)
	if !ok || sym.Name != "t" {
		t.Fatalf("Eval(t) = %v, want symbol t", got)
	}
}

func TestEvalNil(t *testing.T) {
	env := value.NewEnv(nil)
	got, err := Eval(value.Symbol{Name: "nil"}, env)
	if err != nil {
		t.Fatalf("Eval(nil) error: %v", err)
	}
	if !value.IsNil(got) {
		t.Fatalf("Eval(nil) = %v, want NIL", got)
	}
}

func TestEvalNilLiteral(t *testing.T) {
	env := value.NewEnv(nil)
	got, err := Eval(value.NIL, env)
	if err != nil {
		t.Fatalf("Eval(()) error: %v", err)
	}
	if !value.IsNil(got) {
		t.Fatalf("Eval(()) = %v, want NIL", got)
	}
}

func TestEvalSymbolLookup(t *testing.T) {
	env := value.NewEnv(nil)
	env.Define("x", value.Symbol{Name: "a"})
	got, err := Eval(value.Symbol{Name: "x"}, env)
	if err != nil {
		t.Fatalf("Eval(x) error: %v", err)
	}
	sym, ok := got.(value.Symbol)
	if !ok || sym.Name != "a" {
		t.Fatalf("Eval(x) = %v, want symbol a", got)
	}
}

func TestEvalUnboundSymbol(t *testing.T) {
	env := value.NewEnv(nil)
	_, err := Eval(value.Symbol{Name: "missing"}, env)
	if err == nil {
		t.Fatalf("Eval(missing) expected error, got nil")
	}
}

func TestEvalQuoteSymbol(t *testing.T) {
	env := value.NewEnv(nil)
	form := value.List(value.Symbol{Name: "quote"}, value.Symbol{Name: "x"})
	got, err := Eval(form, env)
	if err != nil {
		t.Fatalf("Eval((quote x)) error: %v", err)
	}
	sym, ok := got.(value.Symbol)
	if !ok || sym.Name != "x" {
		t.Fatalf("Eval((quote x)) = %v, want symbol x", got)
	}
}

func TestEvalQuoteList(t *testing.T) {
	env := value.NewEnv(nil)
	quoted := value.List(value.Symbol{Name: "a"}, value.Symbol{Name: "b"})
	form := value.List(value.Symbol{Name: "quote"}, quoted)
	got, err := Eval(form, env)
	if err != nil {
		t.Fatalf("Eval((quote (a b))) error: %v", err)
	}
	if got != quoted {
		t.Fatalf("Eval((quote (a b))) = %v, want %v", got, quoted)
	}
}

func TestEvalQuoteArityError(t *testing.T) {
	env := value.NewEnv(nil)
	form := value.List(value.Symbol{Name: "quote"}, value.Symbol{Name: "a"}, value.Symbol{Name: "b"})
	if _, err := Eval(form, env); err == nil {
		t.Fatalf("Eval((quote a b)) expected error, got nil")
	}
	form2 := value.List(value.Symbol{Name: "quote"})
	if _, err := Eval(form2, env); err == nil {
		t.Fatalf("Eval((quote)) expected error, got nil")
	}
}

func TestEvalCondFirstTruthy(t *testing.T) {
	env := value.NewEnv(nil)
	// (cond (t (quote a)) (t (quote b)))
	clause1 := value.List(value.Symbol{Name: "t"}, value.List(value.Symbol{Name: "quote"}, value.Symbol{Name: "a"}))
	clause2 := value.List(value.Symbol{Name: "t"}, value.List(value.Symbol{Name: "quote"}, value.Symbol{Name: "b"}))
	form := value.List(value.Symbol{Name: "cond"}, clause1, clause2)
	got, err := Eval(form, env)
	if err != nil {
		t.Fatalf("Eval(cond) error: %v", err)
	}
	sym, ok := got.(value.Symbol)
	if !ok || sym.Name != "a" {
		t.Fatalf("Eval(cond) = %v, want symbol a", got)
	}
}

func TestEvalCondSkipsFalseClauses(t *testing.T) {
	env := value.NewEnv(nil)
	// (cond (nil (quote skip)) (t (quote ok)))
	clause1 := value.List(value.Symbol{Name: "nil"}, value.List(value.Symbol{Name: "quote"}, value.Symbol{Name: "skip"}))
	clause2 := value.List(value.Symbol{Name: "t"}, value.List(value.Symbol{Name: "quote"}, value.Symbol{Name: "ok"}))
	form := value.List(value.Symbol{Name: "cond"}, clause1, clause2)
	got, err := Eval(form, env)
	if err != nil {
		t.Fatalf("Eval(cond) error: %v", err)
	}
	sym, ok := got.(value.Symbol)
	if !ok || sym.Name != "ok" {
		t.Fatalf("Eval(cond) = %v, want symbol ok", got)
	}
}

func TestEvalCondNoMatch(t *testing.T) {
	env := value.NewEnv(nil)
	// (cond (nil (quote a)))
	clause := value.List(value.Symbol{Name: "nil"}, value.List(value.Symbol{Name: "quote"}, value.Symbol{Name: "a"}))
	form := value.List(value.Symbol{Name: "cond"}, clause)
	got, err := Eval(form, env)
	if err != nil {
		t.Fatalf("Eval(cond) error: %v", err)
	}
	if !value.IsNil(got) {
		t.Fatalf("Eval(cond) = %v, want NIL", got)
	}
}

func TestEvalCondEmpty(t *testing.T) {
	env := value.NewEnv(nil)
	form := value.List(value.Symbol{Name: "cond"})
	got, err := Eval(form, env)
	if err != nil {
		t.Fatalf("Eval((cond)) error: %v", err)
	}
	if !value.IsNil(got) {
		t.Fatalf("Eval((cond)) = %v, want NIL", got)
	}
}
