package pars

type seqParser struct {
	parsers []Parser
}

//Seq returns a parser that matches all of its given parsers in order or none of them.
func Seq(parsers ...Parser) Parser {
	return &seqParser{parsers: parsers}
}

func (s *seqParser) Parse(src *Reader) (interface{}, error) {
	values := make([]interface{}, len(s.parsers))
	for i, parser := range s.parsers {
		val, err := parser.Parse(src)
		if err != nil {
			unreadParsers(s.parsers[:i], src)
			return nil, seqError{index: i, innerError: err}
		}
		values[i] = val
	}
	return values, nil
}

func (s *seqParser) Unread(src *Reader) {
	unreadParsers(s.parsers, src)
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

//Some returns a parser that matches a given parser zero or more times. Not matching at all is not an error.
func Some(parser Parser) Parser {
	return &someParser{prototype: parser}
}

func (s *someParser) Parse(src *Reader) (interface{}, error) {
	var values []interface{}
	for {
		next := s.prototype.Clone()
		s.used = append(s.used, next)

		nextVal, nextErr := next.Parse(src)
		if nextErr != nil {
			s.used = s.used[:len(s.used)-1]
			break
		}
		values = append(values, nextVal)
	}
	return values, nil
}

func (s *someParser) Unread(src *Reader) {
	unreadParsers(s.used, src)
	s.used = nil
}

func (s *someParser) Clone() Parser {
	return &someParser{prototype: s.prototype.Clone()}
}

//Many returns a parser that matches a given parser one or more times. Not matching at all is an error.
func Many(parser Parser) Parser {
	return SplicingSeq(parser, Some(parser))
}

type orParser struct {
	parsers  []Parser
	selected Parser
}

//Or returns a parser that matches the first of a given set of parsers. A later parser will not be tried if an earlier match was found.
//The returned parser uses the error message of the last parser verbatim.
func Or(parsers ...Parser) Parser {
	return &orParser{parsers: parsers}
}

func (o *orParser) Parse(src *Reader) (val interface{}, err error) {
	for _, parser := range o.parsers {
		val, err = parser.Parse(src)
		if err == nil {
			o.selected = parser
			return
		}
	}
	return
}

func (o *orParser) Unread(src *Reader) {
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

//Except returns a parser that wraps another parser so that it fails if a third, excepted parser would succeed.
func Except(parser, except Parser) Parser {
	return &exceptParser{Parser: parser, except: except}
}

func (e *exceptParser) Parse(src *Reader) (val interface{}, err error) {
	_, err = e.except.Parse(src)
	if err == nil {
		e.except.Unread(src)
		return nil, errExceptionMatched
	}
	val, err = e.Parser.Parse(src)
	return
}

func (e *exceptParser) Clone() Parser {
	return Except(e.Parser.Clone(), e.except.Clone())
}

type optionalParser struct {
	read bool
	Parser
}

//Optional returns a parser that reads exactly one result according to a given other parser. If it fails, the error is discarded and nil is returned.
func Optional(parser Parser) Parser {
	return &optionalParser{Parser: parser}
}

func (o *optionalParser) Parse(src *Reader) (interface{}, error) {
	val, err := o.Parser.Parse(src)
	if err == nil {
		o.read = true
		return val, nil
	}
	return nil, nil
}

func (o *optionalParser) Unread(src *Reader) {
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

//DiscardLeft returns a parser that calls two other parsers but only returns the result of the second parser. Both parsers must succeed.
func DiscardLeft(left, right Parser) Parser {
	return &discardLeftParser{leftParser: left, rightParser: right}
}

func (d *discardLeftParser) Parse(src *Reader) (interface{}, error) {
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

func (d *discardLeftParser) Unread(src *Reader) {
	d.rightParser.Unread(src)
	d.leftParser.Unread(src)
}

func (d *discardLeftParser) Clone() Parser {
	return DiscardLeft(d.leftParser.Clone(), d.rightParser.Clone())
}

type discardRightParser struct {
	leftParser  Parser
	rightParser Parser
}

//DiscardRight returns a parser that calls two other parsers but only returns the result of the first parser. Both parsers must succeed.
func DiscardRight(left, right Parser) Parser {
	return &discardRightParser{leftParser: left, rightParser: right}
}

func (d *discardRightParser) Parse(src *Reader) (interface{}, error) {
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

func (d *discardRightParser) Unread(src *Reader) {
	d.rightParser.Unread(src)
	d.leftParser.Unread(src)
}

func (d *discardRightParser) Clone() Parser {
	return DiscardRight(d.leftParser.Clone(), d.rightParser.Clone())
}

//SplicingSeq returns a parser that works like a Seq but joins slices returned by its subparsers into a single slice.
func SplicingSeq(parsers ...Parser) Parser {
	return Transformer(Seq(parsers...), splice)
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

//Sep returns a parser that parses a sequence of items according to a first parser that are separated by matches of a second parser.
func Sep(item, separator Parser) Parser {
	return SplicingSeq(item, Some(DiscardLeft(separator, item)))
}

type recursiveParser struct {
	parser  Parser
	factory func() Parser
}

//Recursive allows to recursively define a parser in terms of itself.
func Recursive(factory func() Parser) Parser {
	return &recursiveParser{factory: factory}
}

func (r *recursiveParser) Parse(src *Reader) (interface{}, error) {
	r.parser = r.factory()
	val, err := r.parser.Parse(src)
	if err != nil {
		r.parser.Unread(src)
		return nil, err
	}

	return val, nil
}

func (r *recursiveParser) Unread(src *Reader) {
	if r.parser != nil {
		r.parser.Unread(src)
		r.parser = nil
	}
}

func (r *recursiveParser) Clone() Parser {
	return Recursive(r.factory)
}
