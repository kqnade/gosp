package printer

import (
	"strings"

	"github.com/kqnade/gosp/internal/value"
)

func Print(v value.Value) string {
	switch x := v.(type) {
	case value.Nil:
		return "()"
	case value.Symbol:
		return x.Name
	case *value.Pair:
		return printList(x)
	default:
		return "<unknown>"
	}
}

func printList(p *value.Pair) string {
	var parts []string
	for {
		parts = append(parts, Print(p.Car))
		if value.IsNil(p.Cdr) {
			return "(" + strings.Join(parts, " ") + ")"
		}
		p = p.Cdr.(*value.Pair)
	}
}
