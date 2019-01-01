package pars

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type anyRuneParser struct {
	buf []byte
	i   int
}

//NewAnyRune returns a parser that parses a single valid rune. If no such rune can be read, ErrRuneExpected is returned.
func NewAnyRune() Parser {
	return &anyRuneParser{}
}

//ErrRuneExpected is the error returned from an unsuccessful parsing of a parser returned by NewAnyRune.
var ErrRuneExpected = fmt.Errorf("Expected rune")

func (r *anyRuneParser) Parse(src *reader) (interface{}, error) {
	r.buf = make([]byte, utf8.UTFMax)

	r.i = 0
	for ; r.i < len(r.buf); r.i++ {
		_, err := src.Read(r.buf[r.i : r.i+1])
		if err != nil {
			r.i--
			r.Unread(src)
			return nil, err
		}

		if utf8.FullRune(r.buf[0 : r.i+1]) {
			rune := rune(r.buf[0])
			if rune >= utf8.RuneSelf {
				rune, _ = utf8.DecodeRune(r.buf[0 : r.i+1])
			}

			if rune != 0xfffd {
				return rune, nil
			}
			break
		}
	}

	r.Unread(src)
	return nil, ErrRuneExpected
}

func (r *anyRuneParser) Unread(src *reader) {
	if r.i >= 0 && r.buf != nil {
		src.Unread(r.buf[:r.i+1])
		r.buf = nil
		r.i = 0
	}
}

func (r *anyRuneParser) Clone() Parser {
	return &anyRuneParser{}
}

type anyByteParser struct {
	buf  [1]byte
	read bool
}

//NewAnyByte returns a parser that reads exactly one byte from the source.
func NewAnyByte() Parser {
	return &anyByteParser{}
}

func (b *anyByteParser) Parse(src *reader) (interface{}, error) {
	n, err := src.Read(b.buf[:])
	if err != nil {
		return nil, err
	}
	if n != 1 {
		panic("AnyByte read bytes != 1")
	}
	b.read = true
	return b.buf[0], nil
}

func (b *anyByteParser) Unread(src *reader) {
	if b.read {
		src.Unread(b.buf[:])
		b.read = false
	}
}

func (b *anyByteParser) Clone() Parser {
	return &anyByteParser{}
}

type charParser struct {
	expected rune
	anyRuneParser
}

//NewChar returns a parser used to read a single known rune. A different rune is treated as a parsing error.
func NewChar(r rune) Parser {
	return &charParser{expected: r}
}

func (c *charParser) Parse(src *reader) (interface{}, error) {
	val, err := c.anyRuneParser.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("Could not parse expected rune '%c' (0x%x): %v", c.expected, c.expected, err.Error())
	}
	if val, ok := val.(rune); ok {
		if val == c.expected {
			return val, nil
		}
		c.anyRuneParser.Unread(src)
		return nil, fmt.Errorf("Could not parse expected rune '%c' (0x%x): Unexpected rune '%c' (0x%x)", c.expected, c.expected, val, val)
	}
	panic("AnyRune returned type != rune")
}

func (c *charParser) Clone() Parser {
	return NewChar(c.expected)
}

type charPredParser struct {
	pred func(rune) bool
	anyRuneParser
}

//NewCharPred returns a parser that parses a single rune as long as it fulfills the given predicate.
func NewCharPred(pred func(rune) bool) Parser {
	return &charPredParser{pred: pred}
}

func (c *charPredParser) Parse(src *reader) (interface{}, error) {
	val, err := c.anyRuneParser.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("Could not parse expected rune: %v", err.Error())
	}
	if val, ok := val.(rune); ok {
		if c.pred(val) {
			return val, nil
		}
		c.anyRuneParser.Unread(src)
		return nil, fmt.Errorf("Could not parse expected rune: Rune '%c' (0x%x) does not hold predicate", val, val)
	}
	panic("AnyRune returned type != rune")
}

func (c *charPredParser) Clone() Parser {
	return NewCharPred(c.pred)
}

type stringParser struct {
	expected string
	buf      []byte
}

//NewString returns a parser for a single known string. Different strings are treated as a parsing error.
func NewString(expected string) Parser {
	return &stringParser{expected: expected}
}

func (s *stringParser) Parse(src *reader) (val interface{}, err error) {
	s.buf = make([]byte, len([]byte(s.expected)))
	n, err := src.Read(s.buf)

	if n == len(s.buf) && string(s.buf) == s.expected {
		return string(s.buf), nil
	}

	if n == len(s.buf) {
		err = fmt.Errorf("Unexpected string \"%v\"", string(s.buf))
	}

	src.Unread(s.buf[:n])
	s.buf = nil
	return nil, fmt.Errorf("Could not parse expected string \"%v\": %v", s.expected, err)
}

func (s *stringParser) Unread(src *reader) {
	if s.buf != nil {
		src.Unread(s.buf)
		s.buf = nil
	}
}

func (s *stringParser) Clone() Parser {
	return NewString(s.expected)
}

type delimitedStringParser struct {
	Parser
}

//NewDelimitedString returns a parser that parses a string between two identical delimiter strings and returns the value between.
func NewDelimitedString(delimiter string) Parser {
	return &delimitedStringParser{Parser: NewSeq(NewString(delimiter), NewSome(NewExcept(NewAnyRune(), NewString(delimiter))), NewString(delimiter))}
}

func (d *delimitedStringParser) Parse(src *reader) (interface{}, error) {
	val, err := d.Parser.Parse(src)
	if err != nil {
		return nil, err
	}

	values := val.([]interface{})
	runes := values[1].([]interface{})

	builder := strings.Builder{}
	for _, r := range runes {
		builder.WriteRune(r.(rune))
	}
	return builder.String(), nil
}

func (d *delimitedStringParser) Clone() Parser {
	return &delimitedStringParser{Parser: d.Parser.Clone()}
}

type eof struct{}

//EOF is a parser that never yields a value but that succeeds if and only if the source reached EOF
var EOF Parser = eof{}

func (e eof) Parse(src *reader) (interface{}, error) {
	buf := [1]byte{}
	n, err := src.Read(buf[:])
	if err == io.EOF {
		return nil, nil
	}
	if n != 0 {
		err = fmt.Errorf("Found byte 0x%x", buf[0])
		src.Unread(buf[:])
	}
	return nil, fmt.Errorf("Expected EOF: %v", err)
}

func (e eof) Unread(src *reader) {
}

func (e eof) Clone() Parser {
	return e
}

type errorParser struct {
	error
}

//NewError returns a parser that always fails with the given error
func NewError(err error) Parser {
	return errorParser{err}
}

func (e errorParser) Parse(src *reader) (interface{}, error) {
	return nil, e.error
}

func (e errorParser) Unread(src *reader) {
}

func (e errorParser) Clone() Parser {
	return e
}

type intParser struct {
	parsers []Parser
}

//NewInt returns a parser that parses an integer. The parsed integer is converted via strconv.Atoi.
func NewInt() Parser {
	return &intParser{}
}

func (i *intParser) Parse(src *reader) (interface{}, error) {
	buf := bytes.NewBuffer(nil)
	var err error
	for {
		var next Parser
		if buf.Len() == 0 {
			next = NewOr(NewChar('-'), NewCharPred(unicode.IsDigit))
		} else {
			next = NewCharPred(unicode.IsDigit)
		}
		var val interface{}
		val, err = next.Parse(src)
		if err != nil {
			next.Unread(src)
			break
		}
		buf.WriteRune(val.(rune))
		i.parsers = append(i.parsers, next)
	}
	if buf.Len() > 0 {
		val, err := strconv.Atoi(buf.String())
		if err != nil {
			i.Unread(src)
			return nil, fmt.Errorf("Could not parse int: %v", err)
		}
		return val, nil
	}

	return nil, fmt.Errorf("Could not parse int: %v", err)
}

func (i *intParser) Unread(src *reader) {
	for j := len(i.parsers) - 1; j >= 0; j-- {
		i.parsers[j].Unread(src)
	}
	i.parsers = nil
}

func (i *intParser) Clone() Parser {
	return NewInt()
}
