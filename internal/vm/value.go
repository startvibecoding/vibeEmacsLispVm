package vm

import (
	"fmt"
	"strconv"
	"strings"
)

// Value is a runtime value.
type Value interface{}

// Expr is a parsed expression. Parsed expressions and runtime values share the
// same concrete types for this small Lisp.
type Expr interface{}

// Symbol is an interned symbol name.
type Symbol string

// String is a Lisp string value.
type String string

// Number is a Lisp numeric value. MVP arithmetic uses float64 internally while
// preserving integer-looking formatting in Stringify.
type Number float64

// List is a Lisp list.
type List []Value

// NilType is the singleton nil value.
type NilType struct{}

// Nil is the Lisp nil value.
var Nil = NilType{}

// IsNil reports whether v is Lisp nil.
func IsNil(v Value) bool {
	_, ok := v.(NilType)
	return ok || v == nil
}

// Truthy implements Lisp truthiness: only nil is false.
func Truthy(v Value) bool {
	return !IsNil(v)
}

// Stringify returns a readable Lisp representation of v.
func Stringify(v Value) string {
	switch x := v.(type) {
	case nil:
		return "nil"
	case NilType:
		return "nil"
	case Symbol:
		return string(x)
	case String:
		return strconv.Quote(string(x))
	case Number:
		f := float64(x)
		if f == float64(int64(f)) {
			return fmt.Sprintf("%d", int64(f))
		}
		return strconv.FormatFloat(f, 'f', -1, 64)
	case List:
		parts := make([]string, 0, len(x))
		for _, item := range x {
			parts = append(parts, Stringify(item))
		}
		return "(" + strings.Join(parts, " ") + ")"
	case []Value:
		parts := make([]string, 0, len(x))
		for _, item := range x {
			parts = append(parts, Stringify(item))
		}
		return "(" + strings.Join(parts, " ") + ")"
	default:
		return fmt.Sprintf("%v", v)
	}
}
