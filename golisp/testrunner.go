package main

import (
	"os"
	"bufio"
	"bytes"
	"strings"
	"errors"
	"fmt"
	"strconv"
	"github.com/yubrot/golisp"
)

type testcase struct {
	header string
	commandImpl
}

func RunTest(context *golisp.Context, file string) {
	testcases := parseTestcases(file)
	for _, testcase := range testcases {
		err := testcase.run(context)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Test failed at " + testcase.header + ": " + err.Error())
		}
	}
}

func readLines(scanner *bufio.Scanner, length string) string {
	l, err := strconv.Atoi(length)
	if err != nil { panic(err) }
	ret := new(bytes.Buffer)
	for i := 0; i < l; i++ {
		if !scanner.Scan() { panic("scan") }
		if i != 0 { ret.WriteString("\n") }
		ret.WriteString(scanner.Text())
	}
	return ret.String()
}

func parseTestcases(file string) []testcase {
	fp, err := os.Open(file)
	if err != nil { panic(err) }
	defer fp.Close()

	testcases := make([]testcase, 0, 100)
	scanner := bufio.NewScanner(fp)

	for scanner.Scan() {
		header := scanner.Text()
		if !scanner.Scan() { panic("scan") }
		command := strings.Split(scanner.Text(), " ")
		var impl commandImpl

		switch command[0] {
		case "PARSE_SUCCESS":
			i := readLines(scanner, command[1])
			o := readLines(scanner, command[2])
			impl = parseSuccess{i, o}
		case "PARSE_FAILURE":
			i := readLines(scanner, command[1])
			impl = parseFailure{i}
		case "COMPILE_SUCCESS":
			i := readLines(scanner, command[1])
			o := readLines(scanner, command[2])
			impl = compileSuccess{i, o}
		case "COMPILE_FAILURE":
			i := readLines(scanner, command[1])
			impl = compileFailure{i}
		case "EVAL_SUCCESS":
			i := readLines(scanner, command[1])
			o := readLines(scanner, command[2])
			impl = evalSuccess{i, o}
		case "EVAL_FAILURE":
			i := readLines(scanner, command[1])
			impl = evalFailure{i}
		case "EVAL_ALL":
			i := readLines(scanner, command[1])
			impl = evalAll{i}
		default:
			panic("Unknown test command: " + command[0])
		}

		testcases = append(testcases, testcase{header, impl})
	}

	return testcases
}

type commandImpl interface {
	run(context *golisp.Context) error
}

func parseLine(input string) (result golisp.Value, err error) {
	buf := bufio.NewReader(strings.NewReader(input))
	err = errors.New("empty")
	golisp.RunParser(buf, func(expr golisp.Value, e error) error {
		result = expr
		err = e
		return errors.New("dummy")
	})
	return
}

type parseSuccess struct {
	input, result string
}

func (cmd parseSuccess) run(context *golisp.Context) (err error) {
	result, err := parseLine(cmd.input)
	if err == nil && result.Inspect() != cmd.result {
		err = errors.New(result.Inspect())
	}
	return
}

type parseFailure struct {
	input string
}

func (cmd parseFailure) run(context *golisp.Context) (err error) {
	result, err := parseLine(cmd.input)
	if err == nil {
		err = errors.New(result.Inspect())
	} else {
		err = nil
	}
	return
}

type compileSuccess struct {
	input, result string
}

func (cmd compileSuccess) run(context *golisp.Context) (err error) {
	expr, err := parseLine(cmd.input)
	if err == nil {
		var code golisp.Code
		code, err = context.Compile(expr)
		if err == nil && golisp.PrintCode(code) != cmd.result + "\n" {
			err = errors.New(golisp.PrintCode(code))
		}
	}
	return
}

type compileFailure struct {
	input string
}

func (cmd compileFailure) run(context *golisp.Context) (err error) {
	expr, err := parseLine(cmd.input)
	if err == nil {
		var code golisp.Code
		code, err = context.Compile(expr)
		if err == nil {
			err = errors.New(golisp.PrintCode(code))
		} else {
			err = nil
		}
	}
	return
}

type evalSuccess struct {
	input, result string
}

func (cmd evalSuccess) run(context *golisp.Context) (err error) {
	expr, err := parseLine(cmd.input)
	if err == nil {
		var result golisp.Value
		result, err = context.Eval(expr)
		if err == nil && result.Inspect() != cmd.result {
			err = errors.New(result.Inspect())
		}
	}
	return
}

type evalFailure struct {
	input string
}

func (cmd evalFailure) run(context *golisp.Context) (err error) {
	expr, err := parseLine(cmd.input)
	if err == nil {
		var result golisp.Value
		result, err = context.Eval(expr)
		if err == nil {
			err = errors.New(result.Inspect())
		} else {
			err = nil
		}
	}
	return
}

type evalAll struct {
	input string
}

func (cmd evalAll) run(context *golisp.Context) error {
	buf := bufio.NewReader(strings.NewReader(cmd.input))
	return golisp.RunParser(buf, func(expr golisp.Value, err error) error {
		if err == nil {
			_, err = context.Eval(expr)
		}
		return err
	})
}
