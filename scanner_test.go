package pars

import (
	"fmt"
	"testing"
	"unicode"
)

func TestScanner(t *testing.T) {
	r := stringReader("1 2 3 4 5")
	parser := SwallowTrailingWhitespace(CharPred(unicode.IsDigit))

	expected := []rune{'1', '2', '3', '4', '5'}

	s := NewScanner(r, parser)

	for s.Scan() {
		assertValue(t, s.Result(), expected[0])
		expected = expected[1:]
	}

	assertValue(t, len(expected), 0)
	assertError(t, s.Err(), nil)
}

func TestScannerFail(t *testing.T) {
	r := stringReader("ab")
	parser := Char('a')

	s := NewScanner(r, parser)
	assertValue(t, s.Scan(), true)
	assertValue(t, s.Result(), 'a')
	assertError(t, s.Err(), nil)

	assertValue(t, s.Scan(), false)
	assertValue(t, s.Result(), nil)
	assertError(t, s.Err(), fmt.Errorf("Could not parse expected rune 'a' (0x61): Unexpected rune 'b' (0x62)"))

	assertValue(t, s.Scan(), false)
	assertValue(t, s.Result(), nil)
	assertError(t, s.Err(), fmt.Errorf("Could not parse expected rune 'a' (0x61): Unexpected rune 'b' (0x62)"))
}
