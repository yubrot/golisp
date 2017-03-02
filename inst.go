package golisp

type Code []inst

type inst interface {}

type ldc struct {
	value Value
}

type ldv struct {
	name string
}

type ldf struct {
	pattern pattern
	code Code
}

type ldm struct {
	pattern pattern
	code Code
}

type ldb struct {
	name string
}

type sel struct {
	a, b Code
}

type app struct {
	argc int
}

type leave struct {}

type pop struct {}

type def struct {
	name string
}

type set struct {
	name string
}
