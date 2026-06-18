package vm

import (
	"context"
	"testing"
)

func TestBuiltinConcat(t *testing.T) {
	e := New()
	cases := map[string]string{
		`(concat "a" "b")`:       `"ab"`,
		`(concat)`:               `""`,
		`(concat "a")`:           `"a"`,
		`(concat "a" "b" "c")`:   `"abc"`,
		`(concat 'sym "str")`:    `"symstr"`,
		`(concat 42 " items")`:   `"42 items"`,
		`(concat nil "str")`:     `"str"`,
	}
	for src, want := range cases {
		t.Run(src, func(t *testing.T) {
			v, err := e.EvalString(context.Background(), src)
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if got := Stringify(v); got != want {
				t.Fatalf("got %s, want %s", got, want)
			}
		})
	}
}

func TestBuiltinFormat(t *testing.T) {
	e := New()
	cases := map[string]string{
		`(format "hello %s" "world")`:       `"hello world"`,
		`(format "%s %s" "a" "b")`:          `"a b"`,
		`(format "%s" 42)`:                  `"42"`,
		`(format "no args")`:                `"no args"`,
		`(format "%s" nil)`:                 `"nil"`,
		`(format "%s" '(1 2))`:              `"(1 2)"`,
	}
	for src, want := range cases {
		t.Run(src, func(t *testing.T) {
			v, err := e.EvalString(context.Background(), src)
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if got := Stringify(v); got != want {
				t.Fatalf("got %s, want %s", got, want)
			}
		})
	}
}

func TestBuiltinList(t *testing.T) {
	e := New()
	cases := map[string]string{
		`(list)`:           `nil`,
		`(list 1)`:         `(1)`,
		`(list 1 2 3)`:     `(1 2 3)`,
		`(list 'a "b" 3)`:  `(a "b" 3)`,
	}
	for src, want := range cases {
		t.Run(src, func(t *testing.T) {
			v, err := e.EvalString(context.Background(), src)
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if got := Stringify(v); got != want {
				t.Fatalf("got %s, want %s", got, want)
			}
		})
	}
}

func TestBuiltinLength(t *testing.T) {
	e := New()
	cases := map[string]string{
		`(length "")`:           `0`,
		`(length "abc")`:        `3`,
		`(length '())`:          `0`,
		`(length '(1 2 3))`:     `3`,
	}
	for src, want := range cases {
		t.Run(src, func(t *testing.T) {
			v, err := e.EvalString(context.Background(), src)
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if got := Stringify(v); got != want {
				t.Fatalf("got %s, want %s", got, want)
			}
		})
	}
}

func TestBuiltinCons(t *testing.T) {
	e := New()
	cases := map[string]string{
		`(cons 1 nil)`:       `(1)`,
		`(cons 1 '(2 3))`:    `(1 2 3)`,
		`(cons '(1) '(2))`:   `((1) 2)`,
	}
	for src, want := range cases {
		t.Run(src, func(t *testing.T) {
			v, err := e.EvalString(context.Background(), src)
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if got := Stringify(v); got != want {
				t.Fatalf("got %s, want %s", got, want)
			}
		})
	}
}

func TestBuiltinCarCdr(t *testing.T) {
	e := New()
	cases := map[string]string{
		`(car '(1 2 3))`:   `1`,
		`(car nil)`:        `nil`,
		`(cdr '(1 2 3))`:   `(2 3)`,
		`(cdr '(1))`:       `nil`,
		`(cdr nil)`:        `nil`,
	}
	for src, want := range cases {
		t.Run(src, func(t *testing.T) {
			v, err := e.EvalString(context.Background(), src)
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if got := Stringify(v); got != want {
				t.Fatalf("got %s, want %s", got, want)
			}
		})
	}
}

func TestBuiltinNth(t *testing.T) {
	e := New()
	cases := map[string]string{
		`(nth 0 '(a b c))`:   `a`,
		`(nth 2 '(a b c))`:   `c`,
		`(nth 3 '(a b c))`:   `nil`,
		`(nth 0 nil)`:        `nil`,
	}
	for src, want := range cases {
		t.Run(src, func(t *testing.T) {
			v, err := e.EvalString(context.Background(), src)
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if got := Stringify(v); got != want {
				t.Fatalf("got %s, want %s", got, want)
			}
		})
	}
}

func TestBuiltinAppend(t *testing.T) {
	e := New()
	cases := map[string]string{
		`(append)`:                  `nil`,
		`(append '(1 2))`:           `(1 2)`,
		`(append '(1) '(2) '(3))`:   `(1 2 3)`,
		`(append nil '(1))`:         `(1)`,
	}
	for src, want := range cases {
		t.Run(src, func(t *testing.T) {
			v, err := e.EvalString(context.Background(), src)
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if got := Stringify(v); got != want {
				t.Fatalf("got %s, want %s", got, want)
			}
		})
	}
}

func TestBuiltinReverse(t *testing.T) {
	e := New()
	cases := map[string]string{
		`(reverse nil)`:       `nil`,
		`(reverse '(1))`:      `(1)`,
		`(reverse '(1 2 3))`:  `(3 2 1)`,
	}
	for src, want := range cases {
		t.Run(src, func(t *testing.T) {
			v, err := e.EvalString(context.Background(), src)
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if got := Stringify(v); got != want {
				t.Fatalf("got %s, want %s", got, want)
			}
		})
	}
}

func TestBuiltinMember(t *testing.T) {
	e := New()
	cases := map[string]string{
		`(member "b" '("a" "b" "c"))`:   `("b" "c")`,
		`(member "x" '("a" "b" "c"))`:   `nil`,
		`(member 'a '(a b c))`:           `(a b c)`,
	}
	for src, want := range cases {
		t.Run(src, func(t *testing.T) {
			v, err := e.EvalString(context.Background(), src)
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if got := Stringify(v); got != want {
				t.Fatalf("got %s, want %s", got, want)
			}
		})
	}
}

func TestBuiltinAssoc(t *testing.T) {
	e := New()
	cases := map[string]string{
		`(assoc 'b '((a 1) (b 2) (c 3)))`:        `(b 2)`,
		`(assoc "b" '(("a" 1) ("b" 2)))`:         `("b" 2)`,
		`(assoc 'x '((a 1) (b 2)))`:              `nil`,
		`(assoc 'b '(not-a-pair (a 1)))`:         `nil`,
	}
	for src, want := range cases {
		t.Run(src, func(t *testing.T) {
			v, err := e.EvalString(context.Background(), src)
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if got := Stringify(v); got != want {
				t.Fatalf("got %s, want %s", got, want)
			}
		})
	}
}

func TestBuiltinArithmeticMultiArg(t *testing.T) {
	e := New()
	cases := map[string]string{
		`(+ 1 2 3 4)`:    `10`,
		`(- 10 3 2)`:     `5`,
		`(* 2 3 4)`:      `24`,
		`(/ 24 2 3)`:     `4`,
	}
	for src, want := range cases {
		t.Run(src, func(t *testing.T) {
			v, err := e.EvalString(context.Background(), src)
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if got := Stringify(v); got != want {
				t.Fatalf("got %s, want %s", got, want)
			}
		})
	}
}

func TestBuiltinComparisonMultiArg(t *testing.T) {
	e := New()
	cases := map[string]string{
		`(< 1 2 3)`:     `t`,
		`(< 1 3 2)`:     `nil`,
		`(<= 1 1 2)`:    `t`,
		`(> 3 2 1)`:     `t`,
		`(> 3 1 2)`:     `nil`,
		`(>= 3 3 2)`:    `t`,
		`(= 1 1 1)`:     `t`,
		`(= 1 1 2)`:     `nil`,
		`(/= 1 2 3)`:    `t`,
		`(/= 1 2 1)`:    `nil`,
	}
	for src, want := range cases {
		t.Run(src, func(t *testing.T) {
			v, err := e.EvalString(context.Background(), src)
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if got := Stringify(v); got != want {
				t.Fatalf("got %s, want %s", got, want)
			}
		})
	}
}

func TestBuiltinPointMinPointMax(t *testing.T) {
	e := New()
	// Empty buffer: point-min and point-max are both 1
	cases := map[string]string{
		`(progn (erase-buffer) (point-min))`:           `1`,
		`(progn (erase-buffer) (point-max))`:           `1`,
		`(progn (erase-buffer) (insert "abc") (point-min))`: `1`,
		`(progn (erase-buffer) (insert "abc") (point-max))`: `4`,
	}
	for src, want := range cases {
		t.Run(src, func(t *testing.T) {
			v, err := e.EvalString(context.Background(), src)
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if got := Stringify(v); got != want {
				t.Fatalf("got %s, want %s", got, want)
			}
		})
	}
}

func TestBuiltinPointMarker(t *testing.T) {
	e := New()
	cases := map[string]string{
		`(progn (erase-buffer) (insert "abc") (point-marker))`: `#<marker at 4 in *scratch*>`,
	}
	for src, want := range cases {
		t.Run(src, func(t *testing.T) {
			v, err := e.EvalString(context.Background(), src)
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if got := Stringify(v); got != want {
				t.Fatalf("got %s, want %s", got, want)
			}
		})
	}
}

func TestBuiltinCopyMarkerSuccess(t *testing.T) {
	e := New()
	// copy-marker with a position in a buffer
	v, err := e.EvalString(context.Background(), `(progn (erase-buffer) (insert "abc") (setq m (point-marker)) (copy-marker m))`)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if got := Stringify(v); got != "#<marker at 4 in *scratch*>" {
		t.Fatalf("got %s, want marker", got)
	}
}

func TestBuiltinBufferpAndCurrentBuffer(t *testing.T) {
	e := New()
	cases := map[string]string{
		`(bufferp (current-buffer))`:   `t`,
		`(bufferp "not-buffer")`:       `nil`,
		`(bufferp nil)`:                `nil`,
	}
	for src, want := range cases {
		t.Run(src, func(t *testing.T) {
			v, err := e.EvalString(context.Background(), src)
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if got := Stringify(v); got != want {
				t.Fatalf("got %s, want %s", got, want)
			}
		})
	}
}

func TestBuiltinSetBufferSuccess(t *testing.T) {
	e := New()
	v, err := e.EvalString(context.Background(), `(progn (get-buffer-create "target") (set-buffer "target") (erase-buffer) (insert "data") (buffer-name))`)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if got := Stringify(v); got != `"target"` {
		t.Fatalf("got %s, want target", got)
	}
}

func TestBuiltinGetBufferSuccess(t *testing.T) {
	e := New()
	cases := map[string]string{
		`(bufferp (get-buffer "*scratch*"))`:   `t`,
		`(get-buffer "nonexistent")`:           `nil`,
	}
	for src, want := range cases {
		t.Run(src, func(t *testing.T) {
			v, err := e.EvalString(context.Background(), src)
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if got := Stringify(v); got != want {
				t.Fatalf("got %s, want %s", got, want)
			}
		})
	}
}

// Helper function tests
func TestAsNumber(t *testing.T) {
	n, err := asNumber(Number(42))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != Number(42) {
		t.Fatalf("got %v, want 42", n)
	}
	_, err = asNumber(String("42"))
	if err == nil {
		t.Fatal("expected error for string")
	}
}

func TestAsStringDesignator(t *testing.T) {
	s, err := asStringDesignator(String("hello"))
	if err != nil || s != "hello" {
		t.Fatalf("got %q, %v", s, err)
	}
	s, err = asStringDesignator(Symbol("sym"))
	if err != nil || s != "sym" {
		t.Fatalf("got %q, %v", s, err)
	}
	_, err = asStringDesignator(Number(42))
	if err == nil {
		t.Fatal("expected error for number")
	}
}

func TestAsList(t *testing.T) {
	l, err := asList(List{1, 2}, "test")
	if err != nil || len(l) != 2 {
		t.Fatalf("got %v, %v", l, err)
	}
	l, err = asList(Nil, "test")
	if err != nil || len(l) != 0 {
		t.Fatalf("got %v, %v", l, err)
	}
	_, err = asList(Number(42), "test")
	if err == nil {
		t.Fatal("expected error for number")
	}
}

func TestLispBool(t *testing.T) {
	if lispBool(true) != Symbol("t") {
		t.Fatalf("lispBool(true) = %v", lispBool(true))
	}
	if !IsNil(lispBool(false)) {
		t.Fatalf("lispBool(false) = %v", lispBool(false))
	}
}

func TestIsForm(t *testing.T) {
	if !isForm(List{Symbol("a")}, "a") {
		t.Fatal("expected true")
	}
	if isForm(List{Symbol("b")}, "a") {
		t.Fatal("expected false")
	}
}

func TestIsFormWithEmptyList(t *testing.T) {
	if isForm(List{}, "a") {
		t.Fatal("expected false for empty list")
	}
}

func TestEqValues(t *testing.T) {
	if !eqValues(Symbol("a"), Symbol("a")) {
		t.Fatal("expected true for same symbols")
	}
	if eqValues(Symbol("a"), Symbol("b")) {
		t.Fatal("expected false for different symbols")
	}
	if !eqValues(Nil, Nil) {
		t.Fatal("expected true for nil")
	}
	if eqValues(Number(1), Number(1)) {
		t.Fatal("expected false for different numbers (not same object)")
	}
}

func TestEqualValues(t *testing.T) {
	if !equalValues(Number(1), Number(1)) {
		t.Fatal("expected true for equal numbers")
	}
	if !equalValues(String("a"), String("a")) {
		t.Fatal("expected true for equal strings")
	}
	if !equalValues(List{Number(1)}, List{Number(1)}) {
		t.Fatal("expected true for equal lists")
	}
	if equalValues(Number(1), Number(2)) {
		t.Fatal("expected false for different numbers")
	}
}

func TestRawString(t *testing.T) {
	if rawString(String("hello")) != "hello" {
		t.Fatalf("rawString = %q", rawString(String("hello")))
	}
	if rawString(Symbol("sym")) != "sym" {
		t.Fatalf("rawString = %q", rawString(Symbol("sym")))
	}
	if rawString(Number(42)) != "42" {
		t.Fatalf("rawString = %q", rawString(Number(42)))
	}
	if rawString(Nil) != "nil" {
		t.Fatalf("rawString = %q", rawString(Nil))
	}
}

func TestCompareNumbers(t *testing.T) {
	cases := []struct {
		name string
		args []Value
		cmp  func(a, b Number) bool
		want string
	}{
		{"less true", []Value{Number(1), Number(2)}, func(a, b Number) bool { return a < b }, "t"},
		{"less false", []Value{Number(2), Number(1)}, func(a, b Number) bool { return a < b }, "nil"},
		{"less equal", []Value{Number(1), Number(1)}, func(a, b Number) bool { return a <= b }, "t"},
		{"greater", []Value{Number(3), Number(1)}, func(a, b Number) bool { return a > b }, "t"},
		{"greater equal", []Value{Number(1), Number(1)}, func(a, b Number) bool { return a >= b }, "t"},
		{"equal", []Value{Number(1), Number(1)}, func(a, b Number) bool { return a == b }, "t"},
		{"not equal", []Value{Number(1), Number(2)}, func(a, b Number) bool { return a != b }, "t"},
		{"single arg", []Value{Number(1)}, func(a, b Number) bool { return a < b }, "t"},
		{"empty", []Value{}, func(a, b Number) bool { return a < b }, "t"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := compareNumbers(tt.args, tt.cmp)
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if Stringify(got) != tt.want {
				t.Fatalf("got %s, want %s", Stringify(got), tt.want)
			}
		})
	}
}

func TestUniqueBufferName(t *testing.T) {
	e := New()
	// First "tmp" should be "tmp"
	name := e.uniqueBufferName("tmp")
	if name != "tmp" {
		t.Fatalf("first = %q, want tmp", name)
	}
	// Create a buffer named "tmp"
	e.buffers["tmp"] = &Buffer{name: "tmp"}
	// Next should be "tmp<2>"
	name = e.uniqueBufferName("tmp")
	if name != "tmp<2>" {
		t.Fatalf("second = %q, want tmp<2>", name)
	}
}

func TestEnsureCurrentBuffer(t *testing.T) {
	e := New()
	e.current = nil
	buf := e.ensureCurrentBuffer()
	if buf == nil {
		t.Fatal("ensureCurrentBuffer returned nil")
	}
	if buf.name != "*scratch*" {
		t.Fatalf("name = %q, want *scratch*", buf.name)
	}
}
