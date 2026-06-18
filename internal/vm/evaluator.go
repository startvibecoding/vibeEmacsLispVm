package vm

import (
	"context"
	"fmt"
	"sort"
)

// BuiltinFunc is a Go function registered as a Lisp function. Arguments have
// already been evaluated.
type BuiltinFunc func(ctx *EvalContext, args []Value) (Value, error)

// SpecialForm is a Go function registered as a Lisp special form. Arguments are
// passed unevaluated so the form can control evaluation.
type SpecialForm func(ctx *EvalContext, unevaluated []Expr) (Value, error)

// Evaluator evaluates parsed expressions.
type Evaluator struct {
	funcs     map[string]BuiltinFunc
	lispFuncs map[string]*Function
	macros    map[string]*Macro
	specials  map[string]SpecialForm
	globals   *Env
	buffers   map[string]*Buffer
	current   *Buffer
}

// EvalContext carries execution context and lexical environment.
type EvalContext struct {
	Context context.Context

	evaluator *Evaluator
	env       *Env
}

// New creates an evaluator with core forms and builtins registered.
func New() *Evaluator {
	e := &Evaluator{
		funcs:     make(map[string]BuiltinFunc),
		lispFuncs: make(map[string]*Function),
		macros:    make(map[string]*Macro),
		specials:  make(map[string]SpecialForm),
		globals:   NewEnv(nil),
		buffers:   make(map[string]*Buffer),
	}
	scratch := &Buffer{name: "*scratch*", point: 1}
	e.buffers[scratch.name] = scratch
	e.current = scratch
	e.registerCore()
	return e
}

// RegisterFunc registers or replaces a Lisp function.
func (e *Evaluator) RegisterFunc(name string, fn BuiltinFunc) {
	e.funcs[name] = fn
	delete(e.lispFuncs, name)
}

// RegisterSpecial registers or replaces a Lisp special form.
func (e *Evaluator) RegisterSpecial(name string, fn SpecialForm) {
	e.specials[name] = fn
}

// DefineGlobal binds a global variable.
func (e *Evaluator) DefineGlobal(name string, v Value) {
	e.globals.Define(name, v)
}

// FuncNames returns the registered Lisp function names in sorted order.
func (e *Evaluator) FuncNames() []string {
	seen := make(map[string]struct{}, len(e.funcs)+len(e.lispFuncs))
	for name := range e.funcs {
		seen[name] = struct{}{}
	}
	for name := range e.lispFuncs {
		seen[name] = struct{}{}
	}
	return sortedKeys(seen)
}

// SpecialNames returns the registered special form names in sorted order.
func (e *Evaluator) SpecialNames() []string {
	return sortedKeys(e.specials)
}

// GlobalNames returns the defined global variable names in sorted order.
func (e *Evaluator) GlobalNames() []string {
	return e.globals.Names()
}

// EvalString parses and evaluates all top-level expressions in src, returning
// the last value.
func (e *Evaluator) EvalString(ctx context.Context, src string) (Value, error) {
	exprs, err := Parse(src)
	if err != nil {
		return nil, err
	}
	return e.EvalAll(ctx, exprs)
}

// EvalAll evaluates expressions in order, returning the last value.
func (e *Evaluator) EvalAll(ctx context.Context, exprs []Expr) (Value, error) {
	ec := &EvalContext{Context: ctx, evaluator: e, env: NewEnv(e.globals)}
	return ec.EvalAll(exprs)
}

// Eval evaluates one expression in the current lexical context.
func (ctx *EvalContext) Eval(expr Expr) (Value, error) {
	if ctx.Context != nil {
		select {
		case <-ctx.Context.Done():
			return nil, ctx.Context.Err()
		default:
		}
	}

	switch x := expr.(type) {
	case nil:
		return Nil, nil
	case NilType:
		return Nil, nil
	case String, Number:
		return x, nil
	case Symbol:
		name := string(x)
		switch name {
		case "nil":
			return Nil, nil
		case "t":
			return Symbol("t"), nil
		}
		if v, ok := ctx.env.Get(name); ok {
			return v, nil
		}
		return nil, fmt.Errorf("unbound symbol %s", name)
	case List:
		return ctx.evalList(x)
	default:
		return x, nil
	}
}

// EvalAll evaluates expressions in order in the current lexical context.
func (ctx *EvalContext) EvalAll(exprs []Expr) (Value, error) {
	var result Value = Nil
	for _, expr := range exprs {
		v, err := ctx.Eval(expr)
		if err != nil {
			return nil, err
		}
		result = v
	}
	return result, nil
}

// Child returns a child lexical context.
func (ctx *EvalContext) Child() *EvalContext {
	return &EvalContext{
		Context:   ctx.Context,
		evaluator: ctx.evaluator,
		env:       NewEnv(ctx.env),
	}
}

// Define binds a local variable in the current lexical environment.
func (ctx *EvalContext) Define(name string, v Value) {
	ctx.env.Define(name, v)
}

// Set updates a lexical binding.
func (ctx *EvalContext) Set(name string, v Value) {
	ctx.env.Set(name, v)
}

// Get returns a lexical binding.
func (ctx *EvalContext) Get(name string) (Value, bool) {
	return ctx.env.Get(name)
}

func (ctx *EvalContext) evalList(list List) (Value, error) {
	if len(list) == 0 {
		return Nil, nil
	}

	expanded, changed, err := ctx.macroexpand1(list)
	if err != nil {
		return nil, err
	}
	if changed {
		return ctx.Eval(expanded)
	}

	args := make([]Expr, len(list)-1)
	for i := range args {
		args[i] = list[i+1]
	}

	if head, ok := list[0].(Symbol); ok {
		name := string(head)
		if special, ok := ctx.evaluator.specials[name]; ok {
			return special(ctx, args)
		}

		evaluated, err := ctx.evalArgs(args)
		if err != nil {
			return nil, err
		}
		return ctx.callNamedFunction(name, evaluated)
	}

	fn, err := ctx.Eval(list[0])
	if err != nil {
		return nil, err
	}
	evaluated, err := ctx.evalArgs(args)
	if err != nil {
		return nil, err
	}
	return ctx.callFunctionDesignator(fn, evaluated)
}

func (ctx *EvalContext) macroexpand1(expr Expr) (Expr, bool, error) {
	list, ok := expr.(List)
	if !ok || len(list) == 0 {
		return expr, false, nil
	}
	head, ok := list[0].(Symbol)
	if !ok {
		return expr, false, nil
	}
	macro, ok := ctx.evaluator.macros[string(head)]
	if !ok {
		return expr, false, nil
	}
	args := make([]Value, len(list)-1)
	for i := range args {
		args[i] = list[i+1]
	}
	expanded, err := ctx.callMacro(macro, args)
	if err != nil {
		return nil, false, err
	}
	return expanded, true, nil
}

func (ctx *EvalContext) macroexpand(expr Expr) (Expr, error) {
	for {
		expanded, changed, err := ctx.macroexpand1(expr)
		if err != nil {
			return nil, err
		}
		if !changed {
			return expr, nil
		}
		expr = expanded
	}
}

func (ctx *EvalContext) evalArgs(args []Expr) ([]Value, error) {
	evaluated := make([]Value, 0, len(args))
	for _, arg := range args {
		v, err := ctx.Eval(arg)
		if err != nil {
			return nil, err
		}
		evaluated = append(evaluated, v)
	}
	return evaluated, nil
}

func (ctx *EvalContext) callNamedFunction(name string, args []Value) (Value, error) {
	if fn, ok := ctx.evaluator.lispFuncs[name]; ok {
		return ctx.callFunction(fn, args)
	}
	if fn, ok := ctx.evaluator.funcs[name]; ok {
		return fn(ctx, args)
	}
	return nil, fmt.Errorf("undefined function %s", name)
}

func (ctx *EvalContext) callFunctionDesignator(fn Value, args []Value) (Value, error) {
	switch x := fn.(type) {
	case *Function:
		return ctx.callFunction(x, args)
	case Symbol:
		return ctx.callNamedFunction(string(x), args)
	default:
		return nil, fmt.Errorf("invalid function %s", Stringify(fn))
	}
}

func (ctx *EvalContext) callFunction(fn *Function, args []Value) (Value, error) {
	if len(args) != len(fn.params) {
		name := "lambda"
		if fn.name != "" {
			name = fn.name
		}
		return nil, fmt.Errorf("%s expects %d arguments, got %d", name, len(fn.params), len(args))
	}
	callEnv := NewEnv(fn.env)
	for i, param := range fn.params {
		callEnv.Define(param, args[i])
	}
	callCtx := &EvalContext{
		Context:   ctx.Context,
		evaluator: ctx.evaluator,
		env:       callEnv,
	}
	return callCtx.EvalAll(fn.body)
}

func (ctx *EvalContext) callMacro(macro *Macro, args []Value) (Value, error) {
	if len(args) != len(macro.params) {
		name := "macro"
		if macro.name != "" {
			name = macro.name
		}
		return nil, fmt.Errorf("%s expects %d arguments, got %d", name, len(macro.params), len(args))
	}
	callEnv := NewEnv(macro.env)
	for i, param := range macro.params {
		callEnv.Define(param, args[i])
	}
	callCtx := &EvalContext{
		Context:   ctx.Context,
		evaluator: ctx.evaluator,
		env:       callEnv,
	}
	return callCtx.EvalAll(macro.body)
}

func sortedKeys[V any](m map[string]V) []string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
