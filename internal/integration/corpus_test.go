package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kqnade/gosp/internal/eval"
	"github.com/kqnade/gosp/internal/printer"
	"github.com/kqnade/gosp/internal/reader"
)

func TestCorpus(t *testing.T) {
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
			expected, err := os.ReadFile(expectedPath)
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
		})
	}
}
