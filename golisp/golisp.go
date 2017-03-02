package main

//go:generate go-bindata lispboot/boot.lisp

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/yubrot/golisp"
	"os"
)

func main() {
	context := golisp.NewContext()

	if len(os.Args) == 3 && os.Args[1] == "-test" {
		registerBuiltins(context)
		RunTest(context, os.Args[2])

	} else if len(os.Args) == 2 {
		boot(context)
		execFile(context, os.Args[1])
	} else {
		boot(context)
		repl(context)
	}
}

func boot(context *golisp.Context) {
	registerBuiltins(context)

	data, err := Asset("lispboot/boot.lisp")
	if err != nil {
		panic(err)
	}

	buf := bufio.NewReader(bytes.NewReader(data))
	err = exec(context, buf)
	if err != nil {
		panic(err)
	}
}

func execFile(context *golisp.Context, file string) {
	fp, err := os.Open(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer fp.Close()

	buf := bufio.NewReader(fp)
	err = exec(context, buf)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func exec(context *golisp.Context, buf *bufio.Reader) error {
	return golisp.RunParser(buf, func(expr golisp.Value, err error) error {
		if err == nil {
			_, err = context.Eval(expr)
		}
		return err
	})
}

func repl(context *golisp.Context) {
	fmt.Fprintln(os.Stderr, "[golisp REPL]")
	fmt.Fprint(os.Stderr, "> ")

	stdin := bufio.NewReader(os.Stdin)
	golisp.RunParser(stdin, func(expr golisp.Value, err error) (never error) {
		defer fmt.Fprint(os.Stderr, "> ")

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			stdin.ReadLine()
			return
		}

		result, err := context.Eval(expr)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		fmt.Println(result.Inspect())
		return
	})
}
