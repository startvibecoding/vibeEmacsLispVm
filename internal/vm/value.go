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

// Function is a Lisp closure.
type Function struct {
	name   string
	params []string
	body   []Expr
	env    *Env
}

// Macro is a Lisp macro closure.
type Macro struct {
	name   string
	params []string
	body   []Expr
	env    *Env
}

// Buffer is an in-memory Emacs Lisp buffer subset.
type Buffer struct {
	name    string
	text    []rune
	point   int
	killed  bool
	markers []*Marker
}

// Marker points to a position in an in-memory buffer.
type Marker struct {
	buffer *Buffer
	pos    int
}

// NilType is the singleton nil value.
type NilType struct{}

// Nil is the Lisp nil value.
var Nil = NilType{}

// IsNil reports whether v is Lisp nil.
func IsNil(v Value) bool {
	if v == nil {
		return true
	}
	if _, ok := v.(NilType); ok {
		return true
	}
	list, ok := v.(List)
	return ok && len(list) == 0
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
		if len(x) == 0 {
			return "nil"
		}
		parts := make([]string, 0, len(x))
		for _, item := range x {
			parts = append(parts, Stringify(item))
		}
		return "(" + strings.Join(parts, " ") + ")"
	case []Value:
		if len(x) == 0 {
			return "nil"
		}
		parts := make([]string, 0, len(x))
		for _, item := range x {
			parts = append(parts, Stringify(item))
		}
		return "(" + strings.Join(parts, " ") + ")"
	case *Function:
		if x.name != "" {
			return fmt.Sprintf("#<function %s>", x.name)
		}
		return "#<function lambda>"
	case *Macro:
		if x.name != "" {
			return fmt.Sprintf("#<macro %s>", x.name)
		}
		return "#<macro>"
	case *Buffer:
		if x == nil || x.killed {
			return "#<killed buffer>"
		}
		return fmt.Sprintf("#<buffer %s>", x.name)
	case *Marker:
		if x == nil || x.buffer == nil || x.buffer.killed {
			return "#<marker in no buffer>"
		}
		return fmt.Sprintf("#<marker at %d in %s>", x.pos, x.buffer.name)
	default:
		return fmt.Sprintf("%v", v)
	}
}
