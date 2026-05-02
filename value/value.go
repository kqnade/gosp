package value

// Value is the closed set of Lisp values handled by gosp.
//
// The interface is sealed by an unexported method: external packages
// cannot define new Value implementations. Builtin authors should
// dispatch on the existing concrete types (Symbol, *Pair, Nil) rather
// than try to add new ones. Carrying host values into Lisp is currently
// not supported; that capability is planned for v0.2.
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

// Func is the tree-walking interpreter's runtime representation of a
// lambda. It is exported only so it can satisfy Value; its field layout
// is an internal detail subject to change in any v0.x release. Builtin
// authors should treat Func values as opaque.
type Func struct {
	Params []Symbol
	Body   Value
	Env    *Env
	Self   *Symbol
}

// Builtin is a Go-defined Lisp callable. Construct one via
// gosp.Runtime.Register rather than instantiating this struct directly.
type Builtin struct {
	Name string
	Fn   func(args []Value) (Value, error)
}

// Closure is the bytecode VM's runtime representation of a compiled
// lambda. Body holds the compiled body (currently *vm.Code) — kept as
// any so the value package does not import vm. Like Func, this is
// exported only to satisfy Value; its field layout is an internal
// detail subject to change in any v0.x release.
type Closure struct {
	Params []Symbol
	Body   any
	Env    *Env
	Self   *Symbol
}

var NIL Value = Nil{}

func (Symbol) lispValue()   {}
func (Nil) lispValue()      {}
func (*Pair) lispValue()    {}
func (*Func) lispValue()    {}
func (Builtin) lispValue()  {}
func (*Closure) lispValue() {}

func Cons(a, b Value) *Pair {
	return &Pair{Car: a, Cdr: b}
}

func List(xs ...Value) Value {
	out := NIL
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
