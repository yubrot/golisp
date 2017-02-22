package golisp

type code []inst

type inst interface {}

type ldc struct {
	value Value
}

type ldv struct {
	name string
}

type ldf struct {
	pattern pattern
	code code
}

type ldm struct {
	pattern pattern
	code code
}

type ldb struct {
	name string
}

type sel struct {
	a, b code
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
