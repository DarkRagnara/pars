package pars

import (
	"testing"
)

func TestParseStringFunc(t *testing.T) {
	val, err := ParseString("abc", NewString("ab"))
	assertParse(t, val, err, "ab", nil)
}

func TestParseFromReaderFunc(t *testing.T) {
	val, err := ParseFromReader(stringReader("abc"), NewString("ab"))
	assertParse(t, val, err, "ab", nil)
}
