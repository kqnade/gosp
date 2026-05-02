package reader

import (
	"strings"
	"testing"

	"github.com/kqnade/gosp/value"
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

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{name: "open paren", src: "(", want: "unexpected EOF"},
		{name: "close paren", src: ")", want: "unexpected )"},
		{name: "unterminated list", src: "(a", want: "unexpected EOF"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := Tokenize(tt.src)
			if err != nil {
				t.Fatalf("Tokenize returned error: %v", err)
			}
			_, err = Parse(tokens)
			if err == nil {
				t.Fatal("Parse returned nil error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Parse error = %q, want substring %q", err.Error(), tt.want)
			}
		})
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
