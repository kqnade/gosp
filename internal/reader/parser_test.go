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
