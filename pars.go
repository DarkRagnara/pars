package pars

import (
	"strings"
)

type Parser interface {
	Parse(*Reader) (interface{}, error)
}

func ParseString(s string, p Parser) (interface{}, error) {
	r := NewReader(strings.NewReader(s))
	return p.Parse(r)
}

type Rune struct {
}

func (r Rune) Parse(src *Reader) (interface{}, error) {
	return nil, nil
}
