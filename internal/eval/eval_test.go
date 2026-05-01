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
