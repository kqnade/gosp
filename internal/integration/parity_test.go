package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kqnade/gosp/internal/compiler"
	"github.com/kqnade/gosp/internal/eval"
	"github.com/kqnade/gosp/internal/printer"
	"github.com/kqnade/gosp/internal/reader"
	"github.com/kqnade/gosp/value"
	"github.com/kqnade/gosp/internal/vm"
)

func TestParity(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", "testdata", "programs"))
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	matches, err := filepath.Glob(filepath.Join(root, "*.lisp"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("no programs found under %s", root)
	}
	for _, lispPath := range matches {
		name := strings.TrimSuffix(filepath.Base(lispPath), ".lisp")
		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(lispPath)
			if err != nil {
				t.Fatalf("read program: %v", err)
			}
			expectedPath := strings.TrimSuffix(lispPath, ".lisp") + ".expected"
			expectedBytes, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatalf("read expected: %v", err)
			}
			expected := string(expectedBytes)

			forms, err := reader.ReadAll(string(src))
			if err != nil {
				t.Fatalf("ReadAll: %v", err)
			}

			treeOut, err := runTreeWalker(forms)
			if err != nil {
				t.Fatalf("tree-walker: %v", err)
			}
			vmOut, err := runVM(forms)
			if err != nil {
				t.Fatalf("vm: %v", err)
			}

			if treeOut != expected {
				t.Errorf("tree-walker output = %q, want %q", treeOut, expected)
			}
			if vmOut != expected {
				t.Errorf("vm output = %q, want %q", vmOut, expected)
			}
			if treeOut != vmOut {
				t.Errorf("backend mismatch: tree-walker = %q, vm = %q", treeOut, vmOut)
			}
		})
	}
}

func TestParityMetacircular(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", "examples"))
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	src, err := os.ReadFile(filepath.Join(root, "metacircular.lisp"))
	if err != nil {
		t.Fatalf("read program: %v", err)
	}
	expectedBytes, err := os.ReadFile(filepath.Join(root, "metacircular.expected"))
	if err != nil {
		t.Fatalf("read expected: %v", err)
	}
	expected := string(expectedBytes)

	forms, err := reader.ReadAll(string(src))
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	treeOut, err := runTreeWalker(forms)
	if err != nil {
		t.Fatalf("tree-walker: %v", err)
	}
	vmOut, err := runVM(forms)
	if err != nil {
		t.Fatalf("vm: %v", err)
	}

	if treeOut != expected {
		t.Errorf("tree-walker output = %q, want %q", treeOut, expected)
	}
	if vmOut != expected {
		t.Errorf("vm output = %q, want %q", vmOut, expected)
	}
	if treeOut != vmOut {
		t.Errorf("backend mismatch: tree-walker = %q, vm = %q", treeOut, vmOut)
	}
}

func runTreeWalker(forms []value.Value) (string, error) {
	env := eval.NewGlobalEnv()
	result, err := eval.EvalProgram(forms, env)
	if err != nil {
		return "", err
	}
	return printer.Print(result) + "\n", nil
}

func runVM(forms []value.Value) (string, error) {
	code, err := compiler.CompileProgram(forms)
	if err != nil {
		return "", err
	}
	result, err := vm.Run(code, vm.NewGlobalEnv())
	if err != nil {
		return "", err
	}
	return printer.Print(result) + "\n", nil
}
