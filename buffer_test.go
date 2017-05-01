package pars

import (
	"io"
	"testing"
)

func TestIsEmpty(t *testing.T) {
	b := buffer{}
	assertBufferLen(t, b, 0)
	if !b.IsEmpty() {
		t.Error("Constructed buffer not empty")
	}
}

func TestReadFromCurrent(t *testing.T) {
	b := buffer{current: []byte{1, 2, 3}}
	assertBufferLen(t, b, 3)
	read123(t, b)
}

func TestReadFromPrepend(t *testing.T) {
	b := buffer{}
	b.Unread([]byte{1, 2, 3})
	assertBufferLen(t, b, 3)
	read123(t, b)
}

func TestReadNotEverythingFromPrepend(t *testing.T) {
	b := buffer{}
	b.Unread([]byte{1, 2, 3})
	assertBufferLen(t, b, 3)

	buf := make([]byte, 2)
	n, err := b.Read(buf)
	assertRead(t, n, err, 2, nil)
	assertBytes(t, buf, []byte{1, 2})
	assertBytes(t, b.prepend, []byte{3})
}

func TestReadFromBoth(t *testing.T) {
	b := buffer{current: []byte{3}}
	b.Unread([]byte{1, 2})
	assertBufferLen(t, b, 3)
	read123(t, b)
}

func TestReadAsMuchAsPossible(t *testing.T) {
	b := buffer{current: []byte{3}}
	assertBufferLen(t, b, 1)
	b.Unread([]byte{1, 2})
	assertBufferLen(t, b, 3)
	buf := make([]byte, 5)

	n, err := b.Read(buf)
	assertRead(t, n, err, 3, io.EOF)
	assertBytes(t, buf, []byte{1, 2, 3, 0, 0})

	if !b.IsEmpty() {
		t.Error("Constructed buffer not empty")
	}
}

func read123(t *testing.T, b buffer) {
	buf := make([]byte, 3)

	n, err := b.Read(buf)
	assertRead(t, n, err, 3, nil)
	assertBytes(t, buf, []byte{1, 2, 3})

	n, err = b.Read(buf)
	assertRead(t, n, err, 0, io.EOF)

	if !b.IsEmpty() {
		t.Error("Constructed buffer not empty")
	}
}
