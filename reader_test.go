package pars

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestEmptyRead(t *testing.T) {
	r := EOFReader()
	buf := []byte{}
	n, err := r.Read(buf)

	assertRead(t, n, err, 0, nil)
}

func TestEOFRead(t *testing.T) {
	r := EOFReader()
	buf := make([]byte, 1)
	n, err := r.Read(buf)

	assertRead(t, n, err, 0, io.EOF)
}

func TestRead(t *testing.T) {
	r := StringReader("abc")
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

func assertRead(t *testing.T, n int, err error, expectedN int, expectedErr error) {
	if n != expectedN {
		t.Errorf("Expected read count %v, but got %v", expectedN, n)
	}
	if err != expectedErr {
		t.Errorf("Expected error %v, but got %v", expectedErr, err)
	}
}

func assertReader(t *testing.T, r *Reader, buf []byte, err error) {
	if r.lastErr != err {
		t.Errorf("Expected lastErr in reader %v, but got %v", err, r.lastErr)
	}

	assertBytes(t, r.buf[:r.l], buf)
}

func assertBytes(t *testing.T, buf []byte, expectedBuf []byte) {
	if len(expectedBuf) != len(buf) {
		t.Errorf("Expected len in reader %v, but got %v, buffer contains %v", len(expectedBuf), len(buf), buf)
	}

	for i, b := range expectedBuf {
		if buf[i] != b {
			t.Errorf("Expected byte %v in reader buffer to be %v, but got %v", i, b, buf[i])
		}
	}
}

func EOFReader() *Reader {
	return NewReader(bytes.NewReader([]byte{}))
}

func StringReader(s string) *Reader {
	return NewReader(strings.NewReader(s))
}
