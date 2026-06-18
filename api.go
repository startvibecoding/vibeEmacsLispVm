package elispvm

import "github.com/startvibecoding/vibeEmacsLispVm/internal/vm"

// BuiltinFunc is a Go function registered as a Lisp function. Arguments have
// already been evaluated.
type BuiltinFunc = vm.BuiltinFunc

// SpecialForm is a Go function registered as a Lisp special form. Arguments are
// passed unevaluated so the form can control evaluation.
type SpecialForm = vm.SpecialForm

// Evaluator evaluates parsed expressions.
type Evaluator = vm.Evaluator

// EvalContext carries execution context and lexical environment.
type EvalContext = vm.EvalContext

// Env is a lexical environment.
type Env = vm.Env

// Value is a runtime value.
type Value = vm.Value

// Expr is a parsed expression. Parsed expressions and runtime values share the
// same concrete types for this small Lisp.
type Expr = vm.Expr

// Symbol is an interned symbol name.
type Symbol = vm.Symbol

// String is a Lisp string value.
type String = vm.String

// Number is a Lisp numeric value.
type Number = vm.Number

// List is a Lisp list.
type List = vm.List

// NilType is the singleton nil value type.
type NilType = vm.NilType

// Nil is the Lisp nil value.
var Nil = vm.Nil

// Position identifies a location in source text.
type Position = vm.Position

// Error is a parse or evaluation error with optional source position.
type Error = vm.Error

// New creates an evaluator with core forms and builtins registered.
func New() *Evaluator {
	return vm.New()
}

// NewEnv creates an empty environment with an optional parent.
func NewEnv(parent *Env) *Env {
	return vm.NewEnv(parent)
}

// Parse parses all top-level expressions from src.
func Parse(src string) ([]Expr, error) {
	return vm.Parse(src)
}

// IsNil reports whether v is Lisp nil.
func IsNil(v Value) bool {
	return vm.IsNil(v)
}

// Truthy implements Lisp truthiness: only nil is false.
func Truthy(v Value) bool {
	return vm.Truthy(v)
}

// Stringify returns a readable Lisp representation of v.
func Stringify(v Value) string {
	return vm.Stringify(v)
}
