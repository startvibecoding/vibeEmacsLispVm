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
	if got := string(v.(String)); got != "ok:hello world" {
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
		`(if '() "bad" "ok")`:           `"ok"`,
		`(when t (concat "a" "b"))`:     `"ab"`,
		`(unless nil (concat "a" "b"))`: `"ab"`,
		`(cond (nil "bad") ((string= "a" "a") "ok") (t "fallback"))`: `"ok"`,
		`(cond ((string= "a" "b") "bad") ("value"))`:                 `"value"`,
		`(not nil)`:  `t`,
		`(< 1 2 3)`:  `t`,
		`(<= 1 1 2)`: `t`,
		`(> 3 2 1)`:  `t`,
		`(>= 3 3 2)`: `t`,
		`(= 1 1 1)`:  `t`,
		`(/= 1 2 3)`: `t`,
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

func TestEvalArithmetic(t *testing.T) {
	e := New()
	cases := map[string]string{
		`(+)`:           `0`,
		`(+ 1 2 3)`:     `6`,
		`(- 5)`:         `-5`,
		`(- 10 3 2)`:    `5`,
		`(*)`:           `1`,
		`(* 2 3 4)`:     `24`,
		`(/ 8 2 2)`:     `2`,
		`(/ 4)`:         `0.25`,
		`(< 1 2 2)`:     `nil`,
		`(/= 1 2 1)`:    `nil`,
		`(= (+ 1 2) 3)`: `t`,
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

func TestEvalStringComparison(t *testing.T) {
	e := New()
	cases := map[string]string{
		`(string= "a" "a")`:         `t`,
		`(string-equal "a" "a")`:    `t`,
		`(string< "a" "b")`:         `t`,
		`(string-lessp "a" "b")`:    `t`,
		`(string> "b" "a")`:         `t`,
		`(string-greaterp "b" "a")`: `t`,
		`(string= 'name "name")`:    `t`,
		`(string< "b" "a")`:         `nil`,
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

func TestEvalPredicatesAndEquality(t *testing.T) {
	e := New()
	cases := map[string]string{
		`(null nil)`:                        `t`,
		`(null '())`:                        `t`,
		`(symbolp 'name)`:                   `t`,
		`(symbolp nil)`:                     `t`,
		`(stringp "x")`:                     `t`,
		`(numberp 1)`:                       `t`,
		`(listp '(1 2))`:                    `t`,
		`(listp nil)`:                       `t`,
		`(consp '(1 2))`:                    `t`,
		`(consp nil)`:                       `nil`,
		`(atom 'name)`:                      `t`,
		`(atom '(1 2))`:                     `nil`,
		`(eq 'name 'name)`:                  `t`,
		`(eq nil '())`:                      `t`,
		`(equal '(1 "a" (b)) '(1 "a" (b)))`: `t`,
		`(equal '(1 2) '(1 3))`:             `nil`,
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

func TestEvalListBuiltins(t *testing.T) {
	e := New()
	cases := map[string]string{
		`(cons 1 '(2 3))`:                        `(1 2 3)`,
		`(cons 1 nil)`:                           `(1)`,
		`(car '(1 2 3))`:                         `1`,
		`(car nil)`:                              `nil`,
		`(cdr '(1 2 3))`:                         `(2 3)`,
		`(cdr '(1))`:                             `nil`,
		`(nth 0 '(a b c))`:                       `a`,
		`(nth 2 '(a b c))`:                       `c`,
		`(nth 3 '(a b c))`:                       `nil`,
		`(append '(1 2) nil '(3 4))`:             `(1 2 3 4)`,
		`(append)`:                               `nil`,
		`(reverse '(1 2 3))`:                     `(3 2 1)`,
		`(reverse nil)`:                          `nil`,
		`(member "b" '("a" "b" "c"))`:            `("b" "c")`,
		`(member "x" '("a" "b" "c"))`:            `nil`,
		`(assoc 'b '((a 1) (b 2) (c 3)))`:        `(b 2)`,
		`(assoc "b" '(("a" 1) ("b" 2) ("c" 3)))`: `("b" 2)`,
		`(assoc 'b '(not-a-pair (a 1)))`:         `nil`,
		`(equal (append '(1) '(2)) '(1 2))`:      `t`,
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

func TestEvalFunctions(t *testing.T) {
	e := New()
	cases := map[string]string{
		`((lambda (x) (+ x 1)) 4)`:             `5`,
		`(funcall (lambda (x y) (+ x y)) 2 3)`: `5`,
		`(funcall '+ 2 3)`:                     `5`,
		`(apply '+ '(1 2 3))`:                  `6`,
		`(apply '+ 1 '(2 3))`:                  `6`,
		`(let* ((x 2) (y (+ x 3))) (* x y))`:   `10`,
		`(let ((x 10))
		   (let ((add-x (lambda (y) (+ x y))))
		     (let ((x 100))
		       (funcall add-x 5))))`: `15`,
		`(progn
		   (defun add2 (x) (+ x 2))
		   (add2 5))`: `7`,
		`(progn
		   (defun fact (n)
		     (if (= n 0)
		         1
		         (* n (fact (- n 1)))))
		   (fact 5))`: `120`,
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

func TestEvalMacros(t *testing.T) {
	e := New()
	cases := map[string]string{
		"`(a ,(+ 1 2) ,@(list 4 5))": `(a 3 4 5)`,
		`(progn
		   (defmacro twice (x)
		     ` + "`" + `(+ ,x ,x))
		   (twice 5))`: `10`,
		`(progn
		   (defmacro when2 (cond body)
		     ` + "`" + `(if ,cond ,body nil))
		   (let ((x 0))
		     (when2 t (setq x 7))
		     x))`: `7`,
		`(progn
		   (defmacro m1 (x) ` + "`" + `(m2 ,x))
		   (defmacro m2 (x) ` + "`" + `(+ ,x 1))
		   (macroexpand-1 '(m1 4)))`: `(m2 4)`,
		`(progn
		   (defmacro m1 (x) ` + "`" + `(m2 ,x))
		   (defmacro m2 (x) ` + "`" + `(+ ,x 1))
		   (macroexpand '(m1 4)))`: `(+ 4 1)`,
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

func TestEvalBuffersAndMarkers(t *testing.T) {
	e := New()
	cases := map[string]string{
		`(buffer-name)`:              `"*scratch*"`,
		`(bufferp (current-buffer))`: `t`,
		`(progn
		   (erase-buffer)
		   (insert "abc")
		   (buffer-string))`: `"abc"`,
		`(progn
		   (erase-buffer)
		   (insert "abc")
		   (goto-char 2)
		   (insert "X")
		   (list (buffer-string) (point)))`: `("aXbc" 3)`,
		`(progn
		   (erase-buffer)
		   (insert "abcdef")
		   (buffer-substring 2 5))`: `"bcd"`,
		`(progn
		   (erase-buffer)
		   (insert "abcdef")
		   (delete-region 2 5)
		   (buffer-string))`: `"aef"`,
		`(progn
		   (setq b (get-buffer-create "work"))
		   (with-current-buffer b
		     (erase-buffer)
		     (insert "work"))
		   (buffer-name))`: `"*scratch*"`,
		`(progn
		   (setq b (get-buffer-create "work2"))
		   (with-current-buffer b
		     (erase-buffer)
		     (insert "work"))
		   (with-current-buffer "work2"
		     (buffer-string)))`: `"work"`,
		`(progn
		   (setq b (generate-new-buffer "tmp"))
		   (setq c (generate-new-buffer "tmp"))
		   (list (buffer-name b) (buffer-name c)))`: `("tmp" "tmp<2>")`,
		`(progn
		   (setq b (get-buffer-create "kill-me"))
		   (kill-buffer b)
		   (get-buffer "kill-me"))`: `nil`,
		`(progn
		   (erase-buffer)
		   (insert "abcd")
		   (goto-char 3)
		   (setq m (point-marker))
		   (goto-char 1)
		   (insert "XX")
		   (marker-position m))`: `5`,
		`(progn
		   (erase-buffer)
		   (insert "abcdef")
		   (goto-char 5)
		   (setq m (point-marker))
		   (delete-region 2 4)
		   (marker-position m))`: `3`,
		`(progn
		   (setq m (make-marker))
		   (markerp m))`: `t`,
		`(progn
		   (erase-buffer)
		   (insert "abc")
		   (setq m (set-marker (make-marker) 2))
		   (list (marker-position m)
		         (buffer-name (marker-buffer m))))`: `(2 "*scratch*")`,
		`(progn
		   (setq m (set-marker (make-marker) 1))
		   (set-marker m nil)
		   (list (marker-position m) (marker-buffer m)))`: `(nil nil)`,
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

func TestEvalWhileCatchThrow(t *testing.T) {
	e := New()
	v, err := e.EvalString(context.Background(), `
		(let ((i 0)
		      (sum 0))
		  (while (< i 5)
		    (setq i (+ i 1))
		    (setq sum (+ sum i)))
		  sum)
	`)
	if err != nil {
		t.Fatalf("EvalString() error = %v", err)
	}
	if got := Stringify(v); got != `15` {
		t.Fatalf("result = %s, want 15", got)
	}

	v, err = e.EvalString(context.Background(), `
		(catch 'done
		  (let ((i 0))
		    (while (< i 10)
		      (setq i (+ i 1))
		      (when (= i 4)
		        (throw 'done (format "stopped:%s" i)))))
		  "bad")
	`)
	if err != nil {
		t.Fatalf("EvalString() error = %v", err)
	}
	if got := Stringify(v); got != `"stopped:4"` {
		t.Fatalf("result = %s, want stopped", got)
	}
}

func TestRegisterFunc(t *testing.T) {
	e := New()
	e.RegisterFunc("join-with-slash", func(ctx *EvalContext, args []Value) (Value, error) {
		if len(args) != 2 {
			t.Fatalf("args = %d, want 2", len(args))
		}
		return String(string(args[0].(String)) + "/" + string(args[1].(String))), nil
	})

	v, err := e.EvalString(context.Background(), `(join-with-slash "phase" "agent")`)
	if err != nil {
		t.Fatalf("EvalString() error = %v", err)
	}
	if got := string(v.(String)); got != "phase/agent" {
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

func TestRegisteredNames(t *testing.T) {
	e := New()
	e.RegisterFunc("host-fn", func(ctx *EvalContext, args []Value) (Value, error) {
		return Nil, nil
	})
	e.RegisterSpecial("host-special", func(ctx *EvalContext, unevaluated []Expr) (Value, error) {
		return Nil, nil
	})
	e.DefineGlobal("host-global", String("value"))

	if !containsName(e.FuncNames(), "host-fn") {
		t.Fatalf("FuncNames() missing host-fn: %v", e.FuncNames())
	}
	if !containsName(e.SpecialNames(), "host-special") {
		t.Fatalf("SpecialNames() missing host-special: %v", e.SpecialNames())
	}
	if !containsName(e.GlobalNames(), "host-global") {
		t.Fatalf("GlobalNames() missing host-global: %v", e.GlobalNames())
	}

	if _, err := e.EvalString(context.Background(), `(defun lisp-fn () "ok")`); err != nil {
		t.Fatalf("EvalString() defun error = %v", err)
	}
	if !containsName(e.FuncNames(), "lisp-fn") {
		t.Fatalf("FuncNames() missing lisp-fn: %v", e.FuncNames())
	}
}

func containsName(names []string, want string) bool {
	for _, name := range names {
		if name == want {
			return true
		}
	}
	return false
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

func TestEvalSaveCurrentBuffer(t *testing.T) {
	e := New()
	// save-current-buffer should restore current buffer after body
	v, err := e.EvalString(context.Background(), `
		(progn
		  (setq b (get-buffer-create "saved-buf"))
		  (save-current-buffer
		    (set-buffer b)
		    (erase-buffer)
		    (insert "saved-data"))
		  (buffer-name))
	`)
	if err != nil {
		t.Fatalf("EvalString() error = %v", err)
	}
	if got := Stringify(v); got != `"*scratch*"` {
		t.Fatalf("result = %s, want *scratch*", got)
	}

	// Verify the work was done in the other buffer
	v, err = e.EvalString(context.Background(), `(with-current-buffer "saved-buf" (buffer-string))`)
	if err != nil {
		t.Fatalf("EvalString() error = %v", err)
	}
	if got := Stringify(v); got != `"saved-data"` {
		t.Fatalf("result = %s, want saved-data", got)
	}
}

func TestEvalPointMinMax(t *testing.T) {
	e := New()
	cases := map[string]string{
		`(progn (erase-buffer) (point-min))`:                    `1`,
		`(progn (erase-buffer) (point-max))`:                    `1`,
		`(progn (erase-buffer) (insert "abc") (point-min))`:     `1`,
		`(progn (erase-buffer) (insert "abc") (point-max))`:     `4`,
		`(progn (erase-buffer) (insert "abcdef") (point-max))`:  `7`,
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

func TestEvalPointMarker(t *testing.T) {
	e := New()
	v, err := e.EvalString(context.Background(), `
		(progn
		  (erase-buffer)
		  (insert "abc")
		  (point-marker))
	`)
	if err != nil {
		t.Fatalf("EvalString() error = %v", err)
	}
	if got := Stringify(v); got != "#<marker at 4 in *scratch*>" {
		t.Fatalf("result = %s, want marker at 4", got)
	}
}

func TestEvalCopyMarkerSuccess(t *testing.T) {
	e := New()
	v, err := e.EvalString(context.Background(), `
		(progn
		  (erase-buffer)
		  (insert "abc")
		  (goto-char 2)
		  (setq m (point-marker))
		  (copy-marker m))
	`)
	if err != nil {
		t.Fatalf("EvalString() error = %v", err)
	}
	if got := Stringify(v); got != "#<marker at 2 in *scratch*>" {
		t.Fatalf("result = %s, want marker at 2", got)
	}
}

func TestPublicAPIIsNil(t *testing.T) {
	cases := []struct {
		name string
		v    Value
		want bool
	}{
		{"nil", Nil, true},
		{"empty List", List{}, true},
		{"non-empty", List{Symbol("a")}, false},
		{"String", String("x"), false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNil(tt.v); got != tt.want {
				t.Fatalf("IsNil(%v) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

func TestPublicAPITruthy(t *testing.T) {
	cases := []struct {
		name string
		v    Value
		want bool
	}{
		{"nil", Nil, false},
		{"empty List", List{}, false},
		{"non-empty", List{Symbol("a")}, true},
		{"String", String("x"), true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := Truthy(tt.v); got != tt.want {
				t.Fatalf("Truthy(%v) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

func TestPublicAPINewEnv(t *testing.T) {
	e := NewEnv(nil)
	if e == nil {
		t.Fatal("NewEnv(nil) returned nil")
	}
}

func TestEvalAllDirect(t *testing.T) {
	e := New()
	exprs, err := Parse(`1 2 3`)
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}
	v, err := e.EvalAll(context.Background(), exprs)
	if err != nil {
		t.Fatalf("EvalAll error = %v", err)
	}
	if got := Stringify(v); got != "3" {
		t.Fatalf("result = %s, want 3", got)
	}
}

func TestEvalAllEmpty(t *testing.T) {
	e := New()
	v, err := e.EvalAll(context.Background(), nil)
	if err != nil {
		t.Fatalf("EvalAll error = %v", err)
	}
	if got := Stringify(v); got != "nil" {
		t.Fatalf("result = %s, want nil", got)
	}
}

func TestEvalWithCurrentBufferNestedSwitch(t *testing.T) {
	e := New()
	// Test nested with-current-buffer
	v, err := e.EvalString(context.Background(), `
		(progn
		  (setq b1 (get-buffer-create "buf1"))
		  (setq b2 (get-buffer-create "buf2"))
		  (with-current-buffer b1
		    (erase-buffer)
		    (insert "buf1-data")
		    (with-current-buffer b2
		      (erase-buffer)
		      (insert "buf2-data"))
		    (buffer-string)))
	`)
	if err != nil {
		t.Fatalf("EvalString() error = %v", err)
	}
	if got := Stringify(v); got != `"buf1-data"` {
		t.Fatalf("result = %s, want buf1-data", got)
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
		`(/ 1 0)`,
		`(+ "bad")`,
		`(cond nil)`,
		`(throw 'missing 1)`,
		`(string< 1 "a")`,
		`(cons 1 2)`,
		`(car 1)`,
		`(cdr 1)`,
		`(nth "zero" '(a b))`,
		`(nth 1.5 '(a b))`,
		`(append '(a) 1)`,
		`(reverse 1)`,
		`(member 'a 1)`,
		`(assoc 'a 1)`,
		`((lambda (x) x))`,
		`(lambda 1 2)`,
		`(lambda (&optional x) x)`,
		`(defun 1 () 1)`,
		`(defun bad (x x) x)`,
		`(funcall)`,
		`(funcall 1)`,
		`(apply '+ 1)`,
		`(apply '+ 1 2)`,
		`,x`,
		`,@x`,
		"`(a ,@1)",
		`(defmacro 1 () 1)`,
		`(defmacro bad (x x) x)`,
		`(progn (defmacro bad (x) x) (bad))`,
		`(macroexpand-1)`,
		`(buffer-name 1)`,
		`(set-buffer "missing")`,
		`(goto-char 0)`,
		`(goto-char 2)`,
		`(delete-region 2 1)`,
		`(buffer-substring 1 99)`,
		`(copy-marker nil)`,
		`(marker-position 1)`,
		`(set-marker 1 1)`,
		`(set-marker (make-marker) 99)`,
	}
	for _, src := range tests {
		t.Run(src, func(t *testing.T) {
			if _, err := e.EvalString(context.Background(), src); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
