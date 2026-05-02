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

// primitiveNames is the set of McCarthy primitives the VM compiler can
// lower to fixed opcodes. Names in this set need extra bookkeeping when
// they are shadowed at the runtime level so the VM compiler routes
// calls through env lookup instead.
var primitiveNames = map[string]bool{
	"car":  true,
	"cdr":  true,
	"cons": true,
	"atom": true,
	"eq":   true,
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
	// reboundPrimitives tracks primitive names that have been redefined
	// in this Runtime's lifetime, either via Register or via a top-level
	// (label NAME ...). The VM compiler must not lower calls to these
	// names to fixed opcodes — they have to go through env lookup so the
	// override wins. This set is only consulted by the VM backend; the
	// tree-walking evaluator already routes every call through env.
	reboundPrimitives map[string]bool
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
	if primitiveNames[name] {
		r.markReboundPrimitive(name)
	}
}

func (r *Runtime) markReboundPrimitive(name string) {
	if r.reboundPrimitives == nil {
		r.reboundPrimitives = make(map[string]bool)
	}
	r.reboundPrimitives[name] = true
}

// recordTopLevelLabels scans forms for top-level (label NAME EXPR)
// shapes and marks any primitive NAME so the next VM compile pass
// (and any later one) routes through env lookup. Called before
// compilation so the compile pass that contains the label sees the
// override and any subsequent Eval call inherits it.
func (r *Runtime) recordTopLevelLabels(forms []Value) {
	for _, form := range forms {
		pair, ok := form.(*value.Pair)
		if !ok {
			continue
		}
		head, ok := pair.Car.(value.Symbol)
		if !ok || head.Name != "label" {
			continue
		}
		rest, ok := pair.Cdr.(*value.Pair)
		if !ok {
			continue
		}
		name, ok := rest.Car.(value.Symbol)
		if !ok {
			continue
		}
		if primitiveNames[name.Name] {
			r.markReboundPrimitive(name.Name)
		}
	}
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
		r.recordTopLevelLabels(forms)
		var opts []compiler.Option
		if len(r.reboundPrimitives) > 0 {
			names := make([]string, 0, len(r.reboundPrimitives))
			for n := range r.reboundPrimitives {
				names = append(names, n)
			}
			opts = append(opts, compiler.WithRebound(names...))
		}
		code, err := compiler.CompileProgram(forms, opts...)
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
