package pars

import (
	"io"
)

type Reader struct {
	r       io.Reader
	buf     []byte
	l       int
	lastErr error
}

func NewReader(r io.Reader) *Reader {
	return &Reader{r: r, buf: make([]byte, 256)}
}

var _ io.Reader = &Reader{}

func (br *Reader) Read(p []byte) (n int, err error) {
	if br.l == 0 && br.lastErr == io.EOF {
		return 0, io.EOF
	}

	if br.l >= len(p) {
		br.readFromBuf(p)
		return len(p), nil
	}

	read := br.l
	rest := len(p) - read
	copy(p, br.buf[:])

	nFromBuffer, err := br.r.Read(br.buf)
	br.l = nFromBuffer
	br.lastErr = err

	if nFromBuffer >= rest {
		br.readFromBuf(p[read:])
		return len(p), nil
	}

	copy(p[read:], br.buf[:])
	read += br.l
	br.buf = br.buf[:0]
	br.l = 0
	return read, err
}

func (br *Reader) readFromBuf(p []byte) {
	copy(p, br.buf[:len(p)])
	br.buf = br.buf[len(p):]
	br.l -= len(p)
}
