package golisp

import "strconv"

type Num struct {
	Data float64
}

type Sym struct {
	Data string
}

type Str struct {
	Data string
}

type Cons struct {
	Car Value
	Cdr Value
}

type Nil struct{}

type Bool struct {
	Data bool
}

func List(values ...Value) Value {
	var head Value = Nil{}
	for i := range values {
		head = Cons{values[len(values)-1-i], head}
	}
	return head
}

func Slice(value Value) (slice []Value, ok bool) {
	for {
		switch v := value.(type) {
		case Nil:
			ok = true
			return
		case Cons:
			slice = append(slice, v.Car)
			value = v.Cdr
		default:
			return
		}
	}
}

func Test(value Value) bool {
	switch v := value.(type) {
	case Bool:
		return v.Data
	default:
		return true
	}
}

func Quote(v Value) Value {
	return List(Sym{"quote"}, v)
}

func Quasiquote(v Value) Value {
	return List(Sym{"quasiquote"}, v)
}

func Unquote(v Value) Value {
	return List(Sym{"unquote"}, v)
}

func UnquoteSplicing(v Value) Value {
	return List(Sym{"unquote-splicing"}, v)
}

func (num Num) Inspect() string {
	return strconv.FormatFloat(num.Data, 'g', -1, 64)
}

func (sym Sym) Inspect() string {
	return sym.Data
}

func (str Str) Inspect() string {
	return strconv.Quote(str.Data)
}

func (cons Cons) Inspect() string {
	ss, ok := cons.inspectSyntaxSugar()
	if ok {
		return ss
	}
	return "(" + cons.inspectInner() + ")"
}

func (_ Nil) Inspect() string {
	return "()"
}

func (b Bool) Inspect() string {
	if b.Data {
		return "#t"
	} else {
		return "#f"
	}
}

func (cons Cons) inspectSyntaxSugar() (string, bool) {
	if sym, ok := cons.Car.(Sym); ok {
		if cdr, ok := cons.Cdr.(Cons); ok {
			if _, ok := cdr.Cdr.(Nil); ok {
				switch sym.Data {
				case "quote":
					return "'" + cdr.Car.Inspect(), true
				case "quasiquote":
					return "`" + cdr.Car.Inspect(), true
				case "unquote":
					return "," + cdr.Car.Inspect(), true
				case "unquote-splicing":
					return ",@" + cdr.Car.Inspect(), true
				}
			}
		}
	}
	return "", false
}

func (cons Cons) inspectInner() (r string) {
	for {
		r += cons.Car.Inspect()
		switch cdr := cons.Cdr.(type) {
		case Nil:
			return
		case Cons:
			r += " "
			cons = cdr
		default:
			r += " . " + cons.Cdr.Inspect()
			return
		}
	}
	return
}
