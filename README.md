# vibeEmacsLispVm

Tiny embeddable Emacs Lisp subset VM for Go.

This is not a full Emacs Lisp implementation. It is a minimal S-expression
parser/evaluator intended to be embedded by host applications that register
their own Go functions.

## Scope

Supported syntax:

- Lists: `(foo bar)`
- Symbols: `foo`, `:keyword`
- Strings with basic escapes: `"hello\nworld"`
- Numbers: `1`, `3.14`
- Quote shorthand: `'("read" "grep")`
- Line comments: `; comment`

Supported special forms:

- `quote`
- `progn`
- `let`
- `setq`
- `if`
- `when`
- `unless`
- `and`
- `or`

Supported builtins:

- `concat`
- `format` (`%s` only)
- `list`
- `length`
- `=`, `<`, `>`
- `string=`
- `not`

Not supported:

- Full Emacs Lisp runtime
- Macros
- Backquote/comma
- Reader macros
- Vectors
- Buffers, processes, files, shell, network, packages

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

## Tests

```bash
go test ./...
```

## License

MIT License. See [LICENSE](LICENSE).
