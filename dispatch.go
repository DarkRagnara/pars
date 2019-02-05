package pars

type dispatchParser struct {
	clauses []DispatchClause
	parsers []Parser
}

//Dispatch returns a parser that is like a combination of Seq and Or with limited backtracking.
//
//A Dispatch contains multiple clauses consisting of parsers. Dispatch parses by trying the clauses one by one.
//The first matching clause is used, later clauses are not tried. Each clause can contain multiple parsers.
//Clauses are special because they limit the backtracking: If the first parser of a clause matches, that clause
//is selected even if a later parser of that clause fails.
//
//The motivation for limited backtracking is in better error reporting. When an Or parser fails, all you know is that
//not a single parser succeeded. When a Dispatch parser fails after a clause was selected, you know which subclause
//was supposed to be parsed and can return a fitting error message.
func Dispatch(clauses ...DispatchClause) Parser {
	return &dispatchParser{clauses: clauses}
}

func (d *dispatchParser) Parse(src *Reader) (interface{}, error) {
	for _, clause := range d.clauses {
		parsers := clause.Parsers()
		if len(parsers) == 0 {
			continue
		}

		val, selected, err := d.tryParse(src, parsers)
		if selected {
			if err != nil {
				return nil, clause.TransformError(err)
			}
			return clause.TransformResult(val), nil
		}
	}
	return nil, dispatchWithoutMatch{}
}

func (d *dispatchParser) tryParse(src *Reader, parsers []Parser) ([]interface{}, bool, error) {
	val, err := parsers[0].Parse(src)
	if err != nil {
		return nil, false, err
	}

	vals := make([]interface{}, len(parsers))
	vals[0] = val

	for i, parser := range parsers {
		if i == 0 {
			continue
		}

		vals[i], err = parser.Parse(src)
		if err != nil {
			unreadParsers(parsers[:i], src)
			return nil, true, err
		}
	}

	d.parsers = parsers
	return vals, true, nil
}

func (d *dispatchParser) Unread(src *Reader) {
	unreadParsers(d.parsers, src)
	d.parsers = nil
}

func (d *dispatchParser) Clone() Parser {
	return &dispatchParser{clauses: d.clauses}
}

//DispatchClause is the interface of a clause used by Dispatch.
type DispatchClause interface {
	//Parsers returns the parsers of the clause.
	Parsers() []Parser
	//TransformResult allows the DispatchClause to combine the results of its parsers to a single result.
	TransformResult([]interface{}) interface{}
	//TransformError allows the DispatchClause to replace or extend the error returned on a failed match.
	TransformError(error) error
}

//Clause is the most simple DispatchClause. It is just a slice of parsers without any transformations.
type Clause []Parser

var _ DispatchClause = Clause{}

//Parsers returns the parser slice for this clause.
func (c Clause) Parsers() []Parser {
	return c
}

//TransformResult returns the slice of values unchanged.
func (c Clause) TransformResult(val []interface{}) interface{} {
	return val
}

//TransformError returns the given error unchanged.
func (c Clause) TransformError(err error) error {
	return err
}

//DescribeClause extends the error message of a clause so that a custom description is part of the message.
type DescribeClause struct {
	DispatchClause
	description string
}

//TransformError extends the error message of a clause so that a custom description is part of the message.
func (d DescribeClause) TransformError(err error) error {
	return describeClauseError{description: d.description, innerError: err}
}

//StringJoiningClause extends a clause that consists of parsers that return runes or strings so that it
//returnes a single string instead. Slices are handled recursively.
//StringJoiningClause WILL PANIC if any of the parsers return something other than a rune or a string or a
//slice of these types.
type StringJoiningClause struct {
	DispatchClause
}

//TransformResult joins runes and strings together like JoinString.
func (s StringJoiningClause) TransformResult(vals []interface{}) interface{} {
	val, _ := joinToString(vals)
	return val
}
