package pars

import (
	"errors"
	"strings"
	"unicode/utf8"
)

type Parser interface {
	Parse(*Reader) (interface{}, error)
	Unread(*Reader)
	Clone() Parser
}

func ParseString(s string, p Parser) (interface{}, error) {
	r := NewReader(strings.NewReader(s))
	return p.Parse(r)
}

type AnyRune struct {
	buf []byte
	i   int
}

var _ Parser = &AnyRune{}

var ErrRuneExpected = errors.New("Expected rune")

//Parse tries to read a single rune or fails.
func (r *AnyRune) Parse(src *Reader) (interface{}, error) {
	r.buf = make([]byte, utf8.UTFMax)

	r.i = 0
	for ; r.i < len(r.buf); r.i++ {
		_, err := src.Read(r.buf[r.i : r.i+1])
		if err != nil {
			r.i--
			r.Unread(src)
			return nil, err
		}

		if utf8.FullRune(r.buf[0 : r.i+1]) {
			rune := rune(r.buf[0])
			if rune >= utf8.RuneSelf {
				rune, _ = utf8.DecodeRune(r.buf[0 : r.i+1])
			}

			if rune != 0xfffd {
				return rune, nil
			}
			break
		}
	}

	r.Unread(src)
	return nil, ErrRuneExpected
}

func (r *AnyRune) Unread(src *Reader) {
	if r.i >= 0 {
		src.Unread(r.buf[:r.i+1])
		r.buf = nil
		r.i = 0
	}
}

func (r *AnyRune) Clone() Parser {
	return &AnyRune{}
}

type AnyByte struct {
	buf  [1]byte
	read bool
}

var _ Parser = &AnyByte{}

func (b *AnyByte) Parse(src *Reader) (interface{}, error) {
	n, err := src.Read(b.buf[:])
	if err != nil {
		return nil, err
	}
	if n != 1 {
		panic("AnyByte read bytes != 1")
	}
	b.read = true
	return b.buf[0], nil
}

func (b *AnyByte) Unread(src *Reader) {
	if b.read {
		src.Unread(b.buf[:])
		b.read = false
	}
}

func (b *AnyByte) Clone() Parser {
	return &AnyByte{}
}
