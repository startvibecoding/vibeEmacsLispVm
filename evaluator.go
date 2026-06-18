package elispvm

import (
	"context"
	"fmt"
)

// BuiltinFunc is a Go function registered as a Lisp function. Arguments have
// already been evaluated.
type BuiltinFunc func(ctx *EvalContext, args []Value) (Value, error)

// SpecialForm is a Go function registered as a Lisp special form. Arguments are
// passed unevaluated so the form can control evaluation.
type SpecialForm func(ctx *EvalContext, unevaluated []Expr) (Value, error)

// Evaluator evaluates parsed expressions.
type Evaluator struct {
	funcs    map[string]BuiltinFunc
	specials map[string]SpecialForm
	globals  *Env
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
		funcs:    make(map[string]BuiltinFunc),
		specials: make(map[string]SpecialForm),
		globals:  NewEnv(nil),
	}
	e.registerCore()
	return e
}

// RegisterFunc registers or replaces a Lisp function.
func (e *Evaluator) RegisterFunc(name string, fn BuiltinFunc) {
	e.funcs[name] = fn
}

// RegisterSpecial registers or replaces a Lisp special form.
func (e *Evaluator) RegisterSpecial(name string, fn SpecialForm) {
	e.specials[name] = fn
}

// DefineGlobal binds a global variable.
func (e *Evaluator) DefineGlobal(name string, v Value) {
	e.globals.Define(name, v)
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
	head, ok := list[0].(Symbol)
	if !ok {
		return nil, fmt.Errorf("function position must be a symbol, got %s", Stringify(list[0]))
	}
	name := string(head)
	args := make([]Expr, len(list)-1)
	for i := range args {
		args[i] = list[i+1]
	}

	if special, ok := ctx.evaluator.specials[name]; ok {
		return special(ctx, args)
	}
	fn, ok := ctx.evaluator.funcs[name]
	if !ok {
		return nil, fmt.Errorf("undefined function %s", name)
	}

	evaluated := make([]Value, 0, len(args))
	for _, arg := range args {
		v, err := ctx.Eval(arg)
		if err != nil {
			return nil, err
		}
		evaluated = append(evaluated, v)
	}
	return fn(ctx, evaluated)
}
