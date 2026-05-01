package reader

import (
	"testing"

	"github.com/kqnade/gosp/internal/value"
)

func TestParseAtom(t *testing.T) {
	got, err := Parse([]Token{{Kind: TokenSymbol, Lexeme: "a"}})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	want := value.Symbol{Name: "a"}
	if got != want {
		t.Fatalf("Parse = %#v, want %#v", got, want)
	}
}

func TestParseEmptyList(t *testing.T) {
	tokens, err := Tokenize("()")
	if err != nil {
		t.Fatalf("Tokenize returned error: %v", err)
	}
	got, err := Parse(tokens)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if !value.IsNil(got) {
		t.Fatalf("Parse = %#v, want NIL", got)
	}
}

func TestParseList(t *testing.T) {
	tokens, err := Tokenize("(a b)")
	if err != nil {
		t.Fatalf("Tokenize returned error: %v", err)
	}
	got, err := Parse(tokens)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	want := value.List(value.Symbol{Name: "a"}, value.Symbol{Name: "b"})
	if !valuesEqual(got, want) {
		t.Fatalf("Parse = %#v, want %#v", got, want)
	}
}

func TestParseNestedList(t *testing.T) {
	tokens, err := Tokenize("(a (b c) d)")
	if err != nil {
		t.Fatalf("Tokenize returned error: %v", err)
	}
	got, err := Parse(tokens)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	want := value.List(
		value.Symbol{Name: "a"},
		value.List(value.Symbol{Name: "b"}, value.Symbol{Name: "c"}),
		value.Symbol{Name: "d"},
	)
	if !valuesEqual(got, want) {
		t.Fatalf("Parse = %#v, want %#v", got, want)
	}
}

func TestParseQuoteSymbol(t *testing.T) {
	tokens, err := Tokenize("'a")
	if err != nil {
		t.Fatalf("Tokenize returned error: %v", err)
	}
	got, err := Parse(tokens)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	want := value.List(value.Symbol{Name: "quote"}, value.Symbol{Name: "a"})
	if !valuesEqual(got, want) {
		t.Fatalf("Parse = %#v, want %#v", got, want)
	}
}

func TestParseQuoteList(t *testing.T) {
	tokens, err := Tokenize("'(a b)")
	if err != nil {
		t.Fatalf("Tokenize returned error: %v", err)
	}
	got, err := Parse(tokens)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	want := value.List(
		value.Symbol{Name: "quote"},
		value.List(value.Symbol{Name: "a"}, value.Symbol{Name: "b"}),
	)
	if !valuesEqual(got, want) {
		t.Fatalf("Parse = %#v, want %#v", got, want)
	}
}

func valuesEqual(a, b value.Value) bool {
	if value.Eq(a, b) {
		return true
	}

	ap, aok := a.(*value.Pair)
	bp, bok := b.(*value.Pair)
	if !aok || !bok {
		return false
	}
	return valuesEqual(ap.Car, bp.Car) && valuesEqual(ap.Cdr, bp.Cdr)
}
