package golisp

import "strconv"

type pattern struct {
	fixed []string
	rest string
}

func buildPattern(value Value) (pattern pattern) {
	for {
		switch v := value.(type) {
		case Sym:
			pattern.rest = v.Data
			return

		case Nil:
			return

		case Cons:
			if car, ok := v.Car.(Sym); ok {
				pattern.fixed = append(pattern.fixed, car.Data)
				value = v.Cdr
				continue
			} else {
				panic(EvaluationError{"Unsupported pattern: " + v.Car.Inspect()})
			}

		default:
			panic(EvaluationError{"Unsupported pattern: " + v.Inspect()})
		}
	}
}

func (pattern pattern) bind(args []Value, env *Env) {
	if len(args) < len(pattern.fixed) {
		var prefix string
		if pattern.rest == "" { prefix = "at least " }
		panic(EvaluationError{"This function takes " + prefix + strconv.Itoa(len(pattern.fixed)) + " arguments"})
	}
	for _, param := range pattern.fixed {
		env.Def(param, args[0])
		args = args[1:]
	}
	if pattern.rest != "" {
		env.Def(pattern.rest, List(args...))
	}
}

func (pattern pattern) String() string {
	var head Value = Nil{}
	if pattern.rest != "" {
		head = Sym{pattern.rest}
	}

	for i := range pattern.fixed {
		head = Cons{
			Car: Sym{pattern.fixed[len(pattern.fixed)-1-i]},
			Cdr: head,
		}
	}
	return head.Inspect()
}
