package value

import "testing"

func TestEnvLookupDefine(t *testing.T) {
	env := NewEnv(nil)
	want := Symbol{Name: "a"}

	env.Define("x", want)
	got, ok := env.Lookup("x")
	if !ok {
		t.Fatal("Lookup did not find x")
	}
	if got != want {
		t.Fatalf("Lookup = %#v, want %#v", got, want)
	}
}

func TestEnvLookupFallsBackToParent(t *testing.T) {
	parent := NewEnv(nil)
	child := NewEnv(parent)
	want := Symbol{Name: "a"}

	parent.Define("x", want)
	got, ok := child.Lookup("x")
	if !ok {
		t.Fatal("Lookup did not find parent binding")
	}
	if got != want {
		t.Fatalf("Lookup = %#v, want %#v", got, want)
	}
}

func TestEnvDefineShadowsParent(t *testing.T) {
	parent := NewEnv(nil)
	child := NewEnv(parent)
	parent.Define("x", Symbol{Name: "parent"})
	want := Symbol{Name: "child"}

	child.Define("x", want)
	got, ok := child.Lookup("x")
	if !ok {
		t.Fatal("Lookup did not find child binding")
	}
	if got != want {
		t.Fatalf("Lookup = %#v, want %#v", got, want)
	}
}

func TestEnvSetGlobalMutatesRoot(t *testing.T) {
	root := NewEnv(nil)
	child := NewEnv(root)
	grandchild := NewEnv(child)
	root.Define("x", Symbol{Name: "old"})
	want := Symbol{Name: "new"}

	grandchild.SetGlobal("x", want)
	got, ok := root.Lookup("x")
	if !ok {
		t.Fatal("Lookup did not find root binding")
	}
	if got != want {
		t.Fatalf("Lookup = %#v, want %#v", got, want)
	}
}

func TestEnvSetGlobalDoesNotOverwriteChildShadow(t *testing.T) {
	root := NewEnv(nil)
	child := NewEnv(root)
	root.Define("x", Symbol{Name: "root"})
	want := Symbol{Name: "child"}
	child.Define("x", want)

	child.SetGlobal("x", Symbol{Name: "global"})
	got, ok := child.Lookup("x")
	if !ok {
		t.Fatal("Lookup did not find child binding")
	}
	if got != want {
		t.Fatalf("Lookup = %#v, want child shadow %#v", got, want)
	}
}

func TestFuncStoresEnv(t *testing.T) {
	env := NewEnv(nil)
	fn := &Func{Env: env}

	if fn.Env != env {
		t.Fatalf("Env = %#v, want %#v", fn.Env, env)
	}
}
