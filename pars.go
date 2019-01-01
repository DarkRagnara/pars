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

//Parser contains the methods that each parser in this framework has to provide.
type Parser interface {
	//Parse is used for the actual parsing. It reads from the reader and returns the result or an error value.
	//Each parser must remember enough from the call to this method to undo the reading in case of a parsing error that occurs later.
	//When Parse returns with an error, Parse must make sure that all read bytes are unread so that another parser could try to parse them.
	Parse(*reader) (interface{}, error)
	//Unread puts read bytes back to the reader so that they can be read again by other parsers.
	Unread(*reader)
	//Clone creates a parser that works the same as the receiver. This allows to create a single parser as a blueprint for other parsers.
	//Internal state from reading operations should not be cloned.
	Clone() Parser
}

//ParseString is a helper function to directly use a parser on a string.
func ParseString(s string, p Parser) (interface{}, error) {
	r := newReader(strings.NewReader(s))
	return p.Parse(r)
}

//ParseFromReader parses from an io.Reader.
func ParseFromReader(ior io.Reader, p Parser) (interface{}, error) {
	r := newReader(ior)
	return p.Parse(r)
}

//AnyRune is a parser that parses a single valid rune. If no such rune can be read, ErrRuneExpected is returned.
type AnyRune struct {
	buf []byte
	i   int
}

func NewAnyRune() Parser {
	return &AnyRune{}
}

//ErrRuneExpected is the error returned from an unsuccessful AnyRune parsing.
var ErrRuneExpected = fmt.Errorf("Expected rune")

func (r *AnyRune) Parse(src *reader) (interface{}, error) {
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

func (r *AnyRune) Unread(src *reader) {
	if r.i >= 0 && r.buf != nil {
		src.Unread(r.buf[:r.i+1])
		r.buf = nil
		r.i = 0
	}
}

func (r *AnyRune) Clone() Parser {
	return &AnyRune{}
}

//AnyByte is a parser that reads exactly one byte from the source.
type AnyByte struct {
	buf  [1]byte
	read bool
}

func NewAnyByte() Parser {
	return &AnyByte{}
}

func (b *AnyByte) Parse(src *reader) (interface{}, error) {
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

func (b *AnyByte) Unread(src *reader) {
	if b.read {
		src.Unread(b.buf[:])
		b.read = false
	}
}

func (b *AnyByte) Clone() Parser {
	return &AnyByte{}
}

//Char is a parser used to read a single known rune. A different rune is treated as a parsing error.
type Char struct {
	expected rune
	AnyRune
}

func NewChar(r rune) Parser {
	return &Char{expected: r}
}

func (c *Char) Parse(src *reader) (interface{}, error) {
	val, err := c.AnyRune.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("Could not parse expected rune '%c' (0x%x): %v", c.expected, c.expected, err.Error())
	}
	if val, ok := val.(rune); ok {
		if val == c.expected {
			return val, nil
		}
		c.AnyRune.Unread(src)
		return nil, fmt.Errorf("Could not parse expected rune '%c' (0x%x): Unexpected rune '%c' (0x%x)", c.expected, c.expected, val, val)
	}
	panic("AnyRune returned type != rune")
}

func (c *Char) Clone() Parser {
	return NewChar(c.expected)
}

//CharPred parses a single rune as long as it fulfills the given predicate.
type CharPred struct {
	pred func(rune) bool
	AnyRune
}

func NewCharPred(pred func(rune) bool) Parser {
	return &CharPred{pred: pred}
}

func (c *CharPred) Parse(src *reader) (interface{}, error) {
	val, err := c.AnyRune.Parse(src)
	if err != nil {
		return nil, fmt.Errorf("Could not parse expected rune: %v", err.Error())
	}
	if val, ok := val.(rune); ok {
		if c.pred(val) {
			return val, nil
		}
		c.AnyRune.Unread(src)
		return nil, fmt.Errorf("Could not parse expected rune: Rune '%c' (0x%x) does not hold predicate", val, val)
	}
	panic("AnyRune returned type != rune")
}

func (c *CharPred) Clone() Parser {
	return NewCharPred(c.pred)
}

//Seq is a parser that matches all of its given parsers in order or none of them.
type Seq struct {
	parsers []Parser
}

func NewSeq(parsers ...Parser) Parser {
	return &Seq{parsers: parsers}
}

func (s *Seq) Parse(src *reader) (interface{}, error) {
	values := make([]interface{}, len(s.parsers))
	for i, parser := range s.parsers {
		val, err := parser.Parse(src)
		if err != nil {
			for j := i - 1; j >= 0; j-- {
				s.parsers[j].Unread(src)
			}
			return nil, fmt.Errorf("Could not find expected sequence item %v: %v", i, err)
		}
		values[i] = val
	}
	return values, nil
}

func (s *Seq) Unread(src *reader) {
	for i := len(s.parsers) - 1; i >= 0; i-- {
		s.parsers[i].Unread(src)
	}
}

func (s *Seq) Clone() Parser {
	s2 := &Seq{parsers: make([]Parser, len(s.parsers))}
	for i, parser := range s.parsers {
		s2.parsers[i] = parser.Clone()
	}
	return s2
}

//Some matches a given parser zero or more times. Not matching at all is not an error.
type Some struct {
	prototype Parser
	used      []Parser
}

func NewSome(parser Parser) Parser {
	return &Some{prototype: parser}
}

func (s *Some) Parse(src *reader) (interface{}, error) {
	var values []interface{}
	for {
		next := s.prototype.Clone()
		s.used = append(s.used, next)

		nextVal, nextErr := next.Parse(src)
		if nextErr != nil {
			break
		}
		values = append(values, nextVal)
	}
	return values, nil
}

func (s *Some) Unread(src *reader) {
	for i := len(s.used) - 1; i >= 0; i-- {
		s.used[i].Unread(src)
	}
	s.used = nil
}

func (s *Some) Clone() Parser {
	return &Some{prototype: s.prototype.Clone()}
}

//Many matches a given parser one or more times. Not matching at all is an error.
type Many struct {
	Parser
}

func NewMany(parser Parser) Parser {
	return &Many{Parser: NewSeq(parser, NewSome(parser))}
}

func (m *Many) Parse(src *reader) (interface{}, error) {
	val, err := m.Parser.Parse(src)
	if err != nil {
		return nil, err
	}

	results := val.([]interface{})
	values := append([]interface{}{results[0]}, results[1].([]interface{})...)

	return values, nil
}

func (m *Many) Clone() Parser {
	return &Many{Parser: m.Parser.Clone()}
}

//Or is a parser that matches the first of a given set of parsers. A later parser will not be tried if an earlier match was found.
//Or uses the error message of the last parser verbatim.
type Or struct {
	parsers  []Parser
	selected Parser
}

func NewOr(parsers ...Parser) Parser {
	return &Or{parsers: parsers}
}

func (o *Or) Parse(src *reader) (val interface{}, err error) {
	for _, parser := range o.parsers {
		val, err = parser.Parse(src)
		if err == nil {
			o.selected = parser
			return
		}
	}
	return
}

func (o *Or) Unread(src *reader) {
	if o.selected != nil {
		o.selected.Unread(src)
		o.selected = nil
	}
}

func (o *Or) Clone() Parser {
	o2 := &Or{parsers: make([]Parser, len(o.parsers))}
	for i, parser := range o.parsers {
		o2.parsers[i] = parser.Clone()
	}
	return o2
}

//String parses a single known string. Different strings are treated as a parsing error.
type String struct {
	expected string
	buf      []byte
}

func NewString(expected string) Parser {
	return &String{expected: expected}
}

func (s *String) Parse(src *reader) (val interface{}, err error) {
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

func (s *String) Unread(src *reader) {
	if s.buf != nil {
		src.Unread(s.buf)
		s.buf = nil
	}
}

func (s *String) Clone() Parser {
	return NewString(s.expected)
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

//Error is a parser that always fails with the given error
type Error struct {
	error
}

func NewError(err error) Parser {
	return Error{err}
}

func (e Error) Parse(src *reader) (interface{}, error) {
	return nil, e.error
}

func (e Error) Unread(src *reader) {
}

func (e Error) Clone() Parser {
	return e
}

//Int parses an integer. The parsed integer is converted via strconv.Atoi.
type Int struct {
	parsers []Parser
}

func NewInt() Parser {
	return &Int{}
}

func (i *Int) Parse(src *reader) (interface{}, error) {
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

func (i *Int) Unread(src *reader) {
	for j := len(i.parsers) - 1; j >= 0; j-- {
		i.parsers[j].Unread(src)
	}
	i.parsers = nil
}

func (i *Int) Clone() Parser {
	return NewInt()
}
