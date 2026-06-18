# vibeEmacsLispVm

Tiny embeddable Emacs Lisp subset VM for Go.

[中文文档](README_zh.md)

This is not a full Emacs Lisp implementation. It is a minimal S-expression
parser/evaluator intended to be embedded by host applications that register
their own Go functions.

## Packages and Layout

- Root package `elispvm`: public API facade for embedding. It exposes stable
  types, constructors, parsing, formatting, and evaluator methods.
- `internal/vm`: parser, evaluator, runtime values, core forms, and builtin
  functions. This is the implementation package and is intentionally not
  importable by downstream modules.
- `internal/repl`: command-line read-eval-print loop separated from the VM
  library so interactive I/O does not become part of the embedding API.
- `cmd/elispvm`: standalone CLI entrypoint for the REPL and one-shot
  evaluation.

## Scope

Supported syntax:

- Lists: `(foo bar)`
- Symbols: `foo`, `:keyword`
- Strings with basic escapes: `"hello\nworld"`
- Numbers: `1`, `3.14`
- Quote shorthand: `'("read" "grep")`
- Backquote/comma: `` `(a ,b ,@c) ``
- Line comments: `; comment`

Supported special forms:

- `quote`
- `progn`
- `let`
- `let*`
- `setq`
- `if`
- `when`
- `unless`
- `and`
- `or`
- `while`
- `cond`
- `catch`
- `throw`
- `lambda`
- `defun`
- `backquote`
- `comma`
- `comma-splice`
- `defmacro`
- `with-current-buffer`
- `save-current-buffer`

Supported builtins:

- `concat`
- `format` (`%s` only)
- `list`
- `length`
- `cons`, `car`, `cdr`, `nth`, `append`, `reverse`, `member`, `assoc`
- `funcall`, `apply`
- `macroexpand-1`, `macroexpand`
- `+`, `-`, `*`, `/`
- `=`, `/=`, `<`, `<=`, `>`, `>=`
- `eq`, `equal`
- `string=`, `string-equal`, `string-lessp`, `string<`, `string-greaterp`, `string>`
- `not`
- `null`, `symbolp`, `stringp`, `numberp`, `listp`, `consp`, `atom`
- `bufferp`, `buffer-name`, `current-buffer`, `set-buffer`, `get-buffer`,
  `get-buffer-create`, `generate-new-buffer`, `kill-buffer`
- `point`, `point-min`, `point-max`, `goto-char`, `insert`, `delete-region`,
  `buffer-substring`, `buffer-string`, `erase-buffer`
- `markerp`, `make-marker`, `point-marker`, `copy-marker`, `marker-position`,
  `marker-buffer`, `set-marker`

Function support currently accepts fixed argument lists only; `&optional`,
`&rest`, and other full Emacs Lisp lambda-list features are not implemented.

Not supported:

- Full Emacs Lisp runtime
- Full Emacs Lisp lambda-list features
- Reader macros beyond quote/backquote/comma
- Vectors
- Filesystem-backed buffers, windows, frames, processes, files, shell, network,
  packages

## Embedding

```go
package main

import (
	"context"
	"fmt"

	elispvm "github.com/startvibecoding/vibeEmacsLispVm"
)

func main() {
	e := elispvm.New()
	e.RegisterFunc("join", func(ctx *elispvm.EvalContext, args []elispvm.Value) (elispvm.Value, error) {
		a := string(args[0].(elispvm.String))
		b := string(args[1].(elispvm.String))
		return elispvm.String(a + "/" + b), nil
	})

	v, err := e.EvalString(context.Background(), `(concat "hello" " " "world")`)
	if err != nil {
		panic(err)
	}
	fmt.Println(elispvm.Stringify(v))
}
```

Use `RegisterFunc` for normal functions with evaluated arguments. Use
`RegisterSpecial` only when a host function must control evaluation of its
arguments.

Tooling can inspect registered names with `FuncNames`, `SpecialNames`, and
`GlobalNames`. These return sorted copies and do not expose mutable evaluator
state.

More examples are available in [`examples/`](examples/).

## CLI REPL

Build the command:

```bash
make build
```

Start the REPL:

```bash
./bin/elispvm
```

Evaluate source without starting the REPL:

```bash
./bin/elispvm -eval '(concat "hello" " " "world")'
./bin/elispvm -file ./script.el
```

The REPL accepts one expression at a time and supports multi-line lists.
Use `:help` for REPL commands and `:quit` to exit. In an interactive Linux
terminal, press Tab to complete registered functions, special forms, globals,
`nil`, and `t`; when multiple names match, Tab prints the candidates.

Run the sample script:

```bash
./bin/elispvm -file ./examples/scripts/basic.el
```

## Examples

- [`examples/embedding/main.go`](examples/embedding/main.go): embed the VM in a
  Go program and register a host function.
- [`examples/scripts/basic.el`](examples/scripts/basic.el): evaluate a small
  Lisp script with the CLI.

## Make Targets

```bash
make fmt      # gofmt all Go files
make test     # run go test ./...
make vet      # run go vet ./...
make build    # build ./bin/elispvm
make repl     # build and run the REPL
make package  # create a dist/*.tar.gz bundle with binary, README, and LICENSE
make clean    # remove build artifacts
```

## Tests

```bash
go test ./...
```

## License

MIT License. See [LICENSE](LICENSE).
