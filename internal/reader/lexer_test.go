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

func TestTokenizeCommentsAndWhitespace(t *testing.T) {
	got, err := Tokenize(" \tfoo ; skip this\n(bar)\n")
	if err != nil {
		t.Fatalf("Tokenize returned error: %v", err)
	}

	want := []Token{
		{Kind: TokenSymbol, Lexeme: "foo"},
		{Kind: TokenLParen, Lexeme: "("},
		{Kind: TokenSymbol, Lexeme: "bar"},
		{Kind: TokenRParen, Lexeme: ")"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Tokenize = %#v, want %#v", got, want)
	}
}

func TestTokenizeSemicolonAdjacentToSymbol(t *testing.T) {
	got, err := Tokenize("foo;skip\nbar")
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

func TestTokenizeQuoteSymbol(t *testing.T) {
	got, err := Tokenize("'a")
	if err != nil {
		t.Fatalf("Tokenize returned error: %v", err)
	}

	want := []Token{
		{Kind: TokenQuote, Lexeme: "'"},
		{Kind: TokenSymbol, Lexeme: "a"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Tokenize = %#v, want %#v", got, want)
	}
}

func TestTokenizeQuoteList(t *testing.T) {
	got, err := Tokenize("'(a b)")
	if err != nil {
		t.Fatalf("Tokenize returned error: %v", err)
	}

	want := []Token{
		{Kind: TokenQuote, Lexeme: "'"},
		{Kind: TokenLParen, Lexeme: "("},
		{Kind: TokenSymbol, Lexeme: "a"},
		{Kind: TokenSymbol, Lexeme: "b"},
		{Kind: TokenRParen, Lexeme: ")"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Tokenize = %#v, want %#v", got, want)
	}
}
