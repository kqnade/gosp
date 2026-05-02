package example_test

import (
	"testing"

	"github.com/kqnade/gosp"
	"github.com/kqnade/gosp/ext/example"
)

func TestExampleRegister(t *testing.T) {
	for _, tc := range []struct {
		name    string
		backend gosp.Backend
	}{
		{"eval", gosp.BackendEval},
		{"vm", gosp.BackendVM},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rt := gosp.New(gosp.WithBackend(tc.backend))
			example.Register(rt)

			cases := []struct {
				src  string
				want string
			}{
				{"(second '(a b c d))", "b"},
				{"(third '(a b c d))", "c"},
			}
			for _, c := range cases {
				got, err := rt.Eval(c.src)
				if err != nil {
					t.Fatalf("Eval(%q): %v", c.src, err)
				}
				if gosp.Print(got) != c.want {
					t.Fatalf("Eval(%q) = %q, want %q", c.src, gosp.Print(got), c.want)
				}
			}
		})
	}
}

func TestExampleArityError(t *testing.T) {
	rt := gosp.New()
	example.Register(rt)
	if _, err := rt.Eval("(second '(a) '(b))"); err == nil {
		t.Fatalf("expected arity error, got nil")
	}
}

func TestExampleShortList(t *testing.T) {
	rt := gosp.New()
	example.Register(rt)
	if _, err := rt.Eval("(third '(a b))"); err == nil {
		t.Fatalf("expected short-list error, got nil")
	}
}
