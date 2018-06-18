package golisp

func syntaxEnv() *Env {
	env := NewEnv(nil)
	env.Def("def", syntax{syntaxDef{}})
	env.Def("set!", syntax{syntaxSet{}})
	env.Def("begin", syntax{syntaxBegin{}})
	env.Def("if", syntax{syntaxIf{}})
	env.Def("fun", syntax{syntaxFun{}})
	env.Def("macro", syntax{syntaxMacro{}})
	env.Def("builtin", syntax{syntaxBuiltin{}})
	env.Def("quote", syntax{syntaxQuote{}})
	return env
}

type expandAll struct{}
type noexpandFirst struct{}

func (expandAll) Expand(context *Context, args []Value) {
	for i, arg := range args {
		args[i] = context.macroExpand(true, arg)
	}
}

func (noexpandFirst) Expand(context *Context, args []Value) {
	for i, arg := range args {
		if i == 0 {
			continue
		}
		args[i] = context.macroExpand(true, arg)
	}
}

type syntaxDef struct{ noexpandFirst }

func (syntaxDef) Compile(compileEnv *Env, args []Value) Code {
	if len(args) == 2 {
		if sym, ok := args[0].(Sym); ok {
			return append(compile(compileEnv, args[1]), def{sym.Data}, ldc{Nil{}})
		}
	}
	panic(EvaluationError{"Syntax error: expected (def sym x)"})
}

type syntaxSet struct{ noexpandFirst }

func (syntaxSet) Compile(compileEnv *Env, args []Value) Code {
	if len(args) == 2 {
		if sym, ok := args[0].(Sym); ok {
			return append(compile(compileEnv, args[1]), set{sym.Data}, ldc{Nil{}})
		}
	}
	panic(EvaluationError{"Syntax error: expected (set! sym x)"})
}

type syntaxBegin struct{ expandAll }

func (syntaxBegin) Compile(compileEnv *Env, args []Value) Code {
	if len(args) == 0 {
		return Code{ldc{Nil{}}}
	}

	c := compile(compileEnv, args[0])
	for _, arg := range args[1:] {
		c = append(c, pop{})
		c = append(c, compile(compileEnv, arg)...)
	}
	return c
}

type syntaxIf struct{ expandAll }

func (syntaxIf) Compile(compileEnv *Env, args []Value) Code {
	if len(args) == 3 {
		return append(
			compile(compileEnv, args[0]),
			sel{
				append(compile(compileEnv, args[1]), leave{}),
				append(compile(compileEnv, args[2]), leave{}),
			})
	}

	panic(EvaluationError{"Syntax error: expected (if cond then else)"})
}

type syntaxFun struct{ noexpandFirst }

func (syntaxFun) Compile(compileEnv *Env, args []Value) Code {
	if len(args) > 0 {
		pat := buildPattern(args[0])
		body := syntaxBegin{}.Compile(compileEnv, args[1:])
		body = append(body, leave{})
		return Code{ldf{pat, body}}
	}

	panic(EvaluationError{"Syntax error: expected (fun pattern body...)"})
}

type syntaxMacro struct{ noexpandFirst }

func (syntaxMacro) Compile(compileEnv *Env, args []Value) Code {
	if len(args) > 0 {
		pat := buildPattern(args[0])
		body := syntaxBegin{}.Compile(compileEnv, args[1:])
		return Code{ldm{pat, body}}
	}

	panic(EvaluationError{"Syntax error: expected (macro pattern body...)"})
}

type syntaxBuiltin struct{ noexpandFirst }

func (syntaxBuiltin) Compile(compileEnv *Env, args []Value) Code {
	if len(args) == 1 {
		if sym, ok := args[0].(Sym); ok {
			return Code{ldb{sym.Data}}
		}
	}
	panic(EvaluationError{"Syntax error: expected (builtin sym)"})
}

type syntaxQuote struct{ noexpandFirst }

func (syntaxQuote) Compile(compileEnv *Env, args []Value) Code {
	if len(args) == 1 {
		return Code{ldc{args[0]}}
	}
	panic(EvaluationError{"Syntax error: expected (quote expr)"})
}
