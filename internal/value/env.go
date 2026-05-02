package value

type Env struct {
	parent *Env
	vars   map[string]Value
}

func NewEnv(parent *Env) *Env {
	return &Env{
		parent: parent,
		vars:   make(map[string]Value),
	}
}

func (e *Env) Lookup(name string) (Value, bool) {
	for env := e; env != nil; env = env.parent {
		if v, ok := env.vars[name]; ok {
			return v, true
		}
	}
	return nil, false
}

func (e *Env) Define(name string, v Value) {
	e.vars[name] = v
}

func (e *Env) SetGlobal(name string, v Value) {
	root := e
	for root.parent != nil {
		root = root.parent
	}
	root.Define(name, v)
}
