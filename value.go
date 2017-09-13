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

func (_ fun) Inspect() string {
	return "<fun>"
}

func (_ builtin) Inspect() string {
	return "<builtin>"
}

func (_ macro) Inspect() string {
	return "<macro>"
}

func (_ syntax) Inspect() string {
	return "<syntax>"
}

func (vec Vec) Inspect() string {
	return Cons{Sym{"vec"}, List(vec.Payload...)}.Inspect()
}

func (_ fun) procValue()     {}
func (_ builtin) procValue() {}

func (_ macro) metaValue()  {}
func (_ syntax) metaValue() {}
