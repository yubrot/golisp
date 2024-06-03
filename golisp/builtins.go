package main

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"os"
	"strconv"

	. "github.com/yubrot/golisp"
)

func registerBuiltins(context *Context, args []string) {
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
	context.Builtins["vec?"] = builtinTest{"vec?", isVec}

	context.Builtins["+"] = builtinArithmetic{"+", add{}}
	context.Builtins["-"] = builtinArithmetic{"-", sub{}}
	context.Builtins["*"] = builtinArithmetic{"*", mul{}}
	context.Builtins["/"] = builtinArithmetic{"/", div{}}
	context.Builtins["%"] = builtinArithmetic{"%", mod{}}

	context.Builtins["="] = builtinEq{}
	context.Builtins["<"] = builtinCompare{"<", lt}
	context.Builtins[">"] = builtinCompare{">", gt}
	context.Builtins["<="] = builtinCompare{"<=", le}
	context.Builtins[">="] = builtinCompare{">=", ge}

	context.Builtins["call/cc"] = builtinCallCC{}
	context.Builtins["never"] = builtinNever{}

	context.Builtins["str"] = builtinStr{}
	context.Builtins["str-char-at"] = builtinStrCharAt{}
	context.Builtins["str-length"] = builtinStrLength{}
	context.Builtins["str-concat"] = builtinStrConcat{}
	context.Builtins["substr"] = builtinSubstr{}
	context.Builtins["sym->str"] = builtinSymToStr{}
	context.Builtins["num->str"] = builtinNumToStr{}
	context.Builtins["str->num"] = builtinStrToNum{}

	context.Builtins["vec"] = builtinVec{}
	context.Builtins["vec-make"] = builtinVecMake{}
	context.Builtins["vec-length"] = builtinVecLength{}
	context.Builtins["vec-get"] = builtinVecGet{}
	context.Builtins["vec-set!"] = builtinVecSet{}
	context.Builtins["vec-copy!"] = builtinVecCopy{}

	context.Builtins["read-file-text"] = builtinReadFileText{}
	context.Builtins["write-file-text"] = builtinWriteFileText{}
	context.Builtins["read-console-line"] = builtinReadConsoleLine{}
	context.Builtins["write-console"] = builtinWriteConsole{}

	context.Builtins["args"] = builtinArgs{args}

	context.Builtins["eval"] = builtinEval{}
	context.Builtins["macroexpand"] = builtinMacroExpand{"macroexpand", true}
	context.Builtins["macroexpand-1"] = builtinMacroExpand{"macroexpand-1", false}
}

type builtinCons struct{}

func (builtinCons) Run(state *State, args []Value) {
	a, b := takeTwo("cons", args)
	state.Push(Cons{Car: a, Cdr: b})
}

type builtinExit struct{}

func (builtinExit) Run(state *State, args []Value) {
	if len(args) == 0 {
		os.Exit(0)
	}
	if len(args) == 1 {
		if num, ok := args[0].(Num); ok {
			os.Exit(int(num.Data))
		}
	}
	evaluationError("exit takes exitcode")
}

type builtinError struct{}

func (builtinError) Run(state *State, args []Value) {
	if len(args) == 0 {
		evaluationError("error called")
	}
	if len(args) == 1 {
		evaluationError(takeStr("error message", args[0]))
	}
	evaluationError("Builtin function error takes a string argument")
}

type builtinGensym struct {
	id int
}

func (gensym *builtinGensym) Run(state *State, args []Value) {
	takeNone("gensym", args)
	gensym.id++
	state.Push(Sym{Data: fmt.Sprintf("#sym.%v", gensym.id)})
}

type builtinCar struct{}

func (builtinCar) Run(state *State, args []Value) {
	arg := takeOne("car", args)
	cons := takeCons("cons", arg)
	state.Push(cons.Car)
}

type builtinCdr struct{}

func (builtinCdr) Run(state *State, args []Value) {
	arg := takeOne("cdr", args)
	cons := takeCons("cons", arg)
	state.Push(cons.Cdr)
}

type builtinApply struct{}

func (builtinApply) Run(state *State, args []Value) {
	f, fargs := takeTwo("apply", args)
	fslice := takeList("argument list", fargs)
	state.Apply(f, fslice...)
}

type builtinTest struct {
	name string
	cond func(Value) bool
}

func (test builtinTest) Run(state *State, args []Value) {
	arg := takeOne(test.name, args)
	state.Push(Bool{Data: test.cond(arg)})
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

func isVec(value Value) bool {
	_, ok := value.(Vec)
	return ok
}

type builtinArithmetic struct {
	name string
	arithmeticImpl
}

func (arith builtinArithmetic) Run(state *State, args []Value) {
	var nums []float64
	for _, arg := range args {
		nums = append(nums, takeNum("number", arg))
	}
	var result float64
	switch len(nums) {
	case 0:
		var ok bool
		result, ok = arith.zero()
		if !ok {
			evaluationError(arith.name + " takes at least one argument")
		}
	case 1:
		result = arith.one(nums[0])
	default:
		result = nums[0]
		for _, num := range nums[1:] {
			result = arith.fold(result, num)
		}
	}
	state.Push(Num{Data: result})
}

type arithmeticImpl interface {
	zero() (float64, bool)
	one(num float64) float64
	fold(l, r float64) float64
}

type add struct{}

func (add) zero() (float64, bool)     { return 0, true }
func (add) one(num float64) float64   { return num }
func (add) fold(l, r float64) float64 { return l + r }

type sub struct{}

func (sub) zero() (float64, bool)     { return 0, false }
func (sub) one(num float64) float64   { return -num }
func (sub) fold(l, r float64) float64 { return l - r }

type mul struct{}

func (mul) zero() (float64, bool)     { return 1, true }
func (mul) one(num float64) float64   { return num }
func (mul) fold(l, r float64) float64 { return l * r }

type div struct{}

func (div) zero() (float64, bool)     { return 0, false }
func (div) one(num float64) float64   { return 1 / num }
func (div) fold(l, r float64) float64 { return l / r }

type mod struct{}

func (mod) zero() (float64, bool)     { return 0, false }
func (mod) one(num float64) float64   { return num }
func (mod) fold(l, r float64) float64 { return math.Mod(l, r) }

type builtinEq struct{}

func (eq builtinEq) Run(state *State, args []Value) {
	if len(args) >= 1 {
		for _, arg := range args[1:] {
			if !eq.test(args[0], arg) {
				state.Push(Bool{Data: false})
				return
			}
		}
	}
	state.Push(Bool{Data: true})
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
			var nums []float64
			for _, arg := range args[1:] {
				nums = append(nums, takeNum("number", arg))
			}
			for _, r := range nums {
				if !compare.compareNumbers(l, r) {
					state.Push(Bool{Data: false})
					return
				}
				l = r
			}

		case Str:
			l := first.Data
			var strs []string
			for _, arg := range args[1:] {
				strs = append(strs, takeStr("string", arg))
			}
			for _, r := range strs {
				if !compare.compareStrings(l, r) {
					state.Push(Bool{Data: false})
					return
				}
				l = r
			}

		default:
			evaluationError(compare.name + " is only defined for strings or numbers")
		}
	}
	state.Push(Bool{Data: true})
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

func (builtinCallCC) Run(state *State, args []Value) {
	f := takeOne("call/cc", args)
	cont := state.CaptureCont()
	state.Apply(f, cont)
}

type builtinNever struct{}

func (builtinNever) Run(state *State, args []Value) {
	if len(args) > 0 {
		state.ApplyNever(args[0], args[1:]...)
		return
	}
	evaluationError("never takes at least one argument")
}

type builtinStr struct{}

func (builtinStr) Run(state *State, args []Value) {
	var bytes []byte
	for _, arg := range args {
		num := int(takeNum("byte", arg))
		if num < 0 || 255 < num {
			evaluationError("Each byte of string must be inside the range 0-255")
		}
		bytes = append(bytes, byte(num))
	}
	state.Push(Str{Data: string(bytes[:])})
}

type builtinStrCharAt struct{}

func (builtinStrCharAt) Run(state *State, args []Value) {
	str, index := takeTwo("str-char-at", args)
	s := takeStr("string", str)
	i := int(takeNum("index", index))
	if i < 0 || len(s) <= i {
		state.Push(Nil{})
	} else {
		state.Push(Num{Data: float64(s[i])})
	}
}

type builtinStrLength struct{}

func (builtinStrLength) Run(state *State, args []Value) {
	arg := takeOne("str-length", args)
	str := takeStr("string", arg)
	state.Push(Num{Data: float64(len(str))})
}

type builtinStrConcat struct{}

func (builtinStrConcat) Run(state *State, args []Value) {
	var buf bytes.Buffer
	for _, arg := range args {
		buf.WriteString(takeStr("string", arg))
	}
	state.Push(Str{Data: buf.String()})
}

type builtinSubstr struct{}

func (builtinSubstr) Run(state *State, args []Value) {
	s, i, l := takeThree("substr", args)
	str := takeStr("string", s)
	index := int(takeNum("index", i))
	size := int(takeNum("size", l))
	if index < 0 || len(str) < index+size {
		evaluationError("Index out of range")
	}
	state.Push(Str{Data: str[index : index+size]})
}

type builtinSymToStr struct{}

func (builtinSymToStr) Run(state *State, args []Value) {
	arg := takeOne("sym->str", args)
	s := takeSym("symbol", arg)
	state.Push(Str{Data: s})
}

type builtinNumToStr struct{}

func (builtinNumToStr) Run(state *State, args []Value) {
	arg := takeOne("num->str", args)
	n := takeNum("number", arg)
	state.Push(Str{Data: Num{Data: n}.Inspect()})
}

type builtinStrToNum struct{}

func (builtinStrToNum) Run(state *State, args []Value) {
	arg := takeOne("str->num", args)
	s := takeStr("string", arg)
	num, err := strconv.ParseFloat(s, 64)
	if err != nil {
		state.Push(Nil{})
	} else {
		state.Push(Num{Data: num})
	}
}

type builtinVec struct{}

func (builtinVec) Run(state *State, args []Value) {
	state.Push(Vec{Payload: args})
}

type builtinVecMake struct{}

func (builtinVecMake) Run(state *State, args []Value) {
	l, init := takeTwo("vec-make", args)
	length := int(takeNum("length", l))
	slice := make([]Value, length)
	for i := range slice {
		slice[i] = init
	}
	state.Push(Vec{Payload: slice})
}

type builtinVecLength struct{}

func (builtinVecLength) Run(state *State, args []Value) {
	v := takeOne("vec-length", args)
	vec := takeVec("vector", v)
	state.Push(Num{Data: float64(len(vec.Payload))})
}

type builtinVecGet struct{}

func (builtinVecGet) Run(state *State, args []Value) {
	v, n := takeTwo("vec-get", args)
	vec := takeVec("vector", v)
	index := int(takeNum("index", n))
	if index < 0 || len(vec.Payload) <= index {
		state.Push(Nil{})
	} else {
		state.Push(vec.Payload[index])
	}
}

type builtinVecSet struct{}

func (builtinVecSet) Run(state *State, args []Value) {
	v, n, item := takeThree("vec-set!", args)
	vec := takeVec("vector", v)
	index := int(takeNum("index", n))
	if index < 0 || len(vec.Payload) <= index {
		evaluationError("Index out of range")
	} else {
		vec.Payload[index] = item
		state.Push(Nil{})
	}
}

type builtinVecCopy struct{}

func (builtinVecCopy) Run(state *State, args []Value) {
	dest, destS, src, srcS, l := takeFive("vec-copy!", args)
	destVec := takeVec("destination vector", dest)
	destStart := int(takeNum("destination index", destS))
	srcVec := takeVec("source vector", src)
	srcStart := int(takeNum("source index", srcS))
	length := int(takeNum("length", l))

	if 0 <= srcStart && srcStart+length <= len(srcVec.Payload) && 0 <= destStart && destStart+length <= len(destVec.Payload) {
		copy(destVec.Payload[destStart:], srcVec.Payload[srcStart:srcStart+length])
		state.Push(Nil{})
	} else {
		evaluationError("Index out of range")
	}
}

type builtinReadFileText struct{}

func (builtinReadFileText) Run(state *State, args []Value) {
	p := takeOne("read-file-text", args)
	filepath := takeStr("filepath", p)
	contents, err := os.ReadFile(filepath)
	if err == nil {
		state.Push(Cons{Car: Bool{Data: true}, Cdr: Str{Data: string(contents)}})
	} else {
		state.Push(Cons{Car: Bool{Data: false}, Cdr: Str{Data: err.Error()}})
	}
}

type builtinWriteFileText struct{}

func (builtinWriteFileText) Run(state *State, args []Value) {
	p, c := takeTwo("write-file-text", args)
	filepath := takeStr("filepath", p)
	contents := takeStr("contents", c)
	err := os.WriteFile(filepath, []byte(contents), 0666)
	if err == nil {
		state.Push(Cons{Car: Bool{Data: true}, Cdr: Nil{}})
	} else {
		state.Push(Cons{Car: Bool{Data: false}, Cdr: Str{Data: err.Error()}})
	}
}

type builtinReadConsoleLine struct{}

func (builtinReadConsoleLine) Run(state *State, args []Value) {
	takeNone("read-console-line", args)

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		state.Push(Cons{Car: Bool{Data: true}, Cdr: Str{Data: string(scanner.Text())}})
	} else if scanner.Err() == nil {
		state.Push(Cons{Car: Bool{Data: true}, Cdr: Nil{}})
	} else {
		state.Push(Cons{Car: Bool{Data: false}, Cdr: Str{Data: scanner.Err().Error()}})
	}
}

type builtinWriteConsole struct{}

func (builtinWriteConsole) Run(state *State, args []Value) {
	s := takeOne("write-console", args)
	text := takeStr("text", s)
	_, err := fmt.Print(text)
	if err == nil {
		state.Push(Cons{Car: Bool{Data: true}, Cdr: Nil{}})
	} else {
		state.Push(Cons{Car: Bool{Data: false}, Cdr: Str{Data: err.Error()}})
	}
}

type builtinArgs struct {
	args []string
}

func (b builtinArgs) Run(state *State, args []Value) {
	takeNone("args", args)
	var vs []Value
	for _, arg := range b.args {
		vs = append(vs, Str{Data: arg})
	}
	state.Push(List(vs...))
}

type builtinEval struct{}

func (builtinEval) Run(state *State, args []Value) {
	if len(args) == 1 {
		ret, err := state.Context.Eval(args[0])
		if err == nil {
			state.Push(Cons{Car: Bool{Data: true}, Cdr: ret})
		} else {
			state.Push(Cons{Car: Bool{Data: false}, Cdr: Str{Data: err.Error()}})
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
			state.Push(Cons{Car: Bool{Data: true}, Cdr: ret})
		} else {
			state.Push(Cons{Car: Bool{Data: false}, Cdr: Str{Data: err.Error()}})
		}
		return
	}
	evaluationError("Builtin function " + expand.name + " takes one argument")
}

func takeNone(name string, args []Value) {
	if len(args) != 0 {
		evaluationError(name + " takes no arguments")
	}
}

func takeOne(name string, args []Value) Value {
	if len(args) != 1 {
		evaluationError(name + " takes one argument")
	}
	return args[0]
}

func takeTwo(name string, args []Value) (Value, Value) {
	if len(args) != 2 {
		evaluationError(name + " takes two arguments")
	}
	return args[0], args[1]
}

func takeThree(name string, args []Value) (Value, Value, Value) {
	if len(args) != 3 {
		evaluationError(name + " takes three arguments")
	}
	return args[0], args[1], args[2]
}

func takeFive(name string, args []Value) (Value, Value, Value, Value, Value) {
	if len(args) != 5 {
		evaluationError(name + " takes five arguments")
	}
	return args[0], args[1], args[2], args[3], args[4]
}

func takeNum(name string, v Value) float64 {
	ret, ok := v.(Num)
	checkExpected(name, ok, v)
	return ret.Data
}

func takeSym(name string, v Value) string {
	ret, ok := v.(Sym)
	checkExpected(name, ok, v)
	return ret.Data
}

func takeStr(name string, v Value) string {
	ret, ok := v.(Str)
	checkExpected(name, ok, v)
	return ret.Data
}

func takeCons(name string, v Value) Cons {
	ret, ok := v.(Cons)
	checkExpected(name, ok, v)
	return ret
}

func takeList(name string, v Value) []Value {
	ret, ok := Slice(v)
	checkExpected(name, ok, v)
	return ret
}

func takeVec(name string, v Value) Vec {
	ret, ok := v.(Vec)
	checkExpected(name, ok, v)
	return ret
}

func checkExpected(name string, ok bool, v Value) {
	if !ok {
		evaluationError("Expected " + name + " but got " + v.Inspect())
	}
}

func evaluationError(msg string) {
	panic(EvaluationError{Msg: msg})
}
