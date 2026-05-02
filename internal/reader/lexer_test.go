package reader

import (
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  []Token
	}{
		{
			name:  "parens",
			input: "()",
			want: []Token{
				{Kind: TokenLParen, Lexeme: "("},
				{Kind: TokenRParen, Lexeme: ")"},
			},
		},
		{
			name:  "symbols separated by whitespace",
			input: "foo bar",
			want: []Token{
				{Kind: TokenSymbol, Lexeme: "foo"},
				{Kind: TokenSymbol, Lexeme: "bar"},
			},
		},
		{
			name:  "comments and whitespace",
			input: " \tfoo ; skip this\n(bar)\n",
			want: []Token{
				{Kind: TokenSymbol, Lexeme: "foo"},
				{Kind: TokenLParen, Lexeme: "("},
				{Kind: TokenSymbol, Lexeme: "bar"},
				{Kind: TokenRParen, Lexeme: ")"},
			},
		},
		{
			name:  "comment running to EOF without trailing newline",
			input: "foo ; comment without newline",
			want: []Token{
				{Kind: TokenSymbol, Lexeme: "foo"},
			},
		},
		{
			name:  "semicolon adjacent to symbol",
			input: "foo;skip\nbar",
			want: []Token{
				{Kind: TokenSymbol, Lexeme: "foo"},
				{Kind: TokenSymbol, Lexeme: "bar"},
			},
		},
		{
			name:  "quote on symbol",
			input: "'a",
			want: []Token{
				{Kind: TokenQuote, Lexeme: "'"},
				{Kind: TokenSymbol, Lexeme: "a"},
			},
		},
		{
			name:  "quote on list",
			input: "'(a b)",
			want: []Token{
				{Kind: TokenQuote, Lexeme: "'"},
				{Kind: TokenLParen, Lexeme: "("},
				{Kind: TokenSymbol, Lexeme: "a"},
				{Kind: TokenSymbol, Lexeme: "b"},
				{Kind: TokenRParen, Lexeme: ")"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Tokenize(tc.input)
			if err != nil {
				t.Fatalf("Tokenize(%q) returned error: %v", tc.input, err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Tokenize(%q) = %#v, want %#v", tc.input, got, tc.want)
			}
		})
	}
}
