package pars

type seqParser struct {
	parsers []Parser
}

//NewSeq returns a parser that matches all of its given parsers in order or none of them.
//
//Deprecated: Use Seq instead.
func NewSeq(parsers ...Parser) Parser {
	return Seq(parsers...)
}

//Seq returns a parser that matches all of its given parsers in order or none of them.
func Seq(parsers ...Parser) Parser {
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
			return nil, seqError{index: i, innerError: err}
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
//
//Deprecated: Use Some instead.
func NewSome(parser Parser) Parser {
	return Some(parser)
}

//Some returns a parser that matches a given parser zero or more times. Not matching at all is not an error.
func Some(parser Parser) Parser {
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

//NewMany returns a parser that matches a given parser one or more times. Not matching at all is an error.
//
//Deprecated: Use Many instead.
func NewMany(parser Parser) Parser {
	return Many(parser)
}

//Many returns a parser that matches a given parser one or more times. Not matching at all is an error.
func Many(parser Parser) Parser {
	return NewSplicingSeq(parser, NewSome(parser))
}

type orParser struct {
	parsers  []Parser
	selected Parser
}

//NewOr returns a parser that matches the first of a given set of parsers. A later parser will not be tried if an earlier match was found.
//The returned parser uses the error message of the last parser verbatim.
//
//Deprecated: Use Or instead.
func NewOr(parsers ...Parser) Parser {
	return Or(parsers...)
}

//Or returns a parser that matches the first of a given set of parsers. A later parser will not be tried if an earlier match was found.
//The returned parser uses the error message of the last parser verbatim.
func Or(parsers ...Parser) Parser {
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

type exceptParser struct {
	Parser
	except Parser
}

//NewExcept returns a parser that wraps another parser so that it fails if a third, excepted parser would succeed.
//
//Deprecated: Use Except instead.
func NewExcept(parser, except Parser) Parser {
	return Except(parser, except)
}

//Except returns a parser that wraps another parser so that it fails if a third, excepted parser would succeed.
func Except(parser, except Parser) Parser {
	return &exceptParser{Parser: parser, except: except}
}

func (e *exceptParser) Parse(src *reader) (val interface{}, err error) {
	_, err = e.except.Parse(src)
	if err == nil {
		e.except.Unread(src)
		return nil, errExceptionMatched
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
//
//Deprecated: Use Optional instead.
func NewOptional(parser Parser) Parser {
	return Optional(parser)
}

//Optional returns a parser that reads exactly one result according to a given other parser. If it fails, the error is discarded and nil is returned.
func Optional(parser Parser) Parser {
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

type discardLeftParser struct {
	leftParser  Parser
	rightParser Parser
}

//NewDiscardLeft returns a parser that calls two other parsers but only returns the result of the second parser. Both parsers must succeed.
//
//Deprecated: Use DiscardLeft instead.
func NewDiscardLeft(left, right Parser) Parser {
	return DiscardLeft(left, right)
}

//DiscardLeft returns a parser that calls two other parsers but only returns the result of the second parser. Both parsers must succeed.
func DiscardLeft(left, right Parser) Parser {
	return &discardLeftParser{leftParser: left, rightParser: right}
}

func (d *discardLeftParser) Parse(src *reader) (interface{}, error) {
	_, err := d.leftParser.Parse(src)
	if err != nil {
		return nil, err
	}
	val, err := d.rightParser.Parse(src)
	if err != nil {
		d.leftParser.Unread(src)
		return nil, err
	}
	return val, err
}

func (d *discardLeftParser) Unread(src *reader) {
	d.rightParser.Unread(src)
	d.leftParser.Unread(src)
}

func (d *discardLeftParser) Clone() Parser {
	return NewDiscardLeft(d.leftParser.Clone(), d.rightParser.Clone())
}

type discardRightParser struct {
	leftParser  Parser
	rightParser Parser
}

//NewDiscardRight returns a parser that calls two other parsers but only returns the result of the first parser. Both parsers must succeed.
//
//Deprecated: Use DiscardRight instead.
func NewDiscardRight(left, right Parser) Parser {
	return DiscardRight(left, right)
}

//DiscardRight returns a parser that calls two other parsers but only returns the result of the first parser. Both parsers must succeed.
func DiscardRight(left, right Parser) Parser {
	return &discardRightParser{leftParser: left, rightParser: right}
}

func (d *discardRightParser) Parse(src *reader) (interface{}, error) {
	val, err := d.leftParser.Parse(src)
	if err != nil {
		return nil, err
	}
	_, err = d.rightParser.Parse(src)
	if err != nil {
		d.leftParser.Unread(src)
		return nil, err
	}

	return val, nil
}

func (d *discardRightParser) Unread(src *reader) {
	d.rightParser.Unread(src)
	d.leftParser.Unread(src)
}

func (d *discardRightParser) Clone() Parser {
	return NewDiscardRight(d.leftParser.Clone(), d.rightParser.Clone())
}

//NewSplicingSeq returns a parser that works like a Seq but joins slices returned by its subparsers into a single slice.
//
//Deprecated: Use SplicingSeq instead.
func NewSplicingSeq(parsers ...Parser) Parser {
	return SplicingSeq(parsers...)
}

//SplicingSeq returns a parser that works like a Seq but joins slices returned by its subparsers into a single slice.
func SplicingSeq(parsers ...Parser) Parser {
	return NewTransformer(NewSeq(parsers...), splice)
}

func splice(val interface{}) (interface{}, error) {
	results := val.([]interface{})
	values := make([]interface{}, 0, len(results))
	for _, result := range results {
		if resultSlice, ok := result.([]interface{}); ok {
			values = append(values, resultSlice...)
		} else {
			values = append(values, result)
		}
	}
	return values, nil
}

//NewSep returns a parser that parses a sequence of items according to a first parser that are separated by matches of a second parser.
//
//Deprecated: Use Sep instead.
func NewSep(item, separator Parser) Parser {
	return Sep(item, separator)
}

//Sep returns a parser that parses a sequence of items according to a first parser that are separated by matches of a second parser.
func Sep(item, separator Parser) Parser {
	return NewSplicingSeq(item, NewSome(NewDiscardLeft(separator, item)))
}

type recursiveParser struct {
	parser  Parser
	factory func() Parser
}

//NewRecursive allows to recursively define a parser in terms of itself.
//
//Deprecated: Use Recursive instead.
func NewRecursive(factory func() Parser) Parser {
	return Recursive(factory)
}

//Recursive allows to recursively define a parser in terms of itself.
func Recursive(factory func() Parser) Parser {
	return &recursiveParser{factory: factory}
}

func (r *recursiveParser) Parse(src *reader) (interface{}, error) {
	r.parser = r.factory()
	val, err := r.parser.Parse(src)
	if err != nil {
		r.parser.Unread(src)
		return nil, err
	}

	return val, nil
}

func (r *recursiveParser) Unread(src *reader) {
	if r.parser != nil {
		r.parser.Unread(src)
		r.parser = nil
	}
}

func (r *recursiveParser) Clone() Parser {
	return NewRecursive(r.factory)
}
