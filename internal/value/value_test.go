package value

import "testing"

func TestCoreTypesSatisfyValue(t *testing.T) {
	var _ Value = Symbol{Name: "a"}
	var _ Value = NIL
	var _ Value = &Pair{Car: Symbol{Name: "a"}, Cdr: NIL}
}

func TestNILIsSingletonNilValue(t *testing.T) {
	if _, ok := NIL.(Nil); !ok {
		t.Fatalf("NIL = %T, want Nil", NIL)
	}
}

func TestPairStoresCarAndCdr(t *testing.T) {
	car := Symbol{Name: "a"}
	cdr := Symbol{Name: "b"}
	p := &Pair{Car: car, Cdr: cdr}

	if p.Car != car {
		t.Fatalf("Car = %#v, want %#v", p.Car, car)
	}
	if p.Cdr != cdr {
		t.Fatalf("Cdr = %#v, want %#v", p.Cdr, cdr)
	}
}

func TestConsCreatesPair(t *testing.T) {
	car := Symbol{Name: "a"}
	cdr := Symbol{Name: "b"}

	p := Cons(car, cdr)
	if p.Car != car {
		t.Fatalf("Car = %#v, want %#v", p.Car, car)
	}
	if p.Cdr != cdr {
		t.Fatalf("Cdr = %#v, want %#v", p.Cdr, cdr)
	}
}

func TestListCreatesProperList(t *testing.T) {
	list := List(Symbol{Name: "a"}, Symbol{Name: "b"})

	first, ok := list.(*Pair)
	if !ok {
		t.Fatalf("list = %T, want *Pair", list)
	}
	if first.Car != (Symbol{Name: "a"}) {
		t.Fatalf("first.Car = %#v, want a", first.Car)
	}
	second, ok := first.Cdr.(*Pair)
	if !ok {
		t.Fatalf("first.Cdr = %T, want *Pair", first.Cdr)
	}
	if second.Car != (Symbol{Name: "b"}) {
		t.Fatalf("second.Car = %#v, want b", second.Car)
	}
	if !IsNil(second.Cdr) {
		t.Fatalf("second.Cdr = %#v, want NIL", second.Cdr)
	}
}

func TestListWithNoValuesIsNil(t *testing.T) {
	if !IsNil(List()) {
		t.Fatalf("List() = %#v, want NIL", List())
	}
}

func TestIsAtom(t *testing.T) {
	if !IsAtom(Symbol{Name: "a"}) {
		t.Fatal("symbol should be atom")
	}
	if !IsAtom(NIL) {
		t.Fatal("NIL should be atom")
	}
	if IsAtom(Cons(Symbol{Name: "a"}, NIL)) {
		t.Fatal("pair should not be atom")
	}
}

func TestEq(t *testing.T) {
	if !Eq(Symbol{Name: "a"}, Symbol{Name: "a"}) {
		t.Fatal("same symbol names should be eq")
	}
	if Eq(Symbol{Name: "a"}, Symbol{Name: "b"}) {
		t.Fatal("different symbol names should not be eq")
	}
	if !Eq(NIL, NIL) {
		t.Fatal("NIL should be eq to NIL")
	}
	if Eq(Cons(Symbol{Name: "a"}, NIL), Cons(Symbol{Name: "a"}, NIL)) {
		t.Fatal("pairs should not be eq")
	}
}

func TestFuncSatisfiesValueAndStoresFields(t *testing.T) {
	params := []Symbol{{Name: "x"}}
	body := Symbol{Name: "x"}
	self := Symbol{Name: "id"}
	fn := &Func{Params: params, Body: body, Self: &self}

	var _ Value = fn
	if fn.Params[0] != params[0] {
		t.Fatalf("Params = %#v, want %#v", fn.Params, params)
	}
	if fn.Body != body {
		t.Fatalf("Body = %#v, want %#v", fn.Body, body)
	}
	if fn.Self != &self {
		t.Fatalf("Self = %#v, want %#v", fn.Self, &self)
	}
}

func TestBuiltinSatisfiesValueAndStoresFields(t *testing.T) {
	called := false
	builtin := Builtin{
		Name: "atom",
		Fn: func(args []Value) (Value, error) {
			called = true
			return NIL, nil
		},
	}

	var _ Value = builtin
	if builtin.Name != "atom" {
		t.Fatalf("Name = %q, want atom", builtin.Name)
	}
	got, err := builtin.Fn(nil)
	if err != nil {
		t.Fatalf("Fn returned error: %v", err)
	}
	if !called {
		t.Fatal("Fn was not called")
	}
	if !IsNil(got) {
		t.Fatalf("Fn returned %#v, want NIL", got)
	}
}
