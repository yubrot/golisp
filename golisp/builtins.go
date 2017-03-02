package main

import (
	"bytes"
	"fmt"
	. "github.com/yubrot/golisp"
	"math"
	"os"
)

func registerBuiltins(context *Context) {
	context.Builtins["cons"] = builtinCons{}

	context.Builtins["exit"] = builtinExit{}
	context.Builtins["error"] = builtinError{}

	context.Builtins["gensym"] = &builtinGensym{}

	context.Builtins["car"] = builtinCar{}
	context.Builtins["cdr"] = builtinCdr{}

	context.Builtins["apply"] = builtinApply{}

	context.Builtins["num?"] = builtinTest{"num?", isNum}
	context.Builtins["sym?"] = builtinTest{"sym?", isSym}
	context.Builtins["str?"] = builtinTest{"str?", isStr}
	context.Builtins["cons?"] = builtinTest{"cons?", isCons}
	context.Builtins["nil?"] = builtinTest{"nil?", isNil}
	context.Builtins["bool?"] = builtinTest{"bool?", isBool}
	context.Builtins["proc?"] = builtinTest{"proc?", isProc}
	context.Builtins["meta?"] = builtinTest{"meta?", isMeta}

	context.Builtins["+"] = builtinArithmetic{"+", add{}}
	context.Builtins["-"] = builtinArithmetic{"-", sub{}}
	context.Builtins["*"] = builtinArithmetic{"*", mul{}}
	context.Builtins["/"] = builtinArithmetic{"/", div{}}
	context.Builtins["%"] = builtinArithmetic{"%", mod{}}

	context.Builtins["concat"] = builtinConcat{}
	context.Builtins["length"] = builtinLength{}

	context.Builtins["="] = builtinEq{}
	context.Builtins["<"] = builtinCompare{"<", lt}
	context.Builtins[">"] = builtinCompare{">", gt}
	context.Builtins["<="] = builtinCompare{"<=", le}
	context.Builtins[">="] = builtinCompare{">=", ge}

	context.Builtins["call/cc"] = builtinCallCC{}

	context.Builtins["eval"] = builtinEval{}
	context.Builtins["macroexpand"] = builtinMacroExpand{"macroexpand", true}
	context.Builtins["macroexpand-1"] = builtinMacroExpand{"macroexpand-1", false}

	context.Builtins["print"] = builtinPrint{}
	context.Builtins["newline"] = builtinNewline{}

	context.Builtins["inspect"] = builtinInspect{}
}

type builtinCons struct{}

func (_ builtinCons) Run(state *State, args []Value) {
	if len(args) == 2 {
		state.Push(Cons{args[0], args[1]})
		return
	}
	evaluationError("Builtin function cons takes 2 arguments")
}

type builtinExit struct{}

func (_ builtinExit) Run(state *State, args []Value) {
	if len(args) == 0 {
		os.Exit(0)
	}
	if len(args) == 1 {
		if num, ok := args[0].(Num); ok {
			os.Exit(int(num.Data))
		}
	}
	evaluationError("Builtin function exit takes a number argument")
}

type builtinError struct{}

func (_ builtinError) Run(state *State, args []Value) {
	if len(args) == 0 {
		evaluationError("error called")
	}
	if len(args) == 1 {
		if str, ok := args[0].(Str); ok {
			evaluationError(str.Data)
		}
	}
	evaluationError("Builtin function error takes a string argument")
}

type builtinGensym struct {
	id int
}

func (gensym *builtinGensym) Run(state *State, args []Value) {
	if len(args) == 0 {
		gensym.id++
		state.Push(Sym{fmt.Sprintf("#sym.%v", gensym.id)})
		return
	}
	evaluationError("Builtin function gensym takes no arguments")
}

type builtinCar struct{}

func (_ builtinCar) Run(state *State, args []Value) {
	if len(args) == 1 {
		if cons, ok := args[0].(Cons); ok {
			state.Push(cons.Car)
			return
		}
		typeError("Not a cons", args[0])
	}
	evaluationError("Builtin function car takes one argument")
}

type builtinCdr struct{}

func (_ builtinCdr) Run(state *State, args []Value) {
	if len(args) == 1 {
		if cons, ok := args[0].(Cons); ok {
			state.Push(cons.Cdr)
			return
		}
		typeError("Not a cons", args[0])
	}
	evaluationError("Builtin function cdr takes one argument")
}

type builtinApply struct{}

func (_ builtinApply) Run(state *State, args []Value) {
	if len(args) == 2 {
		if slice, ok := Slice(args[1]); ok {
			state.Apply(args[0], slice...)
			return
		}
		typeError("Improper list passed as apply arguments", args[1])
	}
	evaluationError("Builtin function apply takes 2 arguments")
}

type builtinTest struct {
	name string
	cond func(Value) bool
}

func (test builtinTest) Run(state *State, args []Value) {
	if len(args) == 1 {
		state.Push(Bool{test.cond(args[0])})
		return
	}
	evaluationError("Builtin function " + test.name + " takes one argument")
}

func isNum(value Value) bool {
	_, ok := value.(Num)
	return ok
}

func isSym(value Value) bool {
	_, ok := value.(Sym)
	return ok
}

func isStr(value Value) bool {
	_, ok := value.(Str)
	return ok
}

func isCons(value Value) bool {
	_, ok := value.(Cons)
	return ok
}

func isNil(value Value) bool {
	_, ok := value.(Nil)
	return ok
}

func isBool(value Value) bool {
	_, ok := value.(Bool)
	return ok
}

func isProc(value Value) bool {
	_, ok := value.(Proc)
	return ok
}

func isMeta(value Value) bool {
	_, ok := value.(Meta)
	return ok
}

type builtinArithmetic struct {
	name string
	arithmeticImpl
}

func (arith builtinArithmetic) Run(state *State, args []Value) {
	nums := extractNumbers(arith.name, args)
	var ret float64
	switch len(nums) {
	case 0:
		var ok bool
		ret, ok = arith.zero()
		if !ok {
			evaluationError("Operator " + arith.name + " takes at least one argument")
		}
	case 1:
		ret = arith.one(nums[0])
	default:
		ret = nums[0]
		for _, num := range nums[1:] {
			ret = arith.fold(ret, num)
		}
	}
	state.Push(Num{ret})
}

type arithmeticImpl interface {
	zero() (float64, bool)
	one(num float64) float64
	fold(l, r float64) float64
}

type add struct{}

func (_ add) zero() (float64, bool)     { return 0, true }
func (_ add) one(num float64) float64   { return num }
func (_ add) fold(l, r float64) float64 { return l + r }

type sub struct{}

func (_ sub) zero() (float64, bool)     { return 0, false }
func (_ sub) one(num float64) float64   { return -num }
func (_ sub) fold(l, r float64) float64 { return l - r }

type mul struct{}

func (_ mul) zero() (float64, bool)     { return 1, true }
func (_ mul) one(num float64) float64   { return num }
func (_ mul) fold(l, r float64) float64 { return l * r }

type div struct{}

func (_ div) zero() (float64, bool)     { return 0, false }
func (_ div) one(num float64) float64   { return 1 / num }
func (_ div) fold(l, r float64) float64 { return l / r }

type mod struct{}

func (_ mod) zero() (float64, bool)     { return 0, false }
func (_ mod) one(num float64) float64   { return num }
func (_ mod) fold(l, r float64) float64 { return math.Mod(l, r) }

type builtinConcat struct{}

func (_ builtinConcat) Run(state *State, args []Value) {
	strs := extractStrings("concat", args)
	var buf bytes.Buffer
	for _, str := range strs {
		buf.WriteString(str)
	}
	state.Push(Str{buf.String()})
}

type builtinLength struct{}

func (_ builtinLength) Run(state *State, args []Value) {
	if len(args) == 1 {
		if str, ok := args[0].(Str); ok {
			state.Push(Num{float64(len(str.Data))})
			return
		}
	}
	evaluationError("Builtin function length takes a string argument")
}

type builtinEq struct{}

func (eq builtinEq) Run(state *State, args []Value) {
	if len(args) >= 1 {
		for _, arg := range args[1:] {
			if !eq.test(args[0], arg) {
				state.Push(Bool{false})
				return
			}
		}
	}
	state.Push(Bool{true})
}

func (eq builtinEq) test(a, b Value) bool {
	switch a := a.(type) {
	case Num:
		b, ok := b.(Num)
		return ok && a.Data == b.Data

	case Sym:
		b, ok := b.(Sym)
		return ok && a.Data == b.Data

	case Str:
		b, ok := b.(Str)
		return ok && a.Data == b.Data

	case Cons:
		b, ok := b.(Cons)
		return ok && eq.test(a.Car, b.Car) && eq.test(a.Cdr, b.Cdr)

	case Nil:
		_, ok := b.(Nil)
		return ok

	case Bool:
		b, ok := b.(Bool)
		return ok && a.Data == b.Data

	default:
		return false
	}
}

type builtinCompare struct {
	name string
	test func(compareResult int) bool
}

func (compare builtinCompare) Run(state *State, args []Value) {
	if len(args) != 0 {
		switch first := args[0].(type) {
		case Num:
			l := first.Data
			nums := extractNumbers(compare.name, args[1:])
			for _, r := range nums {
				if !compare.compareNumbers(l, r) {
					state.Push(Bool{false})
					return
				}
				l = r
			}

		case Str:
			l := first.Data
			strs := extractStrings(compare.name, args[1:])
			for _, r := range strs {
				if !compare.compareStrings(l, r) {
					state.Push(Bool{false})
					return
				}
				l = r
			}

		default:
			typeError("Operator "+compare.name+" is only defined for strings and numbers", args[0])
		}
	}
	state.Push(Bool{true})
}

func (compare builtinCompare) compareNumbers(l, r float64) bool {
	if l < r {
		return compare.test(-1)
	} else if l > r {
		return compare.test(1)
	} else {
		return compare.test(0)
	}
}

func (compare builtinCompare) compareStrings(l, r string) bool {
	if l < r {
		return compare.test(-1)
	} else if l > r {
		return compare.test(1)
	} else {
		return compare.test(0)
	}
}

func lt(compareResult int) bool { return compareResult == -1 }
func gt(compareResult int) bool { return compareResult == 1 }
func le(compareResult int) bool { return compareResult != 1 }
func ge(compareResult int) bool { return compareResult != -1 }

type builtinCallCC struct{}

func (_ builtinCallCC) Run(state *State, args []Value) {
	if len(args) == 1 {
		cont := state.CaptureCont()
		state.Apply(args[0], cont)
		return
	}
	evaluationError("Builtin function call/cc takes one argument")
}

type builtinEval struct{}

func (_ builtinEval) Run(state *State, args []Value) {
	if len(args) == 1 {
		ret, err := state.Context.Eval(args[0])
		if err == nil {
			state.Push(Cons{Bool{true}, ret})
		} else {
			state.Push(Cons{Bool{false}, Str{err.Error()}})
		}
		return
	}
	evaluationError("Builtin function eval takes one argument")
}

type builtinMacroExpand struct {
	name    string
	recurse bool
}

func (expand builtinMacroExpand) Run(state *State, args []Value) {
	if len(args) == 1 {
		ret, err := state.Context.MacroExpand(expand.recurse, args[0])
		if err == nil {
			state.Push(Cons{Bool{true}, ret})
		} else {
			state.Push(Cons{Bool{false}, Str{err.Error()}})
		}
		return
	}
	evaluationError("Builtin function " + expand.name + " takes one argument")
}

type builtinPrint struct{}

func (_ builtinPrint) Run(state *State, args []Value) {
	for _, arg := range args {
		str, ok := arg.(Str)
		if !ok {
			typeError("Cannot print non-string argument", arg)
		}
		fmt.Print(str.Data)
	}
	state.Push(Nil{})
}

type builtinNewline struct{}

func (_ builtinNewline) Run(state *State, args []Value) {
	if len(args) == 0 {
		fmt.Println()
		state.Push(Nil{})
		return
	}
	evaluationError("Builtin function newline takes no arguments")
}

type builtinInspect struct{}

func (_ builtinInspect) Run(state *State, args []Value) {
	if len(args) == 1 {
		state.Push(Str{args[0].Inspect()})
		return
	}
	evaluationError("Builtin function inspect takes one argument")
}

func extractNumbers(op string, args []Value) []float64 {
	ret := make([]float64, len(args))
	for i, arg := range args {
		switch arg := arg.(type) {
		case Num:
			ret[i] = arg.Data

		default:
			typeError("Operator "+op+" takes number arguments", arg)
		}
	}
	return ret
}

func extractStrings(op string, args []Value) []string {
	ret := make([]string, len(args))
	for i, arg := range args {
		switch arg := arg.(type) {
		case Str:
			ret[i] = arg.Data

		default:
			typeError("Operator "+op+" takes string arguments", arg)
		}
	}
	return ret
}

func evaluationError(msg string) {
	panic(EvaluationError{msg})
}

func typeError(msg string, arg Value) {
	evaluationError(msg + ": " + arg.Inspect())
}
