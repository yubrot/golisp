package golisp

import "fmt"

type EvaluationError struct {
	Msg string
}

type InternalError struct {
	Msg string
}

func (e EvaluationError) Error() string {
	return "Evaluation error: " + e.Msg
}

func (e InternalError) Error() string {
	return "Internal error: " + e.Msg
}

type Context struct {
	toplevel *Env
	Builtins map[string] BuiltinImpl
}

type State struct {
	Cont
	Context *Context
}

type Cont struct {
	stack []Value
	env *Env
	code Code
	dump []dump
}

type dump struct {
	env *Env
	code Code
}

type SyntaxImpl interface {
	Expand(context *Context, args []Value)
	Compile(compileEnv *Env, args []Value) Code
}

type BuiltinImpl interface {
	Run(state *State, args []Value)
}

func compile(compileEnv *Env, expr Value) Code {
	switch expr := expr.(type) {
	case Sym:
		return Code{ldv{expr.Data}}

	case Cons:
		slice, ok := Slice(expr)
		if !ok {
			panic(InternalError{"Improper list: " + expr.Inspect()})
		}

		if syntax, ok := compileEnv.refer(slice[0]).(syntax); ok {
			args := slice[1:]
			return syntax.Compile(compileEnv, args)
		}

		code := Code{}
		for _, v := range slice {
			c := compile(compileEnv, v)
			code = append(code, c...)
		}
		return append(code, app{len(slice) - 1})

	default:
		return Code{ldc{expr}}
	}
}

func (state *State) Push(value Value) {
	state.stack = append(state.stack, value)
}

func (state *State) pop() Value {
	if len(state.stack) == 0 {
		panic(InternalError{"Inconsistent stack"})
	}
	ret := state.stack[len(state.stack)-1]
	state.stack = state.stack[:len(state.stack)-1]
	return ret
}

func (state *State) enter(env *Env, code Code) {
	skipThisFrame := false
	if len(state.code) == 1 {
		_, skipThisFrame = state.code[0].(leave)
	}

	if !skipThisFrame {
		state.dump = append(state.dump, dump{state.env, state.code})
	}
	state.env = env
	state.code = code
}

func (state *State) leave() {
	if len(state.dump) == 0 {
		panic(InternalError{"Inconsistent dump"})
	}
	dump := state.dump[len(state.dump)-1]
	state.dump = state.dump[:len(state.dump)-1]
	state.env = dump.env
	state.code = dump.code
}

func (state *State) Apply(f Value, args ...Value) {
	switch f := f.(type) {
	case fun:
		state.enter(NewEnv(f.env), f.code)
		f.pattern.bind(args, state.env)

	case builtin:
		f.Run(state, args)

	default:
		panic(EvaluationError{"Cannot call: " + f.Inspect()})
	}
}

func (dest *Cont) copy(src Cont) {
	copied := copy(dest.stack, src.stack)
	if copied < len(src.stack) {
		dest.stack = append(dest.stack, src.stack[copied:]...)
	} else {
		dest.stack = dest.stack[:copied]
	}
	dest.env = src.env
	dest.code = src.code
	copied = copy(dest.dump, src.dump)
	if copied < len(src.dump) {
		dest.dump = append(dest.dump, src.dump[copied:]...)
	} else {
		dest.dump = dest.dump[:copied]
	}
}

func (state *State) CaptureCont() Value {
	var cont Cont
	cont.copy(state.Cont)
	return builtin{cont}
}

func (cont Cont) Run(state *State, args []Value) {
	if len(args) > 1 {
		panic(EvaluationError{"Multiple values are not implemented"})
	}

	state.copy(cont)

	if len(args) == 1 {
		state.Push(args[0])
	} else {
		state.Push(Nil{})
	}
}

func (state *State) runInst(i inst) {
	switch i := i.(type) {
	case ldc:
		state.Push(i.value)

	case ldv:
		state.Push(state.env.Get(i.name))

	case ldf:
		state.Push(fun{state.env, i.pattern, i.code})

	case ldm:
		state.Push(macro{state.env, i.pattern, i.code})

	case ldb:
		impl, ok := state.Context.Builtins[i.name]
		if !ok {
			panic(EvaluationError{"Unsupported builtin: " + i.name})
		}
		state.Push(builtin{impl})

	case sel:
		var branchCode Code
		if Test(state.pop()) {
			branchCode = i.a
		} else {
			branchCode = i.b
		}
		state.enter(NewEnv(state.env), branchCode)

	case app:
		args := make([]Value, i.argc)
		for j := 0; j < i.argc; j++ {
			args[i.argc-j-1] = state.pop()
		}
		f := state.pop()
		state.Apply(f, args...)

	case leave:
		state.leave()

	case pop:
		state.pop()

	case def:
		v := state.pop()
		state.env.Def(i.name, v)

	case set:
		v := state.pop()
		state.env.Set(i.name, v)
	}
}

func (state *State) run() Value {
	for len(state.code) != 0 {
		inst := state.code[0]
		state.code = state.code[1:]
		state.runInst(inst)
	}
	return state.pop()
}

func (context *Context) exec(env *Env, code Code) Value {
	state := State{Cont{env: env, code: code}, context}
	return state.run()
}

func (context *Context) macroExpand(recurse bool, expr Value) Value {
	slice, ok := Slice(expr)
	if ok && len(slice) != 0 {
		args := slice[1:]
		switch m := context.toplevel.refer(slice[0]).(type) {
		case macro:
			env := NewEnv(m.env)
			m.pattern.bind(args, env)
			expr = context.exec(env, m.code)
			if !recurse { return expr }
			return context.macroExpand(true, expr)

		case syntax:
			if !recurse { return expr }
			m.Expand(context, args)
			return List(slice...)
		}
	}

	if !recurse { return expr }
	return context.macroExpandChildren(expr)
}

func (context *Context) macroExpandChildren(expr Value) Value {
	cons, ok := expr.(Cons)
	if !ok { return expr }
	return Cons{
		Car: context.macroExpand(true, cons.Car),
		Cdr: context.macroExpandChildren(cons.Cdr),
	}
}

func recoverContext(err *error) {
	if r := recover(); r != nil {
		var ok bool
		*err, ok = r.(error)
		if !ok {
			*err = InternalError{fmt.Sprintf("panic: %v", r)}
		}
	}
	return
}

func NewContext() *Context {
	return &Context{
		toplevel: NewEnv(syntaxEnv()),
		Builtins: map[string]BuiltinImpl{},
	}
}

func (context *Context) Compile(expr Value) (result Code, err error) {
	defer recoverContext(&err)
	result = compile(context.toplevel, expr)
	return
}

func (context *Context) MacroExpand(recurse bool, expr Value) (result Value, err error) {
	defer recoverContext(&err)
	result = context.macroExpand(recurse, expr)
	return
}

func (context *Context) Eval(expr Value) (result Value, err error) {
	defer recoverContext(&err)
	expr = context.macroExpand(true, expr)
	code := compile(context.toplevel, expr)
	result = context.exec(context.toplevel, code)
	return
}
