package reader

import (
	"fmt"

	"github.com/kqnade/gosp/value"
)

func Parse(tokens []Token) (value.Value, error) {
	p := parser{tokens: tokens}
	v, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if p.pos != len(tokens) {
		return nil, fmt.Errorf("gosp: reader: unexpected token %q", tokens[p.pos].Lexeme)
	}
	return v, nil
}

func ReadAll(input string) ([]value.Value, error) {
	tokens, err := Tokenize(input)
	if err != nil {
		return nil, err
	}

	p := parser{tokens: tokens}
	var forms []value.Value
	for p.pos < len(tokens) {
		form, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		forms = append(forms, form)
	}
	return forms, nil
}

type parser struct {
	tokens []Token
	pos    int
}

func (p *parser) parseExpr() (value.Value, error) {
	if p.pos >= len(p.tokens) {
		return nil, fmt.Errorf("gosp: reader: unexpected EOF")
	}

	tok := p.tokens[p.pos]
	p.pos++
	switch tok.Kind {
	case TokenSymbol:
		return value.Symbol{Name: tok.Lexeme}, nil
	case TokenQuote:
		quoted, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		return value.List(value.Symbol{Name: "quote"}, quoted), nil
	case TokenLParen:
		var items []value.Value
		for p.pos < len(p.tokens) && p.tokens[p.pos].Kind != TokenRParen {
			item, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
		if p.pos >= len(p.tokens) {
			return nil, fmt.Errorf("gosp: reader: unexpected EOF")
		}
		p.pos++
		return value.List(items...), nil
	case TokenRParen:
		return nil, fmt.Errorf("gosp: reader: unexpected )")
	default:
		return nil, fmt.Errorf("gosp: reader: unexpected token %q", tok.Lexeme)
	}
}
