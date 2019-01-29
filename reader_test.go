package pars

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestEmptyRead(t *testing.T) {
	r := eofReader()
	buf := []byte{}
	n, err := r.Read(buf)

	assertRead(t, n, err, 0, nil)
}

func TestEOFRead(t *testing.T) {
	r := eofReader()
	buf := make([]byte, 1)
	n, err := r.Read(buf)

	assertRead(t, n, err, 0, io.EOF)
}

func TestRead(t *testing.T) {
	r := stringReader("abc")
	buf := make([]byte, 1)

	n, err := r.Read(buf)
	assertRead(t, n, err, 1, nil)
	assertReader(t, r, []byte{98, 99}, nil)
	assertBytes(t, buf, []byte{97})

	n, err = r.Read(buf)
	assertRead(t, n, err, 1, nil)
	assertReader(t, r, []byte{99}, nil)
	assertBytes(t, buf, []byte{98})

	n, err = r.Read(buf)
	assertRead(t, n, err, 1, nil)
	assertReader(t, r, []byte{}, nil)
	assertBytes(t, buf, []byte{99})

	n, err = r.Read(buf)
	assertRead(t, n, err, 0, io.EOF)
	assertReader(t, r, []byte{}, io.EOF)
}

func eofReader() *Reader {
	return NewReader(bytes.NewReader([]byte{}))
}

func stringReader(s string) *Reader {
	return NewReader(strings.NewReader(s))
}

func byteReader(b []byte) *Reader {
	return NewReader(bytes.NewReader(b))
}
