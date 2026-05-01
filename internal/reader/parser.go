package reader

import (
	"fmt"

	"github.com/kqnade/gosp/internal/value"
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
			return nil, fmt.Errorf("gosp: reader: expected )")
		}
		p.pos++
		return value.List(items...), nil
	default:
		return nil, fmt.Errorf("gosp: reader: unexpected token %q", tok.Lexeme)
	}
}
