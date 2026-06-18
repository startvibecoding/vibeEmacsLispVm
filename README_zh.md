# vibeEmacsLispVm

一个用于 Go 的小型可嵌入 Emacs Lisp 子集 VM。

[English README](README.md)

这不是完整的 Emacs Lisp 实现。它是一个最小化的 S 表达式解析器和求值器，适合被宿主应用嵌入，并由宿主应用注册自己的 Go 函数。

## 包与目录

- 根包 `elispvm`：面向下游使用者的公共 API 门面，暴露稳定的类型、构造函数、解析、格式化和求值器方法。
- `internal/vm`：解析器、求值器、运行时值、核心特殊表单和内置函数的实现包。该包不面向下游模块导入。
- `internal/repl`：命令行 REPL 和行编辑器，避免交互式 I/O 混入嵌入式 VM API。
- `cmd/elispvm`：独立 CLI 入口，支持 REPL 和一次性求值。
- `examples/`：嵌入和脚本示例。

## 支持范围

支持语法：

- 列表：`(foo bar)`
- 符号：`foo`、`:keyword`
- 带基础转义的字符串：`"hello\nworld"`
- 数字：`1`、`3.14`
- quote 简写：`'("read" "grep")`
- 行注释：`; comment`

支持特殊表单：

- `quote`
- `progn`
- `let`
- `setq`
- `if`
- `when`
- `unless`
- `and`
- `or`

支持内置函数：

- `concat`
- `format`（仅支持 `%s`）
- `list`
- `length`
- `=`、`<`、`>`
- `string=`
- `not`

不支持：

- 完整 Emacs Lisp 运行时
- 宏
- backquote/comma
- reader macros
- vector
- buffer、process、file、shell、network、package 等运行时能力

## 嵌入使用

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

普通函数使用 `RegisterFunc` 注册，参数会先被求值。需要控制参数求值时，使用 `RegisterSpecial` 注册特殊表单。

工具可以通过 `FuncNames`、`SpecialNames` 和 `GlobalNames` 查看已注册名称。这些方法返回排序后的副本，不暴露可变的求值器内部状态。

更多示例见 [`examples/`](examples/)。

## CLI REPL

构建命令：

```bash
make build
```

启动 REPL：

```bash
./bin/elispvm
```

不进入 REPL，直接求值：

```bash
./bin/elispvm -eval '(concat "hello" " " "world")'
./bin/elispvm -file ./script.el
```

REPL 每次接收一个表达式，并支持多行列表。使用 `:help` 查看命令，使用 `:quit` 退出。在交互式 Linux 终端中，按 Tab 可以补全已注册函数、特殊表单、全局变量、`nil` 和 `t`；存在多个匹配项时，Tab 会打印候选列表。

运行示例脚本：

```bash
./bin/elispvm -file ./examples/scripts/basic.el
```

## 示例

- [`examples/embedding/main.go`](examples/embedding/main.go)：在 Go 程序中嵌入 VM，并注册宿主函数。
- [`examples/scripts/basic.el`](examples/scripts/basic.el)：用 CLI 执行的小型 Lisp 脚本。

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

## 测试

```bash
go test ./...
```

## 协议

MIT License. See [LICENSE](LICENSE).
