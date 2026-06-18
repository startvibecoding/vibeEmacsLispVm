package vm

import "testing"

func TestIsNil(t *testing.T) {
	cases := []struct {
		name string
		v    Value
		want bool
	}{
		{"nil interface", nil, true},
		{"NilType", Nil, true},
		{"empty List", List{}, true},
		{"non-empty List", List{Symbol("a")}, false},
		{"Symbol nil", Symbol("nil"), false},
		{"String", String(""), false},
		{"Number", Number(0), false},
		{"Symbol", Symbol("t"), false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNil(tt.v); got != tt.want {
				t.Fatalf("IsNil(%v) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

func TestTruthy(t *testing.T) {
	cases := []struct {
		name string
		v    Value
		want bool
	}{
		{"nil", nil, false},
		{"NilType", Nil, false},
		{"empty List", List{}, false},
		{"non-empty List", List{Symbol("a")}, true},
		{"Symbol t", Symbol("t"), true},
		{"String", String("x"), true},
		{"Number", Number(1), true},
		{"Number zero", Number(0), true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := Truthy(tt.v); got != tt.want {
				t.Fatalf("Truthy(%v) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

func TestStringify(t *testing.T) {
	cases := []struct {
		name string
		v    Value
		want string
	}{
		{"nil interface", nil, "nil"},
		{"NilType", Nil, "nil"},
		{"Symbol", Symbol("foo"), "foo"},
		{"String", String("hello"), `"hello"`},
		{"String with escapes", String("a\nb"), "\"a\\nb\""},
		{"Number int", Number(42), "42"},
		{"Number float", Number(3.14), "3.14"},
		{"Number negative", Number(-5), "-5"},
		{"Number zero", Number(0), "0"},
		{"empty List", List{}, "nil"},
		{"List", List{Symbol("a"), Number(1)}, "(a 1)"},
		{"nested List", List{Symbol("a"), List{Symbol("b")}}, "(a (b))"},
		{"Function with name", &Function{name: "foo"}, "#<function foo>"},
		{"Function lambda", &Function{}, "#<function lambda>"},
		{"Macro with name", &Macro{name: "m1"}, "#<macro m1>"},
		{"Macro anonymous", &Macro{}, "#<macro>"},
		{"Buffer", &Buffer{name: "*scratch*"}, "#<buffer *scratch*>"},
		{"Buffer killed", &Buffer{name: "x", killed: true}, "#<killed buffer>"},
		{"Buffer nil", (*Buffer)(nil), "#<killed buffer>"},
		{"Marker in buffer", &Marker{buffer: &Buffer{name: "b"}, pos: 5}, "#<marker at 5 in b>"},
		{"Marker no buffer", &Marker{buffer: nil}, "#<marker in no buffer>"},
		{"Marker killed buffer", &Marker{buffer: &Buffer{name: "b", killed: true}, pos: 1}, "#<marker in no buffer>"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := Stringify(tt.v); got != tt.want {
				t.Fatalf("Stringify(%v) = %q, want %q", tt.v, got, tt.want)
			}
		})
	}
}

func TestStringifyEmptyNil(t *testing.T) {
	// Empty List should stringify as "nil"
	if got := Stringify(List{}); got != "nil" {
		t.Fatalf("Stringify(List{}) = %q, want nil", got)
	}
}

func TestStringifyNilType(t *testing.T) {
	if got := Stringify(NilType{}); got != "nil" {
		t.Fatalf("Stringify(NilType{}) = %q, want nil", got)
	}
}
