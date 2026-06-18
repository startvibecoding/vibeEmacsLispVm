# Elisp 子集扩展计划

本文档记录 `vibeEmacsLispVm` 后续扩展 Emacs Lisp 子集的方向。目标是让 VM 更接近可用的 Elisp 语言核心，同时保持可嵌入、可预测，并且不在 VM core 中内置 IO 或 Emacs runtime 能力。

## 总原则

- 只实现合法 Emacs Lisp 语法的子集，不发明自定义语法。
- 新增 special form、macro、function 名称必须能在 GNU Emacs Lisp Reference Manual 中找到，或者明确属于已支持语法的内部实现细节。
- VM core 可以实现纯语言能力，包括控制流、函数、宏、数据结构、内存 buffer 对象和纯计算内置函数。
- VM core 不内置文件、shell、网络、进程、包加载、window、frame、timer 等 IO 或宿主 UI runtime API。
- IO 能力只应由宿主应用通过注册函数提供，或者由 CLI/REPL 层提供。
- 新增语言行为必须有测试覆盖。
- 保持错误信息明确，优先选择小而确定的语义。
- 允许实现比 GNU Emacs 更窄的语义子集，例如早期数字函数只接受 `Number` 而暂不接受 marker；但不允许新增非 Elisp 的语法或函数名。

## 核对依据

规划中的语法和函数名应优先按 GNU Emacs Lisp Reference Manual 核对：

- [Conditionals](https://www.gnu.org/software/emacs/manual/html_node/elisp/Conditionals.html)
- [Iteration](https://www.gnu.org/software/emacs/manual/html_node/elisp/Iteration.html)
- [Catch and Throw](https://www.gnu.org/software/emacs/manual/html_node/elisp/Catch-and-Throw.html)
- [Comparison of Numbers](https://www.gnu.org/software/emacs/manual/html_node/elisp/Comparison-of-Numbers.html)
- [Arithmetic Operations](https://www.gnu.org/software/emacs/manual/html_node/elisp/Arithmetic-Operations.html)
- [Comparison of Characters and Strings](https://www.gnu.org/software/emacs/manual/html_node/elisp/Text-Comparison.html)
- [Type Predicates](https://www.gnu.org/software/emacs/manual/html_node/elisp/Type-Predicates.html)
- [Predicates on Lists](https://www.gnu.org/software/emacs/manual/html_node/elisp/List_002drelated-Predicates.html)
- [Functions](https://www.gnu.org/software/emacs/manual/html_node/elisp/Functions.html)
- [Macros](https://www.gnu.org/software/emacs/manual/html_node/elisp/Macros.html)
- [Backquote](https://www.gnu.org/software/emacs/manual/html_node/elisp/Backquote.html)
- [Buffers](https://www.gnu.org/software/emacs/manual/html_node/elisp/Buffers.html)
- [Buffer Contents](https://www.gnu.org/software/emacs/manual/html_node/elisp/Buffer-Contents.html)
- [Positions](https://www.gnu.org/software/emacs/manual/html_node/elisp/Positions.html)
- [Markers](https://www.gnu.org/software/emacs/manual/html_node/elisp/Markers.html)

## VM Core 与 REPL/宿主边界

VM core 可以包含：

- parser / reader
- evaluator
- lexical environment
- special forms
- pure builtins
- function and macro support
- in-memory buffer and marker objects
- deterministic runtime values

VM core 不包含：

- `print`
- `message`
- `read`
- `load`
- `require`
- `provide`
- `find-file`
- `insert-file-contents`
- `write-region`
- `shell-command`
- `call-process`
- `start-process`
- `url-retrieve`
- window / frame APIs
- package loading APIs
- filesystem-backed buffer APIs

CLI/REPL 可以负责：

- 读取用户输入
- 打印求值结果
- 处理 REPL 命令，例如 `:help`、`:quit`
- 可选注册调试或交互辅助函数

宿主应用可以通过 `RegisterFunc` / `RegisterSpecial` 显式注册自己的 IO 或副作用函数。VM core 不默认提供这些能力。

## 第一阶段：纯计算核心

优先补齐常用判断、控制流和算术能力。

### 控制流

- `while`
- `cond`
- `catch`
- `throw`

说明：

- `while` 使用标准 Elisp 语法：

  ```elisp
  (while condition
    body...)
  ```

- 嵌套循环跳出不引入自定义 `break` / `continue` 语法，使用 Elisp 风格 `catch` / `throw`：

  ```elisp
  (catch 'done
    (while outer-cond
      (while inner-cond
        (throw 'done "finished"))))
  ```

- 每次循环迭代应检查 `context.Context`，避免不可取消的无限循环。

### 数字函数

- `+`
- `-`
- `*`
- `/`
- `=`
- `/=`
- `<`
- `<=`
- `>`
- `>=`

建议语义：

- GNU Emacs 的相关函数通常接受 number-or-marker；在 marker 阶段完成前，本 VM 先只接受 `Number`。实现 marker 后，可以让数字比较函数接受 marker 作为更完整的 Elisp 子集。
- `+` 零参数返回 `0`。
- `*` 零参数返回 `1`。
- `-` 支持一元取负和多元减法。
- `/` 支持一元倒数和多元除法。
- 除零返回错误。

### 相等判断

- `eq`
- `equal`

建议语义：

- `eq` 先实现为确定性的小子集语义，适用于 `nil`、symbol、string、number 等简单值。
- `equal` 做结构相等，支持 `nil`、symbol、string、number、list。
- 暂不模拟完整 Emacs Lisp 对象身份语义。

### 字符串判断

- `string=`
- `string-equal`
- `string-lessp`
- `string<`
- `string-greaterp`
- `string>`

说明：

- `string-equal` 可作为 `string=` 的别名。
- `string-lessp` 可作为 `string<` 的别名。
- `string-greaterp` 可作为 `string>` 的别名。
- 不规划 `string/=`、`string<=`、`string>=`，因为它们不属于 GNU Emacs Lisp Reference Manual 中列出的字符串比较函数。
- GNU Emacs 允许字符串比较函数接受 string 或 symbol；本 VM 可以先只支持 string，并把 symbol 参数支持作为后续兼容项，但不能新增非 Elisp 函数名。
- 比较规则应尽量接近 Emacs 的字符码比较语义。若实现受 VM 字符模型限制，应在 README 中说明。

### 类型判断

- `null`
- `symbolp`
- `stringp`
- `numberp`
- `listp`
- `consp`
- `atom`

## 第二阶段：列表和基础数据操作

补齐不涉及 IO 的基础列表函数：

- `cons`
- `car`
- `cdr`
- `nth`
- `append`
- `reverse`
- `member`
- `assoc`

建议语义：

- 优先支持 proper list。
- dotted pair 可以后续单独评估；如果暂不支持，应明确报错或保持 parser 不接受相关语法。
- 列表函数应避免修改输入值，保持 VM 行为容易推理。

## 第三阶段：函数能力

支持用户在 Lisp 内定义和调用函数：

- `lambda`
- `defun`
- `funcall`
- `apply`
- `let*`

建议语义：

- 函数闭包捕获词法环境。
- `defun` 应按 Elisp 的函数命名空间实现，不应混入变量值命名空间。
- `apply` 最后一个参数必须是 list。
- 当前 VM 已使用词法环境；后续函数实现应明确这是 `lexical-binding` 风格的 Elisp 子集。暂不实现完整 Emacs Lisp dynamic binding，除非后续明确需要。

## 第四阶段：宏系统

支持合法 Elisp 宏语法：

- `defmacro`
- `macroexpand-1`
- `macroexpand`

宏语义：

- 宏接收未求值参数。
- 宏返回一个 Lisp 表达式。
- 返回的表达式在调用点继续求值。
- 宏本身不拥有任何隐式 IO 能力。
- 如果宏展开后调用宿主注册的副作用函数，该副作用属于宿主注册能力，不属于 VM core 内置能力。

宏系统需要 reader 支持：

- backquote: `` `expr ``
- comma: `,expr`
- comma-splice: `,@expr`

这些是合法 Elisp 语法，可以作为宏阶段的一部分实现。

## 第五阶段：内存 Buffer 和 Marker

可以实现 buffer，但必须限制为内存文本对象子集，不把文件、窗口、进程或 Emacs UI runtime 引入 VM core。

### Buffer 对象

优先支持：

- `bufferp`
- `buffer-name`
- `current-buffer`
- `set-buffer`
- `get-buffer`
- `get-buffer-create`
- `generate-new-buffer`
- `kill-buffer`
- `with-current-buffer`
- `save-current-buffer`

建议语义：

- buffer 是 VM 内部的内存对象，包含名称、文本内容、当前 point 和必要的局部状态。
- `current-buffer` 是 evaluator 执行上下文的一部分，不依赖真实 Emacs 进程。
- `kill-buffer` 只释放 VM 内部 buffer，不触发文件保存、进程处理或用户确认。
- `with-current-buffer` / `save-current-buffer` 应按 Elisp 语法实现为 special form 或 macro，不发明替代语法。

### Buffer 内容和位置

优先支持：

- `point`
- `point-min`
- `point-max`
- `goto-char`
- `insert`
- `delete-region`
- `buffer-substring`
- `buffer-string`
- `erase-buffer`

建议语义：

- 所有操作只作用于 VM 内存 buffer。
- 不实现 text properties，`buffer-substring` 先返回普通 string。
- 位置采用 Elisp 的 1-based buffer position 语义，而不是 Go 字符串下标。
- 字符计数和字节计数的兼容策略需要在实现前明确；早期可以以 rune 位置作为更窄子集，但文档必须说明。

### Marker

优先支持：

- `markerp`
- `make-marker`
- `point-marker`
- `copy-marker`
- `marker-position`
- `marker-buffer`
- `set-marker`

建议语义：

- marker 指向某个 VM buffer 的位置。
- 当 buffer 插入或删除文本时，marker 位置应按 Elisp 规则尽量跟随调整。
- marker 只关联 VM 内存 buffer，不关联真实 Emacs buffer。
- marker 完成后，再考虑让数字比较函数支持 number-or-marker。

### 仍不进入 VM Core 的 Buffer 相关能力

以下能力涉及文件、窗口、显示、用户交互或完整 Emacs runtime，仍不进入 VM core：

- `find-file`
- `find-buffer-visiting`
- `insert-file-contents`
- `write-region`
- `save-buffer`
- `switch-to-buffer`
- `display-buffer`
- `pop-to-buffer`
- `get-buffer-window`
- buffer-local variables 的完整体系
- major mode / minor mode
- hooks
- text properties
- overlays

## 暂不实现

以下内容暂不进入 VM core：

- `cl-loop`
- `cl-lib`
- Common Lisp 风格 `loop`
- 除 quote、backquote、comma、comma-splice 等明确列入的 reader syntax 外的完整 reader 体系
- vector
- hash table
- char table
- overlay / text property
- package system
- byte compiler
- advice system
- hooks
- timers
- process / network / shell / filesystem APIs

`cl-loop` 属于 `cl-lib` 宏体系，不适合作为早期 core 直接实现。Common Lisp 风格 `loop` 不作为 VM 自创语法引入。等基础宏系统稳定后，可以再评估是否以库或可选扩展形式支持 `cl-lib` 的合法 Elisp 子集。

## 文档和测试要求

每次新增语言能力时，应同步更新：

- `README.md`
- `README_zh.md`
- `AGENTS.md`
- 相关 `*_test.go`

每个新增 special form 或 builtin 至少覆盖：

- 正常求值
- 参数数量错误
- 参数类型错误
- 嵌套或短路行为
- context cancellation，如果该能力可能长时间运行
