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

type seqParser struct {
	parsers []Parser
}

//NewSeq returns a parser that matches all of its given parsers in order or none of them.
func NewSeq(parsers ...Parser) Parser {
	return &seqParser{parsers: parsers}
}

func (s *seqParser) Parse(src *reader) (interface{}, error) {
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

func (s *seqParser) Unread(src *reader) {
	for i := len(s.parsers) - 1; i >= 0; i-- {
		s.parsers[i].Unread(src)
	}
}

func (s *seqParser) Clone() Parser {
	s2 := &seqParser{parsers: make([]Parser, len(s.parsers))}
	for i, parser := range s.parsers {
		s2.parsers[i] = parser.Clone()
	}
	return s2
}

type someParser struct {
	prototype Parser
	used      []Parser
}

//NewSome returns a parser that matches a given parser zero or more times. Not matching at all is not an error.
func NewSome(parser Parser) Parser {
	return &someParser{prototype: parser}
}

func (s *someParser) Parse(src *reader) (interface{}, error) {
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

func (s *someParser) Unread(src *reader) {
	for i := len(s.used) - 1; i >= 0; i-- {
		s.used[i].Unread(src)
	}
	s.used = nil
}

func (s *someParser) Clone() Parser {
	return &someParser{prototype: s.prototype.Clone()}
}

type manyParser struct {
	Parser
}

//NewMany returns a parser that matches a given parser one or more times. Not matching at all is an error.
func NewMany(parser Parser) Parser {
	return &manyParser{Parser: NewSeq(parser, NewSome(parser))}
}

func (m *manyParser) Parse(src *reader) (interface{}, error) {
	val, err := m.Parser.Parse(src)
	if err != nil {
		return nil, err
	}

	results := val.([]interface{})
	values := append([]interface{}{results[0]}, results[1].([]interface{})...)

	return values, nil
}

func (m *manyParser) Clone() Parser {
	return &manyParser{Parser: m.Parser.Clone()}
}

type orParser struct {
	parsers  []Parser
	selected Parser
}

//NewOr returns a parser that matches the first of a given set of parsers. A later parser will not be tried if an earlier match was found.
//The returned parser uses the error message of the last parser verbatim.
func NewOr(parsers ...Parser) Parser {
	return &orParser{parsers: parsers}
}

func (o *orParser) Parse(src *reader) (val interface{}, err error) {
	for _, parser := range o.parsers {
		val, err = parser.Parse(src)
		if err == nil {
			o.selected = parser
			return
		}
	}
	return
}

func (o *orParser) Unread(src *reader) {
	if o.selected != nil {
		o.selected.Unread(src)
		o.selected = nil
	}
}

func (o *orParser) Clone() Parser {
	o2 := &orParser{parsers: make([]Parser, len(o.parsers))}
	for i, parser := range o.parsers {
		o2.parsers[i] = parser.Clone()
	}
	return o2
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

type exceptParser struct {
	Parser
	except Parser
}

//ErrExceptionMatched signals that an parser returned by exceptParser matched its exception.
var ErrExceptionMatched = fmt.Errorf("Excepted parser matched")

//NewExcept returns a parser that wraps another parser so that it fails if a third, excepted parser would succeed.
func NewExcept(parser, except Parser) Parser {
	return &exceptParser{Parser: parser, except: except}
}

func (e *exceptParser) Parse(src *reader) (val interface{}, err error) {
	val, err = e.except.Parse(src)
	if err == nil {
		e.except.Unread(src)
		return nil, ErrExceptionMatched
	}
	val, err = e.Parser.Parse(src)
	return
}

func (e *exceptParser) Clone() Parser {
	return NewExcept(e.Parser.Clone(), e.except.Clone())
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

type optionalParser struct {
	read bool
	Parser
}

//NewOptional returns a parser that reads exactly one result according to a given other parser. If it fails, the error is discarded and nil is returned.
func NewOptional(parser Parser) Parser {
	return &optionalParser{Parser: parser}
}

func (o *optionalParser) Parse(src *reader) (interface{}, error) {
	val, err := o.Parser.Parse(src)
	if err == nil {
		o.read = true
		return val, nil
	}
	return nil, nil
}

func (o *optionalParser) Unread(src *reader) {
	if o.read {
		o.Parser.Unread(src)
		o.read = false
	}
}

func (o *optionalParser) Clone() Parser {
	return &optionalParser{Parser: o.Parser.Clone()}
}
