package vm_test

import (
	"testing"

	"github.com/kqnade/gosp/compiler"
	"github.com/kqnade/gosp/value"
	"github.com/kqnade/gosp/vm"
)

// TestVMTailCallStress drives a 100,000-iteration tail-recursive drop
// over a long pair chain. Without TAIL_CALL frame replacement the call
// stack would grow to 100k frames and the test would either run very
// slowly or exhaust process memory; with frame replacement the
// observed frame depth stays bounded by a small constant.
func TestVMTailCallStress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping VM tail-call stress test in -short mode")
	}

	const n = 100_000

	q := value.Symbol{Name: "quote"}
	lambda := value.Symbol{Name: "lambda"}
	label := value.Symbol{Name: "label"}
	cond := value.Symbol{Name: "cond"}
	eq := value.Symbol{Name: "eq"}
	cdr := value.Symbol{Name: "cdr"}
	tSym := value.Symbol{Name: "t"}
	drop := value.Symbol{Name: "drop"}
	xs := value.Symbol{Name: "xs"}
	done := value.Symbol{Name: "done"}
	okSym := value.Symbol{Name: "ok"}

	// Build (x . (x . ... . done)) — 100k cons cells terminating at `done`.
	list := value.Value(done)
	for range n {
		list = value.Cons(value.Symbol{Name: "x"}, list)
	}

	// (label drop (lambda (xs)
	//   (cond ((eq xs (quote done)) (quote ok))
	//         (t (drop (cdr xs))))))
	body := value.List(
		cond,
		value.List(value.List(eq, xs, value.List(q, done)), value.List(q, okSym)),
		value.List(tSym, value.List(drop, value.List(cdr, xs))),
	)
	labelForm := value.List(label, drop, value.List(lambda, value.List(xs), body))
	callForm := value.List(drop, value.List(q, list))

	code, err := compiler.CompileProgram([]value.Value{labelForm, callForm})
	if err != nil {
		t.Fatalf("CompileProgram: %v", err)
	}

	got, maxFrames, err := vm.RunWithStats(code, value.NewEnv(nil))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !value.Eq(got, okSym) {
		t.Errorf("result = %v, want ok", got)
	}
	if maxFrames > 32 {
		t.Errorf("max frame depth = %d for %d iterations, want bounded by ~32", maxFrames, n)
	}
}
