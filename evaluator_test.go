package elispvm

import (
	"context"
	"errors"
	"testing"
)

func TestEvalCoreFormsAndBuiltins(t *testing.T) {
	e := New()
	v, err := e.EvalString(context.Background(), `
		(let ((x "hello")
		      (y "world"))
		  (setq x (concat x " " y))
		  (if (string= x "hello world")
		      (format "ok:%s" x)
		      "bad"))
	`)
	if err != nil {
		t.Fatalf("EvalString() error = %v", err)
	}
	if got := rawString(v); got != "ok:hello world" {
		t.Fatalf("result = %q", got)
	}
}

func TestEvalQuoteAndList(t *testing.T) {
	e := New()
	v, err := e.EvalString(context.Background(), `(list '("read" "grep") (length "abc"))`)
	if err != nil {
		t.Fatalf("EvalString() error = %v", err)
	}
	if got := Stringify(v); got != `(("read" "grep") 3)` {
		t.Fatalf("result = %s", got)
	}
}

func TestEvalConditionals(t *testing.T) {
	e := New()
	cases := map[string]string{
		`(and t "last")`:                `"last"`,
		`(and t nil (missing))`:         `nil`,
		`(or nil "first" (missing))`:    `"first"`,
		`(when t (concat "a" "b"))`:     `"ab"`,
		`(unless nil (concat "a" "b"))`: `"ab"`,
		`(not nil)`:                     `t`,
		`(< 1 2 3)`:                     `t`,
		`(> 3 2 1)`:                     `t`,
		`(= 1 1 1)`:                     `t`,
	}
	for src, want := range cases {
		t.Run(src, func(t *testing.T) {
			v, err := e.EvalString(context.Background(), src)
			if err != nil {
				t.Fatalf("EvalString() error = %v", err)
			}
			if got := Stringify(v); got != want {
				t.Fatalf("result = %s, want %s", got, want)
			}
		})
	}
}

func TestRegisterFunc(t *testing.T) {
	e := New()
	e.RegisterFunc("join-with-slash", func(ctx *EvalContext, args []Value) (Value, error) {
		if len(args) != 2 {
			t.Fatalf("args = %d, want 2", len(args))
		}
		return String(rawString(args[0]) + "/" + rawString(args[1])), nil
	})

	v, err := e.EvalString(context.Background(), `(join-with-slash "phase" "agent")`)
	if err != nil {
		t.Fatalf("EvalString() error = %v", err)
	}
	if got := rawString(v); got != "phase/agent" {
		t.Fatalf("result = %q", got)
	}
}

func TestRegisterSpecialControlsEvaluation(t *testing.T) {
	e := New()
	e.RegisterSpecial("literal-args", func(ctx *EvalContext, unevaluated []Expr) (Value, error) {
		out := make(List, 0, len(unevaluated))
		for _, expr := range unevaluated {
			out = append(out, expr)
		}
		return out, nil
	})

	v, err := e.EvalString(context.Background(), `(literal-args missing-symbol "ok")`)
	if err != nil {
		t.Fatalf("EvalString() error = %v", err)
	}
	if got := Stringify(v); got != `(missing-symbol "ok")` {
		t.Fatalf("result = %s", got)
	}
}

func TestEvalContextCancellation(t *testing.T) {
	e := New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := e.EvalString(ctx, `(concat "a" "b")`)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func TestEvalErrors(t *testing.T) {
	e := New()
	tests := []string{
		`missing`,
		`(missing)`,
		`(if t)`,
		`(let (("bad" 1)) 1)`,
		`(length 1)`,
	}
	for _, src := range tests {
		t.Run(src, func(t *testing.T) {
			if _, err := e.EvalString(context.Background(), src); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
