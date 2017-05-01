package pars

import (
	"io"
	"testing"
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
