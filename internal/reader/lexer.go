package reader

import "unicode"

type TokenKind int

const (
	TokenLParen TokenKind = iota
	TokenRParen
	TokenSymbol
)

type Token struct {
	Kind   TokenKind
	Lexeme string
}

func Tokenize(input string) ([]Token, error) {
	var tokens []Token
	for i := 0; i < len(input); {
		r := rune(input[i])
		switch {
		case unicode.IsSpace(r):
			i++
		case input[i] == '(':
			tokens = append(tokens, Token{Kind: TokenLParen, Lexeme: "("})
			i++
		case input[i] == ')':
			tokens = append(tokens, Token{Kind: TokenRParen, Lexeme: ")"})
			i++
		default:
			start := i
			for i < len(input) && !isDelimiter(rune(input[i])) {
				i++
			}
			tokens = append(tokens, Token{Kind: TokenSymbol, Lexeme: input[start:i]})
		}
	}
	return tokens, nil
}

func isDelimiter(r rune) bool {
	return unicode.IsSpace(r) || r == '(' || r == ')'
}
