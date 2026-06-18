package vm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"
)

func (e *Evaluator) registerCore() {
	e.RegisterSpecial("quote", specialQuote)
	e.RegisterSpecial("backquote", specialBackquote)
	e.RegisterSpecial("comma", specialComma)
	e.RegisterSpecial("comma-splice", specialCommaSplice)
	e.RegisterSpecial("progn", specialProgn)
	e.RegisterSpecial("let", specialLet)
	e.RegisterSpecial("let*", specialLetStar)
	e.RegisterSpecial("setq", specialSetq)
	e.RegisterSpecial("if", specialIf)
	e.RegisterSpecial("when", specialWhen)
	e.RegisterSpecial("unless", specialUnless)
	e.RegisterSpecial("and", specialAnd)
	e.RegisterSpecial("or", specialOr)
	e.RegisterSpecial("while", specialWhile)
	e.RegisterSpecial("cond", specialCond)
	e.RegisterSpecial("catch", specialCatch)
	e.RegisterSpecial("throw", specialThrow)
	e.RegisterSpecial("lambda", specialLambda)
	e.RegisterSpecial("defun", specialDefun)
	e.RegisterSpecial("defmacro", specialDefmacro)
	e.RegisterSpecial("with-current-buffer", specialWithCurrentBuffer)
	e.RegisterSpecial("save-current-buffer", specialSaveCurrentBuffer)

	e.RegisterFunc("concat", builtinConcat)
	e.RegisterFunc("format", builtinFormat)
	e.RegisterFunc("list", builtinList)
	e.RegisterFunc("length", builtinLength)
	e.RegisterFunc("cons", builtinCons)
	e.RegisterFunc("car", builtinCar)
	e.RegisterFunc("cdr", builtinCdr)
	e.RegisterFunc("nth", builtinNth)
	e.RegisterFunc("append", builtinAppend)
	e.RegisterFunc("reverse", builtinReverse)
	e.RegisterFunc("member", builtinMember)
	e.RegisterFunc("assoc", builtinAssoc)
	e.RegisterFunc("funcall", builtinFuncall)
	e.RegisterFunc("apply", builtinApply)
	e.RegisterFunc("macroexpand-1", builtinMacroexpand1)
	e.RegisterFunc("macroexpand", builtinMacroexpand)
	e.RegisterFunc("+", builtinAdd)
	e.RegisterFunc("-", builtinSubtract)
	e.RegisterFunc("*", builtinMultiply)
	e.RegisterFunc("/", builtinDivide)
	e.RegisterFunc("=", builtinNumEqual)
	e.RegisterFunc("/=", builtinNumNotEqual)
	e.RegisterFunc("<", builtinLess)
	e.RegisterFunc("<=", builtinLessEqual)
	e.RegisterFunc(">", builtinGreater)
	e.RegisterFunc(">=", builtinGreaterEqual)
	e.RegisterFunc("eq", builtinEq)
	e.RegisterFunc("equal", builtinEqual)
	e.RegisterFunc("string=", builtinStringEqual)
	e.RegisterFunc("string-equal", builtinStringEqual)
	e.RegisterFunc("string-lessp", builtinStringLess)
	e.RegisterFunc("string<", builtinStringLess)
	e.RegisterFunc("string-greaterp", builtinStringGreater)
	e.RegisterFunc("string>", builtinStringGreater)
	e.RegisterFunc("not", builtinNot)
	e.RegisterFunc("null", builtinNull)
	e.RegisterFunc("symbolp", builtinSymbolp)
	e.RegisterFunc("stringp", builtinStringp)
	e.RegisterFunc("numberp", builtinNumberp)
	e.RegisterFunc("listp", builtinListp)
	e.RegisterFunc("consp", builtinConsp)
	e.RegisterFunc("atom", builtinAtom)
	e.RegisterFunc("bufferp", builtinBufferp)
	e.RegisterFunc("buffer-name", builtinBufferName)
	e.RegisterFunc("current-buffer", builtinCurrentBuffer)
	e.RegisterFunc("set-buffer", builtinSetBuffer)
	e.RegisterFunc("get-buffer", builtinGetBuffer)
	e.RegisterFunc("get-buffer-create", builtinGetBufferCreate)
	e.RegisterFunc("generate-new-buffer", builtinGenerateNewBuffer)
	e.RegisterFunc("kill-buffer", builtinKillBuffer)
	e.RegisterFunc("point", builtinPoint)
	e.RegisterFunc("point-min", builtinPointMin)
	e.RegisterFunc("point-max", builtinPointMax)
	e.RegisterFunc("goto-char", builtinGotoChar)
	e.RegisterFunc("insert", builtinInsert)
	e.RegisterFunc("delete-region", builtinDeleteRegion)
	e.RegisterFunc("buffer-substring", builtinBufferSubstring)
	e.RegisterFunc("buffer-string", builtinBufferString)
	e.RegisterFunc("erase-buffer", builtinEraseBuffer)
	e.RegisterFunc("markerp", builtinMarkerp)
	e.RegisterFunc("make-marker", builtinMakeMarker)
	e.RegisterFunc("point-marker", builtinPointMarker)
	e.RegisterFunc("copy-marker", builtinCopyMarker)
	e.RegisterFunc("marker-position", builtinMarkerPosition)
	e.RegisterFunc("marker-buffer", builtinMarkerBuffer)
	e.RegisterFunc("set-marker", builtinSetMarker)
}

type throwSignal struct {
	tag   Value
	value Value
}

func (s *throwSignal) Error() string {
	return fmt.Sprintf("uncaught throw %s", Stringify(s.tag))
}

func specialQuote(_ *EvalContext, args []Expr) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("quote expects 1 argument, got %d", len(args))
	}
	return args[0], nil
}

func specialBackquote(ctx *EvalContext, args []Expr) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("backquote expects 1 argument, got %d", len(args))
	}
	return evalBackquote(ctx, args[0])
}

func specialComma(_ *EvalContext, args []Expr) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("comma expects 1 argument, got %d", len(args))
	}
	return nil, fmt.Errorf("comma not inside backquote")
}

func specialCommaSplice(_ *EvalContext, args []Expr) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("comma-splice expects 1 argument, got %d", len(args))
	}
	return nil, fmt.Errorf("comma-splice not inside backquote")
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

func specialLetStar(ctx *EvalContext, args []Expr) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("let* expects bindings and body")
	}
	bindings, ok := args[0].(List)
	if !ok {
		return nil, fmt.Errorf("let* bindings must be a list")
	}
	child := ctx.Child()
	for _, raw := range bindings {
		switch b := raw.(type) {
		case Symbol:
			child.Define(string(b), Nil)
		case List:
			if len(b) == 0 || len(b) > 2 {
				return nil, fmt.Errorf("let* binding must be (name value) or name")
			}
			name, ok := b[0].(Symbol)
			if !ok {
				return nil, fmt.Errorf("let* binding name must be a symbol")
			}
			var value Value = Nil
			if len(b) == 2 {
				v, err := child.Eval(b[1])
				if err != nil {
					return nil, err
				}
				value = v
			}
			child.Define(string(name), value)
		default:
			return nil, fmt.Errorf("let* binding must be a symbol or list")
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

func specialWhile(ctx *EvalContext, args []Expr) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("while expects condition and body")
	}
	for {
		cond, err := ctx.Eval(args[0])
		if err != nil {
			return nil, err
		}
		if !Truthy(cond) {
			return Nil, nil
		}
		if _, err := ctx.EvalAll(args[1:]); err != nil {
			return nil, err
		}
	}
}

func specialCond(ctx *EvalContext, args []Expr) (Value, error) {
	for _, arg := range args {
		clause, ok := arg.(List)
		if !ok {
			return nil, fmt.Errorf("cond clause must be a list")
		}
		if len(clause) == 0 {
			return nil, fmt.Errorf("cond clause must not be empty")
		}
		test, err := ctx.Eval(clause[0])
		if err != nil {
			return nil, err
		}
		if Truthy(test) {
			if len(clause) == 1 {
				return test, nil
			}
			body := make([]Expr, len(clause)-1)
			for i := range body {
				body[i] = clause[i+1]
			}
			return ctx.EvalAll(body)
		}
	}
	return Nil, nil
}

func specialCatch(ctx *EvalContext, args []Expr) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("catch expects tag and body")
	}
	tag, err := ctx.Eval(args[0])
	if err != nil {
		return nil, err
	}
	v, err := ctx.EvalAll(args[1:])
	if err == nil {
		return v, nil
	}
	var signal *throwSignal
	if errors.As(err, &signal) && eqValues(signal.tag, tag) {
		return signal.value, nil
	}
	return nil, err
}

func specialThrow(ctx *EvalContext, args []Expr) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("throw expects tag and value")
	}
	tag, err := ctx.Eval(args[0])
	if err != nil {
		return nil, err
	}
	value, err := ctx.Eval(args[1])
	if err != nil {
		return nil, err
	}
	return nil, &throwSignal{tag: tag, value: value}
}

func specialLambda(ctx *EvalContext, args []Expr) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("lambda expects argument list and body")
	}
	params, err := parseParams(args[0], "lambda")
	if err != nil {
		return nil, err
	}
	body := make([]Expr, len(args)-1)
	copy(body, args[1:])
	return &Function{params: params, body: body, env: ctx.env}, nil
}

func specialDefun(ctx *EvalContext, args []Expr) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("defun expects name, argument list, and body")
	}
	name, ok := args[0].(Symbol)
	if !ok {
		return nil, fmt.Errorf("defun name must be a symbol")
	}
	params, err := parseParams(args[1], "defun")
	if err != nil {
		return nil, err
	}
	body := make([]Expr, len(args)-2)
	copy(body, args[2:])
	fn := &Function{name: string(name), params: params, body: body, env: ctx.env}
	ctx.evaluator.lispFuncs[string(name)] = fn
	return name, nil
}

func specialDefmacro(ctx *EvalContext, args []Expr) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("defmacro expects name, argument list, and body")
	}
	name, ok := args[0].(Symbol)
	if !ok {
		return nil, fmt.Errorf("defmacro name must be a symbol")
	}
	params, err := parseParams(args[1], "defmacro")
	if err != nil {
		return nil, err
	}
	body := make([]Expr, len(args)-2)
	copy(body, args[2:])
	macro := &Macro{name: string(name), params: params, body: body, env: ctx.env}
	ctx.evaluator.macros[string(name)] = macro
	return name, nil
}

func specialWithCurrentBuffer(ctx *EvalContext, args []Expr) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("with-current-buffer expects buffer and body")
	}
	bufferValue, err := ctx.Eval(args[0])
	if err != nil {
		return nil, err
	}
	buffer, err := ctx.asLiveBufferDesignator(bufferValue, "with-current-buffer argument")
	if err != nil {
		return nil, err
	}
	prev := ctx.evaluator.current
	ctx.evaluator.current = buffer
	defer func() {
		if prev != nil && !prev.killed {
			ctx.evaluator.current = prev
		} else {
			ctx.evaluator.current = ctx.evaluator.ensureCurrentBuffer()
		}
	}()
	return ctx.EvalAll(args[1:])
}

func specialSaveCurrentBuffer(ctx *EvalContext, args []Expr) (Value, error) {
	prev := ctx.evaluator.current
	defer func() {
		if prev != nil && !prev.killed {
			ctx.evaluator.current = prev
		} else {
			ctx.evaluator.current = ctx.evaluator.ensureCurrentBuffer()
		}
	}()
	return ctx.EvalAll(args)
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

func builtinCons(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("cons expects 2 arguments")
	}
	tail, err := asList(args[1], "cons second argument")
	if err != nil {
		return nil, err
	}
	out := make(List, 0, len(tail)+1)
	out = append(out, args[0])
	out = append(out, tail...)
	return out, nil
}

func builtinCar(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("car expects 1 argument")
	}
	list, err := asList(args[0], "car argument")
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return Nil, nil
	}
	return list[0], nil
}

func builtinCdr(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("cdr expects 1 argument")
	}
	list, err := asList(args[0], "cdr argument")
	if err != nil {
		return nil, err
	}
	if len(list) <= 1 {
		return Nil, nil
	}
	out := make(List, len(list)-1)
	copy(out, list[1:])
	return out, nil
}

func builtinNth(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("nth expects 2 arguments")
	}
	n, err := asNumber(args[0])
	if err != nil {
		return nil, err
	}
	index := int(n)
	if Number(index) != n {
		return nil, fmt.Errorf("nth index must be an integer")
	}
	list, err := asList(args[1], "nth second argument")
	if err != nil {
		return nil, err
	}
	if index < 0 || index >= len(list) {
		return Nil, nil
	}
	return list[index], nil
}

func builtinAppend(_ *EvalContext, args []Value) (Value, error) {
	var out List
	for i, arg := range args {
		list, err := asList(arg, fmt.Sprintf("append argument %d", i+1))
		if err != nil {
			return nil, err
		}
		out = append(out, list...)
	}
	if len(out) == 0 {
		return Nil, nil
	}
	return out, nil
}

func builtinReverse(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("reverse expects 1 argument")
	}
	list, err := asList(args[0], "reverse argument")
	if err != nil {
		return nil, err
	}
	out := make(List, len(list))
	for i := range list {
		out[len(list)-1-i] = list[i]
	}
	if len(out) == 0 {
		return Nil, nil
	}
	return out, nil
}

func builtinMember(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("member expects 2 arguments")
	}
	list, err := asList(args[1], "member second argument")
	if err != nil {
		return nil, err
	}
	for i, item := range list {
		if equalValues(args[0], item) {
			out := make(List, len(list)-i)
			copy(out, list[i:])
			return out, nil
		}
	}
	return Nil, nil
}

func builtinAssoc(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("assoc expects 2 arguments")
	}
	list, err := asList(args[1], "assoc second argument")
	if err != nil {
		return nil, err
	}
	for _, item := range list {
		pair, ok := item.(List)
		if !ok || len(pair) == 0 {
			continue
		}
		if equalValues(args[0], pair[0]) {
			out := make(List, len(pair))
			copy(out, pair)
			return out, nil
		}
	}
	return Nil, nil
}

func builtinFuncall(ctx *EvalContext, args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("funcall expects function and arguments")
	}
	return ctx.callFunctionDesignator(args[0], args[1:])
}

func builtinApply(ctx *EvalContext, args []Value) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("apply expects function and argument list")
	}
	tail, err := asList(args[len(args)-1], "apply last argument")
	if err != nil {
		return nil, err
	}
	callArgs := make([]Value, 0, len(args)-2+len(tail))
	callArgs = append(callArgs, args[1:len(args)-1]...)
	callArgs = append(callArgs, tail...)
	return ctx.callFunctionDesignator(args[0], callArgs)
}

func builtinMacroexpand1(ctx *EvalContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("macroexpand-1 expects 1 argument")
	}
	expanded, _, err := ctx.macroexpand1(args[0])
	if err != nil {
		return nil, err
	}
	return expanded, nil
}

func builtinMacroexpand(ctx *EvalContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("macroexpand expects 1 argument")
	}
	return ctx.macroexpand(args[0])
}

func builtinAdd(_ *EvalContext, args []Value) (Value, error) {
	var sum Number
	for _, arg := range args {
		n, err := asNumber(arg)
		if err != nil {
			return nil, err
		}
		sum += n
	}
	return sum, nil
}

func builtinSubtract(_ *EvalContext, args []Value) (Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("- expects at least 1 argument")
	}
	result, err := asNumber(args[0])
	if err != nil {
		return nil, err
	}
	if len(args) == 1 {
		return -result, nil
	}
	for _, arg := range args[1:] {
		n, err := asNumber(arg)
		if err != nil {
			return nil, err
		}
		result -= n
	}
	return result, nil
}

func builtinMultiply(_ *EvalContext, args []Value) (Value, error) {
	result := Number(1)
	for _, arg := range args {
		n, err := asNumber(arg)
		if err != nil {
			return nil, err
		}
		result *= n
	}
	return result, nil
}

func builtinDivide(_ *EvalContext, args []Value) (Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("/ expects at least 1 argument")
	}
	result, err := asNumber(args[0])
	if err != nil {
		return nil, err
	}
	if len(args) == 1 {
		if result == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return Number(1) / result, nil
	}
	for _, arg := range args[1:] {
		n, err := asNumber(arg)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result /= n
	}
	return result, nil
}

func builtinNumEqual(_ *EvalContext, args []Value) (Value, error) {
	if len(args) < 2 {
		return lispBool(true), nil
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
			return lispBool(false), nil
		}
	}
	return lispBool(true), nil
}

func builtinNumNotEqual(_ *EvalContext, args []Value) (Value, error) {
	if len(args) < 2 {
		return lispBool(true), nil
	}
	seen := make(map[Number]struct{}, len(args))
	for _, arg := range args {
		n, err := asNumber(arg)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[n]; ok {
			return lispBool(false), nil
		}
		seen[n] = struct{}{}
	}
	return lispBool(true), nil
}

func builtinLess(_ *EvalContext, args []Value) (Value, error) {
	return compareNumbers(args, func(a, b Number) bool { return a < b })
}

func builtinLessEqual(_ *EvalContext, args []Value) (Value, error) {
	return compareNumbers(args, func(a, b Number) bool { return a <= b })
}

func builtinGreater(_ *EvalContext, args []Value) (Value, error) {
	return compareNumbers(args, func(a, b Number) bool { return a > b })
}

func builtinGreaterEqual(_ *EvalContext, args []Value) (Value, error) {
	return compareNumbers(args, func(a, b Number) bool { return a >= b })
}

func builtinEq(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("eq expects 2 arguments")
	}
	return lispBool(eqValues(args[0], args[1])), nil
}

func builtinEqual(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("equal expects 2 arguments")
	}
	return lispBool(equalValues(args[0], args[1])), nil
}

func builtinStringEqual(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("string= expects 2 arguments")
	}
	a, err := asStringDesignator(args[0])
	if err != nil {
		return nil, fmt.Errorf("string= first argument %w", err)
	}
	b, err := asStringDesignator(args[1])
	if err != nil {
		return nil, fmt.Errorf("string= second argument %w", err)
	}
	return lispBool(a == b), nil
}

func builtinStringLess(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("string< expects 2 arguments")
	}
	a, err := asStringDesignator(args[0])
	if err != nil {
		return nil, fmt.Errorf("string< first argument %w", err)
	}
	b, err := asStringDesignator(args[1])
	if err != nil {
		return nil, fmt.Errorf("string< second argument %w", err)
	}
	return lispBool(a < b), nil
}

func builtinStringGreater(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("string> expects 2 arguments")
	}
	a, err := asStringDesignator(args[0])
	if err != nil {
		return nil, fmt.Errorf("string> first argument %w", err)
	}
	b, err := asStringDesignator(args[1])
	if err != nil {
		return nil, fmt.Errorf("string> second argument %w", err)
	}
	return lispBool(a > b), nil
}

func builtinNot(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("not expects 1 argument")
	}
	if Truthy(args[0]) {
		return Nil, nil
	}
	return lispBool(true), nil
}

func builtinNull(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("null expects 1 argument")
	}
	return lispBool(IsNil(args[0])), nil
}

func builtinSymbolp(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("symbolp expects 1 argument")
	}
	_, ok := args[0].(Symbol)
	return lispBool(ok || IsNil(args[0])), nil
}

func builtinStringp(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("stringp expects 1 argument")
	}
	_, ok := args[0].(String)
	return lispBool(ok), nil
}

func builtinNumberp(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("numberp expects 1 argument")
	}
	_, ok := args[0].(Number)
	return lispBool(ok), nil
}

func builtinListp(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("listp expects 1 argument")
	}
	_, ok := args[0].(List)
	return lispBool(ok || IsNil(args[0])), nil
}

func builtinConsp(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("consp expects 1 argument")
	}
	list, ok := args[0].(List)
	return lispBool(ok && len(list) > 0), nil
}

func builtinAtom(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("atom expects 1 argument")
	}
	list, ok := args[0].(List)
	return lispBool(!ok || len(list) == 0), nil
}

func builtinBufferp(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("bufferp expects 1 argument")
	}
	_, ok := args[0].(*Buffer)
	return lispBool(ok), nil
}

func builtinBufferName(ctx *EvalContext, args []Value) (Value, error) {
	if len(args) > 1 {
		return nil, fmt.Errorf("buffer-name expects 0 or 1 arguments")
	}
	var buffer *Buffer
	var err error
	if len(args) == 0 {
		buffer = ctx.evaluator.ensureCurrentBuffer()
	} else {
		buffer, err = ctx.asLiveBufferDesignator(args[0], "buffer-name argument")
		if err != nil {
			return nil, err
		}
	}
	return String(buffer.name), nil
}

func builtinCurrentBuffer(ctx *EvalContext, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("current-buffer expects 0 arguments")
	}
	return ctx.evaluator.ensureCurrentBuffer(), nil
}

func builtinSetBuffer(ctx *EvalContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("set-buffer expects 1 argument")
	}
	buffer, err := ctx.asLiveBufferDesignator(args[0], "set-buffer argument")
	if err != nil {
		return nil, err
	}
	ctx.evaluator.current = buffer
	return buffer, nil
}

func builtinGetBuffer(ctx *EvalContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("get-buffer expects 1 argument")
	}
	switch v := args[0].(type) {
	case *Buffer:
		if v == nil || v.killed {
			return Nil, nil
		}
		return v, nil
	default:
		name, err := asStringDesignator(v)
		if err != nil {
			return nil, fmt.Errorf("get-buffer argument %w", err)
		}
		buffer := ctx.evaluator.buffers[name]
		if buffer == nil || buffer.killed {
			return Nil, nil
		}
		return buffer, nil
	}
}

func builtinGetBufferCreate(ctx *EvalContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("get-buffer-create expects 1 argument")
	}
	name, err := asStringDesignator(args[0])
	if err != nil {
		return nil, fmt.Errorf("get-buffer-create argument %w", err)
	}
	buffer := ctx.evaluator.buffers[name]
	if buffer != nil && !buffer.killed {
		return buffer, nil
	}
	return ctx.evaluator.createBuffer(name), nil
}

func builtinGenerateNewBuffer(ctx *EvalContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("generate-new-buffer expects 1 argument")
	}
	base, err := asStringDesignator(args[0])
	if err != nil {
		return nil, fmt.Errorf("generate-new-buffer argument %w", err)
	}
	return ctx.evaluator.createBuffer(ctx.evaluator.uniqueBufferName(base)), nil
}

func builtinKillBuffer(ctx *EvalContext, args []Value) (Value, error) {
	if len(args) > 1 {
		return nil, fmt.Errorf("kill-buffer expects 0 or 1 arguments")
	}
	var buffer *Buffer
	var err error
	if len(args) == 0 {
		buffer = ctx.evaluator.ensureCurrentBuffer()
	} else {
		buffer, err = ctx.asLiveBufferDesignator(args[0], "kill-buffer argument")
		if err != nil {
			return nil, err
		}
	}
	delete(ctx.evaluator.buffers, buffer.name)
	buffer.killed = true
	for _, marker := range buffer.markers {
		marker.buffer = nil
		marker.pos = 0
	}
	buffer.markers = nil
	if ctx.evaluator.current == buffer {
		ctx.evaluator.current = ctx.evaluator.ensureCurrentBuffer()
	}
	return lispBool(true), nil
}

func builtinPoint(ctx *EvalContext, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("point expects 0 arguments")
	}
	return Number(ctx.evaluator.ensureCurrentBuffer().point), nil
}

func builtinPointMin(ctx *EvalContext, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("point-min expects 0 arguments")
	}
	ctx.evaluator.ensureCurrentBuffer()
	return Number(1), nil
}

func builtinPointMax(ctx *EvalContext, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("point-max expects 0 arguments")
	}
	buffer := ctx.evaluator.ensureCurrentBuffer()
	return Number(len(buffer.text) + 1), nil
}

func builtinGotoChar(ctx *EvalContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("goto-char expects 1 argument")
	}
	pos, err := asPosition(args[0], "goto-char argument")
	if err != nil {
		return nil, err
	}
	buffer := ctx.evaluator.ensureCurrentBuffer()
	if err := validateBufferPosition(buffer, pos, true, "goto-char argument"); err != nil {
		return nil, err
	}
	buffer.point = pos
	return Number(pos), nil
}

func builtinInsert(ctx *EvalContext, args []Value) (Value, error) {
	buffer := ctx.evaluator.ensureCurrentBuffer()
	for _, arg := range args {
		text := rawString(arg)
		insertText(buffer, buffer.point, []rune(text))
	}
	return Nil, nil
}

func builtinDeleteRegion(ctx *EvalContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("delete-region expects 2 arguments")
	}
	start, err := asPosition(args[0], "delete-region start")
	if err != nil {
		return nil, err
	}
	end, err := asPosition(args[1], "delete-region end")
	if err != nil {
		return nil, err
	}
	buffer := ctx.evaluator.ensureCurrentBuffer()
	if err := validateRegion(buffer, start, end, "delete-region"); err != nil {
		return nil, err
	}
	deleteText(buffer, start, end)
	return Nil, nil
}

func builtinBufferSubstring(ctx *EvalContext, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("buffer-substring expects 2 arguments")
	}
	start, err := asPosition(args[0], "buffer-substring start")
	if err != nil {
		return nil, err
	}
	end, err := asPosition(args[1], "buffer-substring end")
	if err != nil {
		return nil, err
	}
	buffer := ctx.evaluator.ensureCurrentBuffer()
	if err := validateRegion(buffer, start, end, "buffer-substring"); err != nil {
		return nil, err
	}
	return String(string(buffer.text[start-1 : end-1])), nil
}

func builtinBufferString(ctx *EvalContext, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("buffer-string expects 0 arguments")
	}
	buffer := ctx.evaluator.ensureCurrentBuffer()
	return String(string(buffer.text)), nil
}

func builtinEraseBuffer(ctx *EvalContext, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("erase-buffer expects 0 arguments")
	}
	buffer := ctx.evaluator.ensureCurrentBuffer()
	deleteText(buffer, 1, len(buffer.text)+1)
	return Nil, nil
}

func builtinMarkerp(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("markerp expects 1 argument")
	}
	_, ok := args[0].(*Marker)
	return lispBool(ok), nil
}

func builtinMakeMarker(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("make-marker expects 0 arguments")
	}
	return &Marker{}, nil
}

func builtinPointMarker(ctx *EvalContext, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("point-marker expects 0 arguments")
	}
	buffer := ctx.evaluator.ensureCurrentBuffer()
	return newMarker(buffer, buffer.point), nil
}

func builtinCopyMarker(ctx *EvalContext, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("copy-marker expects 1 or 2 arguments")
	}
	switch v := args[0].(type) {
	case *Marker:
		if v == nil || v.buffer == nil || v.buffer.killed {
			return &Marker{}, nil
		}
		return newMarker(v.buffer, v.pos), nil
	default:
		pos, err := asPosition(v, "copy-marker argument")
		if err != nil {
			return nil, err
		}
		buffer := ctx.evaluator.ensureCurrentBuffer()
		if err := validateBufferPosition(buffer, pos, true, "copy-marker argument"); err != nil {
			return nil, err
		}
		return newMarker(buffer, pos), nil
	}
}

func builtinMarkerPosition(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("marker-position expects 1 argument")
	}
	marker, ok := args[0].(*Marker)
	if !ok {
		return nil, fmt.Errorf("marker-position argument must be a marker")
	}
	if marker.buffer == nil || marker.buffer.killed {
		return Nil, nil
	}
	return Number(marker.pos), nil
}

func builtinMarkerBuffer(_ *EvalContext, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("marker-buffer expects 1 argument")
	}
	marker, ok := args[0].(*Marker)
	if !ok {
		return nil, fmt.Errorf("marker-buffer argument must be a marker")
	}
	if marker.buffer == nil || marker.buffer.killed {
		return Nil, nil
	}
	return marker.buffer, nil
}

func builtinSetMarker(ctx *EvalContext, args []Value) (Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("set-marker expects 2 or 3 arguments")
	}
	marker, ok := args[0].(*Marker)
	if !ok {
		return nil, fmt.Errorf("set-marker first argument must be a marker")
	}
	if IsNil(args[1]) {
		detachMarker(marker)
		return marker, nil
	}
	pos, err := asPosition(args[1], "set-marker position")
	if err != nil {
		return nil, err
	}
	var buffer *Buffer
	if len(args) == 3 {
		buffer, err = ctx.asLiveBufferDesignator(args[2], "set-marker buffer")
		if err != nil {
			return nil, err
		}
	} else {
		buffer = ctx.evaluator.ensureCurrentBuffer()
	}
	if err := validateBufferPosition(buffer, pos, true, "set-marker position"); err != nil {
		return nil, err
	}
	attachMarker(marker, buffer, pos)
	return marker, nil
}

func compareNumbers(args []Value, cmp func(a, b Number) bool) (Value, error) {
	if len(args) < 2 {
		return lispBool(true), nil
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
			return lispBool(false), nil
		}
		prev = next
	}
	return lispBool(true), nil
}

func lispBool(ok bool) Value {
	if ok {
		return Symbol("t")
	}
	return Nil
}

func asNumber(v Value) (Number, error) {
	n, ok := v.(Number)
	if !ok {
		return 0, fmt.Errorf("expected number, got %s", Stringify(v))
	}
	return n, nil
}

func asStringDesignator(v Value) (string, error) {
	switch x := v.(type) {
	case String:
		return string(x), nil
	case Symbol:
		return string(x), nil
	default:
		return "", fmt.Errorf("must be a string or symbol")
	}
}

func asList(v Value, name string) (List, error) {
	if IsNil(v) {
		return nil, nil
	}
	list, ok := v.(List)
	if !ok {
		return nil, fmt.Errorf("%s must be a list", name)
	}
	return list, nil
}

func (e *Evaluator) ensureCurrentBuffer() *Buffer {
	if e.current != nil && !e.current.killed {
		return e.current
	}
	if buffer := e.buffers["*scratch*"]; buffer != nil && !buffer.killed {
		e.current = buffer
		return buffer
	}
	e.current = e.createBuffer("*scratch*")
	return e.current
}

func (e *Evaluator) createBuffer(name string) *Buffer {
	buffer := &Buffer{name: name, point: 1}
	e.buffers[name] = buffer
	return buffer
}

func (e *Evaluator) uniqueBufferName(base string) string {
	if buffer := e.buffers[base]; buffer == nil || buffer.killed {
		return base
	}
	for i := 2; ; i++ {
		name := fmt.Sprintf("%s<%d>", base, i)
		if buffer := e.buffers[name]; buffer == nil || buffer.killed {
			return name
		}
	}
}

func (ctx *EvalContext) asLiveBufferDesignator(v Value, name string) (*Buffer, error) {
	switch x := v.(type) {
	case *Buffer:
		if x == nil || x.killed {
			return nil, fmt.Errorf("%s must be a live buffer", name)
		}
		return x, nil
	default:
		bufferName, err := asStringDesignator(v)
		if err != nil {
			return nil, fmt.Errorf("%s must be a buffer or buffer name", name)
		}
		buffer := ctx.evaluator.buffers[bufferName]
		if buffer == nil || buffer.killed {
			return nil, fmt.Errorf("buffer %q does not exist", bufferName)
		}
		return buffer, nil
	}
}

func asPosition(v Value, name string) (int, error) {
	switch x := v.(type) {
	case Number:
		pos := int(x)
		if Number(pos) != x {
			return 0, fmt.Errorf("%s must be an integer", name)
		}
		return pos, nil
	case *Marker:
		if x == nil || x.buffer == nil || x.buffer.killed {
			return 0, fmt.Errorf("%s marker has no buffer", name)
		}
		return x.pos, nil
	default:
		return 0, fmt.Errorf("%s must be a number or marker", name)
	}
}

func validateBufferPosition(buffer *Buffer, pos int, allowEnd bool, name string) error {
	max := len(buffer.text) + 1
	if allowEnd {
		if pos < 1 || pos > max {
			return fmt.Errorf("%s out of range", name)
		}
		return nil
	}
	if pos < 1 || pos >= max {
		return fmt.Errorf("%s out of range", name)
	}
	return nil
}

func validateRegion(buffer *Buffer, start, end int, name string) error {
	max := len(buffer.text) + 1
	if start < 1 || end < start || end > max {
		return fmt.Errorf("%s region out of range", name)
	}
	return nil
}

func insertText(buffer *Buffer, pos int, text []rune) {
	if len(text) == 0 {
		return
	}
	idx := pos - 1
	out := make([]rune, 0, len(buffer.text)+len(text))
	out = append(out, buffer.text[:idx]...)
	out = append(out, text...)
	out = append(out, buffer.text[idx:]...)
	buffer.text = out
	shiftPositions(buffer, pos, pos, len(text))
}

func deleteText(buffer *Buffer, start, end int) {
	if start == end {
		return
	}
	idxStart := start - 1
	idxEnd := end - 1
	deleted := end - start
	out := make([]rune, 0, len(buffer.text)-deleted)
	out = append(out, buffer.text[:idxStart]...)
	out = append(out, buffer.text[idxEnd:]...)
	buffer.text = out
	shiftPositions(buffer, start, end, -deleted)
}

func shiftPositions(buffer *Buffer, start, end, delta int) {
	buffer.point = shiftedPosition(buffer.point, start, end, delta)
	for _, marker := range buffer.markers {
		if marker.buffer == buffer {
			marker.pos = shiftedPosition(marker.pos, start, end, delta)
		}
	}
}

func shiftedPosition(pos, start, end, delta int) int {
	if delta > 0 {
		if pos >= start {
			return pos + delta
		}
		return pos
	}
	if pos >= end {
		return pos + delta
	}
	if pos > start {
		return start
	}
	return pos
}

func newMarker(buffer *Buffer, pos int) *Marker {
	marker := &Marker{}
	attachMarker(marker, buffer, pos)
	return marker
}

func attachMarker(marker *Marker, buffer *Buffer, pos int) {
	detachMarker(marker)
	marker.buffer = buffer
	marker.pos = pos
	buffer.markers = append(buffer.markers, marker)
}

func detachMarker(marker *Marker) {
	if marker == nil || marker.buffer == nil {
		return
	}
	buffer := marker.buffer
	for i, item := range buffer.markers {
		if item == marker {
			buffer.markers = append(buffer.markers[:i], buffer.markers[i+1:]...)
			break
		}
	}
	marker.buffer = nil
	marker.pos = 0
}

func parseParams(expr Expr, form string) ([]string, error) {
	rawParams, ok := expr.(List)
	if !ok {
		return nil, fmt.Errorf("%s argument list must be a list", form)
	}
	params := make([]string, 0, len(rawParams))
	seen := make(map[string]struct{}, len(rawParams))
	for _, raw := range rawParams {
		param, ok := raw.(Symbol)
		if !ok {
			return nil, fmt.Errorf("%s argument name must be a symbol", form)
		}
		name := string(param)
		if strings.HasPrefix(name, "&") {
			return nil, fmt.Errorf("%s only supports fixed arguments", form)
		}
		if _, ok := seen[name]; ok {
			return nil, fmt.Errorf("%s duplicate argument %s", form, name)
		}
		seen[name] = struct{}{}
		params = append(params, name)
	}
	return params, nil
}

func evalBackquote(ctx *EvalContext, expr Expr) (Value, error) {
	if list, ok := expr.(List); ok {
		if isForm(list, "comma") {
			if len(list) != 2 {
				return nil, fmt.Errorf("comma expects 1 argument")
			}
			return ctx.Eval(list[1])
		}
		if isForm(list, "comma-splice") {
			return nil, fmt.Errorf("comma-splice not inside list")
		}
		out := make(List, 0, len(list))
		for _, item := range list {
			itemList, ok := item.(List)
			if ok && isForm(itemList, "comma-splice") {
				if len(itemList) != 2 {
					return nil, fmt.Errorf("comma-splice expects 1 argument")
				}
				v, err := ctx.Eval(itemList[1])
				if err != nil {
					return nil, err
				}
				splice, err := asList(v, "comma-splice value")
				if err != nil {
					return nil, err
				}
				out = append(out, splice...)
				continue
			}
			v, err := evalBackquote(ctx, item)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		if len(out) == 0 {
			return Nil, nil
		}
		return out, nil
	}
	return expr, nil
}

func isForm(list List, name string) bool {
	if len(list) == 0 {
		return false
	}
	head, ok := list[0].(Symbol)
	return ok && string(head) == name
}

func eqValues(a, b Value) bool {
	if IsNil(a) && IsNil(b) {
		return true
	}
	switch x := a.(type) {
	case Symbol:
		y, ok := b.(Symbol)
		return ok && x == y
	case String:
		y, ok := b.(String)
		return ok && x == y
	case Number:
		y, ok := b.(Number)
		return ok && x == y
	case NilType:
		return IsNil(b)
	case List:
		return false
	case []Value:
		return false
	default:
		at := reflect.TypeOf(a)
		bt := reflect.TypeOf(b)
		if at == nil || bt == nil || at != bt || !at.Comparable() {
			return false
		}
		return a == b
	}
}

func equalValues(a, b Value) bool {
	if eqValues(a, b) {
		return true
	}
	left, ok := a.(List)
	if !ok {
		return false
	}
	right, ok := b.(List)
	if !ok {
		return false
	}
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if !equalValues(left[i], right[i]) {
			return false
		}
	}
	return true
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
