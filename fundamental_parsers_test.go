package pars

import (
	"fmt"
	"io"
	"testing"
	"unicode"
)

func TestParseRune(t *testing.T) {
	r := stringReader("abc")

	aParser := NewAnyRune()
	a, aErr := aParser.Parse(r)
	assertParse(t, a.(rune), aErr, 'a', nil)

	bParser := NewAnyRune()
	b, bErr := bParser.Parse(r)
	assertParse(t, b.(rune), bErr, 'b', nil)

	cParser := NewAnyRune()
	c, cErr := cParser.Parse(r)
	assertParse(t, c.(rune), cErr, 'c', nil)

	eofParser := NewAnyRune()
	eof, eofErr := eofParser.Parse(r)
	assertParse(t, eof, eofErr, nil, io.EOF)
}

func TestParseUTFRune(t *testing.T) {
	r := byteReader([]byte{97, 0xe2, 0x82, 0xac, 99})

	aParser := NewAnyRune()
	a, aErr := aParser.Parse(r)
	assertParse(t, a.(rune), aErr, 'a', nil)

	bParser := NewAnyRune()
	b, bErr := bParser.Parse(r)
	assertParse(t, b.(rune), bErr, '€', nil)

	cParser := bParser.Clone()
	c, cErr := cParser.Parse(r)
	assertParse(t, c.(rune), cErr, 'c', nil)

	eofParser := NewAnyRune()
	eof, eofErr := eofParser.Parse(r)
	assertParse(t, eof, eofErr, nil, io.EOF)
}

func TestPartOfRune(t *testing.T) {
	r := byteReader([]byte{0xe2, 0x82})

	parser := NewAnyRune()
	val, err := parser.Parse(r)
	assertParse(t, val, err, nil, io.EOF)

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{0x82, 0xe2})
}

func TestExpectedRune(t *testing.T) {
	r := byteReader([]byte{0xf5, 0xbf, 0xbf, 0xbf})

	parser := NewAnyRune()
	val, err := parser.Parse(r)
	assertParse(t, val, err, nil, errRuneExpected)

	assertBytes(t, r.buf.current, []byte{0xbf, 0xbf, 0xbf})
	assertBytes(t, r.buf.prepend, []byte{0xf5})
}

func TestParseAnyByte(t *testing.T) {
	r := byteReader([]byte{1, 2, 3})

	parser1 := NewAnyByte()
	val1, err1 := parser1.Parse(r)
	assertParse(t, val1.(byte), err1, byte(1), nil)

	parser2 := NewAnyByte()
	val2, err2 := parser2.Parse(r)
	assertParse(t, val2.(byte), err2, byte(2), nil)

	parser3 := parser2.Clone()
	val3, err3 := parser3.Parse(r)
	assertParse(t, val3.(byte), err3, byte(3), nil)

	parserEOF := NewAnyByte()
	valEOF, errEOF := parserEOF.Parse(r)
	assertParse(t, valEOF, errEOF, nil, io.EOF)

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{})
}

func TestParseAnyByteUnread(t *testing.T) {
	r := stringReader("AB")

	val, err := NewOr(NewSeq(NewAnyByte(), NewChar('C')), NewSeq(NewAnyByte(), NewChar('B'))).Parse(r)
	assertParseSlice(t, val, err, []interface{}{byte('A'), 'B'}, nil)
}

func TestParseExpectedChars(t *testing.T) {
	r := stringReader("aba")

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
	r := stringReader("a")

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
	r := stringReader(" \t")

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
	r := stringReader("a")

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

func TestParseString(t *testing.T) {
	r := stringReader("abcabc")

	abcParser := NewString("abc")
	abcVal1, abcErr1 := abcParser.Parse(r)
	assertParse(t, abcVal1, abcErr1, "abc", nil)

	abcParser2 := abcParser.Clone()
	abcVal2, abcErr2 := abcParser2.Parse(r)
	assertParse(t, abcVal2, abcErr2, "abc", nil)

	parserEOF := abcParser.Clone()
	valEOF, errEOF := parserEOF.Parse(r)
	assertParse(t, valEOF, errEOF, nil, fmt.Errorf("Could not parse expected string \"abc\": EOF"))

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{})
}

func TestParseUnexpectedString(t *testing.T) {
	r := stringReader("abc")

	abdParser := NewString("abd")
	abdVal, abdErr := abdParser.Parse(r)
	assertParse(t, abdVal, abdErr, nil, fmt.Errorf("Could not parse expected string \"abd\": Unexpected string \"abc\""))

	abcParser := NewString("abc")
	abcVal1, abcErr1 := abcParser.Parse(r)
	assertParse(t, abcVal1, abcErr1, "abc", nil)

	parserEOF := abcParser.Clone()
	valEOF, errEOF := parserEOF.Parse(r)
	assertParse(t, valEOF, errEOF, nil, fmt.Errorf("Could not parse expected string \"abc\": EOF"))

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{})
}

func TestParseStringWrongCase(t *testing.T) {
	r := stringReader("ABC")
	val, err := NewString("abc").Parse(r)
	assertParse(t, val, err, nil, fmt.Errorf("Could not parse expected string \"abc\": Unexpected string \"ABC\""))
}

func TestParseStringCI(t *testing.T) {
	r := stringReader("abcabc")

	abcParser := NewStringCI("abc")
	abcVal1, abcErr1 := abcParser.Parse(r)
	assertParse(t, abcVal1, abcErr1, "abc", nil)

	abcParser2 := abcParser.Clone()
	abcVal2, abcErr2 := abcParser2.Parse(r)
	assertParse(t, abcVal2, abcErr2, "abc", nil)

	parserEOF := abcParser.Clone()
	valEOF, errEOF := parserEOF.Parse(r)
	assertParse(t, valEOF, errEOF, nil, fmt.Errorf("Could not parse expected string \"abc\": EOF"))

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{})
}

func TestParseUnexpectedStringCI(t *testing.T) {
	r := stringReader("abc")

	abdParser := NewStringCI("abd")
	abdVal, abdErr := abdParser.Parse(r)
	assertParse(t, abdVal, abdErr, nil, fmt.Errorf("Could not parse expected string \"abd\": Unexpected string \"abc\""))

	abcParser := NewStringCI("abc")
	abcVal1, abcErr1 := abcParser.Parse(r)
	assertParse(t, abcVal1, abcErr1, "abc", nil)

	parserEOF := abcParser.Clone()
	valEOF, errEOF := parserEOF.Parse(r)
	assertParse(t, valEOF, errEOF, nil, fmt.Errorf("Could not parse expected string \"abc\": EOF"))

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{})
}

func TestParseStringCIWrongCase(t *testing.T) {
	r := stringReader("ABC")
	val, err := NewStringCI("abc").Parse(r)
	assertParse(t, val, err, "ABC", nil)
}

func TestParseStringCIUnread(t *testing.T) {
	r := stringReader("ABCD")
	val, err := NewOr(NewDiscardRight(NewStringCI("abc"), EOF), NewDiscardRight(NewString("ABCD"), EOF)).Parse(r)
	assertParse(t, val, err, "ABCD", nil)
}

func TestParseEOF(t *testing.T) {
	r := stringReader("a")

	val1, err1 := EOF.Parse(r)
	assertParse(t, val1, err1, nil, fmt.Errorf("Expected EOF: Found byte 0x61"))

	aVal, aErr := NewChar('a').Parse(r)
	assertParse(t, aVal, aErr, 'a', nil)

	val2, err2 := EOF.Parse(r)
	assertParse(t, val2, err2, nil, nil)

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{})
}

func TestError(t *testing.T) {
	r := stringReader("a")

	val, err := NewError(fmt.Errorf("Expected kanji")).Parse(r)
	assertParse(t, val, err, nil, fmt.Errorf("Expected kanji"))

	assertBytes(t, r.buf.current, []byte{})
	assertBytes(t, r.buf.prepend, []byte{})

	a, aErr := NewChar('a').Parse(r)
	assertParse(t, a, aErr, 'a', nil)
}

func TestParseInt(t *testing.T) {
	r := stringReader("123a-456")

	intParser := NewInt()
	val, err := intParser.Parse(r)
	assertParse(t, val, err, 123, nil)

	aVal, aErr := intParser.Clone().Parse(r)
	assertParse(t, aVal, aErr, nil, fmt.Errorf("Could not parse int: Could not parse expected rune: Rune 'a' (0x61) does not hold predicate"))

	aVal, aErr = NewChar('a').Parse(r)
	assertParse(t, aVal, aErr, 'a', nil)

	val, err = intParser.Clone().Parse(r)
	assertParse(t, val, err, -456, nil)

	eofVal, eofErr := NewInt().Parse(r)
	assertParse(t, eofVal, eofErr, nil, fmt.Errorf("Could not parse int: Could not parse expected rune: EOF"))
}

func TestParseIntTooHuge(t *testing.T) {
	tooLong := "12345678901234567890123456789012345678901234567890"
	r := stringReader(tooLong)

	val, err := NewInt().Parse(r)
	assertParse(t, val, err, nil, fmt.Errorf("Could not parse int: strconv.Atoi: parsing \"%v\": value out of range", tooLong))

	str, err := NewString(tooLong).Parse(r)
	assertParse(t, str, err, tooLong, nil)
}

func TestParseIntMisplacedMinus(t *testing.T) {
	r := stringReader("123-456")

	val, err := NewInt().Parse(r)
	assertParse(t, val, err, 123, nil)

	val, err = NewInt().Parse(r)
	assertParse(t, val, err, -456, nil)
}

func TestParseIntOnlyMinusError(t *testing.T) {
	r := stringReader("--789")

	val, err := NewInt().Parse(r)
	assertParse(t, val, err, nil, fmt.Errorf("Could not parse int: strconv.Atoi: parsing \"-\": invalid syntax"))

	val, err = NewChar('-').Parse(r)
	assertParse(t, val, err, '-', nil)

	val, err = NewInt().Parse(r)
	assertParse(t, val, err, -789, nil)
}

func TestParseBigInt(t *testing.T) {
	tooLong := "12345678901234567890123456789012345678901234567890"

	r := stringReader(tooLong)

	val, err := NewBigInt().Parse(r)
	assertParseBigInt(t, val, err, tooLong, nil)
}

func TestParseBigIntNoCharRead(t *testing.T) {
	r := stringReader("X")
	val, err := NewBigInt().Parse(r)
	assertParseBigInt(t, val, err, nil, fmt.Errorf("Could not parse int: Could not parse expected rune: Rune 'X' (0x58) does not hold predicate"))
}

func TestParseBigIntMisplacedMinus(t *testing.T) {
	r := stringReader("123-456")

	val, err := NewBigInt().Parse(r)
	assertParseBigInt(t, val, err, "123", nil)

	val, err = NewBigInt().Parse(r)
	assertParseBigInt(t, val, err, "-456", nil)
}

func TestParseBigIntOnlyMinusError(t *testing.T) {
	r := stringReader("--789")

	val, err := NewBigInt().Parse(r)
	assertParseBigInt(t, val, err, nil, fmt.Errorf("Could not parse '-' as int"))

	val, err = NewChar('-').Parse(r)
	assertParse(t, val, err, '-', nil)

	val, err = NewBigInt().Parse(r)
	assertParseBigInt(t, val, err, "-789", nil)
}

func TestParseDelimitedString(t *testing.T) {
	r := stringReader("'abc'")
	val, err := NewDelimitedString("'", "'").Parse(r)
	assertParse(t, val, err, "abc", nil)
}

func TestParseDelimitedStringDifferentDelimiters(t *testing.T) {
	r := stringReader("[abc[]")
	val, err := NewDelimitedString("[", "]").Parse(r)
	assertParse(t, val, err, "abc[", nil)
}

func TestParseDelimitedStringMissingEndingDelimiter(t *testing.T) {
	r := stringReader("'abc")
	val, err := NewDelimitedString("'", "'").Parse(r)
	assertParse(t, val, err, nil, fmt.Errorf("Could not parse expected string \"'\": EOF"))
}

func TestParseFloat(t *testing.T) {
	r := stringReader("1.23")
	val, err := NewFloat().Parse(r)
	assertParse(t, val, err, 1.23, nil)
}

func TestParseFloatNegative(t *testing.T) {
	r := stringReader("-1.23")
	val, err := NewFloat().Parse(r)
	assertParse(t, val, err, -1.23, nil)
}

func TestParseFloatPointless(t *testing.T) {
	r := stringReader("123")
	val, err := NewFloat().Parse(r)
	assertParse(t, val, err, 123.0, nil)
}

func TestParseFloatSecondDecimalPoint(t *testing.T) {
	r := stringReader("1.2.3")
	val, err := NewFloat().Parse(r)
	assertParse(t, val, err, 1.2, nil)

	val2, err2 := NewDiscardLeft(NewChar('.'), NewFloat()).Parse(r)
	assertParse(t, val2, err2, 3.0, nil)
}

func TestParseFloatFailsEmptyString(t *testing.T) {
	r := stringReader("")
	val, err := NewFloat().Parse(r)
	assertParse(t, val, err, nil, fmt.Errorf("Could not parse float: Could not parse expected rune: EOF"))
}

func TestParseFloatFails(t *testing.T) {
	r := stringReader("-.")
	val, err := NewFloat().Parse(r)
	assertParse(t, val, err, nil, fmt.Errorf("Could not parse float: strconv.ParseFloat: parsing \"-.\": invalid syntax"))
}

func TestParseFloatUnread(t *testing.T) {
	r := stringReader("-1.23a")
	val, err := NewOr(NewSeq(NewFloat(), NewChar('b')), NewString("-1.23a")).Parse(r)
	assertParse(t, val, err, "-1.23a", nil)
}

func TestParseByte(t *testing.T) {
	r := byteReader([]byte{0, 1, 2, 3})
	val, err := NewByte(0).Parse(r)
	assertParse(t, val, err, byte(0), nil)
}

func TestParseByteOtherByte(t *testing.T) {
	r := byteReader([]byte{0, 1, 2, 3})
	val, err := NewByte(1).Parse(r)
	assertParse(t, val, err, nil, fmt.Errorf("Could not parse expected byte '1': Unexpected byte '0'"))
}
