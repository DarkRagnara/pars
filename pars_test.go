package pars

import (
	"testing"
)

func TestParseStringFunc(t *testing.T) {
	val, err := ParseString("abc", String("ab"))
	assertParse(t, val, err, "ab", nil)
}

func TestParseFromReaderFunc(t *testing.T) {
	val, err := ParseFromReader(stringReader("abc"), String("ab"))
	assertParse(t, val, err, "ab", nil)
}
