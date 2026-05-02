// Package gosp is a minimal McCarthy-style Lisp embeddable as a Go
// library. It exposes a Runtime that can evaluate source through either
// a tree-walking interpreter or a stack-based bytecode VM, and lets host
// programs add Go-defined builtins via Runtime.Register.
//
// The implementation packages (reader, eval, compiler, vm) live under
// internal/ and are intentionally not part of the public API. Builtin
// authors should depend only on package value and this facade.
//
// API stability: gosp is at v0.x; minor versions may break the API.
package gosp

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kqnade/gosp/internal/compiler"
	"github.com/kqnade/gosp/internal/eval"
	"github.com/kqnade/gosp/internal/printer"
	"github.com/kqnade/gosp/internal/reader"
	"github.com/kqnade/gosp/internal/vm"
	"github.com/kqnade/gosp/value"
)

// Value is a Lisp value. Aliased from package value for convenience.
type Value = value.Value

// BuiltinFunc is the Go signature of a Lisp builtin. Argument arity
// must be validated by the function itself.
type BuiltinFunc = func(args []Value) (Value, error)

// Backend selects the execution strategy for a Runtime.
type Backend int

const (
	// BackendEval is the tree-walking evaluator. Default.
	BackendEval Backend = iota
	// BackendVM is the stack-based bytecode VM.
	BackendVM
)

// Option configures a Runtime at construction time.
type Option func(*Runtime)

// WithBackend selects the execution backend.
func WithBackend(b Backend) Option {
	return func(r *Runtime) { r.backend = b }
}

// Runtime owns a persistent global environment and an execution
// backend. Use New to construct one.
//
// Runtime is not safe for concurrent use. Callers that share a Runtime
// across goroutines must serialize all Eval, EvalForm, RunFile, REPL,
// and Register calls externally (e.g. with a sync.Mutex).
type Runtime struct {
	backend Backend
	env     *value.Env
}

// New creates a Runtime with the McCarthy primitives pre-bound.
func New(opts ...Option) *Runtime {
	r := &Runtime{backend: BackendEval}
	for _, opt := range opts {
		opt(r)
	}
	switch r.backend {
	case BackendVM:
		r.env = vm.NewGlobalEnv()
	default:
		r.env = eval.NewGlobalEnv()
	}
	return r
}

// Register adds a Go-defined builtin under name in the global env. It
// shadows any existing binding, including the McCarthy primitives
// (car, cdr, cons, atom, eq).
//
// There is no way to remove a registration; create a new Runtime to
// reset the environment.
func (r *Runtime) Register(name string, fn BuiltinFunc) {
	r.env.Define(name, value.Builtin{Name: name, Fn: fn})
}

// Eval parses src and evaluates each form in order, returning the value
// of the last form. () is returned for empty input.
func (r *Runtime) Eval(src string) (Value, error) {
	forms, err := reader.ReadAll(src)
	if err != nil {
		return nil, err
	}
	if len(forms) == 0 {
		return value.NIL, nil
	}
	return r.evalForms(forms)
}

// EvalForm evaluates a single already-parsed form.
func (r *Runtime) EvalForm(form Value) (Value, error) {
	return r.evalForms([]Value{form})
}

func (r *Runtime) evalForms(forms []Value) (Value, error) {
	switch r.backend {
	case BackendVM:
		code, err := compiler.CompileProgram(forms)
		if err != nil {
			return nil, err
		}
		return vm.Run(code, r.env)
	default:
		return eval.EvalProgram(forms, r.env)
	}
}

// RunFile reads path and evaluates its forms in this Runtime.
func (r *Runtime) RunFile(path string) (Value, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("gosp: %w", err)
	}
	return r.Eval(string(src))
}

// REPL runs an interactive read-eval-print loop on in/out until EOF.
// Each result is printed with Print.
func (r *Runtime) REPL(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	var buf strings.Builder
	for {
		if buf.Len() == 0 {
			fmt.Fprint(out, "> ")
		} else {
			fmt.Fprint(out, "  ")
		}
		if !scanner.Scan() {
			break
		}
		buf.WriteString(scanner.Text())
		buf.WriteByte('\n')
		if !parensBalanced(buf.String()) {
			continue
		}
		src := buf.String()
		buf.Reset()
		forms, err := reader.ReadAll(src)
		if err != nil {
			fmt.Fprintln(out, err)
			continue
		}
		for _, form := range forms {
			v, err := r.EvalForm(form)
			if err != nil {
				fmt.Fprintln(out, err)
				break
			}
			fmt.Fprintln(out, Print(v))
		}
	}
	return scanner.Err()
}

// Print returns the printable representation of v.
func Print(v Value) string {
	return printer.Print(v)
}

func parensBalanced(src string) bool {
	depth := 0
	inComment := false
	for _, r := range src {
		if inComment {
			if r == '\n' {
				inComment = false
			}
			continue
		}
		switch r {
		case ';':
			inComment = true
		case '(':
			depth++
		case ')':
			depth--
		}
	}
	return depth <= 0
}
