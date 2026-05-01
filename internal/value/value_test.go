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
