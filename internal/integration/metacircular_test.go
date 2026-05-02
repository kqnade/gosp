package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kqnade/gosp/eval"
	"github.com/kqnade/gosp/printer"
	"github.com/kqnade/gosp/reader"
)

func TestMetacircular(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", "examples"))
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	src, err := os.ReadFile(filepath.Join(root, "metacircular.lisp"))
	if err != nil {
		t.Fatalf("read program: %v", err)
	}
	expected, err := os.ReadFile(filepath.Join(root, "metacircular.expected"))
	if err != nil {
		t.Fatalf("read expected: %v", err)
	}
	forms, err := reader.ReadAll(string(src))
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	env := eval.NewGlobalEnv()
	result, err := eval.EvalProgram(forms, env)
	if err != nil {
		t.Fatalf("EvalProgram: %v", err)
	}
	got := printer.Print(result) + "\n"
	if got != string(expected) {
		t.Fatalf("output = %q, want %q", got, string(expected))
	}
}
