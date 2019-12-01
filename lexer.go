package golisp

//go:generate go run -mod vendor golang.org/x/tools/cmd/goyacc -o parser.go parser.go.y

import (
	"bufio"
	"errors"
	"strconv"
	"unicode"
)

func RunParser(reader *bufio.Reader, handler func(Value, error) error) (err error) {
	yyErrorVerbose = true
	lex := &lexer{reader: reader}
	for {
		lex.skipSpaces()
		if lex.peek() == eof {
			break
		}
		yyParse(lex)
		result := lex.result
		err = lex.err
		lex.result = nil
		lex.err = nil
		err = handler(result, err)
		if err != nil {
			break
		}
	}
	return
}

type token struct {
	typ int
	lit string
	str string
	num float64
}

type lexer struct {
	reader  *bufio.Reader
	current []rune

	result Value
	err    error
}

func (l *lexer) Lex(lval *yySymType) int {
	lval.tok = l.next()
	return lval.tok.typ
}

func (l *lexer) Error(msg string) {
	if l.err == nil {
		l.err = errors.New(msg)
	}
}

func (l *lexer) next() token {
	l.skipSpaces()
	c := l.read()

	switch {
	case c == eof:
		return l.emit(0)

	case c == '(':
		return l.emit(LPAREN)

	case c == ')':
		return l.emit(RPAREN)

	case c == '[':
		return l.emit(LBRACK)

	case c == ']':
		return l.emit(RBRACK)

	case c == '.':
		return l.emit(DOT)

	case c == '#':
		c = l.read()
		switch c {
		case 't':
			return l.emit(TRUE)
		case 'f':
			return l.emit(FALSE)
		default:
			return l.fail("Unexpected character: " + string(c))
		}

	case c == '\'':
		return l.emit(QUOTE)

	case c == '`':
		return l.emit(QUASIQUOTE)

	case c == ',':
		if l.peek() == '@' {
			l.read()
			return l.emit(UNQUOTE_SPLICING)
		}
		return l.emit(UNQUOTE)

	case unicode.IsLetter(c) || isSpecial(c):
		// ['+' '-'] digit
		if (c == '+' || c == '-') && unicode.IsDigit(l.peek()) {
			op := l.emit(UNUSED)
			r := l.nextNum()
			r.lit = op.lit + r.lit
			if op.lit == "-" {
				r.num = -r.num
			}
			return r
		}

		l.readWhile(func(c rune) bool {
			return unicode.IsLetter(c) || unicode.IsDigit(c) || isSpecial(c)
		})
		return l.emit(SYM)

	case unicode.IsDigit(c):
		l.unread(c)
		return l.nextNum()

	case c == '"':
		l.unread(c)
		return l.nextStr()

	default:
		return l.fail("Unexpected character: " + string(c))
	}
}

func (l *lexer) skipSpaces() {
	for {
		c := l.read()

		switch {
		case unicode.IsSpace(c):
			l.discard()
			continue

		case c == ';':
			l.discard()
			for c != eof && c != '\r' && c != '\n' {
				c = l.read()
				l.discard()
			}
			continue
		}

		l.unread(c)
		break
	}
}

func (l *lexer) nextNum() token {
	l.readWhile(unicode.IsDigit)

	// frac
	if l.peek() == '.' {
		l.read()
		l.readWhile(unicode.IsDigit)
	}

	// exp
	if l.peek() == 'e' || l.peek() == 'E' {
		l.read()
		if l.peek() == '-' || l.peek() == '+' {
			l.read()
		}
		l.readWhile(unicode.IsDigit)
	}

	r := l.emit(NUM)
	num, err := strconv.ParseFloat(r.lit, 64)
	if err != nil {
		panic(err)
	}
	r.num = num
	return r
}

func (l *lexer) nextStr() token {
	l.read() // read '"'
	buf := []rune{}

	for {
		c := l.read()
		switch c {
		case '"':
			r := l.emit(STR)
			r.str = string(buf)
			return r

		case '\\':
			c = l.read()
			switch c {
			case '\\':
				buf = append(buf, '\\')
			case 't':
				buf = append(buf, '\t')
			case 'n':
				buf = append(buf, '\n')
			case '"':
				buf = append(buf, '"')
			default:
				return l.fail("Unsupported escape sequence: " + string(c))
			}

		case eof:
			return l.fail("String is not terminated")

		default:
			buf = append(buf, c)
		}
	}
}

var eof = rune(0)

func (l *lexer) read() rune {
	if l.result != nil {
		return eof
	}

	c, _, err := l.reader.ReadRune()
	if err != nil {
		return eof
	}

	l.current = append(l.current, c)
	return c
}

func (l *lexer) unread(c rune) {
	if c == eof {
		return
	}
	l.reader.UnreadRune()
	l.current = l.current[:len(l.current)-1]
}

func (l *lexer) peek() (c rune) {
	c = l.read()
	l.unread(c)
	return
}

func (l *lexer) readWhile(cond func(rune) bool) {
	for {
		c := l.read()
		if !cond(c) {
			l.unread(c)
			break
		}
	}
}

func (l *lexer) discard() {
	l.current = l.current[:0]
}

func (l *lexer) emit(typ int) token {
	r := token{typ: typ, lit: string(l.current)}
	l.current = l.current[:0]
	return r
}

func (l *lexer) fail(msg string) token {
	l.current = l.current[:0]
	l.err = errors.New(msg)
	return token{typ: UNUSED}
}

func isSpecial(c rune) bool {
	switch c {
	case '!', '$', '%', '&', '*', '+', '-', '.', '/',
		':', '<', '=', '>', '?', '@', '^', '_', '~':
		return true
	default:
		return false
	}
}
