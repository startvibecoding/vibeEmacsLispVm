// Package elispvm implements a tiny embeddable Emacs Lisp subset evaluator.
//
// It intentionally does not implement full Emacs Lisp. The package provides
// S-expression parsing, a small evaluator, core special forms, and a Go-side
// function registration API for host applications.
package elispvm
