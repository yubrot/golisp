package golisp

type UndefinedVariable struct {
	Name string
}

func (e UndefinedVariable) Error() string {
	return "Undefined variable: " + e.Name
}

type Env struct {
	current map[string]Value
	parent  *Env
}

func NewEnv(parent *Env) *Env {
	return &Env{parent: parent, current: map[string]Value{}}
}

func (env *Env) Def(k string, v Value) {
	env.current[k] = v
}

func (env *Env) Set(k string, v Value) {
	for env != nil {
		_, ok := env.current[k]
		if ok {
			env.current[k] = v
			return
		}
		env = env.parent
	}
	panic(UndefinedVariable{k})
}

func (env *Env) Find(k string) Value {
	for env != nil {
		v, ok := env.current[k]
		if ok {
			return v
		}
		env = env.parent
	}
	return nil
}

func (env *Env) Get(k string) Value {
	v := env.Find(k)
	if v != nil {
		return v
	}
	panic(UndefinedVariable{k})
}

func (env *Env) refer(v Value) Value {
	sym, ok := v.(Sym)
	if !ok {
		return nil
	}
	return env.Find(sym.Data)
}
