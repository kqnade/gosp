package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prog.lisp")
	src := "(label cadr (lambda (x) (car (cdr x))))\n(cadr '(a b c))\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("write tmp: %v", err)
	}
	var out bytes.Buffer
	if err := Run(path, &out); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	got := out.String()
	want := "b\n"
	if got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestRunFileMissing(t *testing.T) {
	var out bytes.Buffer
	if err := Run(filepath.Join(t.TempDir(), "no-such.lisp"), &out); err == nil {
		t.Fatalf("expected error for missing file, got nil")
	}
}
