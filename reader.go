package pars

import (
	"io"
)

type reader struct {
	r          io.Reader
	buf        buffer
	bufBackend [256]byte
	lastErr    error
}

func newReader(r io.Reader) *reader {
	reader := &reader{r: r}
	reader.buf.current = reader.bufBackend[0:0]
	return reader
}

var _ io.Reader = &reader{}

func (br *reader) Read(p []byte) (n int, err error) {
	if br.buf.IsEmpty() && br.lastErr == io.EOF {
		return 0, io.EOF
	}

	n, err = br.buf.Read(p)
	if n == len(p) {
		return
	}

	p = p[n:]

	m, lastErr := br.r.Read(br.bufBackend[:])
	br.lastErr = lastErr
	br.buf.current = br.bufBackend[:m]

	n2, err := br.buf.Read(p)
	return n + n2, err
}

func (br *reader) Unread(p []byte) {
	br.buf.Unread(p)
}
