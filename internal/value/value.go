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

type Func struct {
	Params []Symbol
	Body   Value
	Self   *Symbol
}

type Builtin struct {
	Name string
	Fn   func(args []Value) (Value, error)
}

var NIL Value = Nil{}

func (Symbol) lispValue()  {}
func (Nil) lispValue()     {}
func (*Pair) lispValue()   {}
func (*Func) lispValue()   {}
func (Builtin) lispValue() {}

func Cons(a, b Value) *Pair {
	return &Pair{Car: a, Cdr: b}
}

func List(xs ...Value) Value {
	var out Value = NIL
	for i := len(xs) - 1; i >= 0; i-- {
		out = Cons(xs[i], out)
	}
	return out
}

func IsNil(v Value) bool {
	_, ok := v.(Nil)
	return ok
}

func IsAtom(v Value) bool {
	switch v.(type) {
	case Symbol, Nil:
		return true
	default:
		return false
	}
}

func Eq(a, b Value) bool {
	switch av := a.(type) {
	case Symbol:
		bv, ok := b.(Symbol)
		return ok && av.Name == bv.Name
	case Nil:
		return IsNil(b)
	default:
		return false
	}
}
