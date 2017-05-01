package pars

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type Parser interface {
	Parse(*Reader) (interface{}, error)
	Unread(*Reader)
	Clone() Parser
}

func ParseString(s string, p Parser) (interface{}, error) {
	r := NewReader(strings.NewReader(s))
	return p.Parse(r)
}

type AnyRune struct {
	buf []byte
	i   int
}

func NewAnyRune() Parser {
	return &AnyRune{}
}

var ErrRuneExpected = fmt.Errorf("Expected rune")

//Parse tries to read a single rune or fails.
func (r *AnyRune) Parse(src *Reader) (interface{}, error) {
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

func (r *AnyRune) Unread(src *Reader) {
	if r.i >= 0 && r.buf != nil {
		src.Unread(r.buf[:r.i+1])
		r.buf = nil
		r.i = 0
	}
}

func (r *AnyRune) Clone() Parser {
	return &AnyRune{}
}

type AnyByte struct {
	buf  [1]byte
	read bool
}

func NewAnyByte() Parser {
	return &AnyByte{}
}

func (b *AnyByte) Parse(src *Reader) (interface{}, error) {
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

func (b *AnyByte) Unread(src *Reader) {
	if b.read {
		src.Unread(b.buf[:])
		b.read = false
	}
}

func (b *AnyByte) Clone() Parser {
	return &AnyByte{}
}

type Char struct {
	expected rune
	AnyRune
}

func NewChar(r rune) Parser {
	return &Char{expected: r}
}

func (c *Char) Parse(src *Reader) (interface{}, error) {
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

type CharPred struct {
	pred func(rune) bool
	AnyRune
}

func NewCharPred(pred func(rune) bool) Parser {
	return &CharPred{pred: pred}
}

func (c *CharPred) Parse(src *Reader) (interface{}, error) {
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

func (s *Seq) Parse(src *Reader) (interface{}, error) {
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

func (s *Seq) Unread(src *Reader) {
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

//Or is a parser that matches the first of a given set of parsers. A later parser will not be tried if an earlier match was found.
//Or uses the error message of the last parser verbatim.
type Or struct {
	parsers  []Parser
	selected Parser
}

func NewOr(parsers ...Parser) Parser {
	return &Or{parsers: parsers}
}

func (o *Or) Parse(src *Reader) (val interface{}, err error) {
	for _, parser := range o.parsers {
		val, err = parser.Parse(src)
		if err == nil {
			o.selected = parser
			return
		}
	}
	return
}

func (o *Or) Unread(src *Reader) {
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

type String struct {
	expected string
	buf      []byte
}

func NewString(expected string) Parser {
	return &String{expected: expected}
}

func (s *String) Parse(src *Reader) (val interface{}, err error) {
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

func (s *String) Unread(src *Reader) {
	if s.buf != nil {
		src.Unread(s.buf)
		s.buf = nil
	}
}

func (s *String) Clone() Parser {
	return NewString(s.expected)
}
