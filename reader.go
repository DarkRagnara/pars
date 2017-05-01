package pars

import (
	"io"
)

type Reader struct {
	r          io.Reader
	buf        buffer
	bufBackend []byte
	lastErr    error
}

func NewReader(r io.Reader) *Reader {
	bufBackend := make([]byte, 256)
	return &Reader{r: r, bufBackend: bufBackend, buf: buffer{current: bufBackend[0:0]}}
}

var _ io.Reader = &Reader{}

func (br *Reader) Read(p []byte) (n int, err error) {
	if br.buf.IsEmpty() && br.lastErr == io.EOF {
		return 0, io.EOF
	}

	n, err = br.buf.Read(p)
	if n == len(p) {
		return
	}

	p = p[n:]

	m, lastErr := br.r.Read(br.bufBackend)
	br.lastErr = lastErr
	br.buf.current = br.bufBackend[:m]

	n2, err := br.buf.Read(p)
	return n + n2, err
}

func (br *Reader) Unread(p []byte) {
	br.buf.Unread(p)
}
