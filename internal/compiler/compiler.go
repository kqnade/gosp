package compiler

import (
	"fmt"

	"github.com/kqnade/gosp/value"
	"github.com/kqnade/gosp/internal/vm"
)

// Option configures CompileProgram / Compile.
type Option func(*builder)

// WithRebound marks names as already redefined in the runtime's global
// environment. Primitive names listed here are dispatched through env
// lookup instead of being lowered to fixed opcodes, so a host
// registration or a top-level label from an earlier compile pass
// continues to shadow the primitive.
func WithRebound(names ...string) Option {
	return func(b *builder) {
		b.rebound = extendRebound(b.rebound, names...)
	}
}

func Compile(v value.Value, opts ...Option) (*vm.Code, error) {
	c := &builder{code: &vm.Code{}}
	for _, opt := range opts {
		opt(c)
	}
	if err := c.compile(v, false); err != nil {
		return nil, err
	}
	c.emit(vm.OpRet, 0)
	return c.code, nil
}

// CompileProgram compiles a sequence of top-level forms. A leading
// (label NAME EXPR) at top level mutates the global environment via
// OpDefineGlobal; in-expression label still produces a self-bound
// closure. The program returns the value of the last form.
func CompileProgram(forms []value.Value, opts ...Option) (*vm.Code, error) {
	c := &builder{code: &vm.Code{}}
	for _, opt := range opts {
		opt(c)
	}
	if len(forms) == 0 {
		c.emit(vm.OpLoadConst, c.addConst(value.NIL))
		c.emit(vm.OpRet, 0)
		return c.code, nil
	}
	for i, form := range forms {
		if err := c.compileTopLevel(form); err != nil {
			return nil, err
		}
		if i != len(forms)-1 {
			c.emit(vm.OpPop, 0)
		}
	}
	c.emit(vm.OpRet, 0)
	return c.code, nil
}

func (b *builder) compileTopLevel(form value.Value) error {
	if pair, ok := form.(*value.Pair); ok {
		if head, ok := pair.Car.(value.Symbol); ok && head.Name == "label" {
			return b.compileTopLevelLabel(pair.Cdr)
		}
	}
	return b.compile(form, false)
}

func (b *builder) compileTopLevelLabel(form value.Value) error {
	pair, ok := form.(*value.Pair)
	if !ok {
		return fmt.Errorf("gosp: compile: label: missing name")
	}
	name, ok := pair.Car.(value.Symbol)
	if !ok {
		return fmt.Errorf("gosp: compile: label: name must be a symbol")
	}
	rest, ok := pair.Cdr.(*value.Pair)
	if !ok {
		return fmt.Errorf("gosp: compile: label: missing expression")
	}
	if !value.IsNil(rest.Cdr) {
		return fmt.Errorf("gosp: compile: label: too many arguments")
	}
	if err := b.compileLabelBody(rest.Car, name.Name); err != nil {
		return err
	}
	b.emit(vm.OpDefineGlobal, b.addSym(name.Name))
	// Subsequent forms (and any lambda body compiled after this point)
	// must dispatch this name through env lookup, not a fixed opcode.
	b.rebound = extendRebound(b.rebound, name.Name)
	return nil
}

type builder struct {
	code *vm.Code
	// locals tracks names that are lexically bound in the current
	// compilation context (lambda parameters and label self-names).
	// When a primitive name like car/cdr/cons/atom/eq appears in
	// operator position and is shadowed by a local binding, the call
	// must go through env lookup rather than be lowered to a fixed
	// opcode. nil means top level (no local bindings).
	locals map[string]bool
	// rebound tracks names that have been redefined by a prior
	// top-level (label NAME ...) form in the same CompileProgram run.
	// A primitive whose name has been rebound must not be lowered to
	// its fixed opcode — it must dispatch through env lookup so the
	// new binding wins. The set is a snapshot at the time the current
	// builder was constructed; updates happen on the top-level builder
	// only, and lambda bodies inherit a snapshot.
	rebound map[string]bool
}

func (b *builder) isLocal(name string) bool {
	if b.locals == nil {
		return false
	}
	return b.locals[name]
}

func (b *builder) isRebound(name string) bool {
	if b.rebound == nil {
		return false
	}
	return b.rebound[name]
}

// extendBoolSet returns a new set containing every key from parent
// plus the given names. The parent is not modified, so callers can
// safely keep a snapshot reference (used by lambda bodies that inherit
// the locals/rebound state of their enclosing builder without seeing
// later mutations).
func extendBoolSet(parent map[string]bool, names ...string) map[string]bool {
	out := make(map[string]bool, len(parent)+len(names))
	for k, v := range parent {
		out[k] = v
	}
	for _, n := range names {
		out[n] = true
	}
	return out
}

func extendLocals(parent map[string]bool, names ...string) map[string]bool {
	return extendBoolSet(parent, names...)
}

func extendRebound(parent map[string]bool, names ...string) map[string]bool {
	return extendBoolSet(parent, names...)
}

func (b *builder) compile(v value.Value, tail bool) error {
	switch x := v.(type) {
	case value.Nil:
		b.emit(vm.OpLoadConst, b.addConst(value.NIL))
		return nil
	case value.Symbol:
		switch x.Name {
		case "t":
			b.emit(vm.OpLoadConst, b.addConst(x))
		case "nil":
			b.emit(vm.OpLoadConst, b.addConst(value.NIL))
		default:
			b.emit(vm.OpLoadVar, b.addSym(x.Name))
		}
		return nil
	case *value.Pair:
		return b.compilePair(x, tail)
	default:
		return fmt.Errorf("gosp: compile: cannot compile %T", v)
	}
}

func (b *builder) compilePair(p *value.Pair, tail bool) error {
	if head, ok := p.Car.(value.Symbol); ok {
		switch head.Name {
		case "quote":
			return b.compileQuote(p.Cdr)
		case "cond":
			return b.compileCond(p.Cdr, tail)
		case "lambda":
			return b.compileLambda(p.Cdr)
		case "label":
			return b.compileLabel(p.Cdr)
		}
		if !b.isLocal(head.Name) && !b.isRebound(head.Name) {
			switch head.Name {
			case "car":
				return b.compileUnary("car", vm.OpCar, p.Cdr)
			case "cdr":
				return b.compileUnary("cdr", vm.OpCdr, p.Cdr)
			case "cons":
				return b.compileBinary("cons", vm.OpCons, p.Cdr)
			case "atom":
				return b.compileUnary("atom", vm.OpAtom, p.Cdr)
			case "eq":
				return b.compileBinary("eq", vm.OpEq, p.Cdr)
			}
		}
	}
	return b.compileCall(p, tail)
}

func (b *builder) compileLabel(form value.Value) error {
	pair, ok := form.(*value.Pair)
	if !ok {
		return fmt.Errorf("gosp: compile: label: missing name")
	}
	name, ok := pair.Car.(value.Symbol)
	if !ok {
		return fmt.Errorf("gosp: compile: label: name must be a symbol")
	}
	rest, ok := pair.Cdr.(*value.Pair)
	if !ok {
		return fmt.Errorf("gosp: compile: label: missing expression")
	}
	if !value.IsNil(rest.Cdr) {
		return fmt.Errorf("gosp: compile: label: too many arguments")
	}
	if err := b.compileLabelBody(rest.Car, name.Name); err != nil {
		return err
	}
	b.emit(vm.OpMakeLabel, b.addSym(name.Name))
	return nil
}

// compileLabelBody compiles the expression bound by `label`. When the
// expression is a `lambda`, the label name is added to the lambda body's
// lexical scope so recursive calls by name go through env lookup rather
// than being lowered to a primitive opcode (e.g. for `(label car ...)`).
func (b *builder) compileLabelBody(expr value.Value, selfName string) error {
	if exprPair, ok := expr.(*value.Pair); ok {
		if head, ok := exprPair.Car.(value.Symbol); ok && head.Name == "lambda" {
			proto, err := buildFuncProto(exprPair.Cdr, extendLocals(b.locals, selfName), b.rebound)
			if err != nil {
				return err
			}
			idx := len(b.code.Funcs)
			b.code.Funcs = append(b.code.Funcs, proto)
			b.emit(vm.OpMakeClosure, idx)
			return nil
		}
	}
	return b.compile(expr, false)
}

func (b *builder) compileLambda(form value.Value) error {
	proto, err := buildFuncProto(form, b.locals, b.rebound)
	if err != nil {
		return err
	}
	idx := len(b.code.Funcs)
	b.code.Funcs = append(b.code.Funcs, proto)
	b.emit(vm.OpMakeClosure, idx)
	return nil
}

func buildFuncProto(form value.Value, parentLocals, parentRebound map[string]bool) (*vm.FuncProto, error) {
	pair, ok := form.(*value.Pair)
	if !ok {
		return nil, fmt.Errorf("gosp: compile: lambda: missing parameter list")
	}
	body, ok := pair.Cdr.(*value.Pair)
	if !ok {
		return nil, fmt.Errorf("gosp: compile: lambda: missing body")
	}
	if !value.IsNil(body.Cdr) {
		return nil, fmt.Errorf("gosp: compile: lambda: body must be a single expression")
	}
	params, err := parseParams(pair.Car)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(params))
	for i, p := range params {
		names[i] = p.Name
	}
	bb := &builder{
		code:    &vm.Code{},
		locals:  extendLocals(parentLocals, names...),
		rebound: parentRebound,
	}
	if err := bb.compile(body.Car, true); err != nil {
		return nil, err
	}
	bb.emit(vm.OpRet, 0)
	return &vm.FuncProto{Params: params, Code: bb.code}, nil
}

func parseParams(v value.Value) ([]value.Symbol, error) {
	var out []value.Symbol
	for !value.IsNil(v) {
		pair, ok := v.(*value.Pair)
		if !ok {
			return nil, fmt.Errorf("gosp: compile: lambda: parameter list must be proper")
		}
		sym, ok := pair.Car.(value.Symbol)
		if !ok {
			return nil, fmt.Errorf("gosp: compile: lambda: parameter must be a symbol")
		}
		out = append(out, sym)
		v = pair.Cdr
	}
	return out, nil
}

func (b *builder) compileCall(p *value.Pair, tail bool) error {
	args, err := flatArgs(p.Cdr)
	if err != nil {
		return err
	}
	if err := b.compile(p.Car, false); err != nil {
		return err
	}
	for _, a := range args {
		if err := b.compile(a, false); err != nil {
			return err
		}
	}
	op := vm.OpCall
	if tail {
		op = vm.OpTailCall
	}
	b.emit(op, len(args))
	return nil
}

func (b *builder) compileUnary(name string, op vm.Opcode, args value.Value) error {
	xs, err := flatArgs(args)
	if err != nil {
		return err
	}
	if len(xs) != 1 {
		return fmt.Errorf("gosp: compile: %s: wrong number of arguments", name)
	}
	if err := b.compile(xs[0], false); err != nil {
		return err
	}
	b.emit(op, 0)
	return nil
}

func (b *builder) compileBinary(name string, op vm.Opcode, args value.Value) error {
	xs, err := flatArgs(args)
	if err != nil {
		return err
	}
	if len(xs) != 2 {
		return fmt.Errorf("gosp: compile: %s: wrong number of arguments", name)
	}
	if err := b.compile(xs[0], false); err != nil {
		return err
	}
	if err := b.compile(xs[1], false); err != nil {
		return err
	}
	b.emit(op, 0)
	return nil
}

func flatArgs(v value.Value) ([]value.Value, error) {
	var out []value.Value
	for !value.IsNil(v) {
		pair, ok := v.(*value.Pair)
		if !ok {
			return nil, fmt.Errorf("gosp: compile: improper argument list")
		}
		out = append(out, pair.Car)
		v = pair.Cdr
	}
	return out, nil
}

func (b *builder) compileCond(clauses value.Value, tail bool) error {
	var endJumps []int
	for !value.IsNil(clauses) {
		pair, ok := clauses.(*value.Pair)
		if !ok {
			return fmt.Errorf("gosp: compile: cond: malformed clause list")
		}
		clause, ok := pair.Car.(*value.Pair)
		if !ok {
			return fmt.Errorf("gosp: compile: cond: clause must be a list")
		}
		body, ok := clause.Cdr.(*value.Pair)
		if !ok {
			return fmt.Errorf("gosp: compile: cond: clause must have body")
		}
		if !value.IsNil(body.Cdr) {
			return fmt.Errorf("gosp: compile: cond: clause must have exactly one body expression")
		}
		if err := b.compile(clause.Car, false); err != nil {
			return err
		}
		jifPos := len(b.code.Instrs)
		b.emit(vm.OpJumpIfFalse, 0)
		if err := b.compile(body.Car, tail); err != nil {
			return err
		}
		jumpPos := len(b.code.Instrs)
		b.emit(vm.OpJump, 0)
		endJumps = append(endJumps, jumpPos)
		b.code.Instrs[jifPos].Arg = len(b.code.Instrs)
		clauses = pair.Cdr
	}
	b.emit(vm.OpLoadConst, b.addConst(value.NIL))
	end := len(b.code.Instrs)
	for _, p := range endJumps {
		b.code.Instrs[p].Arg = end
	}
	return nil
}

func (b *builder) compileQuote(args value.Value) error {
	pair, ok := args.(*value.Pair)
	if !ok {
		return fmt.Errorf("gosp: compile: quote: wrong number of arguments")
	}
	if !value.IsNil(pair.Cdr) {
		return fmt.Errorf("gosp: compile: quote: wrong number of arguments")
	}
	b.emit(vm.OpLoadConst, b.addConst(pair.Car))
	return nil
}

func (b *builder) emit(op vm.Opcode, arg int) {
	b.code.Instrs = append(b.code.Instrs, vm.Instr{Op: op, Arg: arg})
}

func (b *builder) addConst(v value.Value) int {
	b.code.Consts = append(b.code.Consts, v)
	return len(b.code.Consts) - 1
}

func (b *builder) addSym(name string) int {
	b.code.Syms = append(b.code.Syms, name)
	return len(b.code.Syms) - 1
}
