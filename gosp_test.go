package gosp_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kqnade/gosp"
	"github.com/kqnade/gosp/value"
)

func TestRuntimeEvalDefaultBackendIsEval(t *testing.T) {
	rt := gosp.New()
	got, err := rt.Eval("(car '(a b c))")
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	if gosp.Print(got) != "a" {
		t.Fatalf("Print(got) = %q, want %q", gosp.Print(got), "a")
	}
}

func TestRuntimeBackends(t *testing.T) {
	for _, tc := range []struct {
		name    string
		backend gosp.Backend
	}{
		{"eval", gosp.BackendEval},
		{"vm", gosp.BackendVM},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rt := gosp.New(gosp.WithBackend(tc.backend))
			got, err := rt.Eval("(cons 'a (cons 'b '()))")
			if err != nil {
				t.Fatalf("Eval: %v", err)
			}
			if gosp.Print(got) != "(a b)" {
				t.Fatalf("Print(got) = %q, want %q", gosp.Print(got), "(a b)")
			}
		})
	}
}

func TestRuntimePersistsGlobals(t *testing.T) {
	for _, tc := range []struct {
		name    string
		backend gosp.Backend
	}{
		{"eval", gosp.BackendEval},
		{"vm", gosp.BackendVM},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rt := gosp.New(gosp.WithBackend(tc.backend))
			if _, err := rt.Eval("(label cadr (lambda (x) (car (cdr x))))"); err != nil {
				t.Fatalf("define: %v", err)
			}
			got, err := rt.Eval("(cadr '(a b c))")
			if err != nil {
				t.Fatalf("call: %v", err)
			}
			if gosp.Print(got) != "b" {
				t.Fatalf("Print(got) = %q, want b", gosp.Print(got))
			}
		})
	}
}

func TestRuntimeRegisterBuiltin(t *testing.T) {
	for _, tc := range []struct {
		name    string
		backend gosp.Backend
	}{
		{"eval", gosp.BackendEval},
		{"vm", gosp.BackendVM},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rt := gosp.New(gosp.WithBackend(tc.backend))
			rt.Register("snd", func(args []value.Value) (value.Value, error) {
				if len(args) != 1 {
					return nil, errArity("snd", len(args))
				}
				p, ok := args[0].(*value.Pair)
				if !ok {
					return nil, errType("snd: expected pair")
				}
				rest, ok := p.Cdr.(*value.Pair)
				if !ok {
					return nil, errType("snd: expected at least 2 elements")
				}
				return rest.Car, nil
			})
			got, err := rt.Eval("(snd '(x y z))")
			if err != nil {
				t.Fatalf("Eval: %v", err)
			}
			if gosp.Print(got) != "y" {
				t.Fatalf("Print(got) = %q, want y", gosp.Print(got))
			}
		})
	}
}

func TestRuntimeREPLEvalAndExitOnEOF(t *testing.T) {
	rt := gosp.New()
	in := strings.NewReader("(cons 'a '())\n")
	var out bytes.Buffer
	if err := rt.REPL(in, &out); err != nil {
		t.Fatalf("REPL: %v", err)
	}
	if !strings.Contains(out.String(), "(a)") {
		t.Fatalf("REPL output = %q, want to contain (a)", out.String())
	}
}

type stringErr string

func (s stringErr) Error() string { return string(s) }

func errArity(name string, n int) error { return stringErr("arity error in " + name) }
func errType(msg string) error          { return stringErr(msg) }
