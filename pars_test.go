package pars

import (
	"fmt"
	"io"
	"testing"
	"unicode"
)

func TestParseRune(t *testing.T) {
	r := StringReader("abc")

	aParser := AnyRune{}
	a, aErr := aParser.Parse(r)
	assertParse(t, a.(rune), aErr, 'a', nil)

	bParser := AnyRune{}
	b, bErr := bParser.Parse(r)
	assertParse(t, b.(rune), bErr, 'b', nil)

	cParser := AnyRune{}
	c, cErr := cParser.Parse(r)
	assertParse(t, c.(rune), cErr, 'c', nil)

	eofParser := AnyRune{}
	eof, eofErr := eofParser.Parse(r)
	assertParse(t, eof, eofErr, nil, io.EOF)
}

func TestParseUTFRune(t *testing.T) {
	r := ByteReader([]byte{97, 0xe2, 0x82, 0xac, 99})

	aParser := AnyRune{}
	a, aErr := aParser.Parse(r)
	assertParse(t, a.(rune), aErr, 'a', nil)

	bParser := AnyRune{}
	b, bErr := bParser.Parse(r)
	assertParse(t, b.(rune), bErr, '€', nil)

	cParser := bParser.Clone()
	c, cErr := cParser.Parse(r)
	assertParse(t, c.(rune), cErr, 'c', nil)

	eofParser := AnyRune{}
	eof, eofErr := eofParser.Parse(r)
	assertParse(t, eof, eofErr, nil, io.EOF)
}

func TestPartOfRune(t *testing.T) {
	r := ByteReader([]byte{0xe2, 0x82})

	parser := AnyRune{}
	val, err := parser.Parse(r)
	assertParse(t, val, err, nil, io.EOF)

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{0x82, 0xe2})
}

func TestExpectedRune(t *testing.T) {
	r := ByteReader([]byte{0xf5, 0xbf, 0xbf, 0xbf})

	parser := AnyRune{}
	val, err := parser.Parse(r)
	assertParse(t, val, err, nil, ErrRuneExpected)

	assertBytes(t, r.buf.current, []byte{0xbf, 0xbf, 0xbf})
	assertBytes(t, r.buf.prepend, []byte{0xf5})
}

func TestParseByte(t *testing.T) {
	r := ByteReader([]byte{1, 2, 3})

	parser1 := AnyByte{}
	val1, err1 := parser1.Parse(r)
	assertParse(t, val1.(byte), err1, byte(1), nil)

	parser2 := AnyByte{}
	val2, err2 := parser2.Parse(r)
	assertParse(t, val2.(byte), err2, byte(2), nil)

	parser3 := parser2.Clone()
	val3, err3 := parser3.Parse(r)
	assertParse(t, val3.(byte), err3, byte(3), nil)

	parserEOF := AnyByte{}
	valEOF, errEOF := parserEOF.Parse(r)
	assertParse(t, valEOF, errEOF, nil, io.EOF)

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{})
}

func TestParseExpectedChars(t *testing.T) {
	r := StringReader("aba")

	aParser := NewChar('a')
	aVal, aErr := aParser.Parse(r)
	assertParse(t, aVal.(rune), aErr, 'a', nil)

	bParser := NewChar('b')
	bVal, bErr := bParser.Parse(r)
	assertParse(t, bVal.(rune), bErr, 'b', nil)

	aParser2 := aParser.Clone()
	aVal2, aErr2 := aParser2.Parse(r)
	assertParse(t, aVal2.(rune), aErr2, 'a', nil)

	parserEOF := aParser.Clone()
	valEOF, errEOF := parserEOF.Parse(r)
	assertParse(t, valEOF, errEOF, nil, fmt.Errorf("Could not parse expected rune 'a' (0x61): EOF"))

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{})
}

func TestParseUnexpectedChars(t *testing.T) {
	r := StringReader("a")

	bParser := NewChar('€')
	bVal, bErr := bParser.Parse(r)
	assertParse(t, bVal, bErr, nil, fmt.Errorf("Could not parse expected rune '€' (0x20ac): Unexpected rune 'a' (0x61)"))

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{97})

	aParser := NewChar('a')
	aVal, aErr := aParser.Parse(r)
	assertParse(t, aVal.(rune), aErr, 'a', nil)

	parserEOF := aParser.Clone()
	valEOF, errEOF := parserEOF.Parse(r)
	assertParse(t, valEOF, errEOF, nil, fmt.Errorf("Could not parse expected rune 'a' (0x61): EOF"))

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{})
}

func TestParseExpectedCharPred(t *testing.T) {
	r := StringReader(" \t")

	spaceParser := NewCharPred(unicode.IsSpace)
	spaceVal, spaceErr := spaceParser.Parse(r)
	assertParse(t, spaceVal.(rune), spaceErr, ' ', nil)

	tabParser := spaceParser.Clone()
	tabVal, tabErr := tabParser.Parse(r)
	assertParse(t, tabVal.(rune), tabErr, '\t', nil)

	parserEOF := spaceParser.Clone()
	valEOF, errEOF := parserEOF.Parse(r)
	assertParse(t, valEOF, errEOF, nil, fmt.Errorf("Could not parse expected rune: EOF"))

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{})
}

func TestParseUnexpectedCharPred(t *testing.T) {
	r := StringReader("a")

	spaceParser := NewCharPred(unicode.IsSpace)
	spaceVal, spaceErr := spaceParser.Parse(r)
	assertParse(t, spaceVal, spaceErr, nil, fmt.Errorf("Could not parse expected rune: Rune 'a' (0x61) does not hold predicate"))

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{97})

	aParser := NewCharPred(unicode.IsLetter)
	aVal, aErr := aParser.Parse(r)
	assertParse(t, aVal.(rune), aErr, 'a', nil)

	parserEOF := aParser.Clone()
	valEOF, errEOF := parserEOF.Parse(r)
	assertParse(t, valEOF, errEOF, nil, fmt.Errorf("Could not parse expected rune: EOF"))

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{})
}

func TestParseSeq(t *testing.T) {
	r := StringReader("a€ca€c")

	seqParser := NewSeq(NewChar('a'), NewChar('€'), NewChar('c'))
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

func TestParseSeqFailed(t *testing.T) {
	r := StringReader("a€d")

	seqParser := NewSeq(NewChar('a'), NewChar('€'), NewChar('c'))
	seqVal, seqErr := seqParser.Parse(r)
	assertParse(t, seqVal, seqErr, nil, fmt.Errorf("Could not find expected sequence item 2: Could not parse expected rune 'c' (0x63): Unexpected rune 'd' (0x64)"))

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{0x64, 0xac, 0x82, 0xe2, 0x61})

	seqParser2 := NewSeq(NewChar('a'), NewChar('€'), NewChar('d'))
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
	r := StringReader("aba")

	orParserA := NewOr(NewChar('a'), NewChar('b'))
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
	r := StringReader("c")

	orParserA := NewOr(NewChar('a'), NewChar('b'))
	aVal, aErr := orParserA.Parse(r)
	assertParse(t, aVal, aErr, nil, fmt.Errorf("Could not parse expected rune 'b' (0x62): Unexpected rune 'c' (0x63)"))

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{0x63})

	orParserC := NewOr(NewChar('c'))
	cVal, cErr := orParserC.Parse(r)
	assertParse(t, cVal, cErr, 'c', nil)

	parserEOF := orParserA.Clone()
	valEOF, errEOF := parserEOF.Parse(r)
	assertParse(t, valEOF, errEOF, nil, fmt.Errorf("Could not parse expected rune 'b' (0x62): EOF"))

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{})
}
