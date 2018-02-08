package main

//go:generate go-assets-builder lispboot/boot.lisp -o assets.go

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/yubrot/golisp"
	"os"
)

func main() {
	ctx := golisp.NewContext()

	if len(os.Args) == 0 || len(os.Args) == 1 {
		initContext(ctx, true, []string{})
		repl(ctx)
	} else if os.Args[1] == "-test" {
		initContext(ctx, false, []string{})
		for _, test := range os.Args[2:] {
			RunTest(ctx, test)
		}
	} else {
		var files, args []string
		argsStarted := false
		for _, s := range os.Args[1:] {
			if argsStarted {
				args = append(args, s)
			} else if s == "--" {
				argsStarted = true
			} else {
				files = append(files, s)
			}
		}
		initContext(ctx, true, args)
		for _, file := range files {
			execFile(ctx, file)
		}
	}
}

func initContext(ctx *golisp.Context, boot bool, args []string) {
	registerBuiltins(ctx, args)
	if boot {
		r, err := Assets.Open("/lispboot/boot.lisp")
		if err != nil {
			panic(err)
		}

		buf := bufio.NewReader(r)
		err = exec(ctx, buf)
		if err != nil {
			panic(errors.New("initContext: " + err.Error()))
		}
	}
}

func execFile(context *golisp.Context, file string) {
	fp, err := os.Open(file)
	if err != nil {
		panic(errors.New(file + ": " + err.Error()))
	}
	defer fp.Close()

	buf := bufio.NewReader(fp)
	err = exec(context, buf)
	if err != nil {
		panic(errors.New(file + ": " + err.Error()))
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
