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

func TestExampleErrors(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"arity", "(second '(a) '(b))"},
		{"short list", "(third '(a b))"},
	}
	for _, backend := range []struct {
		name    string
		backend gosp.Backend
	}{
		{"eval", gosp.BackendEval},
		{"vm", gosp.BackendVM},
	} {
		for _, c := range cases {
			t.Run(backend.name+"/"+c.name, func(t *testing.T) {
				rt := gosp.New(gosp.WithBackend(backend.backend))
				example.Register(rt)
				if _, err := rt.Eval(c.src); err == nil {
					t.Fatalf("Eval(%q): expected error, got nil", c.src)
				}
			})
		}
	}
}
