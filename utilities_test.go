package pars

import (
	"fmt"
	"testing"
)

func TestTransformer(t *testing.T) {
	r := stringReader("123")
	val, err := Transformer(Int(), func(v interface{}) (interface{}, error) { return v.(int) * 2, nil }).Parse(r)
	assertParse(t, val, err, 246, nil)
}

func TestFailingTransformer(t *testing.T) {
	r := stringReader("123")
	expectedErr := fmt.Errorf("Some transformer error")
	val, err := Transformer(Int(), func(v interface{}) (interface{}, error) { return nil, expectedErr }).Parse(r)
	assertParse(t, val, err, nil, expectedErr)

	expectedErr = fmt.Errorf("Some parser error")
	val, err = Transformer(DiscardRight(Int(), Error(expectedErr)), func(v interface{}) (interface{}, error) { return v, nil }).Parse(r)
	assertParse(t, val, err, nil, expectedErr)

	val, err = String("123").Parse(r)
	assertParse(t, val, err, "123", nil)
}

func TestTransformerUnread(t *testing.T) {
	r := stringReader("123")
	expectedErr := fmt.Errorf("Forced unread")
	val, err := DiscardRight(Transformer(Int(), func(v interface{}) (interface{}, error) { return v, nil }), Error(expectedErr)).Parse(r)
	assertParse(t, val, err, nil, expectedErr)

	val, err = String("123").Parse(r)
	assertParse(t, val, err, "123", nil)
}

func TestSwallowWhitespace(t *testing.T) {
	r := stringReader(" 123 ")
	val, err := DiscardRight(SwallowWhitespace(Int()), EOF).Parse(r)
	assertParse(t, val, err, 123, nil)
}

func TestSwallowLeadingWhitespace(t *testing.T) {
	r := stringReader(" 123 ")
	val, err := DiscardRight(SwallowLeadingWhitespace(Int()), Seq(String(" "), EOF)).Parse(r)
	assertParse(t, val, err, 123, nil)
}

func TestSwallowTrailingWhitespace(t *testing.T) {
	r := stringReader(" 123 ")
	val, err := DiscardLeft(String(" "), DiscardRight(SwallowTrailingWhitespace(Int()), EOF)).Parse(r)
	assertParse(t, val, err, 123, nil)
}
