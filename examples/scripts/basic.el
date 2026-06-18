; Basic script runnable with:
; ./bin/elispvm -file ./examples/scripts/basic.el

(let ((name "vibeEmacsLispVm")
      (items '("parse" "eval" "embed")))
  (format "%s:%s:%s" name (length items) (concat "ready")))
