package reader

import (
	"reflect"
	"testing"
)

func TestTokenizeParens(t *testing.T) {
	got, err := Tokenize("()")
	if err != nil {
		t.Fatalf("Tokenize returned error: %v", err)
	}

	want := []Token{
		{Kind: TokenLParen, Lexeme: "("},
		{Kind: TokenRParen, Lexeme: ")"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Tokenize = %#v, want %#v", got, want)
	}
}

func TestTokenizeSymbols(t *testing.T) {
	got, err := Tokenize("foo bar")
	if err != nil {
		t.Fatalf("Tokenize returned error: %v", err)
	}

	want := []Token{
		{Kind: TokenSymbol, Lexeme: "foo"},
		{Kind: TokenSymbol, Lexeme: "bar"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Tokenize = %#v, want %#v", got, want)
	}
}
