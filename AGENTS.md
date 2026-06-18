# vibeEmacsLispVm Agent Guide

This repository is a standalone Go library that implements a tiny embeddable
Emacs Lisp subset VM. Keep it small, deterministic, and host-application
agnostic.

## Project Snapshot

- Language: Go
- Module: `github.com/startvibecoding/vibeEmacsLispVm`
- Purpose: minimal Elisp-style S-expression parser/evaluator for embedding in
  other Go applications.
- Primary downstream user: VibeCoding dynamic workflows.
- Current status: early standalone library; no GitHub dependency integration
  yet.

## Design Goals

- Implement a minimal Elisp syntax subset without changing Elisp syntax.
- Provide Go-side extension points through registered functions and special
  forms.
- Avoid workflow-specific logic in this repository.
- Avoid file, shell, network, package loading, process, buffer, or Emacs runtime
  APIs.
- Keep behavior predictable and easy to test.

## Non-Goals

- Do not implement full Emacs Lisp.
- Do not add macros, backquote/comma, reader macros, vectors, packages, buffers,
  markers, processes, or dynamic loading unless explicitly requested.
- Do not add workflow builtins such as `workflow`, `agent`, `parallel`, or
  `result` here. Those belong in the host application via registration.
- Do not depend on third-party Elisp interpreters or parser libraries.

## Current Public API

- `New() *Evaluator`
- `Parse(src string) ([]Expr, error)`
- `(*Evaluator).EvalString(ctx, src)`
- `(*Evaluator).EvalAll(ctx, exprs)`
- `(*Evaluator).RegisterFunc(name, fn)`
- `(*Evaluator).RegisterSpecial(name, fn)`
- `(*Evaluator).DefineGlobal(name, value)`
- `(*Evaluator).FuncNames()`
- `(*Evaluator).SpecialNames()`
- `(*Evaluator).GlobalNames()`
- `NewEnv(parent)`
- `Stringify(value)`
- Runtime value types: `Symbol`, `String`, `Number`, `List`, `Nil`

Registered normal functions receive evaluated arguments. Registered special
forms receive unevaluated expressions and must explicitly evaluate what they
need.

## Supported Language Subset

Syntax:

- Lists: `(foo bar)`
- Symbols: `foo`, `:keyword`
- Strings with basic escapes: `"hello\nworld"`
- Numbers: `1`, `3.14`
- Quote shorthand: `'("read" "grep")`
- Line comments: `; comment`

Special forms:

- `quote`
- `progn`
- `let`
- `setq`
- `if`
- `when`
- `unless`
- `and`
- `or`

Builtins:

- `concat`
- `format` (`%s` only)
- `list`
- `length`
- `=`, `<`, `>`
- `string=`
- `not`

## Repository Layout

- `api.go` - public API facade for downstream imports
- `doc.go` - package documentation for the public root package
- `internal/vm/` - parser, evaluator, runtime values, core forms, and builtins
- `internal/repl/` - command-line read-eval-print loop and line editor
- `cmd/elispvm/` - standalone CLI entrypoint
- `examples/` - small embedding and CLI script examples
- `*_test.go` - public API parser and evaluator tests

## Implementation Rules

- Keep this library independent from VibeCoding internals.
- Do not import VibeCoding packages.
- Do not introduce side-effect APIs into the VM. Hosts can register side-effect
  functions if they choose.
- Do not change parser/tokenizer behavior to support host-specific syntax.
- Prefer small, explicit types and errors over broad abstraction.
- Add tests for every language feature or semantic behavior change.
- Keep exported API stable when possible; if changing it, update `README.md`.

## Testing

Use the local Go toolchain:

```bash
go test ./...
```

If `go` is on PATH, this is equivalent:

```bash
go test ./...
```

Run formatting before finishing changes:

```bash
gofmt -w .
```

## Downstream Integration Notes

VibeCoding should depend on this library only after it is published to GitHub.
The workflow package in VibeCoding should register workflow-specific functions
such as `workflow`, `phase`, `parallel`, `agent`, `result`, and `log` via this
library's registration API.

This repository should remain a generic Elisp subset VM, not a workflow runtime.
