# gosp

A minimal McCarthy-style Lisp implemented in Go, with two execution
backends — a tree-walking evaluator and a stack-based bytecode VM —
that can be cross-validated against a shared corpus.

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

## Run

Run a file with the default tree-walking evaluator:

```sh
go run ./cmd/gosp examples/metacircular.lisp
```

Run the same file through the bytecode VM:

```sh
go run ./cmd/gosp -vm examples/metacircular.lisp
```

Start a REPL (add `-vm` to use the VM):

```sh
go run ./cmd/gosp
> (cons (car '(a b c)) (cdr '(d e)))
(a e)
> (label cadr (lambda (x) (car (cdr x))))
<unknown>
> (cadr '(a b c))
b
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
├── cmd/gosp/                # CLI: file runner and REPL, with -vm
├── internal/
│   ├── value/               # Symbol, Pair, Nil, Func, Builtin, Closure, Env
│   ├── reader/              # lexer + parser
│   ├── printer/             # value pretty-printer
│   ├── eval/                # tree-walker and primitive builtins
│   ├── compiler/            # source → vm.Code
│   ├── vm/                  # opcodes, Code, run loop
│   └── integration/         # corpus, metacircular, parity
├── examples/                # *.lisp samples
└── testdata/programs/       # *.lisp + *.expected
```

## References

- John McCarthy, *Recursive Functions of Symbolic Expressions and
  Their Computation by Machine, Part I* (CACM, 1960).
- Paul Graham, *The Roots of Lisp* (2002).
