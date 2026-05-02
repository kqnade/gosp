package printer

import (
	"testing"

	"github.com/kqnade/gosp/value"
)

func TestPrintAtoms(t *testing.T) {
	tests := []struct {
		name string
		v    value.Value
		want string
	}{
		{name: "nil", v: value.NIL, want: "()"},
		{name: "symbol", v: value.Symbol{Name: "a"}, want: "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Print(tt.v)
			if got != tt.want {
				t.Fatalf("Print = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintProperList(t *testing.T) {
	got := Print(value.List(
		value.Symbol{Name: "a"},
		value.Symbol{Name: "b"},
		value.Symbol{Name: "c"},
	))
	if got != "(a b c)" {
		t.Fatalf("Print = %q, want %q", got, "(a b c)")
	}
}

func TestPrintDottedPair(t *testing.T) {
	got := Print(value.Cons(value.Symbol{Name: "a"}, value.Symbol{Name: "b"}))
	if got != "(a . b)" {
		t.Fatalf("Print = %q, want %q", got, "(a . b)")
	}
}

func TestPrintImproperList(t *testing.T) {
	got := Print(value.Cons(
		value.Symbol{Name: "a"},
		value.Cons(value.Symbol{Name: "b"}, value.Symbol{Name: "c"}),
	))
	if got != "(a b . c)" {
		t.Fatalf("Print = %q, want %q", got, "(a b . c)")
	}
}
