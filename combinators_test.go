package pars

import (
	"fmt"
	"testing"
)

func TestParseSeq(t *testing.T) {
	r := stringReader("a€ca€c")

	seqParser := Seq(Char('a'), Char('€'), Char('c'))
	seqVal, seqErr := seqParser.Parse(r)
	assertParse(t, nil, seqErr, nil, nil)
	assertRunesInSlice(t, seqVal.([]interface{}), "a€c")

	seqParser2 := seqParser.Clone()
	seqVal2, seqErr2 := seqParser2.Parse(r)
	assertParse(t, nil, seqErr2, nil, nil)
	assertRunesInSlice(t, seqVal2.([]interface{}), "a€c")

	parserEOF := seqParser.Clone()
	valEOF, errEOF := parserEOF.Parse(r)
	assertParse(t, valEOF, errEOF, nil, fmt.Errorf("Could not find expected sequence item 0: Could not parse expected rune 'a' (0x61): EOF"))

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{})
}

func TestParseSeqUnread(t *testing.T) {
	r := stringReader("abc")
	seqParser := Or(Seq(Seq(Char('a'), Char('b')), Char('a')), String("abc"))

	val1, err1 := seqParser.Parse(r)
	assertParse(t, val1, err1, "abc", nil)
}

func TestParseSeqFailed(t *testing.T) {
	r := stringReader("a€d")

	seqParser := Seq(Char('a'), Char('€'), Char('c'))
	seqVal, seqErr := seqParser.Parse(r)
	assertParse(t, seqVal, seqErr, nil, fmt.Errorf("Could not find expected sequence item 2: Could not parse expected rune 'c' (0x63): Unexpected rune 'd' (0x64)"))

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{0x64, 0xac, 0x82, 0xe2, 0x61})

	seqParser2 := Seq(Char('a'), Char('€'), Char('d'))
	seqVal2, seqErr2 := seqParser2.Parse(r)
	assertParse(t, nil, seqErr2, nil, nil)
	assertRunesInSlice(t, seqVal2.([]interface{}), "a€d")

	parserEOF := seqParser.Clone()
	valEOF, errEOF := parserEOF.Parse(r)
	assertParse(t, valEOF, errEOF, nil, fmt.Errorf("Could not find expected sequence item 0: Could not parse expected rune 'a' (0x61): EOF"))

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{})
}

func TestParseOr(t *testing.T) {
	r := stringReader("aba")

	orParserA := Or(Char('a'), Char('b'))
	aVal, aErr := orParserA.Parse(r)
	assertParse(t, aVal, aErr, 'a', nil)

	orParserB := orParserA.Clone()
	bVal, bErr := orParserB.Parse(r)
	assertParse(t, bVal, bErr, 'b', nil)

	orParserA2 := orParserA.Clone()
	aVal2, aErr2 := orParserA2.Parse(r)
	assertParse(t, aVal2, aErr2, 'a', nil)

	parserEOF := orParserA.Clone()
	valEOF, errEOF := parserEOF.Parse(r)
	assertParse(t, valEOF, errEOF, nil, fmt.Errorf("Could not parse expected rune 'b' (0x62): EOF"))

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{})
}

func TestParseOrFailed(t *testing.T) {
	r := stringReader("c")

	orParserA := Or(Char('a'), Char('b'))
	aVal, aErr := orParserA.Parse(r)
	assertParse(t, aVal, aErr, nil, fmt.Errorf("Could not parse expected rune 'b' (0x62): Unexpected rune 'c' (0x63)"))

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{0x63})

	orParserC := Or(Char('c'))
	cVal, cErr := orParserC.Parse(r)
	assertParse(t, cVal, cErr, 'c', nil)

	parserEOF := orParserA.Clone()
	valEOF, errEOF := parserEOF.Parse(r)
	assertParse(t, valEOF, errEOF, nil, fmt.Errorf("Could not parse expected rune 'b' (0x62): EOF"))

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{})
}

func TestParseSomeEmptyString(t *testing.T) {
	r := stringReader("")
	val, err := Some(AnyRune()).Parse(r)

	assertParseSlice(t, val, err, []interface{}{}, nil)
}

func TestParseSomeExactlyOneBeforeEOF(t *testing.T) {
	r := stringReader("x")
	val, err := Some(AnyRune()).Parse(r)
	assertParseSlice(t, val, err, []interface{}{'x'}, nil)
}

func TestParseSomeExactlyOneBeforeMismatch(t *testing.T) {
	r := stringReader("xy")
	val, err := Some(Char('x')).Parse(r)
	assertParseSlice(t, val, err, []interface{}{'x'}, nil)
}

func TestParseSomeMultipleHits(t *testing.T) {
	r := stringReader("xxxxy")
	val, err := Some(Char('x')).Parse(r)
	assertParseSlice(t, val, err, []interface{}{'x', 'x', 'x', 'x'}, nil)
}

func TestParseSomeUnreadOnError(t *testing.T) {
	r := stringReader("xy")
	val, err := Seq(Some(Char('x')), Char('y')).Parse(r)
	assertError(t, err, nil)
	values := val.([]interface{})
	assertValueSlice(t, values[0], []interface{}{'x'})
	assertValue(t, values[1], 'y')
}

func TestParseManyEmptyString(t *testing.T) {
	r := stringReader("")
	val, err := Many(AnyRune()).Parse(r)

	assertParse(t, val, err, nil, fmt.Errorf("Could not find expected sequence item 0: EOF"))
}

func TestParseManyExactlyOneBeforeEOF(t *testing.T) {
	r := stringReader("x")
	val, err := Many(AnyRune()).Parse(r)
	assertParseSlice(t, val, err, []interface{}{'x'}, nil)
}

func TestParseManyExactlyOneBeforeMismatch(t *testing.T) {
	r := stringReader("xy")
	val, err := Many(Char('x')).Parse(r)
	assertParseSlice(t, val, err, []interface{}{'x'}, nil)
}

func TestParseManyMultipleHits(t *testing.T) {
	r := stringReader("xxxxy")
	val, err := Many(Char('x')).Parse(r)
	assertParseSlice(t, val, err, []interface{}{'x', 'x', 'x', 'x'}, nil)
}

func TestParseOptionalFails(t *testing.T) {
	r := stringReader("12345")
	val, err := Seq(Optional(Char('-')), String("12345")).Parse(r)
	assertParseSlice(t, val, err, []interface{}{nil, "12345"}, nil)
}

func TestParseOptionalUnread(t *testing.T) {
	r := stringReader("--")
	val, err := Or(Seq(Optional(Char('-')), Int()), String("--")).Parse(r)
	assertParse(t, val, err, "--", nil)
}

func TestParseOptionalSucceeds(t *testing.T) {
	r := stringReader("-12345")
	val, err := Seq(Optional(Char('-')), String("12345")).Parse(r)
	assertParseSlice(t, val, err, []interface{}{'-', "12345"}, nil)
}

func TestParseExceptRegular(t *testing.T) {
	r := stringReader("x")
	val, err := Except(AnyRune(), Char('y')).Parse(r)
	assertParse(t, val, err, 'x', nil)
}

func TestParseExceptException(t *testing.T) {
	r := stringReader("x")
	val, err := Except(AnyRune(), Char('x')).Parse(r)
	assertParse(t, val, err, nil, errExceptionMatched)
}

func TestParseDiscardLeft(t *testing.T) {
	r := stringReader("$15")
	val, err := DiscardLeft(Char('$'), Int()).Parse(r)
	assertParse(t, val, err, 15, nil)
}

func TestParseDiscardLeftLeftFailed(t *testing.T) {
	r := stringReader("$15")
	val, err := DiscardLeft(Char('€'), Int()).Parse(r)
	assertParse(t, val, err, nil, fmt.Errorf("Could not parse expected rune '€' (0x20ac): Unexpected rune '$' (0x24)"))

	val, err = DiscardLeft(Char('$'), Int()).Parse(r)
	assertParse(t, val, err, 15, nil)
}

func TestParseDiscardLeftRightFailed(t *testing.T) {
	r := stringReader("$15")
	val, err := DiscardLeft(Char('$'), Char('0')).Parse(r)
	assertParse(t, val, err, nil, fmt.Errorf("Could not parse expected rune '0' (0x30): Unexpected rune '1' (0x31)"))

	val, err = DiscardLeft(Char('$'), Int()).Parse(r)
	assertParse(t, val, err, 15, nil)
}

func TestParseDiscardLeftUnread(t *testing.T) {
	r := stringReader("$15")
	val, err := Or(Seq(DiscardLeft(Char('$'), Int()), Error(fmt.Errorf("Forced unread"))), String("$15")).Parse(r)
	assertParse(t, val, err, "$15", nil)
}

func TestParseDiscardRight(t *testing.T) {
	r := stringReader("$15")
	val, err := DiscardRight(Char('$'), Int()).Parse(r)
	assertParse(t, val, err, '$', nil)
}

func TestParseDiscardRightLeftFailed(t *testing.T) {
	r := stringReader("$15")
	val, err := DiscardRight(Char('€'), Int()).Parse(r)
	assertParse(t, val, err, nil, fmt.Errorf("Could not parse expected rune '€' (0x20ac): Unexpected rune '$' (0x24)"))

	val, err = DiscardRight(Char('$'), Int()).Parse(r)
	assertParse(t, val, err, '$', nil)
}

func TestParseDiscardRightRightFailed(t *testing.T) {
	r := stringReader("$15")
	val, err := DiscardRight(Char('$'), Char('0')).Parse(r)
	assertParse(t, val, err, nil, fmt.Errorf("Could not parse expected rune '0' (0x30): Unexpected rune '1' (0x31)"))

	val, err = DiscardRight(Char('$'), Int()).Parse(r)
	assertParse(t, val, err, '$', nil)
}

func TestParseDiscardRightUnread(t *testing.T) {
	r := stringReader("$15")
	val, err := Or(Seq(DiscardRight(Char('$'), Int()), Error(fmt.Errorf("Forced unread"))), String("$15")).Parse(r)
	assertParse(t, val, err, "$15", nil)
}

func TestParseSep(t *testing.T) {
	r := stringReader("1,2,3,4,5")
	val, err := Sep(Int(), Char(',')).Parse(r)
	assertParseSlice(t, val, err, []interface{}{1, 2, 3, 4, 5}, nil)
}

func TestParseSepTrailingSeparator(t *testing.T) {
	r := stringReader("1,2,3,4,5,")
	val, err := Sep(Int(), Char(',')).Parse(r)
	assertParseSlice(t, val, err, []interface{}{1, 2, 3, 4, 5}, nil)

	val2, err2 := Char(',').Parse(r)
	assertParse(t, val2, err2, ',', nil)
}

func TestParseSepSingleItem(t *testing.T) {
	r := stringReader("12345")
	val, err := Sep(Int(), Char(',')).Parse(r)
	assertParseSlice(t, val, err, []interface{}{12345}, nil)
}

func TestParseSepFail(t *testing.T) {
	r := stringReader("a,b,c")
	val, err := Sep(Int(), Char(',')).Parse(r)
	assertParse(t, val, err, nil, fmt.Errorf("Could not find expected sequence item 0: Could not parse int: Could not parse expected rune: Rune 'a' (0x61) does not hold predicate"))
}

func TestRecursiveSimple(t *testing.T) {
	r := stringReader("abc")
	val, err := Recursive(func() Parser { return String("abc") }).Parse(r)
	assertParse(t, val, err, "abc", nil)
}

func TestRecursiveSimpleFail(t *testing.T) {
	r := stringReader("abc")
	val, err := Recursive(func() Parser { return Error(fmt.Errorf("Forced failure")) }).Parse(r)
	assertParse(t, val, err, nil, fmt.Errorf("Forced failure"))
}

func recursiveTestParser() Parser {
	return Or(SplicingSeq(Int(), DiscardLeft(Char(';'), Recursive(recursiveTestParser))), Seq(Int()))
}

func TestRecursiveReal(t *testing.T) {
	r := stringReader("123")
	p := recursiveTestParser()
	val, err := p.Clone().Parse(r)
	assertParseSlice(t, val, err, []interface{}{123}, nil)

	r = stringReader("123;234;345")
	val, err = p.Clone().Parse(r)
	assertParseSlice(t, val, err, []interface{}{123, 234, 345}, nil)
}

func TestRecursiveUnread(t *testing.T) {
	r := stringReader("123;456;789")
	val, err := Or(Seq(recursiveTestParser(), Error(fmt.Errorf("Forced unread"))), String("123;456;789")).Parse(r)
	assertParse(t, val, err, "123;456;789", nil)
}
