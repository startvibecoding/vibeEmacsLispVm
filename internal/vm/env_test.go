package vm

import "testing"

func TestNewEnv(t *testing.T) {
	e := NewEnv(nil)
	if e == nil {
		t.Fatal("NewEnv(nil) returned nil")
	}
	if e.parent != nil {
		t.Fatal("NewEnv(nil) should have nil parent")
	}
}

func TestEnvDefineAndGet(t *testing.T) {
	e := NewEnv(nil)
	e.Define("x", Number(42))

	v, ok := e.Get("x")
	if !ok {
		t.Fatal("Get(x) returned false")
	}
	if v != Number(42) {
		t.Fatalf("Get(x) = %v, want 42", v)
	}
}

func TestEnvGetNotFound(t *testing.T) {
	e := NewEnv(nil)
	_, ok := e.Get("missing")
	if ok {
		t.Fatal("Get(missing) should return false")
	}
}

func TestEnvGetFromParent(t *testing.T) {
	parent := NewEnv(nil)
	parent.Define("x", Number(1))

	child := NewEnv(parent)
	child.Define("y", Number(2))

	v, ok := child.Get("x")
	if !ok || v != Number(1) {
		t.Fatalf("child.Get(x) = %v, %v, want 1, true", v, ok)
	}

	v, ok = child.Get("y")
	if !ok || v != Number(2) {
		t.Fatalf("child.Get(y) = %v, %v, want 2, true", v, ok)
	}
}

func TestEnvSetExisting(t *testing.T) {
	e := NewEnv(nil)
	e.Define("x", Number(1))
	e.Set("x", Number(99))

	v, ok := e.Get("x")
	if !ok || v != Number(99) {
		t.Fatalf("Get(x) = %v, %v, want 99, true", v, ok)
	}
}

func TestEnvSetPropagatesToParent(t *testing.T) {
	parent := NewEnv(nil)
	parent.Define("x", Number(1))

	child := NewEnv(parent)
	child.Set("x", Number(99))

	// Should update parent since child doesn't have "x"
	v, ok := parent.Get("x")
	if !ok || v != Number(99) {
		t.Fatalf("parent.Get(x) = %v, %v, want 99, true", v, ok)
	}
}

func TestEnvSetDefinesInCurrentWhenNew(t *testing.T) {
	e := NewEnv(nil)
	e.Set("new", Number(42))

	v, ok := e.Get("new")
	if !ok || v != Number(42) {
		t.Fatalf("Get(new) = %v, %v, want 42, true", v, ok)
	}
}

func TestEnvSetNilEnv(t *testing.T) {
	// Should not panic
	var e *Env
	e.Set("x", Number(1))
}

func TestEnvGetNilEnv(t *testing.T) {
	var e *Env
	_, ok := e.Get("x")
	if ok {
		t.Fatal("nil Env.Get should return false")
	}
}

func TestEnvNames(t *testing.T) {
	parent := NewEnv(nil)
	parent.Define("a", Number(1))
	parent.Define("b", Number(2))

	child := NewEnv(parent)
	child.Define("c", Number(3))

	names := child.Names()
	if len(names) != 3 {
		t.Fatalf("len(Names()) = %d, want 3", len(names))
	}
	// Should be sorted
	expected := []string{"a", "b", "c"}
	for i, name := range names {
		if name != expected[i] {
			t.Fatalf("Names()[%d] = %q, want %q", i, name, expected[i])
		}
	}
}

func TestEnvNamesEmpty(t *testing.T) {
	e := NewEnv(nil)
	names := e.Names()
	if len(names) != 0 {
		t.Fatalf("len(Names()) = %d, want 0", len(names))
	}
}

func TestEnvChildOverridesParent(t *testing.T) {
	parent := NewEnv(nil)
	parent.Define("x", Number(1))

	child := NewEnv(parent)
	child.Define("x", Number(2))

	v, ok := child.Get("x")
	if !ok || v != Number(2) {
		t.Fatalf("child.Get(x) = %v, %v, want 2, true", v, ok)
	}

	// Parent should still have original
	v, ok = parent.Get("x")
	if !ok || v != Number(1) {
		t.Fatalf("parent.Get(x) = %v, %v, want 1, true", v, ok)
	}
}
