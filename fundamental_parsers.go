package pars

import (
	"bytes"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type anyRuneParser struct {
	buf [utf8.UTFMax]byte
	i   int
}

//NewAnyRune returns a parser that parses a single valid rune. If no such rune can be read, ErrRuneExpected is returned.
func NewAnyRune() Parser {
	return &anyRuneParser{i: -1}
}

//ErrRuneExpected is the error returned from an unsuccessful parsing of a parser returned by NewAnyRune.
var ErrRuneExpected = fmt.Errorf("Expected rune")

func (r *anyRuneParser) Parse(src *reader) (interface{}, error) {
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
	if r.i >= 0 {
		src.Unread(r.buf[:r.i+1])
		r.i = -1
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
	silent bool
}

//NewCharPred returns a parser that parses a single rune as long as it fulfills the given predicate.
func NewCharPred(pred func(rune) bool) Parser {
	return &charPredParser{pred: pred}
}

func newSilentCharPred(pred func(rune) bool) Parser {
	return &charPredParser{pred: pred, silent: true}
}

var errCharPredParserSilentFailedPredicateError = fmt.Errorf("Could not parse expected rune: Rune does not hold predicate")

func (c *charPredParser) Parse(src *reader) (interface{}, error) {
	val, err := c.anyRuneParser.Parse(src)
	if err != nil {
		if c.silent {
			return nil, err
		}
		return nil, fmt.Errorf("Could not parse expected rune: %v", err.Error())
	}
	if val, ok := val.(rune); ok {
		if c.pred(val) {
			return val, nil
		}
		c.anyRuneParser.Unread(src)
		if c.silent {
			return nil, errCharPredParserSilentFailedPredicateError
		}
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
	silent   bool
}

var errStringParserSilentError = fmt.Errorf("Could not parse expected string, error silenced")

//NewString returns a parser for a single known string. Different strings are treated as a parsing error.
func NewString(expected string) Parser {
	return &stringParser{expected: expected}
}

func newSilentString(expected string) Parser {
	return &stringParser{expected: expected, silent: true}
}

func (s *stringParser) Parse(src *reader) (val interface{}, err error) {
	s.buf = make([]byte, len([]byte(s.expected)))
	n, err := src.Read(s.buf)

	if n == len(s.buf) && string(s.buf) == s.expected {
		return s.expected, nil
	}

	if n == len(s.buf) {
		if s.silent {
			err = errStringParserSilentError
		} else {
			err = fmt.Errorf("Unexpected string \"%v\"", string(s.buf))
		}
	}

	src.Unread(s.buf[:n])
	s.buf = nil

	if s.silent {
		err = errStringParserSilentError
	} else {
		err = fmt.Errorf("Could not parse expected string \"%v\": %v", s.expected, err)
	}
	return nil, err
}

func (s *stringParser) Unread(src *reader) {
	if s.buf != nil {
		src.Unread(s.buf)
		s.buf = nil
	}
}

func (s *stringParser) Clone() Parser {
	return &stringParser{expected: s.expected, silent: s.silent}
}

type delimitedStringParser struct {
	Parser
}

//NewDelimitedString returns a parser that parses a string between two given delimiter strings and returns the value between.
func NewDelimitedString(beginDelimiter, endDelimiter string) Parser {
	return &delimitedStringParser{Parser: NewDiscardLeft(NewString(beginDelimiter), NewDiscardRight(NewSome(NewExcept(NewAnyRune(), newSilentString(endDelimiter))), NewString(endDelimiter)))}
}

func (d *delimitedStringParser) Parse(src *reader) (interface{}, error) {
	val, err := d.Parser.Parse(src)
	if err != nil {
		return nil, err
	}

	runes := val.([]interface{})

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
	Parser
}

//NewInt returns a parser that parses an integer. The parsed integer is converted via strconv.Atoi.
func NewInt() Parser {
	return &intParser{Parser: newIntegralString()}
}

func (i *intParser) Parse(src *reader) (interface{}, error) {
	val, err := i.Parser.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("Could not parse int: %v", err)
	}

	val, err = strconv.Atoi(val.(string))
	if err != nil {
		i.Unread(src)
		return nil, fmt.Errorf("Could not parse int: %v", err)
	}
	return val, nil
}

func (i *intParser) Clone() Parser {
	return NewInt()
}

type bigIntParser struct {
	Parser
}

//NewBigInt returns a parser that parses an integer. The parsed integer is returned as a math/big.Int.
func NewBigInt() Parser {
	return &bigIntParser{Parser: newIntegralString()}
}

func (i *bigIntParser) Parse(src *reader) (interface{}, error) {
	val, err := i.Parser.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("Could not parse int: %v", err)
	}

	bigInt := big.NewInt(0)
	bigInt, ok := bigInt.SetString(val.(string), 10)
	if ok != true {
		i.Unread(src)
		return nil, fmt.Errorf("Could not parse '%v' as int", val.(string))
	}
	return bigInt, nil
}

func (i *bigIntParser) Clone() Parser {
	return NewBigInt()
}

type integralStringParser struct {
	parsers []Parser
}

func newIntegralString() Parser {
	return &integralStringParser{}
}

func (i *integralStringParser) Parse(src *reader) (interface{}, error) {
	buf := bytes.NewBuffer(nil)
	var err error
	for {
		var next Parser
		if buf.Len() == 0 {
			next = NewOr(NewChar('-'), NewCharPred(unicode.IsDigit))
		} else {
			next = newSilentCharPred(unicode.IsDigit)
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
		return buf.String(), nil
	}

	return nil, err
}

func (i *integralStringParser) Unread(src *reader) {
	for j := len(i.parsers) - 1; j >= 0; j-- {
		i.parsers[j].Unread(src)
	}
	i.parsers = nil
}

func (i *integralStringParser) Clone() Parser {
	return newIntegralString()
}
