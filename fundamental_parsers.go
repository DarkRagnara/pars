package pars

import (
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

//AnyRune returns a parser that parses a single valid rune. If no such rune can be read, ErrRuneExpected is returned.
func AnyRune() Parser {
	return &anyRuneParser{i: -1}
}

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
	return nil, errRuneExpected
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

//AnyByte returns a parser that reads exactly one byte from the source.
func AnyByte() Parser {
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

//Byte returns a parser used to read a single known byte. A different byte is treated as a parsing error.
func Byte(b byte) Parser {
	return Transformer(AnyByte(), func(val interface{}) (interface{}, error) {
		if val, ok := val.(byte); ok {
			if val == b {
				return val, nil
			}
			return nil, byteExpectationError{expected: b, actual: val}
		}
		panic("AnyByte returned type != byte")
	})
}

type charParser struct {
	expected rune
	anyRuneParser
}

//Char returns a parser used to read a single known rune. A different rune is treated as a parsing error.
func Char(r rune) Parser {
	return &charParser{expected: r}
}

func (c *charParser) Parse(src *reader) (interface{}, error) {
	val, err := c.anyRuneParser.Parse(src)
	if err != nil {
		return nil, runeExpectationNoRuneError{expected: c.expected, innerError: err}
	}
	if val, ok := val.(rune); ok {
		if val == c.expected {
			return val, nil
		}
		c.anyRuneParser.Unread(src)
		return nil, runeExpectationError{expected: c.expected, actual: val}
	}
	panic("AnyRune returned type != rune")
}

func (c *charParser) Clone() Parser {
	return Char(c.expected)
}

type charPredParser struct {
	pred func(rune) bool
	anyRuneParser
}

//CharPred returns a parser that parses a single rune as long as it fulfills the given predicate.
func CharPred(pred func(rune) bool) Parser {
	return &charPredParser{pred: pred}
}

func (c *charPredParser) Parse(src *reader) (interface{}, error) {
	val, err := c.anyRuneParser.Parse(src)
	if err != nil {
		return nil, runePredNoRuneError{innerError: err}
	}
	if val, ok := val.(rune); ok {
		if c.pred(val) {
			return val, nil
		}
		c.anyRuneParser.Unread(src)
		return nil, runePredError{actual: val}
	}
	panic("AnyRune returned type != rune")
}

func (c *charPredParser) Clone() Parser {
	return CharPred(c.pred)
}

type stringParser struct {
	expected string
	buf      []byte
}

//String returns a parser for a single known string. Different strings are treated as a parsing error.
func String(expected string) Parser {
	return &stringParser{expected: expected}
}

func (s *stringParser) Parse(src *reader) (val interface{}, err error) {
	s.buf = make([]byte, len(s.expected))
	n, err := src.Read(s.buf)

	actual := string(s.buf)
	if n == len(s.buf) && actual == s.expected {
		return s.expected, nil
	}

	if n == len(s.buf) {
		err = unexpectedStringError{expected: s.expected, actual: actual}
	}

	src.Unread(s.buf[:n])
	s.buf = nil

	err = stringError{expected: s.expected, innerError: err}
	return nil, err
}

func (s *stringParser) Unread(src *reader) {
	if s.buf != nil {
		src.Unread(s.buf)
		s.buf = nil
	}
}

func (s *stringParser) Clone() Parser {
	return &stringParser{expected: s.expected}
}

type stringCIParser struct {
	expected string
	buf      []byte
}

//StringCI returns a case-insensitive parser for a single known string. Different strings are treated as a parsing error.
func StringCI(expected string) Parser {
	return &stringCIParser{expected: expected}
}

func (s *stringCIParser) Parse(src *reader) (val interface{}, err error) {
	s.buf = make([]byte, len(s.expected))
	n, err := src.Read(s.buf)

	actual := string(s.buf)
	if n == len(s.buf) && strings.EqualFold(actual, s.expected) {
		return actual, nil
	}

	if n == len(s.buf) {
		err = unexpectedStringError{expected: s.expected, actual: actual}
	}

	src.Unread(s.buf[:n])
	s.buf = nil

	err = stringError{expected: s.expected, innerError: err}
	return nil, err
}

func (s *stringCIParser) Unread(src *reader) {
	if s.buf != nil {
		src.Unread(s.buf)
		s.buf = nil
	}
}

func (s *stringCIParser) Clone() Parser {
	return &stringCIParser{expected: s.expected}
}

//RunesUntil returns a parser that parses runes as long as the given endCondition parser does not match.
func RunesUntil(endCondition Parser) Parser {
	return Some(Except(AnyRune(), endCondition))
}

//DelimitedString returns a parser that parses a string between two given delimiter strings and returns the value between.
func DelimitedString(beginDelimiter, endDelimiter string) Parser {
	return RunesToString(DiscardLeft(String(beginDelimiter), DiscardRight(RunesUntil(String(endDelimiter)), String(endDelimiter))))
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
		err = eofByteError{actual: buf[0]}
		src.Unread(buf[:])
	}
	return nil, eofOtherError{innerError: err}
}

func (e eof) Unread(src *reader) {
}

func (e eof) Clone() Parser {
	return e
}

type errorParser struct {
	error
}

//Error returns a parser that always fails with the given error
func Error(err error) Parser {
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

//Int returns a parser that parses an integer. The parsed integer is converted via strconv.Atoi.
func Int() Parser {
	return Transformer(newIntegralString(), func(v interface{}) (interface{}, error) {
		val, err := strconv.Atoi(v.(string))
		if err != nil {
			return nil, intError{innerError: err}
		}
		return val, nil
	})
}

//BigInt returns a parser that parses an integer. The parsed integer is returned as a math/big.Int.
func BigInt() Parser {
	return Transformer(newIntegralString(), func(v interface{}) (interface{}, error) {
		bigInt := big.NewInt(0)
		bigInt, ok := bigInt.SetString(v.(string), 10)
		if !ok {
			return nil, intConversionError{actual: v.(string)}
		}
		return bigInt, nil
	})
}

//Float returns a parser that parses a floating point number. The supported format is an optional minus sign followed by digits optionally followed by a decimal point and more digits.
func Float() Parser {
	return Transformer(newFloatNumberString(), func(v interface{}) (interface{}, error) {
		val, err := strconv.ParseFloat(v.(string), 64)
		if err != nil {
			return nil, floatError{innerError: err}
		}
		return val, nil
	})
}

type integralStringParser struct {
	parsers []Parser
}

func newIntegralString() Parser {
	return &integralStringParser{}
}

func (i *integralStringParser) Parse(src *reader) (interface{}, error) {
	buf := strings.Builder{}
	var err error
	for {
		var next Parser
		if buf.Len() == 0 {
			next = Or(Char('-'), CharPred(unicode.IsDigit))
		} else {
			next = CharPred(unicode.IsDigit)
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

	return nil, intError{innerError: err}
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

type floatNumberStringParser struct {
	parsers []Parser
}

func newFloatNumberString() Parser {
	return &floatNumberStringParser{}
}

func (i *floatNumberStringParser) Parse(src *reader) (interface{}, error) {
	buf := strings.Builder{}
	var err error
	var foundDecimalPoint bool
	decimalPointParser := Transformer(Char('.'), func(decimalPoint interface{}) (interface{}, error) {
		foundDecimalPoint = true
		return decimalPoint, nil
	})
	for {
		var next Parser
		if buf.Len() == 0 {
			next = Or(Char('-'), CharPred(unicode.IsDigit))
		} else if !foundDecimalPoint {
			next = Or(CharPred(unicode.IsDigit), decimalPointParser)
		} else {
			next = CharPred(unicode.IsDigit)
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

	return nil, floatError{innerError: err}
}

func (i *floatNumberStringParser) Unread(src *reader) {
	for j := len(i.parsers) - 1; j >= 0; j-- {
		i.parsers[j].Unread(src)
	}
	i.parsers = nil
}

func (i *floatNumberStringParser) Clone() Parser {
	return newFloatNumberString()
}
