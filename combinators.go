package pars

import (
	"fmt"
)

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
