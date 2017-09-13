package main

import (
	"bufio"
	"bytes"
	"fmt"
	. "github.com/yubrot/golisp"
	"io"
	"math"
	"os"
	"strconv"
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
	context.Builtins["port?"] = builtinTest{"port?", isPort}
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
	context.Builtins["str-ref"] = builtinStrRef{}
	context.Builtins["str-bytesize"] = builtinStrBytesize{}
	context.Builtins["str-concat"] = builtinStrConcat{}
	context.Builtins["substr"] = builtinSubstr{}
	context.Builtins["sym->str"] = builtinSymToStr{}
	context.Builtins["num->str"] = builtinNumToStr{}
	context.Builtins["str->num"] = builtinStrToNum{}

	context.Builtins["vec"] = builtinVec{}
	context.Builtins["vec-make"] = builtinVecMake{}
	context.Builtins["vec-ref"] = builtinVecRef{}
	context.Builtins["vec-length"] = builtinVecLength{}
	context.Builtins["vec-set!"] = builtinVecSet{}
	context.Builtins["vec-copy!"] = builtinVecCopy{}

	context.Builtins["open"] = builtinOpen{}
	context.Builtins["close"] = builtinClose{}

	context.Builtins["stdin"] = builtinPort{"stdin", NewPortIn(os.Stdin)}
	context.Builtins["stdout"] = builtinPort{"stdout", NewPortOut(os.Stdout)}
	context.Builtins["stderr"] = builtinPort{"stderr", NewPortOut(os.Stderr)}

	context.Builtins["read-byte"] = builtinReadByte{}
	context.Builtins["read-str"] = builtinReadStr{}
	context.Builtins["read-line"] = builtinReadLine{}

	context.Builtins["write-byte"] = builtinWriteByte{}
	context.Builtins["write-str"] = builtinWriteStr{}
	context.Builtins["write-line"] = builtinWriteLine{}
	context.Builtins["flush"] = builtinFlush{}

	context.Builtins["args"] = builtinArgs{args}

	context.Builtins["eval"] = builtinEval{}
	context.Builtins["macroexpand"] = builtinMacroExpand{"macroexpand", true}
	context.Builtins["macroexpand-1"] = builtinMacroExpand{"macroexpand-1", false}
}

type builtinCons struct{}

func (_ builtinCons) Run(state *State, args []Value) {
	a, b := takeTwo("cons", args)
	state.Push(Cons{a, b})
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
	evaluationError("exit takes exitcode")
}

type builtinError struct{}

func (_ builtinError) Run(state *State, args []Value) {
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
	state.Push(Sym{fmt.Sprintf("#sym.%v", gensym.id)})
}

type builtinCar struct{}

func (_ builtinCar) Run(state *State, args []Value) {
	arg := takeOne("car", args)
	cons := takeCons("cons", arg)
	state.Push(cons.Car)
}

type builtinCdr struct{}

func (_ builtinCdr) Run(state *State, args []Value) {
	arg := takeOne("cdr", args)
	cons := takeCons("cons", arg)
	state.Push(cons.Cdr)
}

type builtinApply struct{}

func (_ builtinApply) Run(state *State, args []Value) {
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
	state.Push(Bool{test.cond(arg)})
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

func isPort(value Value) bool {
	_, ok := value.(Port)
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
	state.Push(Num{result})
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
			var nums []float64
			for _, arg := range args[1:] {
				nums = append(nums, takeNum("number", arg))
			}
			for _, r := range nums {
				if !compare.compareNumbers(l, r) {
					state.Push(Bool{false})
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
					state.Push(Bool{false})
					return
				}
				l = r
			}

		default:
			evaluationError(compare.name + " is only defined for strings or numbers")
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
	f := takeOne("call/cc", args)
	cont := state.CaptureCont()
	state.Apply(f, cont)
}

type builtinNever struct{}

func (_ builtinNever) Run(state *State, args []Value) {
	if len(args) > 0 {
		state.ApplyNever(args[0], args[1:]...)
		return
	}
	evaluationError("never takes at least one argument")
}

type builtinStr struct{}

func (_ builtinStr) Run(state *State, args []Value) {
	var bytes []byte
	for _, arg := range args {
		num := int(takeNum("byte", arg))
		if num < 0 || 255 < num {
			evaluationError("Each byte of string must be inside the range 0-255")
		}
		bytes = append(bytes, byte(num))
	}
	state.Push(Str{string(bytes[:])})
}

type builtinStrRef struct{}

func (_ builtinStrRef) Run(state *State, args []Value) {
	str, index := takeTwo("str-ref", args)
	s := takeStr("string", str)
	i := int(takeNum("index", index))
	if i < 0 || len(s) <= i {
		state.Push(Nil{})
	} else {
		state.Push(Num{float64(s[i])})
	}
}

type builtinStrBytesize struct{}

func (_ builtinStrBytesize) Run(state *State, args []Value) {
	arg := takeOne("str-bytesize", args)
	str := takeStr("string", arg)
	state.Push(Num{float64(len(str))})
}

type builtinStrConcat struct{}

func (_ builtinStrConcat) Run(state *State, args []Value) {
	var buf bytes.Buffer
	for _, arg := range args {
		buf.WriteString(takeStr("string", arg))
	}
	state.Push(Str{buf.String()})
}

type builtinSubstr struct{}

func (_ builtinSubstr) Run(state *State, args []Value) {
	s, i, l := takeThree("substr", args)
	str := takeStr("string", s)
	index := int(takeNum("index", i))
	size := int(takeNum("size", l))
	if index < 0 || len(str) < index+size {
		evaluationError("Index out of range")
	}
	state.Push(Str{str[index : index+size]})
}

type builtinSymToStr struct{}

func (_ builtinSymToStr) Run(state *State, args []Value) {
	arg := takeOne("sym->str", args)
	s := takeSym("symbol", arg)
	state.Push(Str{s})
}

type builtinNumToStr struct{}

func (_ builtinNumToStr) Run(state *State, args []Value) {
	arg := takeOne("sym->str", args)
	n := takeNum("number", arg)
	state.Push(Str{Num{n}.Inspect()})
}

type builtinStrToNum struct{}

func (_ builtinStrToNum) Run(state *State, args []Value) {
	arg := takeOne("str->num", args)
	s := takeStr("string", arg)
	num, err := strconv.ParseFloat(s, 64)
	if err != nil {
		state.Push(Nil{})
	} else {
		state.Push(Num{num})
	}
}

type builtinVec struct{}

func (_ builtinVec) Run(state *State, args []Value) {
	state.Push(Vec{args})
}

type builtinVecMake struct{}

func (_ builtinVecMake) Run(state *State, args []Value) {
	l, init := takeTwo("vec-make", args)
	length := int(takeNum("length", l))
	slice := make([]Value, length)
	for i := range slice {
		slice[i] = init
	}
	state.Push(Vec{slice})
}

type builtinVecRef struct{}

func (_ builtinVecRef) Run(state *State, args []Value) {
	v, n := takeTwo("vec-ref", args)
	vec := takeVec("vector", v)
	index := int(takeNum("index", n))
	if index < 0 || len(vec.Payload) <= index {
		state.Push(Nil{})
	} else {
		state.Push(vec.Payload[index])
	}
}

type builtinVecLength struct{}

func (_ builtinVecLength) Run(state *State, args []Value) {
	v := takeOne("vec-length", args)
	vec := takeVec("vector", v)
	state.Push(Num{float64(len(vec.Payload))})
}

type builtinVecSet struct{}

func (_ builtinVecSet) Run(state *State, args []Value) {
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

func (_ builtinVecCopy) Run(state *State, args []Value) {
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

type builtinOpen struct{}

func (_ builtinOpen) Run(state *State, args []Value) {
	p, m := takeTwo("open", args)
	filepath := takeStr("filepath", p)
	mode := takeStr("mode", m)

	switch mode {
	case "r":
		file, err := os.Open(filepath)
		if err == nil {
			state.Push(Cons{Bool{true}, NewPortIn(file)})
		} else {
			state.Push(Cons{Bool{false}, Str{err.Error()}})
		}
	case "w":
		file, err := os.Create(filepath)
		if err == nil {
			state.Push(Cons{Bool{true}, NewPortOut(file)})
		} else {
			state.Push(Cons{Bool{false}, Str{err.Error()}})
		}
	default:
		evaluationError("Unsupported mode for open: " + mode)
	}
}

type builtinClose struct{}

func (_ builtinClose) Run(state *State, args []Value) {
	p := takeOne("close", args)
	port := takePort("port", p)
	err := port.Close()

	if err == nil {
		state.Push(Cons{Bool{true}, Nil{}})
	} else {
		state.Push(Cons{Bool{false}, Str{err.Error()}})
	}
}

type builtinPort struct {
	name string
	port Port
}

func (bp builtinPort) Run(state *State, args []Value) {
	takeNone(bp.name, args)
	state.Push(bp.port)
}

func eofOrError(err error) Value {
	if err == io.EOF {
		return Cons{Bool{true}, Sym{"eof"}}
	} else {
		return Cons{Bool{false}, Str{err.Error()}}
	}
}

type builtinReadByte struct{}

func (_ builtinReadByte) Run(state *State, args []Value) {
	p := takeOne("read-byte", args)
	r := takePortIn(takePort("port", p))

	b, err := r.ReadByte()
	if err == nil {
		state.Push(Cons{Bool{true}, Num{float64(b)}})
	} else {
		state.Push(eofOrError(err))
	}
}

type builtinReadStr struct{}

func (_ builtinReadStr) Run(state *State, args []Value) {
	s, p := takeTwo("read-str", args)
	size := int(takeNum("size", s))
	r := takePortIn(takePort("port", p))

	bytes := make([]byte, size)
	n, err := r.Read(bytes)
	if err == nil {
		state.Push(Cons{Bool{true}, Str{string(bytes[:n])}})
	} else {
		state.Push(eofOrError(err))
	}
}

type builtinReadLine struct{}

func (_ builtinReadLine) Run(state *State, args []Value) {
	p := takeOne("read-line", args)
	r := takePortIn(takePort("port", p))

	line, err := r.ReadBytes('\n')
	if err == nil {
		state.Push(Cons{Bool{true}, Str{string(line[:len(line)-1])}})
	} else {
		state.Push(eofOrError(err))
	}
}

type builtinWriteByte struct{}

func (_ builtinWriteByte) Run(state *State, args []Value) {
	b, p := takeTwo("write-byte", args)
	w := takePortOut(takePort("port", p))
	err := w.WriteByte(byte(takeNum("byte", b)))
	if err == nil {
		state.Push(Cons{Bool{true}, Num{1}})
	} else {
		state.Push(Cons{Bool{false}, Str{err.Error()}})
	}
}

type builtinWriteStr struct{}

func (_ builtinWriteStr) Run(state *State, args []Value) {
	s, p := takeTwo("write-str", args)
	w := takePortOut(takePort("port", p))
	str := takeStr("string", s)
	n, err := w.WriteString(str)
	if err == nil || n > 0 {
		state.Push(Cons{Bool{true}, Num{float64(n)}})
	} else {
		state.Push(Cons{Bool{false}, Str{err.Error()}})
	}
}

type builtinWriteLine struct{}

func (_ builtinWriteLine) Run(state *State, args []Value) {
	s, p := takeTwo("write-line", args)
	w := takePortOut(takePort("port", p))
	str := takeStr("string", s)
	n, err := w.WriteString(str)
	if n == len(str) {
		err = w.WriteByte('\n')
		if err == nil {
			n += 1
			err = w.Flush()
		}
	}
	if err == nil || n > 0 {
		state.Push(Cons{Bool{true}, Num{float64(n)}})
	} else {
		state.Push(Cons{Bool{false}, Str{err.Error()}})
	}
}

type builtinFlush struct{}

func (_ builtinFlush) Run(state *State, args []Value) {
	p := takeOne("flush", args)
	w := takePortOut(takePort("port", p))
	err := w.Flush()
	if err == nil {
		state.Push(Cons{Bool{true}, Nil{}})
	} else {
		state.Push(Cons{Bool{false}, Str{err.Error()}})
	}
}

type builtinArgs struct {
	args []string
}

func (b builtinArgs) Run(state *State, args []Value) {
	takeNone("args", args)
	var vs []Value
	for _, arg := range b.args {
		vs = append(vs, Str{arg})
	}
	state.Push(List(vs...))
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

func takePort(name string, v Value) Port {
	ret, ok := v.(Port)
	checkExpected(name, ok, v)
	return ret
}

func takePortIn(p Port) *bufio.Reader {
	reader := p.Reader()
	if reader == nil {
		evaluationError("port is not available for reading")
	}
	return reader
}

func takePortOut(p Port) *bufio.Writer {
	writer := p.Writer()
	if writer == nil {
		evaluationError("port is not available for writing")
	}
	return writer
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
	panic(EvaluationError{msg})
}
