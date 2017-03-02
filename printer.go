package golisp

import (
	"bytes"
	"strconv"
)

func PrintCode(code Code) string {
	printer := &codePrinter{id: 0}
	printer.putBlock("entry", code)
	return printer.print()
}

type codePrinter struct {
	id int
	blocks []*bytes.Buffer
}

func (printer *codePrinter) print() string {
	ret := new(bytes.Buffer)
	for _, block := range printer.blocks {
		ret.Write(block.Bytes())
	}
	return ret.String()
}

func (printer *codePrinter) putBlock(header string, code Code) string {
	id := "[" + strconv.Itoa(printer.id) + " " + header + "]"
	block := &bytes.Buffer{}
	printer.id += 1
	printer.blocks = append(printer.blocks, block)

	block.WriteString(id + "\n")
	for _, i := range code {
		block.WriteString("  ")
		block.WriteString(printer.putInst(i))
		block.WriteString("\n")
	}
	return id
}

func (printer *codePrinter) putInst(i inst) string {
	switch i := i.(type) {
	case ldc:
		return "ldc " + i.value.Inspect()

	case ldv:
		return "ldv " + i.name

	case ldf:
		block := printer.putBlock("fun " + i.pattern.String(), i.code)
		return "ldf " + block

	case ldm:
		block := printer.putBlock("macro " + i.pattern.String(), i.code)
		return "ldm " + block

	case ldb:
		return "ldb " + i.name

	case sel:
		a := printer.putBlock("then", i.a)
		b := printer.putBlock("else", i.b)
		return "sel " + a + " " + b

	case app:
		return "app " + strconv.Itoa(i.argc)

	case leave:
		return "leave"

	case pop:
		return "pop"

	case def:
		return "def " + i.name

	case set:
		return "set " + i.name

	default:
		panic(InternalError{"Unknown inst"})
	}
}
