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

func TestRunInteractiveSingleLine(t *testing.T) {
	in := bytes.NewBufferString("(quote a)\n")
	var out bytes.Buffer
	if err := RunInteractive(in, &out); err != nil {
		t.Fatalf("RunInteractive error: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("a\n")) {
		t.Fatalf("output = %q, want to contain %q", out.String(), "a\n")
	}
}

func TestRunInteractiveMultiLine(t *testing.T) {
	// Form spans two lines; parens only balance after second line.
	in := bytes.NewBufferString("(quote\n  a)\n")
	var out bytes.Buffer
	if err := RunInteractive(in, &out); err != nil {
		t.Fatalf("RunInteractive error: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("a\n")) {
		t.Fatalf("output = %q, want to contain %q", out.String(), "a\n")
	}
}

func TestRunInteractiveSyntaxError(t *testing.T) {
	// Stray ) is a syntax error — should print error and continue, then EOF cleanly.
	in := bytes.NewBufferString(")\n(quote a)\n")
	var out bytes.Buffer
	if err := RunInteractive(in, &out); err != nil {
		t.Fatalf("RunInteractive error: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("a\n")) {
		t.Fatalf("output = %q, want to recover and produce %q", out.String(), "a\n")
	}
}
