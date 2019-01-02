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
	assertParse(t, val, err, nil, fmt.Errorf("Result transformation failed: %v", expectedErr))

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
