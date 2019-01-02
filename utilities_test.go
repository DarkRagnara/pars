package pars

import (
	"fmt"
	"testing"
)

func TestTransformer(t *testing.T) {
	r := stringReader("123")
	val, err := NewTransformer(NewInt(), func(v interface{}) (interface{}, error) { return v.(int) * 2, nil }).Parse(r)
	assertParse(t, val, err, 246, nil)
}

func TestFailingTransformer(t *testing.T) {
	r := stringReader("123")
	expectedErr := fmt.Errorf("Some transformer error")
	val, err := NewTransformer(NewInt(), func(v interface{}) (interface{}, error) { return nil, expectedErr }).Parse(r)
	assertParse(t, val, err, nil, expectedErr)

	expectedErr = fmt.Errorf("Some parser error")
	val, err = NewTransformer(NewDiscardRight(NewInt(), NewError(expectedErr)), func(v interface{}) (interface{}, error) { return v, nil }).Parse(r)
	assertParse(t, val, err, nil, expectedErr)

	val, err = NewString("123").Parse(r)
	assertParse(t, val, err, "123", nil)
}

func TestTransformerUnread(t *testing.T) {
	r := stringReader("123")
	expectedErr := fmt.Errorf("Forced unread")
	val, err := NewDiscardRight(NewTransformer(NewInt(), func(v interface{}) (interface{}, error) { return v, nil }), NewError(expectedErr)).Parse(r)
	assertParse(t, val, err, nil, expectedErr)

	val, err = NewString("123").Parse(r)
	assertParse(t, val, err, "123", nil)
}

func TestSwallowWhitespace(t *testing.T) {
	r := stringReader(" 123 ")
	val, err := NewDiscardRight(NewSwallowWhitespace(NewInt()), EOF).Parse(r)
	assertParse(t, val, err, 123, nil)
}

func TestSwallowLeadingWhitespace(t *testing.T) {
	r := stringReader(" 123 ")
	val, err := NewDiscardRight(NewSwallowLeadingWhitespace(NewInt()), NewSeq(NewString(" "), EOF)).Parse(r)
	assertParse(t, val, err, 123, nil)
}

func TestSwallowTrailingWhitespace(t *testing.T) {
	r := stringReader(" 123 ")
	val, err := NewDiscardLeft(NewString(" "), NewDiscardRight(NewSwallowTrailingWhitespace(NewInt()), EOF)).Parse(r)
	assertParse(t, val, err, 123, nil)
}
