package pars

import (
	"fmt"
	"testing"
)

func TestEmptyDispatchFails(t *testing.T) {
	r := stringReader("")
	val, err := Dispatch().Parse(r)
	assertParse(t, val, err, nil, dispatchWithoutMatch{})
}

func TestNoMatchingClause(t *testing.T) {
	r := stringReader("a")
	val, err := Dispatch(Clause{Char('b')}, Clause{Char('c')}).Parse(r)
	assertParse(t, val, err, nil, dispatchWithoutMatch{})
}

func TestSimpleMatchingClause(t *testing.T) {
	r := stringReader("a")
	val, err := Dispatch(Clause{Char('b')}, Clause{Char('a')}).Parse(r)
	assertParse(t, nil, err, nil, nil)
	assertRunesInSlice(t, val.([]interface{}), "a")
}

func TestMultiParserMatchingClause(t *testing.T) {
	r := stringReader("aAa")
	val, err := Dispatch(Clause{Char('b')}, Clause{Char('a'), Char('A'), Char('a')}).Parse(r)
	assertParse(t, nil, err, nil, nil)
	assertRunesInSlice(t, val.([]interface{}), "aAa")
}

func TestStringJoiningClause(t *testing.T) {
	r := stringReader("aAa")
	val, err := Dispatch(StringJoiningClause{Clause{Char('a'), Char('A'), Char('a')}}).Parse(r)
	assertParse(t, val, err, "aAa", nil)
}

func TestClauseSelection(t *testing.T) {
	r := stringReader("aAa")
	val, err := Dispatch(Clause{Char('a'), Char('b')}, Clause{Char('a'), Char('A'), Char('a')}).Parse(r)
	assertParse(t, val, err, nil, fmt.Errorf("Could not parse expected rune 'b' (0x62): Unexpected rune 'A' (0x41)"))
}

func TestErrorTransformingClause(t *testing.T) {
	r := stringReader("aAa")
	val, err := Dispatch(DescribeClause{DispatchClause: Clause{Char('a'), Char('b')}, description: "ab"}).Parse(r)
	assertParse(t, val, err, nil, fmt.Errorf("ab expected: Could not parse expected rune 'b' (0x62): Unexpected rune 'A' (0x41)"))
}

func TestDispatchUnreadClause(t *testing.T) {
	r := stringReader("aAa")
	val, err := Dispatch(Clause{Char('a'), Char('A'), Char('a'), Char('A')}).Parse(r)
	assertParse(t, val, err, nil, fmt.Errorf("Could not parse expected rune 'A' (0x41): EOF"))

	val, err = String("aAa").Parse(r)
	assertParse(t, val, err, "aAa", nil)
}

func TestDispatchUnread(t *testing.T) {
	r := stringReader("aAa")
	val, err := Or(DiscardRight(Dispatch(Clause{Char('a'), Char('A')}), Char('b')), String("aAa")).Parse(r)
	assertParse(t, val, err, "aAa", nil)
}
