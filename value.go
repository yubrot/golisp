package golisp

type Value interface {
	Inspect() string
}

type Proc interface {
	procValue()
}

type Meta interface {
	metaValue()
}

type fun struct {
	env     *Env
	pattern pattern
	code    Code
}

type builtin struct {
	BuiltinImpl
}

type macro struct {
	env     *Env
	pattern pattern
	code    Code
}

type syntax struct {
	SyntaxImpl
}

type Vec struct {
	Payload []Value
}

func (fun) Inspect() string {
	return "<fun>"
}

func (builtin) Inspect() string {
	return "<builtin>"
}

func (macro) Inspect() string {
	return "<macro>"
}

func (syntax) Inspect() string {
	return "<syntax>"
}

func (vec Vec) Inspect() string {
	return Cons{Sym{"vec"}, List(vec.Payload...)}.Inspect()
}

func (fun) procValue()     {}
func (builtin) procValue() {}

func (macro) metaValue()  {}
func (syntax) metaValue() {}
