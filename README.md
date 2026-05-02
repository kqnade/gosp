# gosp

A minimal McCarthy-style Lisp implemented in Go. Runs as a standalone
CLI and embeds as a Go library. Two execution backends — a
tree-walking evaluator and a stack-based bytecode VM — are
cross-validated against a shared corpus.

## Language

Strict McCarthy 1960 / Roots of Lisp:

- **Atoms** are symbols only. `t` is true; `()` / `nil` is false. There
  are no number, string, or boolean literals.
- **Data**: atoms and pairs (`Cons` cells).
- **Primitives** (5): `car`, `cdr`, `cons`, `atom`, `eq`.
- **Special forms** (4): `quote`, `cond`, `lambda`, `label`.
- `'x` is reader sugar for `(quote x)`.
- `;` starts a line comment.
- Truthiness: only `()` is false.
- `(label F EXPR)` at the top level mutates the global env; nested
  inside an expression it is McCarthy's local recursive binding.

## Build

```sh
go build ./...
```

## Run as a CLI

Run a file with the default tree-walking evaluator:

```sh
go run ./cmd/gosp examples/metacircular.lisp
```

Run the same file through the bytecode VM:

```sh
go run ./cmd/gosp -vm examples/metacircular.lisp
```

Start a REPL (add `-vm` to use the VM, `Ctrl-D` to exit):

```sh
go run ./cmd/gosp
> (cons (car '(a b c)) (cdr '(d e)))
(a e)
> (label cadr (lambda (x) (car (cdr x))))
<unknown>
> (cadr '(a b c))
b
```

## Embed as a library

`gosp` ships a small public API for hosting the language in a Go
program and registering Go-defined builtins.

```sh
go get github.com/kqnade/gosp
```

```go
package main

import (
    "fmt"

    "github.com/kqnade/gosp"
    "github.com/kqnade/gosp/value"
)

func main() {
    rt := gosp.New() // default: tree-walking; gosp.WithBackend(gosp.BackendVM) for the VM

    // Add a host-defined builtin reachable from Lisp.
    rt.Register("greet", func(args []value.Value) (value.Value, error) {
        if len(args) != 1 {
            return nil, fmt.Errorf("greet: arity")
        }
        return value.Cons(value.Symbol{Name: "hello"}, value.Cons(args[0], value.NIL)), nil
    })

    v, _ := rt.Eval("(greet 'world)")
    fmt.Println(gosp.Print(v)) // (hello world)
}
```

The full runnable version lives in [`examples/embed/`](examples/embed/).

### Public API surface (v0.x)

The public API is intentionally small. Only these packages have
stability commitments — everything under `internal/` is implementation
detail.

| Package | Use |
|---|---|
| `github.com/kqnade/gosp` | `Runtime`, `New`, `WithBackend`, `Register`, `Eval`, `EvalForm`, `RunFile`, `REPL`, `Print` |
| `github.com/kqnade/gosp/value` | `Value`, `Symbol`, `Pair`, `Nil`, `Func`, `Builtin`, `Closure`, `Env`, `Cons`, `List`, `NIL`, `IsNil`, `IsAtom`, `Eq` |
| `github.com/kqnade/gosp/ext/...` | Optional curated extensions (start with [`ext/example/`](ext/example/)) |

> **API stability:** gosp is at v0.x. Minor versions may break the
> public API while the language and extension story stabilize. v1.0
> will lock the contract.

### Writing an extension

Extensions are plain Go packages exposing `Register(rt *gosp.Runtime)`.
[`ext/example/`](ext/example/) is the minimal template — copy it to
add your own builtins (numbers, strings, I/O, ...).

```go
package mathext

import (
    "github.com/kqnade/gosp"
    "github.com/kqnade/gosp/value"
)

func Register(rt *gosp.Runtime) {
    rt.Register("zero?", func(args []value.Value) (value.Value, error) {
        // ...
    })
}
```

Use it from your host:

```go
rt := gosp.New()
mathext.Register(rt)
```

## Examples

```lisp
; First two arguments composed via cons.
(cons (car '(a b c)) (cdr '(d e)))
; => (a e)

; Curried lambda.
(((lambda (x) (lambda (y) (cons x (cons y '()))))
   'a)
  'b)
; => (a b)

; Recursive walker via label.
((label ff (lambda (x)
   (cond ((atom x) x)
         (t (ff (car x))))))
 '((a b) c))
; => a
```

See `examples/metacircular.lisp` for McCarthy's `eval` written in
itself, and `testdata/programs/` for the small parity corpus.

## Testing

```sh
go test -race ./...
```

A parity test (`internal/integration/parity_test.go`) runs each
program in `testdata/programs/` through both backends and asserts
they agree with each other and with the `*.expected` file. A 100k
tail-call stress test lives in `internal/vm/stress_test.go` and is
skipped under `-short`.

## Layout

```text
gosp/
├── gosp.go                  # public facade: Runtime, New, Register, Eval, REPL
├── value/                   # public types: Value, Symbol, Pair, Nil, Env, Builtin, ...
├── ext/                     # curated extensions (Register(rt))
│   └── example/             # minimal template for derivative extension packages
├── examples/                # *.lisp samples + embed/ Go example
├── cmd/gosp/                # CLI: file runner and REPL, with -vm
├── internal/                # implementation detail (no stability guarantees)
│   ├── reader/              # lexer + parser
│   ├── printer/             # value pretty-printer
│   ├── eval/                # tree-walker and primitive builtins
│   ├── compiler/            # source → vm.Code
│   ├── vm/                  # opcodes, Code, run loop
│   └── integration/         # corpus, metacircular, parity
└── testdata/programs/       # *.lisp + *.expected
```

## References

- John McCarthy, *Recursive Functions of Symbolic Expressions and
  Their Computation by Machine, Part I* (CACM, 1960).
- Paul Graham, *The Roots of Lisp* (2002).
