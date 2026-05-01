package value

// Value is the closed set of Lisp values handled by gosp.
type Value interface {
	lispValue()
}

type Symbol struct {
	Name string
}

type Nil struct{}

type Pair struct {
	Car Value
	Cdr Value
}

var NIL Value = Nil{}

func (Symbol) lispValue() {}
func (Nil) lispValue()    {}
func (*Pair) lispValue()  {}
