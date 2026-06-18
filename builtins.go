package elispvm

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func (e *Evaluator) registerCore() {
	e.RegisterSpecial("quote", specialQuote)
	e.RegisterSpecial("progn", specialProgn)
	e.RegisterSpecial("let", specialLet)
	e.RegisterSpecial("setq", specialSetq)
	e.RegisterSpecial("if", specialIf)
	e.RegisterSpecial("when", specialWhen)
	e.RegisterSpecial("unless", specialUnless)
	e.RegisterSpecial("and", specialAnd)
	e.RegisterSpecial("or", specialOr)

	e.RegisterFunc("concat", builtinConcat)
	e.RegisterFunc("format", builtinFormat)
	e.RegisterFunc("list", builtinList)
	e.RegisterFunc("length", builtinLength)
	e.RegisterFunc("=", builtinNumEqual)
	e.RegisterFunc("<", builtinLess)
	e.RegisterFunc(">", builtinGreater)
	e.RegisterFunc("string=", builtinStringEqual)
	e.RegisterFunc("not", builtinNot)
}

func specialQuote(_ *EvalContext, args []Expr) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("quote expects 1 argument, got %d", len(args))
	}
	return args[0], nil
}

func specialProgn(ctx *EvalContext, args []Expr) (Value, error) {
	return ctx.EvalAll(args)
}

func specialLet(ctx *EvalContext, args []Expr) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("let expects bindings and body")
	}
	bindings, ok := args[0].(List)
	if !ok {
		return nil, fmt.Errorf("let bindings must be a list")
	}
	child := ctx.Child()
	for _, raw := range bindings {
		switch b := raw.(type) {
		case Symbol:
			child.Define(string(b), Nil)
		case List:
			if len(b) == 0 || len(b) > 2 {
				return nil, fmt.Errorf("let binding must be (name value) or name")
			}
			name, ok := b[0].(Symbol)
			if !ok {
				return nil, fmt.Errorf("let binding name must be a symbol")
			}
			var value Value = Nil
			if len(b) == 2 {
				v, err := ctx.Eval(b[1])
				if err != nil {
					return nil, err
				}
				value = v
			}
			child.Define(string(name), value)
		default:
			return nil, fmt.Errorf("let binding must be a symbol or list")
		}
	}
	return child.EvalAll(args[1:])
}

func specialSetq(ctx *EvalContext, args []Expr) (Value, error) {
	if len(args)%2 != 0 {
		return nil, fmt.Errorf("setq expects symbol/value pairs")
	}
	var result Value = Nil
	for i := 0; i < len(args); i += 2 {
		name, ok := args[i].(Symbol)
		if !ok {
			return nil, fmt.Errorf("setq target must be a symbol")
		}
		v, err := ctx.Eval(args[i+1])
		if err != nil {
			return nil, err
		}
		ctx.Set(string(name), v)
		result = v
	}
	return result, nil
}

func specialIf(ctx *EvalContext, args []Expr) (Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("if expects 2 or 3 arguments")
	}
	cond, err := ctx.Eval(args[0])
	if err != nil {
		return nil, err
	}
	if Truthy(cond) {
		return ctx.Eval(args[1])
	}
	if len(args) == 3 {
		return ctx.Eval(args[2])
	}
	return Nil, nil
}

func specialWhen(ctx *EvalContext, args []Expr) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("when expects condition and body")
	}
	cond, err := ctx.Eval(args[0])
	if err != nil {
		return nil, err
	}
	if Truthy(cond) {
		return ctx.EvalAll(args[1:])
	}
	return Nil, nil
}

func specialUnless(ctx *EvalContext, args []Expr) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("unless expects condition and body")
	}
	cond, err := ctx.Eval(args[0])
	if err != nil {
		return nil, err
	}
	if !Truthy(cond) {
		return ctx.EvalAll(args[1:])
	}
	return Nil, nil
}

func specialAnd(ctx *EvalContext, args []Expr) (Value, error) {
	var result Value = Symbol("t")
	for _, arg := range args {
		v, err := ctx.Eval(arg)
		if err != nil {
			return nil, err
		}
		if !Truthy(v) {
			return Nil, nil
		}
		result = v
	}
	return result, nil
}

func specialOr(ctx *EvalContext, args []Expr) (Value, error) {
	for _, arg := range args {
		v, err := ctx.Eval(arg)
		if err != nil {
			return nil, err
		}
		if Truthy(v) {
			return v, nil
		}
	}
	return Nil, nil
}

func builtinConcat(_ *EvalContext, args []Value) (Value, error) {
	var sb strings.Builder
	for _, arg := range args {
		sb.WriteString(rawString(arg))
	}
	return String(sb.String()), nil
}

func builtinFormat(_ *EvalContext, args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("format expects a format string")
	}
	format, ok := args[0].(String)
	if !ok {
		return nil, fmt.Errorf("format first argument must be a string")
	}
	out := string(format)
	for _, arg := range args[1:] {
		idx := strings.Index(out, "%s")
		if idx < 0 {
			break
		}
		out = out[:idx] + rawString(arg) + out[idx+2:]
	}
	out = strings.ReplaceAll(out, "%%", "%")
	return String(out), nil
}

func builtinList(_ *EvalContext, args []Value) (Value, error) {
	return List(args), nil
}

func builtinLength(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("length expects 1 argument")
	}
	switch v := args[0].(type) {
	case String:
		return Number(utf8.RuneCountInString(string(v))), nil
	case List:
		return Number(len(v)), nil
	default:
		return nil, fmt.Errorf("length unsupported for %s", Stringify(v))
	}
}

func builtinNumEqual(_ *EvalContext, args []Value) (Value, error) {
	if len(args) < 2 {
		return Symbol("t"), nil
	}
	first, err := asNumber(args[0])
	if err != nil {
		return nil, err
	}
	for _, arg := range args[1:] {
		n, err := asNumber(arg)
		if err != nil {
			return nil, err
		}
		if n != first {
			return Nil, nil
		}
	}
	return Symbol("t"), nil
}

func builtinLess(_ *EvalContext, args []Value) (Value, error) {
	return compareNumbers(args, func(a, b Number) bool { return a < b })
}

func builtinGreater(_ *EvalContext, args []Value) (Value, error) {
	return compareNumbers(args, func(a, b Number) bool { return a > b })
}

func builtinStringEqual(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("string= expects 2 arguments")
	}
	a, ok := args[0].(String)
	if !ok {
		return nil, fmt.Errorf("string= first argument must be a string")
	}
	b, ok := args[1].(String)
	if !ok {
		return nil, fmt.Errorf("string= second argument must be a string")
	}
	if a == b {
		return Symbol("t"), nil
	}
	return Nil, nil
}

func builtinNot(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("not expects 1 argument")
	}
	if Truthy(args[0]) {
		return Nil, nil
	}
	return Symbol("t"), nil
}

func compareNumbers(args []Value, cmp func(a, b Number) bool) (Value, error) {
	if len(args) < 2 {
		return Symbol("t"), nil
	}
	prev, err := asNumber(args[0])
	if err != nil {
		return nil, err
	}
	for _, arg := range args[1:] {
		next, err := asNumber(arg)
		if err != nil {
			return nil, err
		}
		if !cmp(prev, next) {
			return Nil, nil
		}
		prev = next
	}
	return Symbol("t"), nil
}

func asNumber(v Value) (Number, error) {
	n, ok := v.(Number)
	if !ok {
		return 0, fmt.Errorf("expected number, got %s", Stringify(v))
	}
	return n, nil
}

func rawString(v Value) string {
	switch x := v.(type) {
	case String:
		return string(x)
	case Symbol:
		return string(x)
	case NilType:
		return "nil"
	default:
		return Stringify(v)
	}
}
