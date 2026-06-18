package vm

import (
	"context"
	"testing"
)

func newTestEvaluator() *Evaluator {
	return New()
}

func evalExpect(t *testing.T, e *Evaluator, src string, want string) {
	t.Helper()
	v, err := e.EvalString(context.Background(), src)
	if err != nil {
		t.Fatalf("EvalString(%q) error = %v", src, err)
	}
	if got := Stringify(v); got != want {
		t.Fatalf("EvalString(%q) = %s, want %s", src, got, want)
	}
}

func evalExpectError(t *testing.T, e *Evaluator, src string) {
	t.Helper()
	_, err := e.EvalString(context.Background(), src)
	if err == nil {
		t.Fatalf("EvalString(%q) expected error", src)
	}
}

func TestEvalSelfEvaluating(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, "42", "42")
	evalExpect(t, e, "3.14", "3.14")
	evalExpect(t, e, `"hello"`, `"hello"`)
	evalExpect(t, e, "nil", "nil")
}

func TestEvalSymbolLookup(t *testing.T) {
	e := newTestEvaluator()
	e.DefineGlobal("my-var", String("value"))
	evalExpect(t, e, "my-var", `"value"`)
}

func TestEvalUnboundSymbol(t *testing.T) {
	e := newTestEvaluator()
	evalExpectError(t, e, "unbound-var")
}

func TestEvalTAndNil(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, "t", "t")
	evalExpect(t, e, "nil", "nil")
}

func TestEvalProgn(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(progn 1 2 3)`, "3")
	evalExpect(t, e, `(progn)`, "nil")
}

func TestEvalLet(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(let ((x 1) (y 2)) (+ x y))`, "3")
	evalExpect(t, e, `(let () 42)`, "42")
	evalExpect(t, e, `(let ((x 1)) (let ((x 2)) x))`, "2")
}

func TestEvalLetStar(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(let* ((x 1) (y (+ x 1))) y)`, "2")
}

func TestEvalSetq(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(let ((x 0)) (setq x 42) x)`, "42")
	evalExpect(t, e, `(let ((x 0)) (setq x 1 x 2) x)`, "2")
}

func TestEvalIf(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(if t "yes" "no")`, `"yes"`)
	evalExpect(t, e, `(if nil "yes" "no")`, `"no"`)
	evalExpect(t, e, `(if t "yes")`, `"yes"`)
	evalExpect(t, e, `(if nil "yes")`, "nil")
	evalExpect(t, e, `(if 0 "yes" "no")`, `"yes"`)
}

func TestEvalWhenUnless(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(when t "ok")`, `"ok"`)
	evalExpect(t, e, `(when nil "ok")`, "nil")
	evalExpect(t, e, `(unless t "ok")`, "nil")
	evalExpect(t, e, `(unless nil "ok")`, `"ok"`)
}

func TestEvalAndOr(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(and t t)`, "t")
	evalExpect(t, e, `(and t nil)`, "nil")
	evalExpect(t, e, `(and nil (error))`, "nil")
	evalExpect(t, e, `(or nil t)`, "t")
	evalExpect(t, e, `(or nil nil)`, "nil")
	evalExpect(t, e, `(or t (error))`, "t")
	evalExpect(t, e, `(and)`, "t")
	evalExpect(t, e, `(or)`, "nil")
}

func TestEvalCond(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(cond (t "a"))`, `"a"`)
	evalExpect(t, e, `(cond (nil "a") (t "b"))`, `"b"`)
	evalExpect(t, e, `(cond (nil "a") (nil "b"))`, "nil")
	evalExpect(t, e, `(cond ("value"))`, `"value"`)
}

func TestEvalWhile(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(let ((i 0) (s 0)) (while (< i 3) (setq s (+ s i)) (setq i (+ i 1))) s)`, "3")
}

func TestEvalCatchThrow(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(catch 'tag (throw 'tag 42))`, "42")
	evalExpect(t, e, `(catch 'tag 1 2 3)`, "3")
}

func TestEvalLambda(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `((lambda (x) (* x x)) 5)`, "25")
	evalExpect(t, e, `((lambda () 42))`, "42")
}

func TestEvalDefun(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(progn (defun double (x) (+ x x)) (double 7))`, "14")
	evalExpect(t, e, `(progn (defun fact (n) (if (= n 0) 1 (* n (fact (- n 1))))) (fact 5))`, "120")
}

func TestEvalFuncall(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(funcall '+ 1 2)`, "3")
	evalExpect(t, e, `(funcall (lambda (x) x) 42)`, "42")
}

func TestEvalApply(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(apply '+ '(1 2 3))`, "6")
	evalExpect(t, e, `(apply '+ 1 '(2 3))`, "6")
	evalExpect(t, e, `(apply '+ 1 2 '(3))`, "6")
}

func TestEvalEmptyList(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, "()", "nil")
}

func TestEvalQuote(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(quote (a b c))`, "(a b c)")
	evalExpect(t, e, `'x`, "x")
	evalExpect(t, e, `'(1 2)`, "(1 2)")
}

func TestEvalBackquote(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, "`(a ,(+ 1 2))", "(a 3)")
	evalExpect(t, e, "`(a ,@(list 1 2))", "(a 1 2)")
	evalExpect(t, e, "`(a b)", "(a b)")
}

func TestEvalDefmacro(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(progn (defmacro m (x) `+"`"+`(+ ,x 1)) (m 4))`, "5")
}

func TestEvalMacroexpand(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(progn (defmacro m (x) `+"`"+`(+ ,x 1)) (macroexpand-1 '(m 4)))`, "(+ 4 1)")
	evalExpect(t, e, `(progn (defmacro m1 (x) `+"`"+`(m2 ,x)) (defmacro m2 (x) `+"`"+`(+ ,x 1)) (macroexpand '(m1 4)))`, "(+ 4 1)")
}

func TestEvalArithmetic(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(+ 1 2)`, "3")
	evalExpect(t, e, `(- 5 2)`, "3")
	evalExpect(t, e, `(* 3 4)`, "12")
	evalExpect(t, e, `(/ 10 2)`, "5")
	evalExpect(t, e, `(+)`, "0")
	evalExpect(t, e, `(*)`, "1")
	evalExpect(t, e, `(- 5)`, "-5")
	evalExpect(t, e, `(/ 4)`, "0.25")
}

func TestEvalComparison(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(= 1 1)`, "t")
	evalExpect(t, e, `(= 1 2)`, "nil")
	evalExpect(t, e, `(/= 1 2)`, "t")
	evalExpect(t, e, `(/= 1 1)`, "nil")
	evalExpect(t, e, `(< 1 2)`, "t")
	evalExpect(t, e, `(< 2 1)`, "nil")
	evalExpect(t, e, `(<= 1 1)`, "t")
	evalExpect(t, e, `(> 2 1)`, "t")
	evalExpect(t, e, `(>= 1 1)`, "t")
}

func TestEvalStringOps(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(concat "a" "b")`, `"ab"`)
	evalExpect(t, e, `(concat)`, `""`)
	evalExpect(t, e, `(format "hello %s %s" "a" "b")`, `"hello a b"`)
	evalExpect(t, e, `(length "abc")`, "3")
	evalExpect(t, e, `(length '(1 2 3))`, "3")
	evalExpect(t, e, `(length "")`, "0")
}

func TestEvalPredicates(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(null nil)`, "t")
	evalExpect(t, e, `(null '())`, "t")
	evalExpect(t, e, `(null 1)`, "nil")
	evalExpect(t, e, `(not nil)`, "t")
	evalExpect(t, e, `(not t)`, "nil")
	evalExpect(t, e, `(symbolp 'x)`, "t")
	evalExpect(t, e, `(symbolp nil)`, "t")
	evalExpect(t, e, `(symbolp 1)`, "nil")
	evalExpect(t, e, `(stringp "x")`, "t")
	evalExpect(t, e, `(stringp 1)`, "nil")
	evalExpect(t, e, `(numberp 1)`, "t")
	evalExpect(t, e, `(numberp "x")`, "nil")
	evalExpect(t, e, `(listp '(1))`, "t")
	evalExpect(t, e, `(listp nil)`, "t")
	evalExpect(t, e, `(listp 1)`, "nil")
	evalExpect(t, e, `(consp '(1))`, "t")
	evalExpect(t, e, `(consp nil)`, "nil")
	evalExpect(t, e, `(atom 'x)`, "t")
	evalExpect(t, e, `(atom '(1))`, "nil")
}

func TestEvalEqEqual(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(eq 'x 'x)`, "t")
	evalExpect(t, e, `(eq nil '())`, "t")
	evalExpect(t, e, `(equal '(1 2) '(1 2))`, "t")
	evalExpect(t, e, `(equal '(1 2) '(1 3))`, "nil")
}

func TestEvalListBuiltins(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(cons 1 '(2 3))`, "(1 2 3)")
	evalExpect(t, e, `(cons 1 nil)`, "(1)")
	evalExpect(t, e, `(car '(1 2))`, "1")
	evalExpect(t, e, `(cdr '(1 2))`, "(2)")
	evalExpect(t, e, `(nth 0 '(a b c))`, "a")
	evalExpect(t, e, `(nth 3 '(a b c))`, "nil")
	evalExpect(t, e, `(append '(1) '(2) '(3))`, "(1 2 3)")
	evalExpect(t, e, `(reverse '(1 2 3))`, "(3 2 1)")
	evalExpect(t, e, `(member "b" '("a" "b" "c"))`, `("b" "c")`)
	evalExpect(t, e, `(assoc 'b '((a 1) (b 2)))`, "(b 2)")
}

func TestEvalStringComparison(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(string= "a" "a")`, "t")
	evalExpect(t, e, `(string= "a" "b")`, "nil")
	evalExpect(t, e, `(string< "a" "b")`, "t")
	evalExpect(t, e, `(string> "b" "a")`, "t")
	evalExpect(t, e, `(string-equal "a" "a")`, "t")
	evalExpect(t, e, `(string-lessp "a" "b")`, "t")
	evalExpect(t, e, `(string-greaterp "b" "a")`, "t")
}

func TestEvalBufferBasic(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(buffer-name)`, `"*scratch*"`)
	evalExpect(t, e, `(bufferp (current-buffer))`, "t")
	evalExpect(t, e, `(progn (erase-buffer) (insert "abc") (buffer-string))`, `"abc"`)
	evalExpect(t, e, `(progn (erase-buffer) (insert "abcdef") (buffer-substring 2 5))`, `"bcd"`)
	evalExpect(t, e, `(progn (erase-buffer) (insert "abcdef") (delete-region 2 5) (buffer-string))`, `"aef"`)
}

func TestEvalPointAndGoto(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(progn (erase-buffer) (insert "abc") (point))`, "4")
	evalExpect(t, e, `(progn (erase-buffer) (insert "abc") (goto-char 2) (point))`, "2")
	evalExpect(t, e, `(progn (erase-buffer) (point-min))`, "1")
	evalExpect(t, e, `(progn (erase-buffer) (insert "abc") (point-max))`, "4")
}

func TestEvalWithCurrentBuffer(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(progn (setq b (get-buffer-create "work")) (with-current-buffer b (erase-buffer) (insert "work")) (buffer-name))`, `"*scratch*"`)
	evalExpect(t, e, `(progn (setq b (get-buffer-create "work2")) (with-current-buffer b (erase-buffer) (insert "data")) (with-current-buffer "work2" (buffer-string)))`, `"data"`)
}

func TestEvalSaveCurrentBuffer(t *testing.T) {
	e := newTestEvaluator()
	// save-current-buffer should restore the current buffer after body
	evalExpect(t, e, `(progn (setq b (get-buffer-create "saved")) (save-current-buffer (set-buffer b) (erase-buffer) (insert "saved")) (buffer-name))`, `"*scratch*"`)
	// Verify the work was done in the other buffer
	evalExpect(t, e, `(with-current-buffer "saved" (buffer-string))`, `"saved"`)
}

func TestEvalGenerateNewBuffer(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(progn (setq b (generate-new-buffer "tmp")) (setq c (generate-new-buffer "tmp")) (list (buffer-name b) (buffer-name c)))`, `("tmp" "tmp<2>")`)
}

func TestEvalKillBuffer(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(progn (setq b (get-buffer-create "kill-me")) (kill-buffer b) (get-buffer "kill-me"))`, "nil")
}

func TestEvalMarkers(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(progn (setq m (make-marker)) (markerp m))`, "t")
	evalExpect(t, e, `(progn (erase-buffer) (insert "abc") (goto-char 2) (setq m (point-marker)) (marker-position m))`, "2")
	evalExpect(t, e, `(progn (erase-buffer) (insert "abc") (setq m (set-marker (make-marker) 2)) (marker-position m))`, "2")
	evalExpect(t, e, `(progn (setq m (set-marker (make-marker) 1)) (set-marker m nil) (marker-position m))`, "nil")
	evalExpect(t, e, `(progn (erase-buffer) (insert "abc") (setq m (set-marker (make-marker) 2)) (buffer-name (marker-buffer m)))`, `"*scratch*"`)
}

func TestEvalCopyMarker(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(progn (erase-buffer) (insert "abc") (goto-char 2) (setq m (point-marker)) (setq c (copy-marker m)) (marker-position c))`, "2")
}

func TestEvalMarkerAdjusts(t *testing.T) {
	e := newTestEvaluator()
	// Marker should adjust after insert
	evalExpect(t, e, `(progn (erase-buffer) (insert "abcd") (goto-char 3) (setq m (point-marker)) (goto-char 1) (insert "XX") (marker-position m))`, "5")
	// Marker should adjust after delete
	evalExpect(t, e, `(progn (erase-buffer) (insert "abcdef") (goto-char 5) (setq m (point-marker)) (delete-region 2 4) (marker-position m))`, "3")
}

func TestEvalContextCancellation(t *testing.T) {
	e := newTestEvaluator()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := e.EvalString(ctx, `(+ 1 2)`)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestRegisterFunc(t *testing.T) {
	e := newTestEvaluator()
	e.RegisterFunc("my-fn", func(ctx *EvalContext, args []Value) (Value, error) {
		return String("custom"), nil
	})
	evalExpect(t, e, `(my-fn)`, `"custom"`)
}

func TestRegisterSpecial(t *testing.T) {
	e := newTestEvaluator()
	e.RegisterSpecial("my-special", func(ctx *EvalContext, unevaluated []Expr) (Value, error) {
		return Number(len(unevaluated)), nil
	})
	evalExpect(t, e, `(my-special a b c)`, "3")
}

func TestDefineGlobal(t *testing.T) {
	e := newTestEvaluator()
	e.DefineGlobal("g", Number(42))
	evalExpect(t, e, "g", "42")
}

func TestFuncNamesIncludesRegistered(t *testing.T) {
	e := newTestEvaluator()
	e.RegisterFunc("custom-fn", func(ctx *EvalContext, args []Value) (Value, error) {
		return Nil, nil
	})
	found := false
	for _, name := range e.FuncNames() {
		if name == "custom-fn" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("FuncNames() missing custom-fn: %v", e.FuncNames())
	}
}

func TestFuncNamesIncludesDefun(t *testing.T) {
	e := newTestEvaluator()
	e.EvalString(context.Background(), `(defun my-defun () 1)`)
	found := false
	for _, name := range e.FuncNames() {
		if name == "my-defun" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("FuncNames() missing my-defun: %v", e.FuncNames())
	}
}

func TestSpecialNames(t *testing.T) {
	e := newTestEvaluator()
	names := e.SpecialNames()
	if len(names) == 0 {
		t.Fatal("SpecialNames() is empty")
	}
	// Check some expected specials
	expected := []string{"quote", "let", "if", "progn"}
	for _, want := range expected {
		found := false
		for _, name := range names {
			if name == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("SpecialNames() missing %s", want)
		}
	}
}

func TestGlobalNames(t *testing.T) {
	e := newTestEvaluator()
	e.DefineGlobal("a", Number(1))
	e.DefineGlobal("b", Number(2))
	names := e.GlobalNames()
	if len(names) < 2 {
		t.Fatalf("GlobalNames() = %v, want at least 2", names)
	}
}

// Error cases
func TestEvalErrors(t *testing.T) {
	e := newTestEvaluator()
	cases := []string{
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
	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			evalExpectError(t, e, src)
		})
	}
}

func TestEvalClosureCapture(t *testing.T) {
	e := newTestEvaluator()
	// Closure should capture variable by reference
	evalExpect(t, e, `(let ((x 10)) (let ((add-x (lambda (y) (+ x y)))) (let ((x 100)) (funcall add-x 5))))`, "15")
}

func TestEvalNestedMacros(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(progn (defmacro m1 (x) `+"`"+`(m2 ,x)) (defmacro m2 (x) `+"`"+`(+ ,x 1)) (m1 4))`, "5")
}

func TestEvalWithCurrentBufferStringArg(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(progn (get-buffer-create "strbuf") (with-current-buffer "strbuf" (erase-buffer) (insert "str-data")) (with-current-buffer "strbuf" (buffer-string)))`, `"str-data"`)
}

func TestEvalGetBufferCreate(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(bufferp (get-buffer-create "newbuf"))`, "t")
}

func TestEvalSetBuffer(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(progn (get-buffer-create "target") (set-buffer "target") (erase-buffer) (insert "data") (buffer-name))`, `"target"`)
}

func TestEvalGetBuffer(t *testing.T) {
	e := newTestEvaluator()
	evalExpect(t, e, `(bufferp (get-buffer "*scratch*"))`, "t")
	evalExpect(t, e, `(get-buffer "nonexistent")`, "nil")
}
